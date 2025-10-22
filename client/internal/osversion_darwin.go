//go:build darwin

package internal

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// getOSVersionLinux returns the version of the operating system for Linux
func getOSVersionLinux() string {
	Log.Debug("Starting Linux distribution identification")

	// Helper function to read a file and return its content as string
	readFile := func(filename string) (string, error) {
		data, err := os.ReadFile(filename)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(data)), nil
	}

	// 1. Try to read /etc/os-release (most modern method)
	osReleaseData, err := readFile("/etc/os-release")
	if err == nil {
		Log.Debug("Arquivo /etc/os-release encontrado, processando...")
		osRelease := make(map[string]string)
		lines := strings.Split(osReleaseData, "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			value := strings.Trim(strings.TrimSpace(parts[1]), `"`)
			osRelease[key] = value
		}

		// Extract relevant information
		name, ok := osRelease["NAME"]
		if !ok {
			name = osRelease["ID"]
		}
		if name == "" {
			name = "Linux"
		}

		version, ok := osRelease["VERSION"]
		if !ok {
			version = osRelease["VERSION_ID"]
		}
		if version == "" {
			Log.Debug("Version not found in /etc/os-release")
		}

		if name != "" {
			if version != "" {
				Log.Debugf("Distribution identified via /etc/os-release: %s %s", name, version)
				return fmt.Sprintf("%s %s", name, version)
			}
			Log.Debugf("Distribution identified via /etc/os-release: %s", name)
			return name
		}
	} else {
		Log.Debugf("/etc/os-release file not found or error reading: %v", err)
	}

	// 2. Try to read /etc/lsb-release (for Debian/Ubuntu based distributions)
	lsbReleaseData, err := readFile("/etc/lsb-release")
	if err == nil {
		Log.Debug("Arquivo /etc/lsb-release encontrado, processando...")
		lsbRelease := make(map[string]string)
		lines := strings.Split(lsbReleaseData, "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			value := strings.Trim(strings.TrimSpace(parts[1]), `"`)
			lsbRelease[key] = value
		}

		name := lsbRelease["DISTRIB_ID"]
		version := lsbRelease["DISTRIB_RELEASE"]
		if name != "" {
			if version != "" {
				Log.Debugf("Distribution identified via /etc/lsb-release: %s %s", name, version)
				return fmt.Sprintf("%s %s", name, version)
			}
			Log.Debugf("Distribution identified via /etc/lsb-release: %s", name)
			return name
		}
	} else {
		Log.Debugf("/etc/lsb-release file not found or error reading: %v", err)
	}

	// 3. Try to read /etc/redhat-release (for Red Hat based distributions)
	redhatReleaseData, err := readFile("/etc/redhat-release")
	if err == nil {
		Log.Debugf("Arquivo /etc/redhat-release encontrado: %s", strings.TrimSpace(redhatReleaseData))
		return strings.TrimSpace(redhatReleaseData)
	} else {
		Log.Debugf("/etc/redhat-release file not found or error reading: %v", err)
	}

	// 4. Fallback para uname -r
	Log.Debug("No distribution file found, using fallback uname -r")
	cmd := exec.Command("uname", "-r")
	output, err := cmd.Output()
	if err != nil {
		Log.Warnf("Error executing uname: %v", err)
		return "Linux Unknown"
	}

	version := strings.TrimSpace(string(output))
	Log.Debugf("Kernel version detected via uname: %s", version)
	return "Linux " + version
}
