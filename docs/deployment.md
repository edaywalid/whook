# Deployment

whook is a single static Go binary with SQLite built in, so most deploys are a
matter of running one container or one binary.

## Docker image

Prebuilt multi-arch images (`linux/amd64`, `linux/arm64`) are published to GitHub
Container Registry on every push to `main` and every `v*` tag.

```sh
docker run -p 8080:8080 -v whook-data:/data ghcr.io/edaywalid/whook:latest
```

Tags: `latest`, `main`, `sha-<short>`, and on releases the semver tags
(`1.2.3`, `1.2`).

## Docker Compose

The bundled `docker-compose.yml` builds from source and persists data in a named
volume.

```sh
cp .env.example .env   # set WHOOK_ADMIN_TOKEN and WHOOK_SECRET_KEY
docker compose up --build -d
```

The gateway listens on `http://localhost:8080`.

## From source

Requires Go 1.25 or newer.

```sh
go install github.com/edaywalid/whook/cmd/whook@latest
# or, from a clone:
go build -o whook ./cmd/whook
./whook
```

It defaults to a SQLite file (`whook.db`) in the working directory and applies its
migrations on startup.

## Postgres

SQLite is single-writer, which is fine for one instance. To scale out, point
`WHOOK_DATABASE_URL` at Postgres:

```sh
WHOOK_DATABASE_URL=postgres://whook:whook@localhost:5432/whook?sslmode=disable
```

The gateway applies its Postgres migrations on startup and claims the delivery
queue with `SELECT ... FOR UPDATE SKIP LOCKED`, so many workers and many instances
share one queue without ever delivering the same event twice.

With the bundled compose file, start the optional Postgres service and set the DSN
(host `postgres`):

```sh
docker compose --profile postgres up
```

## Observability

`GET /metrics` exposes Prometheus metrics: ingest rate, ingest write latency,
delivery outcomes, and dead-letter volume. It sits behind the admin token, so
point your scraper at it with the `Authorization: Bearer <token>` header.

`GET /healthz` is an unauthenticated liveness check suitable for load balancers and
orchestrators.

## Graceful shutdown

On `SIGINT`/`SIGTERM` the gateway stops accepting new work, lets in-flight
deliveries finish (up to a drain timeout), and then exits, so a deploy or restart
does not cut a delivery off mid-flight.
