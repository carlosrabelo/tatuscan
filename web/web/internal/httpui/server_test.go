package httpui_test

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
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

	for _, path := range []string{"/healthz"} {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, path, nil))
		if rr.Code != 200 {
			t.Fatalf("%s: %d %s", path, rr.Code, rr.Body.String())
		}
	}
}
