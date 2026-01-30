package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"thingies/internal/db"
	"thingies/internal/models"
)

// tasksToJSON converts a slice of tasks to their JSON representation
func tasksToJSON(tasks []models.Task) []models.TaskJSON {
	result := make([]models.TaskJSON, len(tasks))
	for i, t := range tasks {
		result[i] = t.ToJSON()
	}
	return result
}

// writeJSONTasks writes tasks as JSON to the response
func writeJSONTasks(w http.ResponseWriter, tasks []models.Task) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasksToJSON(tasks))
}

// writeJSONError writes an error response
func writeJSONError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// handleToday returns today's tasks
func (s *Server) handleToday(w http.ResponseWriter, r *http.Request) {
	tasks, err := s.db.ListTasks(db.TaskFilter{Today: true})
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSONTasks(w, tasks)
}

// handleInbox returns inbox tasks
func (s *Server) handleInbox(w http.ResponseWriter, r *http.Request) {
	tasks, err := s.db.GetInboxTasks()
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSONTasks(w, tasks)
}

// handleAnytime returns anytime tasks
func (s *Server) handleAnytime(w http.ResponseWriter, r *http.Request) {
	tasks, err := s.db.GetAnytimeTasks()
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSONTasks(w, tasks)
}

// handleUpcoming returns upcoming scheduled tasks
func (s *Server) handleUpcoming(w http.ResponseWriter, r *http.Request) {
	tasks, err := s.db.GetUpcomingTasks()
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSONTasks(w, tasks)
}

// handleSomeday returns someday tasks
func (s *Server) handleSomeday(w http.ResponseWriter, r *http.Request) {
	tasks, err := s.db.GetSomedayTasks()
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSONTasks(w, tasks)
}

// handleLogbook returns completed tasks
func (s *Server) handleLogbook(w http.ResponseWriter, r *http.Request) {
	limit := 50 // default
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	tasks, err := s.db.GetLogbook(limit)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSONTasks(w, tasks)
}

// handleDeadlines returns tasks with upcoming deadlines
func (s *Server) handleDeadlines(w http.ResponseWriter, r *http.Request) {
	days := 7 // default
	if daysStr := r.URL.Query().Get("days"); daysStr != "" {
		if parsed, err := strconv.Atoi(daysStr); err == nil && parsed > 0 {
			days = parsed
		}
	}

	tasks, err := s.db.GetDeadlines(days)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSONTasks(w, tasks)
}
