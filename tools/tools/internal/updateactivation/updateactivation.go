package updateactivation

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/carlosrabelo/tatuscan/tools/tools/internal/apiclient"
	"github.com/carlosrabelo/tatuscan/tools/tools/internal/i18n"
)

var hostnamePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)ifmt[-_]?(\d+)$`),
	regexp.MustCompile(`(?i)ifmt[-_]?(\d+)`),
	regexp.MustCompile(`(?i)m(\d+)$`),
	regexp.MustCompile(`(?i)m(\d+)`),
	regexp.MustCompile(`(\d+)$`),
}

var digitsRE = regexp.MustCompile(`\d+`)

// Result summarizes a CSV-driven activation update.
type Result struct {
	TotalHosts   int
	HostsWithNum int
	Matches      int
	Updated      int
	Errors       []string
}

// LoadReport indexes NUMERO -> DATA DA CARGA from a CSV.
func LoadReport(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return loadReport(f)
}

func loadReport(r io.Reader) (map[string]string, error) {
	reader := csv.NewReader(r)
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read csv header: %w", err)
	}
	numIdx, dateIdx := -1, -1
	for i, col := range header {
		col = strings.TrimSpace(col)
		if i == 0 {
			col = strings.TrimPrefix(col, "\uFEFF")
		}
		switch strings.ToUpper(col) {
		case "NUMERO":
			numIdx = i
		case "DATA DA CARGA":
			dateIdx = i
		}
	}
	if numIdx < 0 || dateIdx < 0 {
		return nil, fmt.Errorf("csv must have NUMERO and DATA DA CARGA columns")
	}
	out := map[string]string{}
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read csv row: %w", err)
		}
		if numIdx >= len(row) || dateIdx >= len(row) {
			continue
		}
		num := NormalizeNumber(row[numIdx])
		date := strings.TrimSpace(row[dateIdx])
		if num == "" || date == "" {
			continue
		}
		out[num] = date
	}
	return out, nil
}

// NormalizeNumber keeps digits and strips leading zeros.
func NormalizeNumber(value string) string {
	var b strings.Builder
	for _, r := range value {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	digits := b.String()
	if digits == "" {
		return ""
	}
	normalized := strings.TrimLeft(digits, "0")
	if normalized == "" {
		return "0"
	}
	return normalized
}

// ExtractHostnameNumber returns the numeric id from hostnames like IFMT-1234 or m1234.
func ExtractHostnameNumber(hostname string) string {
	hostname = strings.TrimSpace(hostname)
	if hostname == "" {
		return ""
	}
	for _, re := range hostnamePatterns {
		if m := re.FindStringSubmatch(hostname); len(m) > 1 {
			return NormalizeNumber(m[1])
		}
	}
	groups := digitsRE.FindAllString(hostname, -1)
	if len(groups) == 0 {
		return ""
	}
	return NormalizeNumber(groups[len(groups)-1])
}

// ParseDateISO converts common date strings to yyyy-mm-dd.
func ParseDateISO(value string) (string, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", false
	}
	for _, layout := range []string{"02/01/2006", "2006-01-02", "02-01-2006"} {
		if t, err := time.Parse(layout, value); err == nil {
			return t.Format("2006-01-02"), true
		}
	}
	slog.Warn("could not convert date to ISO", "value", value)
	return "", false
}

// NormalizeAPIDate reduces an API datetime to yyyy-mm-dd.
func NormalizeAPIDate(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if strings.HasSuffix(value, "Z") {
		value = strings.TrimSuffix(value, "Z") + "+00:00"
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		if t, err := time.Parse(layout, value); err == nil {
			return t.Format("2006-01-02")
		}
	}
	slog.Debug("unexpected API date", "value", value)
	return ""
}

// Process PATCHes computer_activation when the CSV number matches the hostname.
func Process(ctx context.Context, c *apiclient.Client, index map[string]string, cat i18n.Catalog) (Result, error) {
	items, err := c.ListMachines(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("list inventory: %w", err)
	}
	var res Result
	res.TotalHosts = len(items)
	for _, item := range items {
		num := ExtractHostnameNumber(item.Hostname)
		if num == "" {
			continue
		}
		res.HostsWithNum++
		reportDate, ok := index[num]
		if !ok {
			continue
		}
		iso, ok := ParseDateISO(reportDate)
		if !ok {
			continue
		}
		res.Matches++
		current := ""
		if item.ComputerActivation != nil {
			current = NormalizeAPIDate(*item.ComputerActivation)
		}
		if current == iso {
			continue
		}
		if err := c.PatchActivation(ctx, item.MachineID, iso); err != nil {
			msg := cat.T("update.err", item.MachineID, err)
			slog.Error(msg)
			res.Errors = append(res.Errors, msg)
			continue
		}
		res.Updated++
	}
	return res, nil
}
