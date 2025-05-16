package apiclient

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// SetLogLevel configures the default slog logger from a Python-style level name.
func SetLogLevel(name string) error {
	var level slog.Level
	switch strings.ToUpper(strings.TrimSpace(name)) {
	case "", "INFO":
		level = slog.LevelInfo
	case "DEBUG":
		level = slog.LevelDebug
	case "WARNING", "WARN":
		level = slog.LevelWarn
	case "ERROR", "CRITICAL":
		level = slog.LevelError
	default:
		return fmt.Errorf("invalid log level %q", name)
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))
	return nil
}
