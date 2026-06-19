package api

import (
	"html/template"
	"net/http"

	"github.com/edaywalid/whook/internal/storage"
)

var dashboardTmpl = template.Must(template.New("dashboard").Parse(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>whook</title>
<style>
  :root { color-scheme: light dark; }
  body { font: 15px/1.5 ui-sans-serif, system-ui, sans-serif; margin: 0; padding: 2rem; max-width: 960px; }
  h1 { font-size: 1.4rem; margin: 0 0 0.25rem; }
  p.sub { color: #888; margin: 0 0 1.5rem; }
  table { border-collapse: collapse; width: 100%; }
  th, td { text-align: left; padding: 0.5rem 0.75rem; border-bottom: 1px solid #8883; }
  th { font-size: 0.8rem; text-transform: uppercase; letter-spacing: 0.04em; color: #888; }
  td.id { font-family: ui-monospace, monospace; font-size: 0.85rem; }
  .status { font-size: 0.8rem; padding: 0.1rem 0.5rem; border-radius: 999px; border: 1px solid #8884; }
  .received { color: #2a7; } .rejected { color: #c44; }
  .empty { color: #888; padding: 2rem 0; }
</style>
</head>
<body>
  <h1>whook</h1>
  <p class="sub">Captured events</p>
  {{if .Events}}
  <table>
    <thead><tr><th>ID</th><th>Source</th><th>Status</th><th>Method</th><th>Received</th></tr></thead>
    <tbody>
    {{range .Events}}
      <tr>
        <td class="id">{{.ID}}</td>
        <td>{{.Source}}</td>
        <td><span class="status {{.Status}}">{{.Status}}</span></td>
        <td>{{.Method}}</td>
        <td>{{.ReceivedAt.Format "2006-01-02 15:04:05"}}</td>
      </tr>
    {{end}}
    </tbody>
  </table>
  {{else}}
  <p class="empty">No events captured yet. Point a provider at <code>/ingest/{source}</code>.</p>
  {{end}}
</body>
</html>`))

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

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := dashboardTmpl.Execute(w, map[string]any{"Events": summaries}); err != nil {
		s.logger.Error("rendering dashboard", "error", err)
	}
}
