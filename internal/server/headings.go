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

	w.WriteHeader(http.StatusOK)
	s.writeJSON(w, map[string]string{
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

	w.WriteHeader(http.StatusOK)
	s.writeJSON(w, map[string]string{
		"status": "updated",
		"uuid":   uuid,
		"title":  req.Title,
	})
}
