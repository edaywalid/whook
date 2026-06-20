# Contributing

Thanks for your interest in whook. This is a single Go module with a small
TanStack Start landing page under `web/`. Issues and pull requests are welcome.

## Prerequisites

- Go 1.25 or newer
- Docker (optional, for the Postgres integration tests and the image build)
- Node 20+ (only if you work on the landing page in `web/`)

## Getting set up

```sh
git clone https://github.com/edaywalid/whook
cd whook
go build ./...
go test ./...
go run ./cmd/whook
```

The gateway runs on `http://localhost:8080` against a local SQLite file.

## Tests and checks

Run the same checks CI runs before opening a pull request:

```sh
go test -race ./...
go vet ./...
golangci-lint run ./...
```

The Postgres storage tests are skipped unless `WHOOK_TEST_POSTGRES_URL` is set.
To run them locally:

```sh
docker run -d --name whook-pg -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=whook_test -p 5433:5432 postgres:16-alpine
export WHOOK_TEST_POSTGRES_URL="postgres://postgres:postgres@localhost:5433/whook_test?sslmode=disable"
go test ./internal/storage/postgres/...
```

A good test exercises a module through its public interface and asserts on
observable behavior, not internals. Transport and storage are injected as fakes
so logic is exercised deterministically. Follow the table-driven style already in
the codebase.

## Database migrations

Migrations live under `db/migrations/{sqlite,postgres}` and are embedded into the
binary, then applied on startup. They are SQL only, one schema change per logical
migration, with both `migrate:up` and `migrate:down` sections.

Create the SQLite file with dbmate so the timestamp prefix is correct, then add
the matching Postgres file by hand (the dialects differ, BYTEA vs BLOB, BOOLEAN vs
INTEGER, TIMESTAMPTZ vs TIMESTAMP):

```sh
dbmate --migrations-dir db/migrations/sqlite new <name>
```

## Code style

- Run `gofmt -s -w .` and keep `golangci-lint` clean.
- Wrap errors with context (`fmt.Errorf("doing x: %w", err)`); never discard them.
- Comments are short and only where they add value.
- Match the surrounding code's conventions.

## Commits and pull requests

- Commit in small, logical pieces rather than one large change.
- Write commit messages and PR descriptions in a plain, direct voice.
- Keep the description focused on what changed and why.

## Landing page

The marketing site is in `web/` (TanStack Start, React 19, Tailwind v4).

```sh
cd web
npm install
npm run dev    # http://localhost:3000
npm run build
```
