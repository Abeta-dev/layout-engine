// Package migrations embeds versioned SQL schema migrations for SQLite and PostgreSQL.
package migrations

import "embed"

// Files holds all embedded SQL migration scripts under sqlite/ and postgres/.
//
//go:embed sqlite/*.sql postgres/*.sql
var Files embed.FS
