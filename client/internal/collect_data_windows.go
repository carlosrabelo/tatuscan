//go:build windows

package internal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/StackExchange/wmi"
)

// virtualInterfacePatterns lists prefixes/suffixes of virtual network interfaces
var virtualInterfacePatterns = []string{
	// Linux (kept for function compatibility)
	"docker", "veth", "br-", "tun", "tap", "vmnet", "macvlan", "ipvlan", "wg", "wireguard", "dummy",
	// Windows
	"virtual", "vpn", "hyper-v", "vmware", "virtualbox", "teredo",
}

// isVirtualInterface checks if an interface is virtual based on its name
func isVirtualInterface(name string) bool {
	nameLower := strings.ToLower(name)
	for _, pattern := range virtualInterfacePatterns {
		if strings.Contains(nameLower, pattern) {
			return true
		}
	}
	return false
}

// isLocallyAdministeredMAC indicates if MAC has "locally administered" bit (typical of virtuals)
func isLocallyAdministeredMAC(hw net.HardwareAddr) bool {
	if len(hw) == 0 {
		return false
	}
	return (hw[0] & 0x02) == 0x02
}

// collectData collects machine information for Windows
func CollectData() (MachineInfo, error) {
	Log.Info("Starting data collection")
	info := MachineInfo{Timestamp: time.Now().Format(time.RFC3339)}

	// Hostname and basic OS
	Log.Debug("Collecting basic host information")
	info.OS = runtime.GOOS
	var err error
	info.Hostname, err = os.Hostname()
	if err != nil {
		Log.Warnf("Error to collect hostname: %v", err)
		info.Hostname = "Unknown"
	}
	Log.Debugf("OS detected: %s, Hostname: %s", info.OS, info.Hostname)

	// OS Version
	info.OSVersion = getOSVersionWindows()
	Log.Debugf("OSVersion detected: %s", info.OSVersion)

	// IP Address and MAC Addresses
	Log.Debug("Collecting MAC and IP addresses")
	macAddresses, err := collectMACsWindows()
	if err != nil {
		Log.Errorf("Error to collect MACs: %v", err)
		return info, fmt.Errorf("failed to collect MAC addresses: %v", err)
	}

	// Collect IP using net.Interfaces() (considering only non-virtual and UP NICs)
	Log.Debug("Starting IP collection on Windows")
	var ipAddress string
	interfaces, err := net.Interfaces()
	if err != nil {
		Log.Warnf("Error to collect network interfaces: %v", err)
	} else {
		for _, iface := range interfaces {
			if iface.Name == "" || (iface.Flags&net.FlagLoopback) != 0 || (iface.Flags&net.FlagUp) == 0 || isVirtualInterface(iface.Name) {
				continue
			}
			addrs, err := iface.Addrs()
			if err != nil {
				continue
			}
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
					ipAddress = ipnet.IP.String()
					Log.Debugf("IP found: %s", ipAddress)
					break
				}
			}
			if ipAddress != "" {
				break
			}
		}
	}

	if ipAddress == "" {
		Log.Warnf("No valid IPv4 address found")
	}
	info.IP = ipAddress

	if len(macAddresses) == 0 {
		Log.Errorf("No physical MAC address found; failed to generate MachineID")
		return info, fmt.Errorf("no physical MAC address available")
	}

	// Machine ID generation: Use all physical MAC addresses
	Log.Debug("Generating MachineID based on physical MACs")
	sort.Strings(macAddresses) // Sort for consistency
	idInput := strings.Join(macAddresses, "|")
	Log.Debugf("MACs used for MachineID: %s", idInput)
	hash := sha256.Sum256([]byte(idInput))
	info.MachineID = hex.EncodeToString(hash[:])
	Log.Debugf("MachineID generated: %s", info.MachineID)

	// Collect common metrics (CPU, Memory)
	commonInfo := collectCommonMetrics()
	info.CPUPercent = commonInfo.CPUPercent
	info.MemoryTotalMB = commonInfo.MemoryTotalMB
	info.MemoryUsedMB = commonInfo.MemoryUsedMB

	Log.Debugf("Data collected: %+v", info)
	return info, nil
}

