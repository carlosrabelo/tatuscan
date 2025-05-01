package model

import "time"

// Inventory is a machine inventory row.
type Inventory struct {
	MachineID          string     `json:"machine_id"`
	Hostname           string     `json:"hostname"`
	IP                 string     `json:"ip"`
	OS                 string     `json:"os"`
	OSVersion          *string    `json:"os_version"`
	CPUPercent         float64    `json:"cpu_percent"`
	MemoryTotalMB      int64      `json:"memory_total_mb"`
	MemoryUsedMB       *int64     `json:"memory_used_mb"`
	ComputerModel      *string    `json:"computer_model"`
	ComputerActivation *time.Time `json:"computer_activation"`
	ActivationDays     *int       `json:"activation_days"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          *time.Time `json:"updated_at"`
}

// OSCount is an OS distribution bucket.
type OSCount struct {
	OS    string `json:"os"`
	Count int    `json:"count"`
}

// VersionCount is an OS version distribution bucket.
type VersionCount struct {
	Version string `json:"version"`
	Count   int    `json:"count"`
}

// AgeCount is an age-range distribution bucket.
type AgeCount struct {
	Range string `json:"range"`
	Count int    `json:"count"`
}

// OnlineStats is fleet online/offline counts.
type OnlineStats struct {
	Online  int    `json:"online"`
	Offline int    `json:"offline"`
	After   string `json:"after"`
}
