package server

import (
	"encoding/json"
	"net/http"

	"thingies/internal/things"
)

// headingUpdateRequest represents the body for PATCH /headings/:uuid
type headingUpdateRequest struct {
	Title string `json:"title"`
}

// handleDeleteHeading handles DELETE /headings/{uuid}
func (s *Server) handleDeleteHeading(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if uuid == "" {
		s.writeError(w, http.StatusBadRequest, "missing uuid")
		return
	}

	if err := things.DeleteHeading(uuid); err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]string{
		"status": "deleted",
		"uuid":   uuid,
	})
}

// handleUpdateHeading handles PATCH /headings/{uuid}
func (s *Server) handleUpdateHeading(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if uuid == "" {
		s.writeError(w, http.StatusBadRequest, "missing uuid")
		return
	}

	var req headingUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Title == "" {
		s.writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	if err := things.RenameHeading(uuid, req.Title); err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]string{
		"status": "updated",
		"uuid":   uuid,
		"title":  req.Title,
	})
}

// writeJSON writes a JSON response
func (s *Server) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes an error JSON response
func (s *Server) writeError(w http.ResponseWriter, status int, message string) {
	s.writeJSON(w, status, map[string]string{
		"error": message,
	})
}
