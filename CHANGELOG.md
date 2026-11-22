# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project follows
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `*` segment in filter body paths to match against every element of an array, e.g. `data.assignees.*.id`.

## [0.1.0] - 2026-06-29

First tagged release. whook captures inbound webhooks durably, acknowledges the
provider with a fast 2xx, and forwards each event to one or more destinations
with automatic retries, dead-lettering, and replay.

### Added

- Durable capture that stores every inbound request before the provider is acknowledged.
- Exponential-backoff retries with a per-delivery budget, honoring `Retry-After` on 429, then dead-lettering.
- Fan-out to many destinations, each with its own filter, retry state, and status.
- Signature verification for Stripe, GitHub, Shopify, Slack, and Linear, selected per source.
- Idempotent capture keyed by a configurable per-source dedup header.
- Server-rendered dashboard with an event list, per-event detail and attempt history, replay, and a dead-letter view.
- HTTP API for sources, destinations, events, replay, and dead letters behind a single admin bearer token.
- SQLite by default with a pure-Go driver, and Postgres behind the same storage interface.
- Prometheus metrics, request rate limiting, and an age- and count-based retention pruner.
- `--version` flag, with the version stamped into release binaries and Docker images.

[Unreleased]: https://github.com/edaywalid/whook/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/edaywalid/whook/releases/tag/v0.1.0
