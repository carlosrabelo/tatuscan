package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/carlosrabelo/tatuscan/api/api/internal/i18n"
	"github.com/carlosrabelo/tatuscan/api/api/internal/service"
	"github.com/carlosrabelo/tatuscan/api/api/internal/store"
)

// Server is the HTTP API.
type Server struct {
	svc          *service.Service
	store        *store.Store
	logger       *slog.Logger
	mux          *http.ServeMux
	apiToken     string
	offlineAfter time.Duration
	defaultLang  string
}

// Options configures optional API server behavior.
type Options struct {
	APIToken     string
	OfflineAfter time.Duration
	DefaultLang  string
}

// New creates the API server and registers routes.
func New(svc *service.Service, st *store.Store, logger *slog.Logger, opts Options) *Server {
	if opts.OfflineAfter <= 0 {
		opts.OfflineAfter = 2 * time.Hour
	}
	s := &Server{
		svc:          svc,
		store:        st,
		logger:       logger,
		mux:          http.NewServeMux(),
		apiToken:     strings.TrimSpace(opts.APIToken),
		offlineAfter: opts.OfflineAfter,
		defaultLang:  i18n.Parse(opts.DefaultLang),
	}
	s.routes()
	return s
}

// Handler returns the root handler.
func (s *Server) Handler() http.Handler { return s.mux }

func (s *Server) routes() {
	s.mux.HandleFunc("GET /api/health", s.health)
	s.mux.HandleFunc("GET /api/machines", s.listMachines)
	s.mux.HandleFunc("GET /api/inventory", s.listMachines)
	s.mux.HandleFunc("POST /api/machines", s.createOrUpdate)
	s.mux.HandleFunc("PATCH /api/machines/{id}", s.patchMachine)
	s.mux.HandleFunc("DELETE /api/machines/{id}", s.deleteMachine)
	s.mux.HandleFunc("GET /api/stats/os", s.statsOS)
	s.mux.HandleFunc("GET /api/stats/versions", s.statsVersions)
}

func (s *Server) cat() i18n.Catalog {
	return i18n.New(s.defaultLang)
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(v)
}

func (s *Server) writeError(w http.ResponseWriter, err error) {
	if se, ok := err.(*service.Error); ok {
		if se.StatusCode >= 500 {
			s.logger.Error("service error", "err", se.Message)
			s.writeJSON(w, se.StatusCode, map[string]string{"error": s.cat().T("err.internal")})
			return
		}
		s.writeJSON(w, se.StatusCode, map[string]string{"error": se.Message})
		return
	}
	s.logger.Error("unexpected error", "err", err)
	s.writeJSON(w, 500, map[string]string{"error": s.cat().T("err.internal")})
}
