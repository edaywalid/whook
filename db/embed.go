// Package db embeds the SQL migrations so the gateway ships as a single binary
// and applies them on startup.
package db

import "embed"

// Migrations holds the dbmate-authored migration files.
//
//go:embed migrations/*.sql
var Migrations embed.FS
