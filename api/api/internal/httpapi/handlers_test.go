package httpapi_test

import (
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
