//go:build windows || linux || darwin

package internal

import (
	"os"
	"strings"
	"time"
)

const (
	DefaultInterval    = 60 * time.Second
	DefaultServerURL   = "http://127.0.0.1:8040"
	EnvServerURL       = "TATUSCAN_URL"
	EnvCollectInterval = "TATUSCAN_INTERVAL"
	EnvAPIToken        = "TATUSCAN_API_TOKEN"
	AgentVersion       = "0.0.1"
)

// GetServerURL retrieves the base server URL from the environment variable.
// When unset (and no .env), defaults to DefaultServerURL.
func GetServerURL() string {
	Log.Debug("Getting ServerURL from environment variable")
	base := strings.TrimSpace(os.Getenv(EnvServerURL))
	if base == "" {
		base = DefaultServerURL
		Log.Debugf("%s not set; using default %s", EnvServerURL, DefaultServerURL)
	}
	base = strings.TrimRight(base, "/")
	url := base + "/api/machines"
	Log.Debugf("Final ServerURL: %s", url)
	return url
}

// ResolveInterval resolves the collection interval from flag, env, or default.
func ResolveInterval(flagVal string) time.Duration {
	if flagVal != "" {
		if d, err := time.ParseDuration(flagVal); err == nil {
			return d
		}
		Log.Fatalf("Invalid value for -interval: %s", flagVal)
	}
	if env := strings.TrimSpace(os.Getenv(EnvCollectInterval)); env != "" {
		if d, err := time.ParseDuration(env); err == nil {
			return d
		}
		Log.Fatalf("Invalid value for %s: %s", EnvCollectInterval, env)
	}
	return DefaultInterval
}
