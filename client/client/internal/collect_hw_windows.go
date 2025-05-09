//go:build windows

package internal

import (
	"fmt"
	"strings"
	"time"

	"github.com/StackExchange/wmi"
)

// computerModelWindows reads Win32_ComputerSystem.Model via WMI.
func computerModelWindows() string {
	type Win32_ComputerSystem struct {
		Model *string
	}
	var rows []Win32_ComputerSystem
	q := wmi.CreateQuery(&rows, "")
	if err := wmi.Query(q, &rows); err != nil {
		Log.Debugf("computer model WMI failed: %v", err)
		return ""
	}
	for _, r := range rows {
		if r.Model == nil {
			continue
		}
		model := strings.TrimSpace(*r.Model)
		if model != "" {
			return model
		}
	}
	return ""
}

// computerActivationWindows reads Win32_OperatingSystem.InstallDate via WMI.
func computerActivationWindows() string {
	type Win32_OperatingSystem struct {
		InstallDate *string
	}
	var rows []Win32_OperatingSystem
	q := wmi.CreateQuery(&rows, "")
	if err := wmi.Query(q, &rows); err != nil {
		Log.Debugf("install date WMI failed: %v", err)
		return ""
	}
	for _, r := range rows {
		if r.InstallDate == nil || strings.TrimSpace(*r.InstallDate) == "" {
			continue
		}
		t, err := parseWMIDateTime(*r.InstallDate)
		if err != nil {
			Log.Debugf("parse InstallDate %q: %v", *r.InstallDate, err)
			continue
		}
		return t.UTC().Format(time.RFC3339)
	}
	return ""
}

// parseWMIDateTime parses yyyymmddHHMMSS[.mmmmmm][±UUU] WMI datetime.
func parseWMIDateTime(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if len(s) < 14 {
		return time.Time{}, fmt.Errorf("WMI datetime too short")
	}
	return time.ParseInLocation("20060102150405", s[:14], time.Local)
}
