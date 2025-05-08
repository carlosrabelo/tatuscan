//go:build darwin

package internal

import (
	"os"
	"os/exec"
	"strings"
	"time"
)

// computerModelDarwin returns hw.model (e.g. MacBookPro18,1).
func computerModelDarwin() string {
	out, err := exec.Command("sysctl", "-n", "hw.model").Output()
	if err != nil {
		Log.Debugf("computer model unavailable: %v", err)
		return ""
	}
	return strings.TrimSpace(string(out))
}

// computerActivationDarwin approximates first setup using /.InstalledOrNot
// or /var/db/.AppleSetupDone ModTime when present.
func computerActivationDarwin() string {
	for _, path := range []string{"/var/db/.AppleSetupDone", "/var/db/receipts"} {
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
	return ""
}
