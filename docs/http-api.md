# HTTP API

The gateway exposes one public ingest endpoint, a JSON management API, and a thin
server-rendered dashboard. With `WHOOK_ADMIN_TOKEN` set, every endpoint except
ingest and health requires the bearer token (see
[Configuration](configuration.md#authentication)).

## Endpoints

| Method & path | Purpose |
| --- | --- |
| `POST /ingest/{source}` | Capture an inbound webhook (public) |
| `GET /healthz` | Liveness check (public) |
| `GET /metrics` | Prometheus metrics (ingest, delivery, dead-letter) |
| `GET /events` | List captured events (`?source=`, `?status=`, `?limit=`, `?offset=`) |
| `GET /events/{id}` | Event detail with per-destination delivery state and attempt history |
| `POST /events/{id}/replay` | Re-deliver an event (optional `?destination_id=`) |
| `GET /deadletters` | List deliveries that exhausted their retry budget |
| `GET /sources` | List registered sources |
| `POST /sources` | Register a source |
| `PATCH /sources/{name}` | Update a source |
| `GET /destinations?source=` | List a source's destinations |
| `POST /destinations` | Create a destination |
| `PATCH /destinations/{id}` | Update a destination (including enable/disable) |

## Capturing a webhook

```sh
curl -X POST localhost:8080/ingest/stripe \
  -H 'Content-Type: application/json' \
  -d '{"type":"payment.succeeded","amount":4900}'
# -> {"event_id":"evt_9f3a21c8b4e07d56"}
```

The gateway responds with a fast `200` and the event id once the request is
durably stored. A request that fails signature verification is still captured
(with status `rejected`) and returns `401`. An unregistered source returns `404`,
and a body over the size limit returns `413`.

## Sources

A source is the provider integration that webhooks POST to. Register it before
sending.

```sh
curl -X POST localhost:8080/sources -d '{
  "name": "stripe",
  "verifier": "stripe",
  "secret": "whsec_...",
  "dedup_header": "Stripe-Id"
}'
```

- `verifier` (optional): one of `stripe`, `github`, `shopify`, `slack`, or `linear`.
  Requests that fail verification are captured but never delivered.
- `secret` (optional, write-only): the provider signing secret, stored encrypted.
- `dedup_header` (optional): a header carrying the provider's event id. Re-sent
  events with the same key collapse into one.

## Destinations

A destination is where matching events are forwarded. One source can fan out to
many destinations, each tracked independently.

```sh
curl -X POST localhost:8080/destinations -d '{
  "source": "stripe",
  "url": "https://billing.internal/webhooks",
  "filter_spec": { "body_equals": { "type": "payment.succeeded" } }
}'
```

### Filter spec

A destination only receives events that match its `filter_spec`. All clauses must
match (logical AND); an empty or absent spec matches everything. Body clauses
address a JSON field by a dot-separated path (e.g. `data.object.status`).

| Clause | Example | Matches when |
| --- | --- | --- |
| `body_equals` | `{"type": "payment.succeeded"}` | the field equals the value |
| `header_equals` | `{"X-Github-Event": "push"}` | the header equals the value (case-insensitive name) |
| `body_exists` | `["data.object.id"]` | the field is present |
| `body_in` | `{"type": ["a", "b"]}` | the field value is one of the list |
| `body_prefix` | `{"type": "payment."}` | the field value starts with the prefix |
| `body_not` | `{"type": "payment.failed"}` | the field does not equal the value |

## Replay

```sh
# Replay to every destination of the event
curl -X POST localhost:8080/events/evt_9f3a.../replay

# Replay to a single destination
curl -X POST 'localhost:8080/events/evt_9f3a.../replay?destination_id=dst_...'
```

A replay enqueues a fresh delivery with its own retry budget and is recorded as a
replay attempt, leaving the original history intact. This is how you recover a
dead-lettered event after fixing the underlying cause.

## Dashboard

`GET /ui` serves a dashboard that lists events, links to a per-event detail page
(payload, per-destination delivery state, attempt history, and a replay button),
and has a dead-letter view. With an admin token set, pass it as
`Authorization: Bearer <token>` or `?token=<token>`.
