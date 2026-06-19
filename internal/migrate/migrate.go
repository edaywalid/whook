// Package migrate applies the embedded dbmate migrations to a database on
// startup, so the gateway stays a single self-contained binary.
package migrate

import (
	"database/sql"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/edaywalid/whook/db"
)

// Run applies any not-yet-applied migrations in version order. It is safe to
// call on every startup; already-applied migrations are skipped.
func Run(conn *sql.DB) error {
	if _, err := conn.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	applied, err := appliedVersions(conn)
	if err != nil {
		return err
	}

	files, err := fs.Glob(db.Migrations, "migrations/*.sql")
	if err != nil {
		return fmt.Errorf("list migrations: %w", err)
	}
	sort.Strings(files)

	for _, f := range files {
		version := versionOf(f)
		if applied[version] {
			continue
		}
		raw, err := fs.ReadFile(db.Migrations, f)
		if err != nil {
			return fmt.Errorf("read %s: %w", f, err)
		}
		up := upSection(string(raw))
		if err := applyOne(conn, version, up); err != nil {
			return fmt.Errorf("apply %s: %w", version, err)
		}
	}
	return nil
}

func applyOne(conn *sql.DB, version, up string) error {
	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(up); err != nil {
		return err
	}
	if _, err := tx.Exec(`INSERT INTO schema_migrations (version) VALUES (?)`, version); err != nil {
		return err
	}
	return tx.Commit()
}

func appliedVersions(conn *sql.DB) (map[string]bool, error) {
	rows, err := conn.Query(`SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[string]bool{}
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		out[v] = true
	}
	return out, rows.Err()
}

// versionOf extracts the 14-digit timestamp prefix from a migration filename.
func versionOf(path string) string {
	base := path[strings.LastIndex(path, "/")+1:]
	if i := strings.Index(base, "_"); i > 0 {
		return base[:i]
	}
	return base
}

// upSection returns the SQL between "-- migrate:up" and "-- migrate:down".
func upSection(content string) string {
	_, rest, found := strings.Cut(content, "-- migrate:up")
	if !found {
		return strings.TrimSpace(content)
	}
	if up, _, found := strings.Cut(rest, "-- migrate:down"); found {
		return strings.TrimSpace(up)
	}
	return strings.TrimSpace(rest)
}
