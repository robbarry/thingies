package server

import (
	"encoding/json"
	"net/http"

	"thingies/internal/things"
)

// TaskCreateRequest is the request body for creating a task
type TaskCreateRequest struct {
	Title    string `json:"title"`
	Notes    string `json:"notes,omitempty"`
	When     string `json:"when,omitempty"`
	Deadline string `json:"deadline,omitempty"`
	Tags     string `json:"tags,omitempty"`
	List     string `json:"list,omitempty"`
	Heading  string `json:"heading,omitempty"`
}

// TaskUpdateRequest is the request body for updating a task
type TaskUpdateRequest struct {
	Title    string `json:"title,omitempty"`
	Notes    string `json:"notes,omitempty"`
	When     string `json:"when,omitempty"`
	Deadline string `json:"deadline,omitempty"`
	Tags     string `json:"tags,omitempty"`
}

// ProjectCreateRequest is the request body for creating a project
type ProjectCreateRequest struct {
	Title    string   `json:"title"`
	Notes    string   `json:"notes,omitempty"`
	When     string   `json:"when,omitempty"`
	Deadline string   `json:"deadline,omitempty"`
	Tags     string   `json:"tags,omitempty"`
	Area     string   `json:"area,omitempty"`
	ToDos    []string `json:"todos,omitempty"`
}

// APIResponse is a standard API response
type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes a JSON error response
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, APIResponse{Success: false, Message: message})
}

// writeSuccess writes a JSON success response
func writeSuccess(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusOK, APIResponse{Success: true, Message: message})
}

// handleCreateTask handles POST /tasks
func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var req TaskCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	params := things.AddParams{
		Title:    req.Title,
		Notes:    req.Notes,
		When:     req.When,
		Deadline: req.Deadline,
		Tags:     req.Tags,
		List:     req.List,
		Heading:  req.Heading,
	}

	url := things.BuildAddURL(params)
	if err := things.OpenURL(url); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create task: "+err.Error())
		return
	}

	writeSuccess(w, "task created")
}

// handleUpdateTask handles PATCH /tasks/{uuid}
func (s *Server) handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if uuid == "" {
		writeError(w, http.StatusBadRequest, "task UUID is required")
		return
	}

	var req TaskUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	params := things.TaskUpdateParams{
		UUID:     uuid,
		Name:     req.Title,
		Notes:    req.Notes,
		When:     req.When,
		DueDate:  req.Deadline,
		TagNames: req.Tags,
	}

	// Specific dates need an auth token for the URL scheme
	if things.IsSpecificDate(req.When) {
		token, err := s.db.GetAuthToken()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to get auth token: "+err.Error())
			return
		}
		params.AuthToken = token
	}

	if err := things.UpdateTask(params); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update task: "+err.Error())
		return
	}

	writeSuccess(w, "task updated")
}

// handleCompleteTask handles POST /tasks/{uuid}/complete
func (s *Server) handleCompleteTask(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if uuid == "" {
		writeError(w, http.StatusBadRequest, "task UUID is required")
		return
	}

	if err := things.CompleteTask(uuid); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to complete task: "+err.Error())
		return
	}

	writeSuccess(w, "task completed")
}

// handleCancelTask handles POST /tasks/{uuid}/cancel
func (s *Server) handleCancelTask(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if uuid == "" {
		writeError(w, http.StatusBadRequest, "task UUID is required")
		return
	}

	if err := things.CancelTask(uuid); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to cancel task: "+err.Error())
		return
	}

	writeSuccess(w, "task canceled")
}

// handleDeleteTask handles DELETE /tasks/{uuid}
func (s *Server) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if uuid == "" {
		writeError(w, http.StatusBadRequest, "task UUID is required")
		return
	}

	if err := things.DeleteTask(uuid); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete task: "+err.Error())
		return
	}

	writeSuccess(w, "task deleted")
}

// handleMoveTaskToToday handles POST /tasks/{uuid}/move-to-today
func (s *Server) handleMoveTaskToToday(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if uuid == "" {
		writeError(w, http.StatusBadRequest, "task UUID is required")
		return
	}

	if err := things.MoveTaskToToday(uuid); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to move task to today: "+err.Error())
		return
	}

	writeSuccess(w, "task moved to today")
}

// handleMoveTaskToSomeday handles POST /tasks/{uuid}/move-to-someday
func (s *Server) handleMoveTaskToSomeday(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if uuid == "" {
		writeError(w, http.StatusBadRequest, "task UUID is required")
		return
	}

	// Use UpdateTask with When="someday" which moves to the Someday list
	params := things.TaskUpdateParams{
		UUID: uuid,
		When: "someday",
	}

	if err := things.UpdateTask(params); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to move task to someday: "+err.Error())
		return
	}

	writeSuccess(w, "task moved to someday")
}

// handleCreateProject handles POST /projects
func (s *Server) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	var req ProjectCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	params := things.AddProjectParams{
		Title:    req.Title,
		Notes:    req.Notes,
		When:     req.When,
		Deadline: req.Deadline,
		Tags:     req.Tags,
		Area:     req.Area,
		ToDos:    req.ToDos,
	}

	url := things.BuildAddProjectURL(params)
	if err := things.OpenURL(url); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create project: "+err.Error())
		return
	}

	writeSuccess(w, "project created")
}
