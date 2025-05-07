package store

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseStoredTimeNaiveUsesLocation(t *testing.T) {
	loc, err := time.LoadLocation("America/Cuiaba")
	if err != nil {
		t.Fatal(err)
	}
	got, err := parseStoredTime("2024-01-15 10:30:00", loc)
	if err != nil {
		t.Fatal(err)
	}
	if _, offset := got.Zone(); offset == 0 && got.Location() == time.UTC {
		// Must not treat naive as UTC when loc is Cuiaba
		if got.Hour() == 10 && got.Location() == time.UTC {
			t.Fatalf("naive time treated as UTC: %v", got)
		}
	}
	// Wall clock in Cuiaba should be 10:30
	local := got.In(loc)
	if local.Hour() != 10 || local.Minute() != 30 {
		t.Fatalf("want 10:30 in Cuiaba, got %v", local)
	}
}

func TestParseStoredTimeLegacyFormats(t *testing.T) {
	loc := time.UTC
	cases := []string{
		"2024-01-15T10:30:00Z",
		"2024-01-15T10:30:00.123456789Z",
		"2024-01-15 10:30:00",
		"2024-01-15 10:30:00.123456",
		"2024-01-15 10:30:00.123456-04:00",
		"2024-01-15",
	}
	for _, c := range cases {
		if _, err := parseStoredTime(c, loc); err != nil {
			t.Fatalf("parse %q: %v", c, err)
		}
	}
}

func TestNormalizeLegacyDatetimes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.db")
	st, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })

	loc, err := time.LoadLocation("America/Cuiaba")
	if err != nil {
		t.Fatal(err)
	}
	st.SetLocation(loc)

	_, err = st.db.Exec(`
		INSERT INTO inventory (
			machine_id, hostname, ip, os, cpu_percent, memory_total_mb,
			computer_activation, created_at, updated_at
		) VALUES (
			'm1', 'host1', '10.0.0.1', 'linux', 1.0, 1024,
			'2020-06-01 12:00:00.123456',
			'2024-01-15 10:30:00.654321',
			'2024-02-01 08:00:00'
		), (
			'm2', 'bad', '10.0.0.2', 'linux', 1.0, 1024,
			NULL,
			'not-a-date',
			NULL
		)`)
	if err != nil {
		t.Fatal(err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	res, err := st.NormalizeLegacyDatetimes(context.Background(), loc, logger)
	if err != nil {
		t.Fatal(err)
	}
	if res.Updated != 1 {
		t.Fatalf("updated=%d want 1", res.Updated)
	}
	if res.Skipped != 1 {
		t.Fatalf("skipped=%d want 1", res.Skipped)
	}

	res, err = st.NormalizeLegacyDatetimes(context.Background(), loc, logger)
	if err != nil {
		t.Fatal(err)
	}
	if res.Updated != 0 {
		t.Fatalf("second run updated=%d want 0", res.Updated)
	}

	inv, err := st.GetByID(context.Background(), "m1")
	if err != nil {
		t.Fatal(err)
	}
	local := inv.CreatedAt.In(loc)
	if local.Hour() != 10 || local.Minute() != 30 {
		t.Fatalf("created_at wall clock wrong after migrate: %v", local)
	}
}
