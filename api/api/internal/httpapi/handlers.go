package httpapi

import (
	"encoding/json"
	"errors"
	"io"
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
