//go:build darwin

package internal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"
)

var darwinVirtualIF = []string{
	"lo", "bridge", "awdl", "llw", "utun", "ipsec", "gif", "stf", "ap", "vmenet", "vmnet",
}

func isVirtualDarwinIF(name string) bool {
	n := strings.ToLower(name)
	for _, p := range darwinVirtualIF {
		if n == p || strings.HasPrefix(n, p) {
			return true
		}
	}
	return false
}

func platformUUIDDarwin() string {
	out, err := exec.Command("ioreg", "-rd1", "-c", "IOPlatformExpertDevice").Output()
	if err != nil {
		return ""
	}
	re := regexp.MustCompile(`"IOPlatformUUID"\s*=\s*"([^"]+)"`)
	m := re.FindSubmatch(out)
	if len(m) < 2 {
		return ""
	}
	return string(m[1])
}

// CollectData collects machine information for macOS.
func CollectData() (MachineInfo, error) {
	Log.Info("Starting data collection (darwin)")
	info := MachineInfo{Timestamp: time.Now().Format(time.RFC3339), OS: runtime.GOOS}

	var err error
	info.Hostname, err = os.Hostname()
	if err != nil {
		info.Hostname = "Unknown"
	}
	info.OSVersion = getOSVersionDarwin()

	var macAddresses []string
	var ipAddress string

	interfaces, err := net.Interfaces()
	if err != nil {
		return info, fmt.Errorf("failed to collect network interfaces: %v", err)
	}
	sort.Slice(interfaces, func(i, j int) bool { return interfaces[i].Index < interfaces[j].Index })

	for _, iface := range interfaces {
		if iface.Name == "" || iface.HardwareAddr.String() == "" {
			continue
		}
		if (iface.Flags&net.FlagLoopback) != 0 || (iface.Flags&net.FlagUp) == 0 {
			continue
		}
		if isVirtualDarwinIF(iface.Name) {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		hasIP := false
		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if !ok || ipnet.IP.IsLoopback() {
				continue
			}
			if v4 := ipnet.IP.To4(); v4 != nil {
				hasIP = true
				if ipAddress == "" {
					ipAddress = v4.String()
				}
				break
			}
			if ipnet.IP.To16() != nil && !ipnet.IP.IsLinkLocalUnicast() {
				hasIP = true
				if ipAddress == "" {
					ipAddress = ipnet.IP.String()
				}
			}
		}
		if !hasIP {
			continue
		}
		// Prefer globally administered MACs; still keep interface if only local admin
		mac := iface.HardwareAddr.String()
		if !isLocallyAdministeredMAC(iface.HardwareAddr) {
			macAddresses = append(macAddresses, mac)
		}
	}

	info.IP = ipAddress
	if info.IP == "" {
		info.IP = "0.0.0.0"
	}

	if len(macAddresses) == 0 {
		if uuid := platformUUIDDarwin(); uuid != "" {
			Log.Infof("Using IOPlatformUUID for MachineID (no physical MAC)")
			sum := sha256.Sum256([]byte(uuid))
			info.MachineID = hex.EncodeToString(sum[:])
		} else {
			return info, fmt.Errorf("no physical MAC or platform UUID available")
		}
	} else {
		sort.Strings(macAddresses)
		sum := sha256.Sum256([]byte(strings.Join(macAddresses, "|")))
		info.MachineID = hex.EncodeToString(sum[:])
	}

	common := collectCommonMetrics()
	info.CPUPercent = common.CPUPercent
	info.MemoryTotalMB = common.MemoryTotalMB
	info.MemoryUsedMB = common.MemoryUsedMB

	info.ComputerModel = computerModelDarwin()
	info.ComputerActivation = computerActivationDarwin()
	return info, nil
}

func isLocallyAdministeredMAC(hw net.HardwareAddr) bool {
	if len(hw) == 0 {
		return false
	}
	return (hw[0] & 0x02) == 0x02
}
