package httpui

import (
	"io/fs"
	"log/slog"
	"net/http"
	"time"

	"github.com/carlosrabelo/tatuscan/web/web/internal/apiclient"
	"github.com/carlosrabelo/tatuscan/web/web/internal/i18n"
)

// Server is the web UI.
type Server struct {
	api          *apiclient.Client
	logger       *slog.Logger
	tmplFS       fs.FS
	mux          *http.ServeMux
	offlineAfter time.Duration
	defaultLang  string
}

// Options configures optional UI server behavior.
type Options struct {
	OfflineAfter time.Duration
	DefaultLang  string
}

// New creates the UI server with embedded templates.
func New(api *apiclient.Client, logger *slog.Logger, opts Options) *Server {
	return NewWithFS(api, templateFS, logger, opts)
}

// NewWithFS creates the UI server with a custom template FS (tests).
func NewWithFS(api *apiclient.Client, tmplFS fs.FS, logger *slog.Logger, opts Options) *Server {
	if opts.OfflineAfter <= 0 {
		opts.OfflineAfter = 2 * time.Hour
	}
	s := &Server{
		api:          api,
		logger:       logger,
		tmplFS:       tmplFS,
		mux:          http.NewServeMux(),
		offlineAfter: opts.OfflineAfter,
		defaultLang:  i18n.Parse(opts.DefaultLang),
	}
	s.routes()
	return s
}

func (s *Server) Handler() http.Handler { return s.mux }

func (s *Server) routes() {
	s.mux.HandleFunc("GET /healthz", s.healthz)
}

func (s *Server) healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
