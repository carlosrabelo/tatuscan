package httpapi

import (
	"net/http"
)

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
