//go:build darwin

package internal

import (
	"os/exec"
	"strings"
)

// getOSVersionDarwin returns macOS product name and version via sw_vers.
func getOSVersionDarwin() string {
	name := runTrimmed("sw_vers", "-productName")
	version := runTrimmed("sw_vers", "-productVersion")
	if name == "" {
		name = "macOS"
	}
	if version == "" {
		return name
	}
	return name + " " + version
}

func runTrimmed(name string, args ...string) string {
	out, err := exec.Command(name, args...).Output()
	if err != nil {
		Log.Debugf("%s failed: %v", name, err)
		return ""
	}
	return strings.TrimSpace(string(out))
}
