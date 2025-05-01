package store

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// Store provides inventory persistence.
type Store struct {
	db  *sql.DB
	loc *time.Location // used when reading naive legacy timestamps
}

// Open opens a SQLite database and ensures schema exists.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`PRAGMA foreign_keys = ON; PRAGMA busy_timeout = 5000; PRAGMA journal_mode = WAL`); err != nil {
		_ = db.Close()
		return nil, err
	}
	s := &Store{db: db, loc: time.UTC}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

// SetLocation sets the timezone for parsing naive legacy datetimes.
func (s *Store) SetLocation(loc *time.Location) {
	if loc != nil {
		s.loc = loc
	}
}

// Close closes the database.
func (s *Store) Close() error {
	return s.db.Close()
}

// DB exposes the underlying *sql.DB (health checks).
func (s *Store) DB() *sql.DB {
	return s.db
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS inventory (
    machine_id TEXT PRIMARY KEY NOT NULL,
    hostname TEXT NOT NULL,
    ip TEXT NOT NULL,
    os TEXT NOT NULL,
    os_version TEXT,
    cpu_percent REAL NOT NULL,
    memory_total_mb INTEGER NOT NULL,
    memory_used_mb INTEGER,
    computer_model TEXT,
    computer_activation TEXT,
    activation_days INTEGER,
    created_at TEXT NOT NULL,
    updated_at TEXT
);`)
	return err
}
