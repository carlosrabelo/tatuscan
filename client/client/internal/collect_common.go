//go:build windows || linux || darwin

package internal

import (
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

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
