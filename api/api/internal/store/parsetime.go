package store

import (
	"fmt"
	"strings"
	"time"
)

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
