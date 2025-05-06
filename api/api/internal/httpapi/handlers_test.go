package httpapi_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/carlosrabelo/tatuscan/api/api/internal/httpapi"
	"github.com/carlosrabelo/tatuscan/api/api/internal/service"
	"github.com/carlosrabelo/tatuscan/api/api/internal/store"
)

func newTestServer(t *testing.T) http.Handler {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "api.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	svc, err := service.New(st, "America/Cuiaba")
	if err != nil {
		t.Fatal(err)
	}
	return httpapi.New(svc, st, slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})), httpapi.Options{}).Handler()
}

func newTestServerWithOpts(t *testing.T, opts httpapi.Options) http.Handler {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "api.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	svc, err := service.New(st, "America/Cuiaba")
	if err != nil {
		t.Fatal(err)
	}
	return httpapi.New(svc, st, slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})), opts).Handler()
}

func TestHealth(t *testing.T) {
	h := newTestServer(t)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/health", nil))
	if rr.Code != 200 {
		t.Fatalf("health: %d", rr.Code)
	}
}

func TestListEmpty(t *testing.T) {
	h := newTestServer(t)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/machines", nil))
	if rr.Code != 200 {
		t.Fatalf("list: %d %s", rr.Code, rr.Body.String())
	}
}

func TestCreateAndList(t *testing.T) {
	h := newTestServer(t)
	body := map[string]any{
		"machine_id": "abc", "hostname": "pc1", "ip": "10.0.0.1",
		"os": "linux", "cpu_percent": 1.5, "memory_total_mb": 4096,
	}
	b, _ := json.Marshal(body)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/machines", bytes.NewReader(b)))
	if rr.Code != 201 {
		t.Fatalf("create: %d %s", rr.Code, rr.Body.String())
	}
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/inventory", nil))
	if rr.Code != 200 {
		t.Fatalf("list: %d", rr.Code)
	}
}

func TestInvalidJSON(t *testing.T) {
	h := newTestServer(t)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/machines", bytes.NewReader([]byte("{bad"))))
	if rr.Code != 400 {
		t.Fatalf("want 400, got %d %s", rr.Code, rr.Body.String())
	}
}

func TestStatsOnline(t *testing.T) {
	h := newTestServer(t)
	body := map[string]any{
		"machine_id": "abc", "hostname": "pc1", "ip": "10.0.0.1",
		"os": "linux", "cpu_percent": 1.5, "memory_total_mb": 4096,
	}
	b, _ := json.Marshal(body)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/machines", bytes.NewReader(b)))
	if rr.Code != 201 {
		t.Fatalf("create: %d", rr.Code)
	}
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/stats/online?after=1h", nil))
	if rr.Code != 200 {
		t.Fatalf("stats online: %d %s", rr.Code, rr.Body.String())
	}
}
