package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"thingies/internal/db"
	"thingies/internal/models"
)

// handleListTasks handles GET /tasks
func (s *Server) handleListTasks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	filter := db.TaskFilter{
		Status:        query.Get("status"),
		Area:          query.Get("area"),
		Project:       query.Get("project"),
		Tag:           query.Get("tag"),
		Today:         query.Get("today") == "true",
		IncludeFuture: query.Get("include-future") == "true",
	}

	tasks, err := s.db.ListTasks(filter)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeTasksJSON(w, tasks)
}

// handleGetTask handles GET /tasks/{uuid}
func (s *Server) handleGetTask(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if uuid == "" {
		s.writeError(w, http.StatusBadRequest, "missing task uuid")
		return
	}

	task, err := s.db.GetTask(uuid)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			s.writeError(w, http.StatusNotFound, err.Error())
			return
		}
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, task.ToJSON())
}

// handleSearchTasks handles GET /tasks/search
func (s *Server) handleSearchTasks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	q := query.Get("q")
	if q == "" {
		s.writeError(w, http.StatusBadRequest, "missing required query parameter: q")
		return
	}

	includeNotes := query.Get("in-notes") == "true"
	includeFuture := query.Get("include-future") == "true"

	tasks, err := s.db.Search(q, includeNotes, includeFuture)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeTasksJSON(w, tasks)
}

// writeTasksJSON converts tasks to JSON and writes the response
func (s *Server) writeTasksJSON(w http.ResponseWriter, tasks []models.Task) {
	result := make([]models.TaskJSON, 0, len(tasks))
	for _, t := range tasks {
		result = append(result, t.ToJSON())
	}
	s.writeJSON(w, result)
}

// writeJSON writes a JSON response
func (s *Server) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// writeError writes a JSON error response
func (s *Server) writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
