//go:build darwin

package internal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

// virtualInterfacePatterns lists prefixes/suffixes of virtual network interfaces
var virtualInterfacePatterns = []string{
	// Linux
	"docker", "veth", "br-", "tun", "tap", "vmnet", "macvlan", "ipvlan", "wg", "wireguard", "dummy",
	// Windows (kept only for function compatibility)
	"Virtual", "VPN", "Hyper-V", "VMware", "VirtualBox", "Teredo",
}

// isVirtualInterface checks if an interface is virtual based on its name
func isVirtualInterface(name string) bool {
	nameLower := strings.ToLower(name)
	for _, pattern := range virtualInterfacePatterns {
		if strings.Contains(nameLower, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

// isLocallyAdministeredMAC returns true if the MAC has the "locally administered" bit set
func isLocallyAdministeredMAC(hw net.HardwareAddr) bool {
	if len(hw) == 0 {
		return false
	}
	// Bit 1 (0x02) from first octet indicates "locally administered"
	return (hw[0] & 0x02) == 0x02
}

// isVirtualLinuxBySysfs checks /sys/class/net/<iface> symlink for "/virtual/" path
func isVirtualLinuxBySysfs(name string) bool {
	p := filepath.Join("/sys/class/net", name)
	link, err := os.Readlink(p)
	if err == nil && strings.Contains(link, "/virtual/") {
		return true
	}
	// fallback by additional patterns (case sysfs not accessible)
	return isVirtualInterface(name)
}

// collectData collects machine information for Linux
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
	Log.Debug("Running collection for Linux")
	info.OSVersion = getOSVersionLinux()
	Log.Debugf("OSVersion detected: %s", info.OSVersion)

	// IP Address and MAC Addresses
	Log.Debug("Collecting MAC and IP addresses")
	var macAddresses []string
	var ipAddress string

	interfaces, err := net.Interfaces()
	if err != nil {
		Log.Errorf("Error to collect network interfaces: %v", err)
		return info, fmt.Errorf("failed to collect network interfaces: %v", err)
	}
	Log.Debug("Network interfaces detected:")
	sort.Slice(interfaces, func(i, j int) bool {
		return interfaces[i].Index < interfaces[j].Index
	})
	foundValidInterface := false
	for _, iface := range interfaces {
		if iface.Name == "" {
			Log.Debugf("Interface without name, ignored")
			continue
		}

		// Basic flags
		if iface.HardwareAddr.String() == "" {
			Log.Debugf("Interface %s ignored: empty MAC", iface.Name)
			continue
		}
		if (iface.Flags & net.FlagLoopback) != 0 {
			Log.Debugf("Interface %s ignored: loopback", iface.Name)
			continue
		}
		if (iface.Flags & net.FlagUp) == 0 {
			Log.Debugf("Interface %s ignored: interface DOWN", iface.Name)
			continue
		}

		// Virtual by name/sysfs
		if isVirtualLinuxBySysfs(iface.Name) {
			Log.Debugf("Interface %s ignored: virtual (sysfs/pattern)", iface.Name)
			continue
		}

		// Locally administered MAC - typical of virtuals/containers
		if isLocallyAdministeredMAC(iface.HardwareAddr) {
			Log.Debugf("Interface %s ignored: locally administered MAC (%s)", iface.Name, iface.HardwareAddr)
			continue
		}

		// Valid IP (IPv4 non-loopback)
		addrs, err := iface.Addrs()
		if err != nil {
			Log.Errorf("Error to collect addresses from interface %s: %v", iface.Name, err)
			continue
		}
		hasValidIP := false
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
				hasValidIP = true
				if ipAddress == "" {
					ipAddress = ipnet.IP.String()
					Log.Debugf("Selecionada interface %s com IP %s", iface.Name, ipAddress)
				}
				break
			}
		}
		if !hasValidIP {
			Log.Debugf("Interface %s ignored: no valid IPv4", iface.Name)
			continue
		}

		// MAC coletado
		mac := iface.HardwareAddr.String()
		macAddresses = append(macAddresses, mac)
		Log.Debugf("Physical MAC included: %s (interface %s)", mac, iface.Name)
		foundValidInterface = true
	}
	if !foundValidInterface {
		Log.Warnf("No valid physical network interface found")
		return info, fmt.Errorf("no valid physical network interface found")
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
