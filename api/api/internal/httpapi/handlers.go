package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/carlosrabelo/tatuscan/api/api/internal/i18n"
)

func (s *Server) root(w http.ResponseWriter, r *http.Request) {
	cat := s.cat()
	info := map[string]any{
		"service": "tatuscan-api",
		"panel":   "http://127.0.0.1:8050/",
		"lang":    cat.Locale(),
		"endpoints": []string{
			"GET /api/health",
			"GET /api/machines",
			"GET /api/inventory",
			"POST /api/machines",
			"PATCH /api/machines/{id}",
			"DELETE /api/machines/{id}",
			"GET /api/stats/os",
			"GET /api/stats/versions",
			"GET /api/stats/age",
			"GET /api/stats/online",
		},
	}
	if strings.Contains(r.Header.Get("Accept"), "text/html") {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(w, rootPage(cat))
		return
	}
	s.writeJSON(w, http.StatusOK, info)
}

func rootPage(cat i18n.Catalog) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="%s">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>%s</title>
  <style>
    body { margin:0; padding:32px; font-family: system-ui, sans-serif; background:#f5f7fa; color:#2c3e50; }
    .card { max-width:640px; background:#fff; border-radius:12px; padding:24px; box-shadow:0 6px 18px rgba(0,0,0,.06); }
    a { color:#2563eb; }
    code { background:#f1f5f9; padding:2px 6px; border-radius:6px; }
    ul { line-height:1.8; }
  </style>
</head>
<body>
  <div class="card">
    <h1>%s</h1>
    <p>%s</p>
    <p><a href="http://127.0.0.1:8050/">%s</a></p>
    <h2>%s</h2>
    <ul>
      <li><a href="/api/health"><code>GET /api/health</code></a></li>
      <li><a href="/api/machines"><code>GET /api/machines</code></a></li>
      <li><a href="/api/inventory"><code>GET /api/inventory</code></a></li>
      <li><a href="/api/stats/os"><code>GET /api/stats/os</code></a></li>
      <li><a href="/api/stats/versions"><code>GET /api/stats/versions</code></a></li>
      <li><a href="/api/stats/age"><code>GET /api/stats/age</code></a></li>
      <li><a href="/api/stats/online"><code>GET /api/stats/online</code></a></li>
    </ul>
  </div>
</body>
</html>
`, cat.HTMLLang(), cat.T("root.title"), cat.T("root.title"), cat.T("root.lead"), cat.T("root.panel"), cat.T("root.endpoints"))
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	if err := s.store.DB().PingContext(r.Context()); err != nil {
		s.logger.Error("health check failed", "err", err)
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"status": "unhealthy"})
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}

func (s *Server) listMachines(w http.ResponseWriter, r *http.Request) {
	items, err := s.svc.ListAll(r.Context(), "hostname", "asc")
	if err != nil {
		s.writeError(w, err)
		return
	}
	out := make([]map[string]any, 0, len(items))
	for _, inv := range items {
		out = append(out, s.svc.SerializeInventory(inv))
	}
	s.writeJSON(w, http.StatusOK, map[string]any{"items": out, "count": len(out)})
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request) (map[string]any, error) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MiB
	dec := json.NewDecoder(r.Body)
	var data map[string]any
	if err := dec.Decode(&data); err != nil {
		if errors.Is(err, io.EOF) {
			return map[string]any{}, nil
		}
		return nil, err
	}
	if data == nil {
		data = map[string]any{}
	}
	return data, nil
}

func (s *Server) createOrUpdate(w http.ResponseWriter, r *http.Request) {
	data, err := decodeJSONBody(w, r)
	if err != nil {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": s.cat().T("err.invalid_json")})
		return
	}
	s.logger.Debug("received inventory", "data", data)

	inv, created, err := s.svc.CreateOrUpdate(r.Context(), data)
	if err != nil {
		s.logger.Error("service error", "err", err)
		s.writeError(w, err)
		return
	}
	msg := s.cat().T("msg.updated")
	status := http.StatusOK
	if created {
		msg = s.cat().T("msg.created")
		status = http.StatusCreated
	}
	s.logger.Info("inventory upsert", "created", created, "hostname", inv.Hostname)
	s.writeJSON(w, status, map[string]any{
		"message": msg,
		"item":    s.svc.SerializeInventory(inv),
	})
}

func (s *Server) patchMachine(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	data, err := decodeJSONBody(w, r)
	if err != nil {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": s.cat().T("err.invalid_json")})
		return
	}
	inv, err := s.svc.PartialUpdate(r.Context(), id, data)
	if err != nil {
		s.logger.Error("service error", "err", err)
		s.writeError(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]any{
		"message": s.cat().T("msg.patched"),
		"item":    s.svc.SerializeInventory(inv),
	})
}

func (s *Server) deleteMachine(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.svc.Delete(r.Context(), id); err != nil {
		s.logger.Error("service error", "err", err)
		s.writeError(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"message": s.cat().T("msg.deleted")})
}

func (s *Server) statsOS(w http.ResponseWriter, r *http.Request) {
	items, err := s.svc.OSDistribution(r.Context())
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) statsVersions(w http.ResponseWriter, r *http.Request) {
	topN := 8
	if v := r.URL.Query().Get("top"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			topN = n
		}
	}
	if topN > 100 {
		topN = 100
	}
	items, err := s.svc.VersionDistribution(r.Context(), topN)
	if err != nil {
		s.writeError(w, err)
		return
	}
	other := s.cat().T("stats.other")
	for i := range items {
		if items[i].Version == "Other" || items[i].Version == "Outros" {
			items[i].Version = other
		}
	}
	s.writeJSON(w, http.StatusOK, map[string]any{"items": items, "top": topN})
}

func (s *Server) statsAge(w http.ResponseWriter, r *http.Request) {
	items, err := s.svc.AgeDistribution(r.Context())
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) statsOnline(w http.ResponseWriter, r *http.Request) {
	after := s.offlineAfter
	if v := r.URL.Query().Get("after"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil || d <= 0 {
			s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": s.cat().T("err.invalid_after")})
			return
		}
		after = d
	}
	stats, err := s.svc.OnlineDistribution(r.Context(), after)
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, stats)
}
