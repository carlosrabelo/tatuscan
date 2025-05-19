package deleteolder

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/carlosrabelo/tatuscan/tools/tools/internal/apiclient"
	"github.com/carlosrabelo/tatuscan/tools/tools/internal/i18n"
)

// Result is a dry-run or delete summary.
type Result struct {
	DuplicateHosts int
	Removed        int
	Marked         int
	Errors         []string
}

// Process deletes older inventory rows that share a hostname, keeping the newest.
func Process(ctx context.Context, c *apiclient.Client, dryRun bool, out io.Writer, cat i18n.Catalog) (Result, error) {
	items, err := c.ListMachines(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("list inventory: %w", err)
	}

	groups := map[string][]apiclient.Machine{}
	for _, item := range items {
		host := strings.TrimSpace(item.Hostname)
		if host == "" {
			continue
		}
		groups[host] = append(groups[host], item)
	}

	var res Result
	for host, recs := range groups {
		if len(recs) <= 1 {
			continue
		}
		res.DuplicateHosts++
		ordered := sortNewestFirst(recs)
		keep := ordered[0]
		fmt.Fprintln(out, cat.T("delete.dup", host, len(recs)))
		fmt.Fprintln(out, cat.T("delete.keep", keep.MachineID, displayTime(keep, cat)))
		for _, drop := range ordered[1:] {
			fmt.Fprintln(out, cat.T("delete.drop", drop.MachineID, displayTime(drop, cat)))
			if dryRun {
				res.Marked++
				continue
			}
			if err := c.DeleteMachine(ctx, drop.MachineID); err != nil {
				msg := cat.T("delete.err", drop.MachineID, err)
				slog.Error(msg)
				res.Errors = append(res.Errors, msg)
				continue
			}
			res.Removed++
		}
	}
	return res, nil
}

func sortNewestFirst(recs []apiclient.Machine) []apiclient.Machine {
	out := append([]apiclient.Machine(nil), recs...)
	sort.SliceStable(out, func(i, j int) bool {
		ui, ci := recTimes(out[i])
		uj, cj := recTimes(out[j])
		if !ui.Equal(uj) {
			return ui.After(uj)
		}
		return ci.After(cj)
	})
	return out
}

func recTimes(m apiclient.Machine) (updated, created time.Time) {
	updated = parseAPITime(deref(m.UpdatedAt))
	created = parseAPITime(deref(m.CreatedAt))
	if updated.IsZero() {
		updated = created
	}
	return updated, created
}

func displayTime(m apiclient.Machine, cat i18n.Catalog) string {
	t := parseAPITime(deref(m.UpdatedAt))
	if t.IsZero() {
		t = parseAPITime(deref(m.CreatedAt))
	}
	if t.IsZero() {
		return cat.T("delete.unknown")
	}
	return t.Format(time.RFC3339)
}

func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func parseAPITime(value string) time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}
	}
	if strings.HasSuffix(value, "Z") {
		value = strings.TrimSuffix(value, "Z") + "+00:00"
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		if t, err := time.Parse(layout, value); err == nil {
			return t
		}
	}
	slog.Debug("unexpected datetime format", "value", value)
	return time.Time{}
}
