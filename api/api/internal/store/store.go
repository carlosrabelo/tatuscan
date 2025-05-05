package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/carlosrabelo/tatuscan/api/api/internal/model"

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

var sortColumns = map[string]string{
	"hostname":            "hostname",
	"os":                  "os",
	"os_version":          "os_version",
	"created_at":          "created_at",
	"updated_at":          "updated_at",
	"computer_activation": "computer_activation",
}

// ListAll returns inventories ordered by whitelist column.
func (s *Store) ListAll(ctx context.Context, orderBy, direction string) ([]model.Inventory, error) {
	col, ok := sortColumns[orderBy]
	if !ok {
		col = "hostname"
	}
	dir := "ASC"
	if strings.EqualFold(direction, "desc") {
		dir = "DESC"
	}
	q := fmt.Sprintf(`SELECT machine_id, hostname, ip, os, os_version, cpu_percent,
		memory_total_mb, memory_used_mb, computer_model, computer_activation,
		activation_days, created_at, updated_at
		FROM inventory ORDER BY %s %s`, col, dir)

	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return s.scanInventories(rows)
}

// GetByID returns one inventory or sql.ErrNoRows.
func (s *Store) GetByID(ctx context.Context, machineID string) (model.Inventory, error) {
	row := s.db.QueryRowContext(ctx, `SELECT machine_id, hostname, ip, os, os_version, cpu_percent,
		memory_total_mb, memory_used_mb, computer_model, computer_activation,
		activation_days, created_at, updated_at
		FROM inventory WHERE machine_id = ?`, machineID)
	return s.scanInventory(row)
}

// Insert creates a new inventory row.
func (s *Store) Insert(ctx context.Context, inv model.Inventory) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO inventory (
		machine_id, hostname, ip, os, os_version, cpu_percent, memory_total_mb,
		memory_used_mb, computer_model, computer_activation, activation_days,
		created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		inv.MachineID, inv.Hostname, inv.IP, inv.OS, nullStr(inv.OSVersion),
		inv.CPUPercent, inv.MemoryTotalMB, nullInt64(inv.MemoryUsedMB),
		nullStr(inv.ComputerModel), nullTime(inv.ComputerActivation),
		nullInt(inv.ActivationDays), formatTime(inv.CreatedAt), nullTime(inv.UpdatedAt),
	)
	return err
}

// Update overwrites an inventory row.
func (s *Store) Update(ctx context.Context, inv model.Inventory) error {
	_, err := s.db.ExecContext(ctx, `UPDATE inventory SET
		hostname = ?, ip = ?, os = ?, os_version = ?, cpu_percent = ?,
		memory_total_mb = ?, memory_used_mb = ?, computer_model = ?,
		computer_activation = ?, activation_days = ?, updated_at = ?
		WHERE machine_id = ?`,
		inv.Hostname, inv.IP, inv.OS, nullStr(inv.OSVersion), inv.CPUPercent,
		inv.MemoryTotalMB, nullInt64(inv.MemoryUsedMB), nullStr(inv.ComputerModel),
		nullTime(inv.ComputerActivation), nullInt(inv.ActivationDays),
		nullTime(inv.UpdatedAt), inv.MachineID,
	)
	return err
}

// Delete removes an inventory row.
func (s *Store) Delete(ctx context.Context, machineID string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM inventory WHERE machine_id = ?`, machineID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

type scannable interface {
	Scan(dest ...any) error
}

func (s *Store) scanInventory(row scannable) (model.Inventory, error) {
	var inv model.Inventory
	var osVersion, modelName, activation, updated sql.NullString
	var memUsed sql.NullInt64
	var actDays sql.NullInt64
	var created string

	err := row.Scan(
		&inv.MachineID, &inv.Hostname, &inv.IP, &inv.OS, &osVersion,
		&inv.CPUPercent, &inv.MemoryTotalMB, &memUsed, &modelName, &activation,
		&actDays, &created, &updated,
	)
	if err != nil {
		return inv, err
	}
	inv.OSVersion = nullStringPtr(osVersion)
	inv.ComputerModel = nullStringPtr(modelName)
	if memUsed.Valid {
		v := memUsed.Int64
		inv.MemoryUsedMB = &v
	}
	if actDays.Valid {
		v := int(actDays.Int64)
		inv.ActivationDays = &v
	}
	ct, err := parseStoredTime(created, s.loc)
	if err != nil {
		return inv, fmt.Errorf("created_at: %w", err)
	}
	inv.CreatedAt = ct
	if activation.Valid && activation.String != "" {
		t, err := parseStoredTime(activation.String, s.loc)
		if err == nil {
			inv.ComputerActivation = &t
		}
	}
	if updated.Valid && updated.String != "" {
		t, err := parseStoredTime(updated.String, s.loc)
		if err == nil {
			inv.UpdatedAt = &t
		}
	}
	return inv, nil
}

func (s *Store) scanInventories(rows *sql.Rows) ([]model.Inventory, error) {
	var out []model.Inventory
	for rows.Next() {
		inv, err := s.scanInventory(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, inv)
	}
	return out, rows.Err()
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func nullTime(t *time.Time) any {
	if t == nil {
		return nil
	}
	return formatTime(*t)
}

func nullStr(s *string) any {
	if s == nil {
		return nil
	}
	return *s
}

func nullInt64(v *int64) any {
	if v == nil {
		return nil
	}
	return *v
}

func nullInt(v *int) any {
	if v == nil {
		return nil
	}
	return *v
}

func nullStringPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	s := ns.String
	return &s
}

// OSDistribution groups by OS.
func (s *Store) OSDistribution(ctx context.Context) ([]model.OSCount, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT COALESCE(os, '-'), COUNT(*) FROM inventory
		GROUP BY os ORDER BY COUNT(*) DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.OSCount
	for rows.Next() {
		var item model.OSCount
		if err := rows.Scan(&item.OS, &item.Count); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}
