// Package migrate applies the embedded dbmate migrations to a database on
// startup, so the gateway stays a single self-contained binary. Migrations are
// split by dialect (sqlite, postgres) since the schema SQL differs.
package migrate

import (
	"database/sql"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/edaywalid/whook/db"
)

// Dialect names the migration set and the placeholder style to apply.
type Dialect struct {
	dir         string
	placeholder string
}

var (
	// SQLite uses migrations/sqlite and "?" placeholders.
	SQLite = Dialect{dir: "migrations/sqlite", placeholder: "?"}
	// Postgres uses migrations/postgres and "$1" placeholders.
	Postgres = Dialect{dir: "migrations/postgres", placeholder: "$1"}
)

// Run applies any not-yet-applied migrations for the dialect in version order.
// It is safe to call on every startup; already-applied migrations are skipped.
func Run(conn *sql.DB, d Dialect) error {
	if _, err := conn.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	applied, err := appliedVersions(conn)
	if err != nil {
		return err
	}

	files, err := fs.Glob(db.Migrations, d.dir+"/*.sql")
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
		if err := applyOne(conn, version, upSection(string(raw)), d.placeholder); err != nil {
			return fmt.Errorf("apply %s: %w", version, err)
		}
	}
	return nil
}

func applyOne(conn *sql.DB, version, up, placeholder string) error {
	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Exec one statement at a time: the Postgres driver does not run a
	// multi-statement string through a single Exec.
	for _, stmt := range splitStatements(up) {
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("statement %q: %w", truncate(stmt, 60), err)
		}
	}
	if _, err := tx.Exec(`INSERT INTO schema_migrations (version) VALUES (`+placeholder+`)`, version); err != nil {
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

// splitStatements breaks a migration body into individual statements on `;`.
// Line comments are stripped first so a semicolon inside a comment does not
// split a statement. The schema uses plain DDL with no semicolons in string
// literals, so this is correct here.
func splitStatements(up string) []string {
	var stripped strings.Builder
	for line := range strings.Lines(up) {
		if i := strings.Index(line, "--"); i >= 0 {
			line = line[:i] + "\n"
		}
		stripped.WriteString(line)
	}

	var out []string
	for raw := range strings.SplitSeq(stripped.String(), ";") {
		if stmt := strings.TrimSpace(raw); stmt != "" {
			out = append(out, stmt)
		}
	}
	return out
}

func truncate(s string, n int) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
