// Package config loads gateway configuration from environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all runtime settings for the gateway.
type Config struct {
	// HTTP listen address, e.g. ":8080".
	Addr string

	// DatabaseURL points at the SQLite file, e.g. "file:whook.db".
	DatabaseURL string

	// AdminToken guards every endpoint except ingest and health. Empty means
	// the management API is unprotected, which is only sensible for local dev.
	AdminToken string

	// MaxBodyBytes is the largest inbound request body the ingest endpoint
	// will accept. Larger requests are rejected with 413 and not captured.
	MaxBodyBytes int64

	// SecretKey encrypts provider signing secrets at rest. Required once a
	// source configures a signing secret.
	SecretKey string

	// Delivery worker settings.
	Workers      int
	PollInterval time.Duration

	// Retention. RetentionAge is how long received events are kept; zero
	// disables pruning entirely. FailedRetentionAge keeps rejected and
	// dead-lettered events longer for debugging; it defaults to RetentionAge.
	RetentionAge       time.Duration
	FailedRetentionAge time.Duration
	PruneInterval      time.Duration

	// Ingest rate limiting, per source. IngestRate is requests per second; zero
	// disables limiting. IngestBurst is the bucket size; zero defaults to one
	// second of capacity.
	IngestRate  float64
	IngestBurst int

	// Rate-limiter eviction. The limiter keeps a token bucket per source key;
	// EvictInterval is how often idle buckets are swept and IdleTTL is how long a
	// bucket may go unused before it is removed. Without this the key map grows
	// without bound when requests arrive under many distinct source names.
	IngestEvictInterval time.Duration
	IngestIdleTTL       time.Duration
}

// Load reads configuration from the environment, applying defaults.
func Load() (Config, error) {
	c := Config{
		Addr:         envStr("WHOOK_ADDR", ":8080"),
		DatabaseURL:  envStr("WHOOK_DATABASE_URL", "file:whook.db?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"),
		AdminToken:   envStr("WHOOK_ADMIN_TOKEN", ""),
		SecretKey:    envStr("WHOOK_SECRET_KEY", ""),
		MaxBodyBytes: envInt64("WHOOK_MAX_BODY_BYTES", 1<<20), // 1 MiB
		Workers:      int(envInt64("WHOOK_WORKERS", 4)),
		PollInterval: envDuration("WHOOK_POLL_INTERVAL", time.Second),

		RetentionAge:       envDuration("WHOOK_RETENTION_AGE", 0),
		FailedRetentionAge: envDuration("WHOOK_FAILED_RETENTION_AGE", 0),
		PruneInterval:      envDuration("WHOOK_PRUNE_INTERVAL", time.Hour),

		IngestRate:  envFloat("WHOOK_INGEST_RATE", 0),
		IngestBurst: int(envInt64("WHOOK_INGEST_BURST", 0)),

		IngestEvictInterval: envDuration("WHOOK_INGEST_EVICT_INTERVAL", 10*time.Minute),
		IngestIdleTTL:       envDuration("WHOOK_INGEST_IDLE_TTL", time.Hour),
	}

	// Failed events are kept at least as long as received ones.
	if c.FailedRetentionAge < c.RetentionAge {
		c.FailedRetentionAge = c.RetentionAge
	}

	if c.Workers < 1 {
		return Config{}, fmt.Errorf("WHOOK_WORKERS must be at least 1, got %d", c.Workers)
	}
	if c.MaxBodyBytes < 1 {
		return Config{}, fmt.Errorf("WHOOK_MAX_BODY_BYTES must be positive, got %d", c.MaxBodyBytes)
	}
	return c, nil
}

func envStr(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}

func envFloat(key string, def float64) float64 {
	v, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return def
	}
	return f
}

func envInt64(key string, def int64) int64 {
	v, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return def
	}
	return n
}

func envDuration(key string, def time.Duration) time.Duration {
	v, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}
