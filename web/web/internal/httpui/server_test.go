package httpui_test

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/carlosrabelo/tatuscan/web/web/internal/apiclient"
	"github.com/carlosrabelo/tatuscan/web/web/internal/httpui"
)

func TestPages(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/machines", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"items": []any{}, "count": 0})
	})
	mux.HandleFunc("/api/stats/os", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"items": []any{}})
	})
	mux.HandleFunc("/api/stats/versions", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"items": []any{}})
	})
	mux.HandleFunc("/api/stats/age", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"items": []any{}})
	})
	mux.HandleFunc("/api/stats/online", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"online": 0, "offline": 0, "after": "2h0m0s"})
	})
	apiSrv := httptest.NewServer(mux)
	t.Cleanup(apiSrv.Close)

	h := httpui.New(apiclient.New(apiSrv.URL), slog.New(slog.NewTextHandler(os.Stderr, nil)), httpui.Options{
		OfflineAfter: 2 * time.Hour,
	}).Handler()

	for _, path := range []string{"/healthz", "/", "/report/", "/charts/"} {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, path, nil))
		if rr.Code != 200 {
			t.Fatalf("%s: %d %s", path, rr.Code, rr.Body.String())
		}
	}
}

func TestAPIDownShowsBanner(t *testing.T) {
	h := httpui.New(apiclient.New("http://127.0.0.1:1"), slog.New(slog.NewTextHandler(os.Stderr, nil)), httpui.Options{}).Handler()
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	if rr.Code != 200 {
		t.Fatalf("status %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Could not reach the API") {
		t.Fatalf("expected API error banner, body=%s", rr.Body.String())
	}
}

func TestLocaleFromConfig(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/machines", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"items": []any{}, "count": 0})
	})
	mux.HandleFunc("/api/stats/os", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"items": []any{}})
	})
	mux.HandleFunc("/api/stats/versions", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"items": []any{}})
	})
	mux.HandleFunc("/api/stats/age", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"items": []any{}})
	})
	mux.HandleFunc("/api/stats/online", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"online": 0, "offline": 0, "after": "2h0m0s"})
	})
	apiSrv := httptest.NewServer(mux)
	t.Cleanup(apiSrv.Close)

	en := httpui.New(apiclient.New(apiSrv.URL), slog.New(slog.NewTextHandler(os.Stderr, nil)), httpui.Options{
		DefaultLang: "en",
	}).Handler()
	rr := httptest.NewRecorder()
	en.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	if !strings.Contains(rr.Body.String(), `lang="en"`) || !strings.Contains(rr.Body.String(), "Report") {
		t.Fatalf("en page: %s", rr.Body.String())
	}
	if strings.Contains(rr.Body.String(), "Português") || strings.Contains(rr.Header().Get("Set-Cookie"), "tatuscan_lang") {
		t.Fatalf("language switcher leaked: %s cookie=%s", rr.Body.String(), rr.Header().Get("Set-Cookie"))
	}

	pt := httpui.New(apiclient.New(apiSrv.URL), slog.New(slog.NewTextHandler(os.Stderr, nil)), httpui.Options{
		DefaultLang: "pt",
	}).Handler()
	rr = httptest.NewRecorder()
	pt.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/?lang=en", nil))
	if !strings.Contains(rr.Body.String(), `lang="pt-BR"`) || !strings.Contains(rr.Body.String(), "Relatório") {
		t.Fatalf("pt page ignores query: %s", rr.Body.String())
	}
}

func TestReportShowsEnrichedFields(t *testing.T) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/machines", func(w http.ResponseWriter, _ *http.Request) {
		used := int64(1024)
		model := "ThinkPad"
		_ = json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{{
				"machine_id": "m1", "hostname": "lab-01", "ip": "10.0.0.5",
				"os": "linux", "os_version": "Ubuntu 22.04",
				"cpu_percent": 12.3, "memory_total_mb": 8192, "memory_used_mb": used,
				"computer_model": model, "computer_activation": now,
				"updated_at": now, "created_at": now,
			}},
			"count": 1,
		})
	})
	apiSrv := httptest.NewServer(mux)
	t.Cleanup(apiSrv.Close)

	h := httpui.New(apiclient.New(apiSrv.URL), slog.New(slog.NewTextHandler(os.Stderr, nil)), httpui.Options{
		OfflineAfter: 2 * time.Hour,
	}).Handler()
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/report/", nil))
	if rr.Code != 200 {
		t.Fatalf("status %d", rr.Code)
	}
	body := rr.Body.String()
	for _, want := range []string{"10.0.0.5", "ThinkPad", "12.3%", "Online", "1024 / 8192 MB"} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q in body", want)
		}
	}
}
