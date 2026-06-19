package api

import (
	"errors"
	"html/template"
	"net/http"

	"github.com/edaywalid/whook/internal/dispatch"
	"github.com/edaywalid/whook/internal/storage"
)

var uiTmpl = template.Must(template.New("ui").
	Funcs(template.FuncMap{"token": withToken}).
	Parse(uiTemplates))

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	events, err := s.store.ListEvents(r.Context(), storage.EventFilter{Limit: 100})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "loading dashboard")
		return
	}

	summaries := make([]eventSummary, 0, len(events))
	for _, e := range events {
		summaries = append(summaries, toSummary(e))
	}
	s.render(w, "dashboard", map[string]any{
		"Events": summaries,
		"Token":  s.presentedToken(r),
	})
}

func (s *Server) handleEventPage(w http.ResponseWriter, r *http.Request) {
	eventID := r.PathValue("id")
	event, err := s.store.GetEvent(r.Context(), eventID)
	if errors.Is(err, storage.ErrNotFound) {
		http.Error(w, "event not found", http.StatusNotFound)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "loading event")
		return
	}

	attempts, err := s.store.ListAttempts(r.Context(), eventID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "loading attempts")
		return
	}
	deliveries, err := s.store.DeliveriesForEvent(r.Context(), eventID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "loading deliveries")
		return
	}

	detail := toDetail(event, attempts, deliveries, s.destinationURLs(r.Context(), event.Source))
	s.render(w, "detail", map[string]any{
		"Detail": detail,
		"Token":  s.presentedToken(r),
	})
}

func (s *Server) handleUIReplay(w http.ResponseWriter, r *http.Request) {
	eventID := r.PathValue("id")
	event, err := s.store.GetEvent(r.Context(), eventID)
	if errors.Is(err, storage.ErrNotFound) {
		http.Error(w, "event not found", http.StatusNotFound)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "loading event")
		return
	}
	if _, err := dispatch.Enqueue(r.Context(), s.store, event, true, s.now()); err != nil {
		writeError(w, http.StatusInternalServerError, "queuing replay")
		return
	}
	http.Redirect(w, r, withToken("/ui/events/"+eventID, s.presentedToken(r)), http.StatusSeeOther)
}

func (s *Server) handleDeadLettersPage(w http.ResponseWriter, r *http.Request) {
	dead, err := s.store.ListDeadLetters(r.Context(), 100, 0)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "loading dead letters")
		return
	}

	out := make([]deadLetterDTO, 0, len(dead))
	for _, p := range dead {
		url := ""
		if dest, err := s.store.GetDestination(r.Context(), p.DestinationID); err == nil {
			url = dest.URL
		}
		out = append(out, toDeadLetterDTO(p, url))
	}
	s.render(w, "deadletters", map[string]any{
		"DeadLetters": out,
		"Token":       s.presentedToken(r),
	})
}

func (s *Server) render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := uiTmpl.ExecuteTemplate(w, name, data); err != nil {
		s.logger.Error("rendering ui", "template", name, "error", err)
	}
}

func withToken(path, token string) string {
	if token == "" {
		return path
	}
	return path + "?token=" + token
}