// collectMACsWindows collects physical MACs on Windows.
// 1) Try WMI with broad filter (MACAddress != NULL).
// 2) Filter virtuals / disabled / locally-administered in Go.
// 3) If WMI fails or returns empty, fallback via net.Interfaces().
func collectMACsWindows() ([]string, error) {
	// --- Attempt 1: WMI (broad query) ---
	type adapter struct {
		MACAddress      *string
		NetEnabled      *bool
		PhysicalAdapter *bool
		Name            *string
	}

	Log.Debug("Querying Win32_NetworkAdapter via WMI (broad query)")
	var result []adapter

	q := wmi.CreateQuery(&result, `WHERE MACAddress IS NOT NULL`)
	wmiErr := wmi.Query(q, &result)
	if wmiErr == nil {
		macs := make([]string, 0, len(result))
		for _, r := range result {
			if r.MACAddress == nil || *r.MACAddress == "" {
				continue
			}
			name := ""
			if r.Name != nil {
				name = *r.Name
			}

			// Filter virtuals by known name
			if isVirtualInterface(name) {
				Log.Debugf("Ignoring MAC (virtual by name): %s (%s)", *r.MACAddress, name)
				continue
			}

			// Filter disabled (if field exists)
			if r.NetEnabled != nil && !*r.NetEnabled {
				Log.Debugf("Ignoring MAC (NetEnabled=false): %s (%s)", *r.MACAddress, name)
				continue
			}

			// Prefer physical when we can know
			if r.PhysicalAdapter != nil && !*r.PhysicalAdapter {
				Log.Debugf("Ignoring MAC (not physical): %s (%s)", *r.MACAddress, name)
				continue
			}

			// Normalize separator and avoid locally administered MACs
			normalized := strings.ReplaceAll(*r.MACAddress, "-", ":")
			if hw, err := net.ParseMAC(normalized); err == nil && isLocallyAdministeredMAC(hw) {
				Log.Debugf("Ignoring MAC (locally administered): %s (%s)", normalized, name)
				continue
			}

			macs = append(macs, normalized)
			Log.Debugf("Physical MAC included (WMI): %s (%s)", normalized, name)
		}

		if len(macs) > 0 {
			return macs, nil
		}
		Log.Warn("WMI returned empty after filters; proceeding to fallback via net.Interfaces()")
	} else {
		Log.Warnf("WMI query failed (%v); proceeding to fallback via net.Interfaces()", wmiErr)
	}

	// --- Fallback: net.Interfaces() ---
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to collect interfaces in fallback: %w", err)
	}

	var macs []string
	for _, iface := range ifaces {
		if iface.Name == "" {
			continue
		}
		if (iface.Flags&net.FlagLoopback) != 0 || (iface.Flags&net.FlagUp) == 0 {
			continue
		}
		if iface.HardwareAddr.String() == "" {
			continue
		}
		if isVirtualInterface(iface.Name) {
			Log.Debugf("Ignoring MAC (virtual by name): %s (%s)", iface.HardwareAddr, iface.Name)
			continue
		}
		if isLocallyAdministeredMAC(iface.HardwareAddr) {
			Log.Debugf("Ignoring MAC (locally administered): %s (%s)", iface.HardwareAddr, iface.Name)
			continue
		}
		macs = append(macs, iface.HardwareAddr.String())
		Log.Debugf("Physical MAC included (fallback): %s (%s)", iface.HardwareAddr, iface.Name)
	}

	if len(macs) == 0 {
		return nil, fmt.Errorf("no physical MAC address enabled was found (WMI and fallback)")
	}
	return macs, nil
}
