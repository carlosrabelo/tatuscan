package deleteolder

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/carlosrabelo/tatuscan/tools/tools/internal/apiclient"
	"github.com/carlosrabelo/tatuscan/tools/tools/internal/i18n"
)

func strp(s string) *string { return &s }

func TestSortNewestFirst(t *testing.T) {
	recs := []apiclient.Machine{
		{MachineID: "old", UpdatedAt: strp("2020-01-01T00:00:00Z")},
		{MachineID: "new", UpdatedAt: strp("2024-06-01T12:00:00Z")},
		{MachineID: "mid", CreatedAt: strp("2022-01-01T00:00:00Z")},
	}
	got := sortNewestFirst(recs)
	if got[0].MachineID != "new" || got[2].MachineID != "old" {
		t.Fatalf("order: %s %s %s", got[0].MachineID, got[1].MachineID, got[2].MachineID)
	}
}

func TestProcessDryRunAndDelete(t *testing.T) {
	var deleted []string
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/machines", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{"machine_id": "keep", "hostname": "pc", "updated_at": "2024-01-02T00:00:00Z"},
				{"machine_id": "drop", "hostname": "pc", "updated_at": "2020-01-01T00:00:00Z"},
				{"machine_id": "solo", "hostname": "other", "updated_at": "2024-01-01T00:00:00Z"},
			},
		})
	})
	mux.HandleFunc("DELETE /api/machines/{id}", func(w http.ResponseWriter, r *http.Request) {
		deleted = append(deleted, r.PathValue("id"))
		w.WriteHeader(http.StatusOK)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	c := apiclient.New(srv.URL+"/api", "")

	dry, err := Process(context.Background(), c, true, io.Discard, i18n.New(i18n.EN))
	if err != nil {
		t.Fatal(err)
	}
	if dry.DuplicateHosts != 1 || dry.Marked != 1 || dry.Removed != 0 || len(deleted) != 0 {
		t.Fatalf("dry-run: %+v deleted=%v", dry, deleted)
	}

	live, err := Process(context.Background(), c, false, io.Discard, i18n.New(i18n.EN))
	if err != nil {
		t.Fatal(err)
	}
	if live.Removed != 1 || strings.Join(deleted, ",") != "drop" {
		t.Fatalf("live: %+v deleted=%v", live, deleted)
	}
}