const uiTemplates = `
{{define "head"}}<!doctype html>
<html lang="en"><head>
<meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">
<title>whook</title>
<style>
  :root { color-scheme: light dark; }
  body { font: 15px/1.5 ui-sans-serif, system-ui, sans-serif; margin: 0; padding: 2rem; max-width: 960px; }
  a { color: #0f766e; text-decoration: none; } a:hover { text-decoration: underline; }
  h1 { font-size: 1.4rem; margin: 0 0 0.25rem; }
  h2 { font-size: 1rem; margin: 1.5rem 0 0.5rem; }
  p.sub { color: #888; margin: 0 0 1.5rem; }
  nav { margin-bottom: 1.5rem; font-size: 0.9rem; }
  table { border-collapse: collapse; width: 100%; }
  th, td { text-align: left; padding: 0.5rem 0.75rem; border-bottom: 1px solid #8883; vertical-align: top; }
  th { font-size: 0.8rem; text-transform: uppercase; letter-spacing: 0.04em; color: #888; }
  td.id, code, pre { font-family: ui-monospace, monospace; }
  td.id { font-size: 0.85rem; }
  pre { background: #8881; padding: 0.75rem; border-radius: 6px; overflow-x: auto; font-size: 0.85rem; }
  .pill { font-size: 0.78rem; padding: 0.1rem 0.5rem; border-radius: 999px; border: 1px solid #8884; }
  .received, .delivered, .success { color: #2a7; }
  .rejected, .dead_letter, .permanent { color: #c44; }
  .pending, .retryable { color: #c80; }
  .empty { color: #888; padding: 2rem 0; }
  button { font: inherit; padding: 0.4rem 0.9rem; border-radius: 6px; border: 1px solid #0f766e; background: #0f766e; color: #fff; cursor: pointer; }
</style></head><body>{{end}}

{{define "foot"}}</body></html>{{end}}

{{define "dashboard"}}{{template "head"}}
  <h1>whook</h1>
  <p class="sub">Captured events</p>
  <nav><a href="{{token "/ui/deadletters" .Token}}">Dead letters &rarr;</a></nav>
  {{if .Events}}
  <table>
    <thead><tr><th>ID</th><th>Source</th><th>Status</th><th>Method</th><th>Received</th></tr></thead>
    <tbody>
    {{range .Events}}
      <tr>
        <td class="id"><a href="{{token (printf "/ui/events/%s" .ID) $.Token}}">{{.ID}}</a></td>
        <td>{{.Source}}</td>
        <td><span class="pill {{.Status}}">{{.Status}}</span></td>
        <td>{{.Method}}</td>
        <td>{{.ReceivedAt.Format "2006-01-02 15:04:05"}}</td>
      </tr>
    {{end}}
    </tbody>
  </table>
  {{else}}
  <p class="empty">No events captured yet. Point a provider at <code>/ingest/{source}</code>.</p>
  {{end}}
{{template "foot"}}{{end}}

{{define "detail"}}{{template "head"}}
  {{with .Detail}}
  <nav><a href="{{token "/ui" $.Token}}">&larr; Events</a></nav>
  <h1>{{.ID}}</h1>
  <p class="sub">{{.Source}} &middot; {{.Method}} {{.Path}} &middot; <span class="pill {{.Status}}">{{.Status}}</span> &middot; {{.ReceivedAt.Format "2006-01-02 15:04:05"}}</p>

  <form method="post" action="{{token (printf "/ui/events/%s/replay" .ID) $.Token}}">
    <button type="submit">Replay</button>
  </form>

  <h2>Deliveries</h2>
  {{if .Deliveries}}
  <table>
    <thead><tr><th>Destination</th><th>State</th><th>Attempts</th></tr></thead>
    <tbody>
    {{range .Deliveries}}
      <tr><td>{{if .DestinationURL}}{{.DestinationURL}}{{else}}{{.DestinationID}}{{end}}</td>
      <td><span class="pill {{.State}}">{{.State}}</span></td><td>{{.AttemptNumber}}</td></tr>
    {{end}}
    </tbody>
  </table>
  {{else}}<p class="empty">No deliveries queued.</p>{{end}}

  <h2>Attempt history</h2>
  {{if .Attempts}}
  <table>
    <thead><tr><th>#</th><th>Outcome</th><th>Status</th><th>Duration</th><th>When</th></tr></thead>
    <tbody>
    {{range .Attempts}}
      <tr><td>{{.AttemptNumber}}</td><td><span class="pill {{.Outcome}}">{{.Outcome}}</span></td>
      <td>{{.StatusCode}}</td><td>{{.DurationMS}}ms</td><td>{{.AttemptedAt.Format "15:04:05"}}</td></tr>
    {{end}}
    </tbody>
  </table>
  {{else}}<p class="empty">No attempts yet.</p>{{end}}

  <h2>Body</h2>
  <pre>{{.Body}}</pre>
  {{end}}
{{template "foot"}}{{end}}

{{define "deadletters"}}{{template "head"}}
  <nav><a href="{{token "/ui" .Token}}">&larr; Events</a></nav>
  <h1>Dead letters</h1>
  <p class="sub">Deliveries that exhausted their retry budget</p>
  {{if .DeadLetters}}
  <table>
    <thead><tr><th>Event</th><th>Destination</th><th>Attempts</th></tr></thead>
    <tbody>
    {{range .DeadLetters}}
      <tr><td class="id"><a href="{{token (printf "/ui/events/%s" .EventID) $.Token}}">{{.EventID}}</a></td>
      <td>{{if .DestinationURL}}{{.DestinationURL}}{{else}}{{.DestinationID}}{{end}}</td>
      <td>{{.AttemptNumber}}</td></tr>
    {{end}}
    </tbody>
  </table>
  {{else}}<p class="empty">No dead letters. Everything delivered.</p>{{end}}
{{template "foot"}}{{end}}
`
