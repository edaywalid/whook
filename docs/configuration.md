# Configuration

whook is configured entirely through environment variables, so it drops cleanly
into containers and 12-factor setups. Every value has a sensible default; you can
run it with no configuration at all for local use.

## Reference

| Variable | Default | Purpose |
| --- | --- | --- |
| `WHOOK_ADDR` | `:8080` | HTTP listen address |
| `WHOOK_DATABASE_URL` | `file:whook.db?...` | Storage DSN. A `postgres://` or `postgresql://` URL selects Postgres; anything else is SQLite |
| `WHOOK_ADMIN_TOKEN` | empty | Bearer token for the management API and dashboard. Empty disables auth (local only) |
| `WHOOK_SECRET_KEY` | empty | Passphrase used to encrypt provider signing secrets at rest |
| `WHOOK_MAX_BODY_BYTES` | `1048576` | Max inbound body size; larger requests are rejected with 413 |
| `WHOOK_WORKERS` | `4` | Concurrent delivery workers |
| `WHOOK_POLL_INTERVAL` | `1s` | How often workers poll the queue |
| `WHOOK_RETENTION_AGE` | empty | Delete received events older than this (e.g. `720h`). Empty disables pruning |
| `WHOOK_FAILED_RETENTION_AGE` | = retention | Keep rejected and dead-lettered events longer |
| `WHOOK_PRUNE_INTERVAL` | `1h` | How often the retention pruner runs |
| `WHOOK_INGEST_RATE` | `0` | Max ingest requests per second per source. `0` disables limiting |
| `WHOOK_INGEST_BURST` | = rate | Token-bucket burst size for ingest limiting |

## Authentication

`POST /ingest/{source}` and `GET /healthz` are always public. Everything else (the
management API, the dashboard, `/metrics`) requires the admin token once
`WHOOK_ADMIN_TOKEN` is set. Pass it as `Authorization: Bearer <token>`, or as a
`?token=` query parameter for the browser dashboard. Leaving the token empty
disables the check and is only appropriate for local development.

## Signing secrets at rest

When a source configures a signing secret, it is encrypted with AES-256-GCM using
a key derived from `WHOOK_SECRET_KEY`. The API never returns a stored secret; it
only reports whether one is set. Keep `WHOOK_SECRET_KEY` stable, rotating it
invalidates existing encrypted secrets.

## Retention

Captured events are kept forever unless `WHOOK_RETENTION_AGE` is set. When it is,
a background pruner deletes received events past that age, along with their
attempts and queue rows. Rejected and dead-lettered events are kept until the
longer `WHOOK_FAILED_RETENTION_AGE` so failures stay around for debugging. Set
`WHOOK_RETENTION_AGE` to an empty value to disable pruning entirely.

## Rate limiting

`WHOOK_INGEST_RATE` caps how many webhooks a single source can send per second,
with an independent token bucket per source. Over-limit requests are shed early
with a 429 and a `Retry-After` header, before the body is read. Leave it at `0` to
accept everything.

## Storage backends

SQLite is the default and needs nothing extra. For larger or multi-instance
deploys, point `WHOOK_DATABASE_URL` at Postgres. See
[Deployment](deployment.md#postgres) for details.
