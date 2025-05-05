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
