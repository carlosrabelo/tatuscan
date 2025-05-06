package httpapi_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
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

func TestRootIndex(t *testing.T) {
	h := newTestServer(t)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept", "application/json")
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("json root: %d %s", rr.Code, rr.Body.String())
	}
	var info map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &info); err != nil {
		t.Fatal(err)
	}
	if info["service"] != "tatuscan-api" {
		t.Fatalf("service: %v", info["service"])
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept", "text/html")
	h.ServeHTTP(rr, req)
	if rr.Code != 200 || !strings.Contains(rr.Body.String(), "JSON API only") {
		t.Fatalf("html root en: %d %s", rr.Code, rr.Body.String())
	}
	if strings.Contains(rr.Body.String(), "/?lang=") {
		t.Fatalf("language switcher leaked: %s", rr.Body.String())
	}

	hpt := newTestServerWithOpts(t, httpapi.Options{DefaultLang: "pt"})
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/?lang=en", nil)
	req.Header.Set("Accept", "text/html")
	hpt.ServeHTTP(rr, req)
	if rr.Code != 200 || !strings.Contains(rr.Body.String(), "só a API JSON") {
		t.Fatalf("html root pt from config: %d %s", rr.Code, rr.Body.String())
	}
}

func TestHealthAndCRUD(t *testing.T) {
	h := newTestServer(t)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/health", nil))
	if rr.Code != 200 {
		t.Fatalf("health: %d", rr.Code)
	}

	body := map[string]any{
		"machine_id": "abc", "hostname": "pc1", "ip": "10.0.0.1",
		"os": "linux", "cpu_percent": 1.5, "memory_total_mb": 4096,
	}
	b, _ := json.Marshal(body)
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/machines", bytes.NewReader(b)))
	if rr.Code != 201 {
		t.Fatalf("create: %d %s", rr.Code, rr.Body.String())
	}

	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/inventory", nil))
	if rr.Code != 200 {
		t.Fatalf("list: %d", rr.Code)
	}
	var list map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &list)
	if int(list["count"].(float64)) != 1 {
		t.Fatalf("count=%v", list["count"])
	}

	patch, _ := json.Marshal(map[string]any{"activation_days": 30})
	rr = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/machines/abc", bytes.NewReader(patch))
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("patch: %d %s", rr.Code, rr.Body.String())
	}

	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/stats/os", nil))
	if rr.Code != 200 {
		t.Fatalf("stats os: %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodDelete, "/api/machines/abc", nil))
	if rr.Code != 200 {
		t.Fatalf("delete: %d %s", rr.Code, rr.Body.String())
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

func TestAPITokenRequired(t *testing.T) {
	h := newTestServerWithOpts(t, httpapi.Options{APIToken: "secret"})
	body := map[string]any{
		"machine_id": "abc", "hostname": "pc1", "ip": "10.0.0.1",
		"os": "linux", "cpu_percent": 1.5, "memory_total_mb": 4096,
	}
	b, _ := json.Marshal(body)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/machines", bytes.NewReader(b)))
	if rr.Code != 401 {
		t.Fatalf("want 401 without token, got %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/machines", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer secret")
	h.ServeHTTP(rr, req)
	if rr.Code != 201 {
		t.Fatalf("want 201 with token, got %d %s", rr.Code, rr.Body.String())
	}

	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/machines", nil))
	if rr.Code != 200 {
		t.Fatalf("GET should stay open: %d", rr.Code)
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
	var stats map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &stats)
	if int(stats["online"].(float64)) != 1 || int(stats["offline"].(float64)) != 0 {
		t.Fatalf("unexpected stats: %v", stats)
	}
}
