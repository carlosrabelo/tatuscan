//go:build windows

package internal

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v3/host"
)

// getOSVersionWindows returns a friendly string based on PlatformVersion
// Examples of PlatformVersion: "6.1.7601", "6.3.9600", "10.0.19045", "10.0.22631"
func getOSVersionWindows() string {
	info, err := host.Info()
	if err != nil {
		// use the global logger from agent
		if Log != nil {
			Log.Warnf("Failed to get host.Info(): %v", err)
		}
		return "Windows Unknown"
	}

	pv := strings.TrimSpace(info.PlatformVersion)
	if pv == "" {
		return "Windows Unknown"
	}

	// Normalize and separate in parts "major.minor.build"
	parts := strings.SplitN(pv, ".", 3)
	major, minor, build := "", "", ""
	if len(parts) > 0 {
		major = parts[0]
	}
	if len(parts) > 1 {
		minor = parts[1]
	}
	if len(parts) > 2 {
		// sometimes comes as "19045" or "19045 Build 19045" - we keep only the first numeric token
		build = strings.Fields(parts[2])[0]
	}

	base := fmt.Sprintf("%s.%s", major, minor) // ex.: "10.0", "6.1"

	switch base {
	case "6.1":
		return "Windows 7"
	case "6.2":
		return "Windows 8"
	case "6.3":
		return "Windows 8.1"
	case "10.0":
		// Heuristic: Windows 11 (Build >= 22000)
		if b, err := strconv.Atoi(build); err == nil && b >= 22000 {
			return "Windows 11"
		}
		return "Windows 10"
	default:
		// Generic fallback
		return "Windows Unknown"
	}
}
