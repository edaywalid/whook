-- migrate:up

-- Inbound sources. A provider integration must be registered before it can
-- send. Per-source config lives here: which verifier to run, the encrypted
-- signing secret, and which header carries a dedup key.
CREATE TABLE sources (
    id                       TEXT PRIMARY KEY,
    name                     TEXT NOT NULL UNIQUE,
    verifier                 TEXT,
    signing_secret_encrypted BLOB,
    dedup_header             TEXT,
    created_at               TIMESTAMP NOT NULL
);

-- Every inbound request, captured raw. status is 'received' for accepted
-- requests and 'rejected' for ones that failed signature verification.
CREATE TABLE events (
    id          TEXT PRIMARY KEY,
    source      TEXT NOT NULL,
    received_at TIMESTAMP NOT NULL,
    source_ip   TEXT,
    method      TEXT NOT NULL,
    path        TEXT NOT NULL,
    query       TEXT,
    headers     TEXT NOT NULL,
    body        BLOB,
    dedup_key   TEXT,
    status      TEXT NOT NULL
);

-- Idempotency: a provider re-sending the same event (same dedup key for a
-- source) resolves to one event. NULL dedup keys never collide.
CREATE UNIQUE INDEX idx_events_source_dedup
    ON events (source, dedup_key)
    WHERE dedup_key IS NOT NULL;

CREATE INDEX idx_events_source_received ON events (source, received_at);

-- Where to forward events for a source. filter_spec is structured JSON; an
-- empty spec matches everything.
CREATE TABLE destinations (
    id              TEXT PRIMARY KEY,
    source          TEXT NOT NULL,
    url             TEXT NOT NULL,
    enabled         INTEGER NOT NULL DEFAULT 1,
    backoff_config  TEXT,
    filter_spec     TEXT,
    created_at      TIMESTAMP NOT NULL
);

CREATE INDEX idx_destinations_source ON destinations (source);

-- One row per delivery attempt, per (event, destination). The audit trail.
CREATE TABLE delivery_attempts (
    id               TEXT PRIMARY KEY,
    event_id         TEXT NOT NULL,
    destination_id   TEXT NOT NULL,
    attempt_number   INTEGER NOT NULL,
    status_code      INTEGER,
    response_snippet TEXT,
    duration_ms      INTEGER NOT NULL,
    outcome          TEXT NOT NULL,
    is_replay        INTEGER NOT NULL DEFAULT 0,
    attempted_at     TIMESTAMP NOT NULL
);

CREATE INDEX idx_attempts_event ON delivery_attempts (event_id);

-- The pending-delivery queue. One row tracks the next attempt for an
-- (event, destination) pair plus its terminal state. locked_until is the
-- worker lease so two workers never deliver the same row.
CREATE TABLE pending_deliveries (
    id              TEXT PRIMARY KEY,
    event_id        TEXT NOT NULL,
    destination_id  TEXT NOT NULL,
    attempt_number  INTEGER NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMP NOT NULL,
    locked_until    TIMESTAMP,
    state           TEXT NOT NULL DEFAULT 'pending',
    created_at      TIMESTAMP NOT NULL,
    updated_at      TIMESTAMP NOT NULL
);

CREATE INDEX idx_pending_due ON pending_deliveries (state, next_attempt_at);

-- migrate:down

DROP TABLE pending_deliveries;
DROP TABLE delivery_attempts;
DROP TABLE destinations;
DROP INDEX idx_events_source_dedup;
DROP TABLE events;
DROP TABLE sources;
