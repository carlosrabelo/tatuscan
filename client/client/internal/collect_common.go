//go:build windows || linux || darwin

package internal

import (
	"net"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// isEligibleIP reports whether ip may identify a host for collection:
// non-loopback IPv4, or non-loopback non-link-local IPv6 (including ULA).
func isEligibleIP(ip net.IP) bool {
	if ip == nil || ip.IsLoopback() {
		return false
	}
	if ip.To4() != nil {
		return true
	}
	return ip.To16() != nil && !ip.IsLinkLocalUnicast()
}

// hasEligibleIP prefers the first eligible IPv4, else the first eligible IPv6.
func hasEligibleIP(addrs []net.Addr) (string, bool) {
	var firstV6 string
	for _, addr := range addrs {
		ipnet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		ip := ipnet.IP
		if !isEligibleIP(ip) {
			continue
		}
		if ip.To4() != nil {
			return ip.String(), true
		}
		if firstV6 == "" {
			firstV6 = ip.String()
		}
	}
	if firstV6 != "" {
		return firstV6, true
	}
	return "", false
}

// collectCommonMetrics collects CPU and memory usage
func collectCommonMetrics() MachineMetrics {
	Log.Debug("Collecting CPU usage")
	// Non-zero interval needed for a meaningful first sample.
	cpuPercent, err := cpu.Percent(200*time.Millisecond, false)
	if err != nil {
		Log.Errorf("Failed to collect CPU usage: %v", err)
	}
	var cpuValue float64
	if len(cpuPercent) > 0 {
		cpuValue = cpuPercent[0]
		Log.Debugf("CPU usage: %.2f%%", cpuValue)
	}

	Log.Debug("Collecting memory information")
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		Log.Errorf("Failed to collect memory info: %v", err)
	}
	var memTotal, memUsed uint64
	if memInfo != nil {
		memTotal = memInfo.Total / (1024 * 1024)
		memUsed = memInfo.Used / (1024 * 1024)
		Log.Debugf("Memory Total: %d MB, Used: %d MB", memTotal, memUsed)
	}

	return MachineMetrics{
		CPUPercent:    cpuValue,
		MemoryTotalMB: memTotal,
		MemoryUsedMB:  memUsed,
	}
}
