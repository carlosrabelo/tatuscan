package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"
)

// Config holds web panel configuration.
type Config struct {
	Port         string
	APIURL       string
	LogLevel     string
	OfflineAfter time.Duration
	DefaultLang  string
}

// Load reads configuration from environment.
// A local .env file is optional: when missing, built-in defaults apply.
func Load() (Config, error) {
	_ = loadDotEnv(".env")
	cfg := Config{
		Port:        firstNonEmpty(os.Getenv("TATUSCAN_PORT"), os.Getenv("PORT"), "8050"),
		APIURL:      strings.TrimRight(firstNonEmpty(os.Getenv("TATUSCAN_API_URL"), "http://127.0.0.1:8040"), "/"),
		LogLevel:    firstNonEmpty(os.Getenv("LOG_LEVEL"), "INFO"),
		DefaultLang: firstNonEmpty(os.Getenv("TATUSCAN_LANG"), "en"),
	}
	if cfg.APIURL == "" {
		return Config{}, fmt.Errorf("TATUSCAN_API_URL is required")
	}
	offlineAfter, err := parseDurationEnv("TATUSCAN_OFFLINE_AFTER", 2*time.Hour)
	if err != nil {
		return Config{}, err
	}
	cfg.OfflineAfter = offlineAfter
	return cfg, nil
}

func parseDurationEnv(key string, def time.Duration) (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return def, nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	}
	if d <= 0 {
		return 0, fmt.Errorf("invalid %s: must be positive", key)
	}
	return d, nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// loadDotEnv loads KEY=VALUE pairs from path into the process environment.
// Missing file is ignored. Existing env vars are never overwritten.
func loadDotEnv(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		if key == "" {
			continue
		}
		if len(val) >= 2 {
			if (val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'') {
				val = val[1 : len(val)-1]
			}
		}
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, val)
		}
	}
	return nil
}

// NewLogger builds a text slog logger from LOG_LEVEL-style names.
func NewLogger(level string) *slog.Logger {
	var lv slog.Level
	switch strings.ToUpper(level) {
	case "DEBUG":
		lv = slog.LevelDebug
	case "WARN", "WARNING":
		lv = slog.LevelWarn
	case "ERROR":
		lv = slog.LevelError
	default:
		lv = slog.LevelInfo
	}
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: lv}))
}
