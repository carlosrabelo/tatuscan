package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Config holds API server configuration.
type Config struct {
	Port         string
	DBPath       string
	Timezone     string
	LogLevel     string
	APIToken     string
	OfflineAfter time.Duration
	DefaultLang  string
}

// Load reads configuration from environment variables.
// A local .env file is optional: when missing, built-in defaults apply.
func Load() (Config, error) {
	_ = loadDotEnv(".env")
	cfg := Config{
		Port:        firstNonEmpty(os.Getenv("TATUSCAN_PORT"), os.Getenv("PORT"), "8040"),
		Timezone:    firstNonEmpty(os.Getenv("TIMEZONE"), "America/Cuiaba"),
		LogLevel:    firstNonEmpty(os.Getenv("LOG_LEVEL"), "INFO"),
		APIToken:    strings.TrimSpace(os.Getenv("TATUSCAN_API_TOKEN")),
		DefaultLang: firstNonEmpty(os.Getenv("TATUSCAN_LANG"), "en"),
	}

	offlineAfter, err := parseDurationEnv("TATUSCAN_OFFLINE_AFTER", 2*time.Hour)
	if err != nil {
		return Config{}, err
	}
	cfg.OfflineAfter = offlineAfter

	dbPath, err := resolveDBPath()
	if err != nil {
		return Config{}, err
	}
	cfg.DBPath = dbPath
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

func resolveDBPath() (string, error) {
	if uri := strings.TrimSpace(os.Getenv("SQLALCHEMY_DATABASE_URI")); uri != "" {
		path, err := sqlitePathFromURI(uri)
		if err != nil {
			return "", err
		}
		return path, nil
	}

	// Local default is /tmp so bare binary runs work without .env.
	// Docker/K8s override via TATUSCAN_DB_DIR=/data (or a volume path).
	dbDir := firstNonEmpty(os.Getenv("TATUSCAN_DB_DIR"), "/tmp")
	dbFile := firstNonEmpty(os.Getenv("TATUSCAN_DB_FILE"), "tatuscan.db")
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		return "", fmt.Errorf("create database directory %s: %w", dbDir, err)
	}
	return filepath.Join(dbDir, dbFile), nil
}

func sqlitePathFromURI(uri string) (string, error) {
	if !strings.HasPrefix(uri, "sqlite:") {
		return "", fmt.Errorf("unsupported database URI (only sqlite supported): %s", uri)
	}
	rest := strings.TrimPrefix(uri, "sqlite:")
	if i := strings.Index(rest, "?"); i >= 0 {
		rest = rest[:i]
	}
	rest = strings.TrimPrefix(rest, "///")
	rest = strings.TrimPrefix(rest, "//")
	if rest == "" || rest == ":memory:" {
		return "", fmt.Errorf("empty or unsupported sqlite path in URI")
	}
	if !strings.HasPrefix(rest, "/") {
		// relative path like ./data/tatuscan.db
		abs, err := filepath.Abs(rest)
		if err != nil {
			return "", err
		}
		rest = abs
	}
	dir := filepath.Dir(rest)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create database directory %s: %w", dir, err)
	}
	return rest, nil
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

// PortInt returns the listen port as int when needed.
func (c Config) PortInt() int {
	n, err := strconv.Atoi(c.Port)
	if err != nil {
		return 8040
	}
	return n
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
