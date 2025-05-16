package updateactivation

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/carlosrabelo/tatuscan/tools/tools/internal/apiclient"
	"github.com/carlosrabelo/tatuscan/tools/tools/internal/i18n"
)

func TestNormalizeAndExtract(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"IFMT-0123", "123"},
		{"ifmt_99", "99"},
		{"m0042", "42"},
		{"host-7", "7"},
		{"pc0000", "0"},
		{"no-digits", ""},
		{"", ""},
	}
	for _, tt := range tests {
		if got := ExtractHostnameNumber(tt.in); got != tt.want {
			t.Errorf("ExtractHostnameNumber(%q)=%q want %q", tt.in, got, tt.want)
		}
	}
	if got := NormalizeNumber("AB-0010"); got != "10" {
		t.Fatalf("NormalizeNumber: %q", got)
	}
}

func TestParseDates(t *testing.T) {
	iso, ok := ParseDateISO("15/01/2024")
	if !ok || iso != "2024-01-15" {
		t.Fatalf("br date: %s %v", iso, ok)
	}
	iso, ok = ParseDateISO("2024-01-15")
	if !ok || iso != "2024-01-15" {
		t.Fatalf("iso date: %s %v", iso, ok)
	}
	if _, ok := ParseDateISO("not-a-date"); ok {
		t.Fatal("expected parse failure")
	}
	if got := NormalizeAPIDate("2024-01-15T10:00:00Z"); got != "2024-01-15" {
		t.Fatalf("api date: %s", got)
	}
}

func TestLoadReportAndProcess(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "relatorio.csv")
	csv := "NUMERO,DATA DA CARGA\n0123,01/02/2024\n99,skip-me\n"
	if err := os.WriteFile(path, []byte(csv), 0o644); err != nil {
		t.Fatal(err)
	}
	index, err := LoadReport(path)
	if err != nil {
		t.Fatal(err)
	}
	if index["123"] != "01/02/2024" {
		t.Fatalf("index: %v", index)
	}

	bomPath := filepath.Join(dir, "bom.csv")
	if err := os.WriteFile(bomPath, []byte("\uFEFFNUMERO,DATA DA CARGA\n42,02/02/2024\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	bomIndex, err := LoadReport(bomPath)
	if err != nil || bomIndex["42"] != "02/02/2024" {
		t.Fatalf("bom csv: %v err=%v", bomIndex, err)
	}

	var patched []string
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/machines", func(w http.ResponseWriter, r *http.Request) {
		act := "2020-01-01T00:00:00Z"
		same := "2024-02-01T00:00:00-04:00"
		_ = json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{"machine_id": "a", "hostname": "IFMT-0123", "computer_activation": act},
				{"machine_id": "b", "hostname": "IFMT-0123-dup", "computer_activation": same},
				{"machine_id": "c", "hostname": "plain"},
			},
		})
	})
	mux.HandleFunc("PATCH /api/machines/{id}", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		patched = append(patched, r.PathValue("id")+":"+string(body))
		w.WriteHeader(http.StatusOK)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	res, err := Process(context.Background(), apiclient.New(srv.URL+"/api", ""), index, i18n.New(i18n.EN))
	if err != nil {
		t.Fatal(err)
	}
	if res.TotalHosts != 3 || res.HostsWithNum != 2 || res.Matches != 2 || res.Updated != 1 {
		t.Fatalf("result: %+v", res)
	}
	if len(patched) != 1 || !strings.Contains(patched[0], `"computer_activation":"2024-02-01"`) {
		t.Fatalf("patched: %v", patched)
	}
}
