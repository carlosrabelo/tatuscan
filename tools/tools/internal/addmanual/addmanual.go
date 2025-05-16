package addmanual

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/carlosrabelo/tatuscan/tools/tools/internal/apiclient"
)

// Entry is a manual inventory row.
type Entry struct {
	Hostname      string
	OS            string
	OSVersion     string
	MachineID     string
	IP            string
	CPUPercent    float64
	MemoryTotalMB int64
	MemoryUsedMB  int64
	Activation    string
	Salt          string
}

// GenerateMachineID hashes hostname + optional salt (same as the Python tool).
func GenerateMachineID(hostname, salt string) string {
	sum := sha256.Sum256(append([]byte(strings.TrimSpace(hostname)), []byte(salt)...))
	return hex.EncodeToString(sum[:])
}

// NormalizeActivation converts known date formats to yyyy-mm-dd, or returns the raw value.
func NormalizeActivation(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	for _, layout := range []string{"2006-01-02", "02/01/2006", "2006/01/02"} {
		if t, err := time.Parse(layout, value); err == nil {
			return t.Format("2006-01-02")
		}
	}
	return value
}

// Payload builds the POST body. machine_id is generated when omitted.
func Payload(e Entry) map[string]any {
	id := strings.TrimSpace(e.MachineID)
	if id == "" {
		id = GenerateMachineID(e.Hostname, e.Salt)
	}
	body := map[string]any{
		"machine_id":      id,
		"hostname":        e.Hostname,
		"ip":              e.IP,
		"os":              e.OS,
		"os_version":      e.OSVersion,
		"cpu_percent":     e.CPUPercent,
		"memory_total_mb": e.MemoryTotalMB,
		"memory_used_mb":  e.MemoryUsedMB,
	}
	if act := NormalizeActivation(e.Activation); act != "" {
		body["computer_activation"] = act
	}
	return body
}

// Send POSTs the entry, or logs the payload when dryRun is set.
func Send(ctx context.Context, c *apiclient.Client, e Entry, dryRun bool) error {
	body := Payload(e)
	if dryRun {
		slog.Info("dry-run POST /machines", "payload", body)
		return nil
	}
	msg, err := c.CreateMachine(ctx, body)
	if err != nil {
		return fmt.Errorf("send %s: %w", e.Hostname, err)
	}
	slog.Info("machine processed", "hostname", e.Hostname, "message", msg)
	return nil
}
