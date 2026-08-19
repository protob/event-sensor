// Package db opens the SQLite database with the pragmas the schema depends on and runs
// the embedded migrations. Tests call the same two functions, so they exercise the same
// connection settings as the binary.
package db

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// foreign_keys is a per-connection pragma in SQLite, so it must live in the DSN (modernc
// honors _pragma). A one-off db.Exec would only set it on whichever pooled connection
// served it, leaving ON DELETE RESTRICT able to no-op on the others. busy_timeout, WAL and
// _txlock=immediate are here for the same per-connection reason.
const dsnParams = "?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_txlock=immediate"

// Open opens the SQLite file at path with a single writer connection.
func Open(path string) (*sql.DB, error) {
	conn, err := sql.Open("sqlite", path+dsnParams)
	if err != nil {
		return nil, fmt.Errorf("open database %s: %w", path, err)
	}
	// Single writer: serializes writes and avoids SQLITE_BUSY. The network fetch happens
	// before the transaction opens, so this connection is only ever held for a small local
	// write, and the data ceiling is one person's concert history.
	conn.SetMaxOpenConns(1)
	return conn, nil
}

// Migrate applies the embedded goose migrations. They are safe to re-apply.
func Migrate(conn *sql.DB) error {
	goose.SetBaseFS(migrationsFS)
	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}
	if err := goose.Up(conn, "migrations"); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}
	return nil
}
