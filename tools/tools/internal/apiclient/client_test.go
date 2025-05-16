package apiclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResolveAPIBase(t *testing.T) {
	t.Setenv("TATUSCAN_URL", "")
	if got := ResolveAPIBase(""); got != DefaultAPIBase {
		t.Fatalf("default: got %s", got)
	}
	if got := ResolveAPIBase("http://example:9/api/"); got != "http://example:9/api" {
		t.Fatalf("flag: got %s", got)
	}
	if got := ResolveAPIBase("http://example:9"); got != "http://example:9/api" {
		t.Fatalf("flag without /api: got %s", got)
	}
	t.Setenv("TATUSCAN_URL", "http://host:8040")
	if got := ResolveAPIBase("ignored"); got != "http://host:8040/api" {
		t.Fatalf("env without /api: got %s", got)
	}
	t.Setenv("TATUSCAN_URL", "http://host:8040/api")
	if got := ResolveAPIBase(""); got != "http://host:8040/api" {
		t.Fatalf("env with /api: got %s", got)
	}
}

func TestResolveToken(t *testing.T) {
	t.Setenv("TATUSCAN_API_TOKEN", "  secret  ")
	if got := ResolveToken(); got != "secret" {
		t.Fatalf("got %q", got)
	}
}

func TestListMachinesFallbackAndAuth(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/machines", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "gone", http.StatusNotFound)
	})
	mux.HandleFunc("GET /api/inventory", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{{"machine_id": "m1", "hostname": "h1"}},
		})
	})
	mux.HandleFunc("DELETE /api/machines/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok" {
			http.Error(w, "no", http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	c := New(srv.URL+"/api", "tok")
	items, err := c.ListMachines(context.Background())
	if err != nil || len(items) != 1 || items[0].MachineID != "m1" {
		t.Fatalf("list: %+v err=%v", items, err)
	}
	if err := c.DeleteMachine(context.Background(), "m1"); err != nil {
		t.Fatal(err)
	}
}

func TestHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusForbidden)
	}))
	t.Cleanup(srv.Close)
	c := New(srv.URL, "")
	_, err := c.ListMachines(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if st, ok := StatusOf(err); !ok || st != http.StatusForbidden {
		t.Fatalf("status: %v ok=%v err=%v", st, ok, err)
	}
}

func TestSetLogLevel(t *testing.T) {
	if err := SetLogLevel("DEBUG"); err != nil {
		t.Fatal(err)
	}
	if err := SetLogLevel("nope"); err == nil {
		t.Fatal("expected invalid level")
	}
}
