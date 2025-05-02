package service

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/carlosrabelo/tatuscan/api/api/internal/model"
	"github.com/carlosrabelo/tatuscan/api/api/internal/store"
)

// Service implements inventory business rules.
type Service struct {
	store *store.Store
	loc   *time.Location
}

// New creates a Service.
func New(st *store.Store, timezone string) (*Service, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("load timezone %q: %w", timezone, err)
	}
	return &Service{store: st, loc: loc}, nil
}

// Location returns the configured timezone.
func (s *Service) Location() *time.Location { return s.loc }

// ListAll lists inventories with sorting.
func (s *Service) ListAll(ctx context.Context, orderBy, direction string) ([]model.Inventory, error) {
	return s.store.ListAll(ctx, orderBy, direction)
}

// GetByID returns inventory or NotFound.
func (s *Service) GetByID(ctx context.Context, machineID string) (model.Inventory, error) {
	inv, err := s.store.GetByID(ctx, machineID)
	if err == sql.ErrNoRows {
		return model.Inventory{}, NotFound("Inventory", machineID)
	}
	if err != nil {
		return model.Inventory{}, Database(err.Error())
	}
	return inv, nil
}

// CreateOrUpdate upserts inventory from a JSON-like map.
func (s *Service) CreateOrUpdate(ctx context.Context, data map[string]any) (model.Inventory, bool, error) {
	required := []string{"machine_id", "hostname", "ip", "os", "cpu_percent", "memory_total_mb"}
	var missing []string
	for _, k := range required {
		v, ok := data[k]
		if !ok || v == nil {
			missing = append(missing, k)
		}
	}
	if len(missing) > 0 {
		return model.Inventory{}, false, Validation(fmt.Sprintf("Missing required fields: %v", missing))
	}

	machineID, err := asString(data["machine_id"])
	if err != nil || strings.TrimSpace(machineID) == "" {
		return model.Inventory{}, false, Validation("invalid machine_id")
	}
	machineID = strings.TrimSpace(machineID)
	hostname, err := asString(data["hostname"])
	if err != nil || strings.TrimSpace(hostname) == "" {
		return model.Inventory{}, false, Validation("invalid hostname")
	}
	ip, err := asString(data["ip"])
	if err != nil || strings.TrimSpace(ip) == "" {
		return model.Inventory{}, false, Validation("invalid ip")
	}
	osName, err := asString(data["os"])
	if err != nil || strings.TrimSpace(osName) == "" {
		return model.Inventory{}, false, Validation("invalid os")
	}
	cpu, err := asFloat(data["cpu_percent"])
	if err != nil {
		return model.Inventory{}, false, Validation("invalid cpu_percent")
	}
	memTotal, err := asInt64(data["memory_total_mb"])
	if err != nil {
		return model.Inventory{}, false, Validation("invalid memory_total_mb")
	}

	existing, err := s.store.GetByID(ctx, machineID)
	if err != nil && err != sql.ErrNoRows {
		return model.Inventory{}, false, Database(err.Error())
	}

	now := time.Now().In(s.loc)

	if err == nil {
		existing.Hostname = truncate(hostname, 100)
		existing.IP = truncate(ip, 45)
		existing.OS = truncate(osName, 100)
		if _, ok := data["os_version"]; ok {
			existing.OSVersion = optionalString(data, "os_version")
		}
		existing.CPUPercent = cpu
		existing.MemoryTotalMB = memTotal
		if v, ok := data["memory_used_mb"]; ok {
			if v == nil {
				existing.MemoryUsedMB = nil
			} else if n, err := asInt64(v); err == nil {
				existing.MemoryUsedMB = &n
			}
		}
		if _, ok := data["computer_model"]; ok {
			existing.ComputerModel = optionalString(data, "computer_model")
		}
		if _, ok := data["computer_activation"]; ok {
			dt, err := s.parseDatetime(data["computer_activation"])
			if err != nil {
				return model.Inventory{}, false, err
			}
			existing.ComputerActivation = dt
		}
		if _, ok := data["activation_days"]; ok {
			days, err := optionalInt(data["activation_days"])
			if err != nil {
				return model.Inventory{}, false, Validation("invalid activation_days")
			}
			existing.ActivationDays = days
		}
		existing.UpdatedAt = &now
		if err := s.store.Update(ctx, existing); err != nil {
			return model.Inventory{}, false, Database(err.Error())
		}
		return existing, false, nil
	}

	inv := model.Inventory{
		MachineID:     machineID,
		Hostname:      truncate(hostname, 100),
		IP:            truncate(ip, 45),
		OS:            truncate(osName, 100),
		OSVersion:     optionalString(data, "os_version"),
		CPUPercent:    cpu,
		MemoryTotalMB: memTotal,
		ComputerModel: optionalString(data, "computer_model"),
		CreatedAt:     now,
	}
	if v, ok := data["memory_used_mb"]; ok && v != nil {
		if n, err := asInt64(v); err == nil {
			inv.MemoryUsedMB = &n
		}
	}
	if _, ok := data["computer_activation"]; ok {
		dt, err := s.parseDatetime(data["computer_activation"])
		if err != nil {
			return model.Inventory{}, false, err
		}
		inv.ComputerActivation = dt
	}
	if _, ok := data["activation_days"]; ok {
		days, err := optionalInt(data["activation_days"])
		if err != nil {
			return model.Inventory{}, false, Validation("invalid activation_days")
		}
		inv.ActivationDays = days
	}
	if err := s.store.Insert(ctx, inv); err != nil {
		return model.Inventory{}, false, Database(err.Error())
	}
	return inv, true, nil
}

