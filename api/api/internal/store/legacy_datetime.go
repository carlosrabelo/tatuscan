package store

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

// MigrateResult summarizes a legacy datetime normalization pass.
type MigrateResult struct {
	Updated int
	Skipped int
}

// NormalizeLegacyDatetimes rewrites Flask/SQLAlchemy datetime strings in inventory
// to canonical RFC3339Nano UTC used by the Go API.
//
// Naive timestamps (no offset) are interpreted in loc (app TIMEZONE), matching
// Flask/pytz behavior. Unparseable rows are skipped (logged) so one bad row
// cannot block API startup.
//
// Transitional: idempotent. Remove once all databases are migrated.
// Skip entirely with env TATUSCAN_SKIP_LEGACY_DATETIME_MIGRATE=1.
func (s *Store) NormalizeLegacyDatetimes(ctx context.Context, loc *time.Location, logger *slog.Logger) (MigrateResult, error) {
	if loc == nil {
		loc = time.UTC
	}
	if logger == nil {
		logger = slog.Default()
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT machine_id, computer_activation, created_at, updated_at
		FROM inventory`)
	if err != nil {
		return MigrateResult{}, fmt.Errorf("legacy datetime scan: %w", err)
	}
	defer rows.Close()

	type row struct {
		id         string
		activation sql.NullString
		created    string
		updated    sql.NullString
	}
	var batch []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.id, &r.activation, &r.created, &r.updated); err != nil {
			return MigrateResult{}, err
		}
		batch = append(batch, r)
	}
	if err := rows.Err(); err != nil {
		return MigrateResult{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return MigrateResult{}, err
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, `
		UPDATE inventory
		SET computer_activation = ?, created_at = ?, updated_at = ?
		WHERE machine_id = ?`)
	if err != nil {
		return MigrateResult{}, err
	}
	defer stmt.Close()

	var result MigrateResult
	for _, r := range batch {
		created, err := parseStoredTime(r.created, loc)
		if err != nil {
			logger.Warn("legacy datetime skip", "machine_id", r.id, "field", "created_at", "value", r.created, "err", err)
			result.Skipped++
			continue
		}
		newCreated := formatTime(created)

		var newActivation any
		if r.activation.Valid && strings.TrimSpace(r.activation.String) != "" {
			t, err := parseStoredTime(r.activation.String, loc)
			if err != nil {
				logger.Warn("legacy datetime skip field", "machine_id", r.id, "field", "computer_activation", "value", r.activation.String, "err", err)
				newActivation = r.activation.String // keep original
			} else {
				newActivation = formatTime(t)
			}
		}

		var newUpdated any
		if r.updated.Valid && strings.TrimSpace(r.updated.String) != "" {
			t, err := parseStoredTime(r.updated.String, loc)
			if err != nil {
				logger.Warn("legacy datetime skip field", "machine_id", r.id, "field", "updated_at", "value", r.updated.String, "err", err)
				newUpdated = r.updated.String
			} else {
				newUpdated = formatTime(t)
			}
		}

		needs := !isCanonicalTime(r.created) ||
			(r.activation.Valid && r.activation.String != "" && !isCanonicalTime(r.activation.String)) ||
			(r.updated.Valid && r.updated.String != "" && !isCanonicalTime(r.updated.String))
		if !needs {
			continue
		}

		if _, err := stmt.ExecContext(ctx, newActivation, newCreated, newUpdated, r.id); err != nil {
			logger.Warn("legacy datetime update failed", "machine_id", r.id, "err", err)
			result.Skipped++
			continue
		}
		result.Updated++
	}

	if err := tx.Commit(); err != nil {
		return MigrateResult{}, err
	}
	return result, nil
}

func isCanonicalTime(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return true
	}
	_, err := time.Parse(time.RFC3339Nano, s)
	return err == nil
}

// parseStoredTime accepts Go API formats and Flask/SQLAlchemy SQLite dumps.
// Values with an explicit offset use that offset; naive values use loc.
func parseStoredTime(s string, loc *time.Location) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("empty time")
	}
	if loc == nil {
		loc = time.UTC
	}
	if strings.HasSuffix(s, "Z") {
		s = strings.TrimSuffix(s, "Z") + "+00:00"
	}

	aware := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02 15:04:05.999999-07:00",
		"2006-01-02 15:04:05-07:00",
	}
	for _, f := range aware {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}

	naive := []string{
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05.999999",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05.999999999",
		"2006-01-02T15:04:05.999999",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}
	for _, f := range naive {
		if t, err := time.ParseInLocation(f, s, loc); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse time %q", s)
}
