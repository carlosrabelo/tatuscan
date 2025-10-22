//go:build windows || linux || darwin

package internal

// MachineInfo represents the collected machine data
type MachineInfo struct {
	MachineID     string  `json:"machine_id"`
	Hostname      string  `json:"hostname"`
	IP            string  `json:"ip"`
	OS            string  `json:"os"`
	OSVersion     string  `json:"os_version"`
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryTotalMB uint64  `json:"memory_total_mb"`
	MemoryUsedMB  uint64  `json:"memory_used_mb"`
	Timestamp     string  `json:"timestamp"`
}

// MachineMetrics holds common machine metrics
type MachineMetrics struct {
	CPUPercent    float64
	MemoryTotalMB uint64
	MemoryUsedMB  uint64
}