// SerializeInventory converts inventory to API JSON map with TZ ISO times.
func (s *Service) SerializeInventory(inv model.Inventory) map[string]any {
	return map[string]any{
		"machine_id":          inv.MachineID,
		"hostname":            inv.Hostname,
		"ip":                  inv.IP,
		"os":                  inv.OS,
		"os_version":          inv.OSVersion,
		"cpu_percent":         inv.CPUPercent,
		"memory_total_mb":     inv.MemoryTotalMB,
		"memory_used_mb":      inv.MemoryUsedMB,
		"computer_model":      inv.ComputerModel,
		"computer_activation": s.toISO(inv.ComputerActivation),
		"activation_days":     inv.ActivationDays,
		"created_at":          s.toISO(&inv.CreatedAt),
		"updated_at":          s.toISO(inv.UpdatedAt),
	}
}

func (s *Service) toISO(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.In(s.loc).Format(time.RFC3339Nano)
}

func (s *Service) parseDatetime(value any) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	switch v := value.(type) {
	case time.Time:
		t := v
		if t.Location() == nil {
			t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), s.loc)
		} else {
			t = t.In(s.loc)
		}
		return &t, nil
	case float64:
		t := time.Unix(int64(v), 0).In(s.loc)
		return &t, nil
	case int:
		t := time.Unix(int64(v), 0).In(s.loc)
		return &t, nil
	case int64:
		t := time.Unix(v, 0).In(s.loc)
		return &t, nil
	case string:
		str := strings.TrimSpace(v)
		if str == "" {
			return nil, nil
		}
		if strings.HasSuffix(str, "Z") {
			str = strings.TrimSuffix(str, "Z") + "+00:00"
		}
		if t, err := time.Parse(time.RFC3339Nano, str); err == nil {
			t = t.In(s.loc)
			return &t, nil
		}
		if t, err := time.Parse(time.RFC3339, str); err == nil {
			t = t.In(s.loc)
			return &t, nil
		}
		for _, layout := range []string{"2006-01-02", "02/01/2006"} {
			if t, err := time.ParseInLocation(layout, str, s.loc); err == nil {
				return &t, nil
			}
		}
		return nil, Validation(fmt.Sprintf("Invalid datetime format: %q", v))
	default:
		return nil, Validation(fmt.Sprintf("Invalid datetime format: %v", value))
	}
}

func truncate(s string, n int) string {
	if n <= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return string([]rune(s)[:n])
}

func asString(v any) (string, error) {
	if v == nil {
		return "", fmt.Errorf("empty")
	}
	switch t := v.(type) {
	case string:
		return t, nil
	case fmt.Stringer:
		return t.String(), nil
	default:
		return fmt.Sprint(v), nil
	}
}

func asFloat(v any) (float64, error) {
	switch t := v.(type) {
	case float64:
		return t, nil
	case float32:
		return float64(t), nil
	case int:
		return float64(t), nil
	case int64:
		return float64(t), nil
	case string:
		return strconv.ParseFloat(t, 64)
	default:
		return 0, fmt.Errorf("not a number")
	}
}

func asInt64(v any) (int64, error) {
	switch t := v.(type) {
	case float64:
		return int64(t), nil
	case int:
		return int64(t), nil
	case int64:
		return t, nil
	case string:
		return strconv.ParseInt(t, 10, 64)
	default:
		return 0, fmt.Errorf("not an int")
	}
}

func optionalString(data map[string]any, key string) *string {
	v, ok := data[key]
	if !ok || v == nil {
		return nil
	}
	s, err := asString(v)
	if err != nil {
		return nil
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}

func optionalInt(v any) (*int, error) {
	if v == nil {
		return nil, nil
	}
	n, err := asInt64(v)
	if err != nil {
		return nil, err
	}
	i := int(n)
	return &i, nil
}
