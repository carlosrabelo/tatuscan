//go:build linux

package internal

import (
	"os"
	"strings"
	"time"
)

var linuxModelPlaceholders = []string{
	"to be filled by o.e.m.",
	"not specified",
	"not available",
	"default string",
	"system product name",
	"none",
	"unknown",
}

// computerModelLinux reads DMI product name when available.
func computerModelLinux() string {
	data, err := os.ReadFile("/sys/class/dmi/id/product_name")
	if err != nil {
		Log.Debugf("computer model unavailable: %v", err)
		return ""
	}
	model := strings.TrimSpace(string(data))
	if model == "" || isLinuxModelPlaceholder(model) {
		return ""
	}
	return model
}

func isLinuxModelPlaceholder(model string) bool {
	lower := strings.ToLower(model)
	for _, p := range linuxModelPlaceholders {
		if lower == p {
			return true
		}
	}
	return false
}

// computerActivationLinux approximates first-boot / install time.
// Prefer ModTime of /etc/machine-id (created at install or first boot);
// fall back to ModTime of "/".
func computerActivationLinux() string {
	for _, path := range []string{"/etc/machine-id", "/"} {
		fi, err := os.Stat(path)
		if err != nil {
			continue
		}
		t := fi.ModTime().UTC()
		if t.IsZero() || t.Year() < 2000 {
			continue
		}
		return t.Format(time.RFC3339)
	}
	Log.Debug("computer activation unavailable on linux")
	return ""
}
