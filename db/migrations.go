// Package db carries the schema migrations that ship with the binary.
package db

import "embed"

// Migrations holds every migration file, compiled into the binary rather than
// read from disk. The server resolves no paths at startup as a result, so it
// behaves the same whether it is run from the source tree or from the
// container, and a build can never be paired with a stale migrations directory.
//
//go:embed migrations/*.sql
var Migrations embed.FS
