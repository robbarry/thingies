package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"thingies/internal/db"
	"thingies/internal/things"
)

// Config holds server configuration
type Config struct {
	Host string
	Port int
}

// Server wraps an HTTP server with Things DB access
type Server struct {
	config     Config
	httpServer *http.Server
	db         *db.ThingsDB
}

// New creates a new server instance
func New(cfg Config, thingsDB *db.ThingsDB) *Server {
	s := &Server{
		config: cfg,
		db:     thingsDB,
	}

	mux := http.NewServeMux()
	s.registerRoutes(mux)

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      s.withMiddleware(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s
}

// registerRoutes sets up the HTTP routes
func (s *Server) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", s.handleHealth)

	// Project write endpoints
	mux.HandleFunc("POST /projects", s.handleCreateProject)
	mux.HandleFunc("PATCH /projects/{uuid}", s.handleUpdateProject)
	mux.HandleFunc("POST /projects/{uuid}/complete", s.handleCompleteProject)
	mux.HandleFunc("POST /projects/{uuid}/cancel", s.handleCancelProject)
	mux.HandleFunc("DELETE /projects/{uuid}", s.handleDeleteProject)
}

// withMiddleware wraps the handler with middleware
func (s *Server) withMiddleware(handler http.Handler) http.Handler {
	// Apply middleware in reverse order (last applied runs first)
	handler = s.corsMiddleware(handler)
	handler = s.loggingMiddleware(handler)
	return handler
}

// loggingMiddleware logs incoming requests
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		log.Printf("%s %s %d %s", r.Method, r.URL.Path, wrapped.statusCode, time.Since(start))
	})
}

// corsMiddleware adds CORS headers for local development
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// responseWriter wraps http.ResponseWriter to capture the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// handleHealth handles the health check endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]string{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(response)
}

// Start starts the HTTP server
func (s *Server) Start() error {
	log.Printf("Starting server on %s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down server...")
	return s.httpServer.Shutdown(ctx)
}

// DB returns the database connection
func (s *Server) DB() *db.ThingsDB {
	return s.db
}

// Addr returns the server address
func (s *Server) Addr() string {
	return s.httpServer.Addr
}

// createProjectRequest represents the request body for creating a project
type createProjectRequest struct {
	Title    string   `json:"title"`
	Notes    string   `json:"notes,omitempty"`
	When     string   `json:"when,omitempty"`
	Deadline string   `json:"deadline,omitempty"`
	Tags     string   `json:"tags,omitempty"`
	Area     string   `json:"area,omitempty"`
	Todos    []string `json:"todos,omitempty"`
	Headings []struct {
		Title string   `json:"title"`
		Todos []string `json:"todos,omitempty"`
	} `json:"headings,omitempty"`
}

// updateProjectRequest represents the request body for updating a project
type updateProjectRequest struct {
	Title    string `json:"title,omitempty"`
	Notes    string `json:"notes,omitempty"`
	Deadline string `json:"deadline,omitempty"`
	Tags     string `json:"tags,omitempty"`
}

// handleCreateProject creates a new project via Things URL scheme
func (s *Server) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	var req createProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		s.jsonError(w, "title is required", http.StatusBadRequest)
		return
	}

	// Build todos list including headings
	var allTodos []string
	allTodos = append(allTodos, req.Todos...)

	// Headings in Things URL scheme are prefixed with "# "
	for _, heading := range req.Headings {
		allTodos = append(allTodos, "# "+heading.Title)
		allTodos = append(allTodos, heading.Todos...)
	}

	params := things.AddProjectParams{
		Title:    req.Title,
		Notes:    req.Notes,
		When:     req.When,
		Deadline: req.Deadline,
		Tags:     req.Tags,
		Area:     req.Area,
		ToDos:    allTodos,
	}

	url := things.BuildAddProjectURL(params)
	if err := things.OpenURL(url); err != nil {
		s.jsonError(w, "failed to create project: "+err.Error(), http.StatusInternalServerError)
		return
	}

	s.jsonResponse(w, map[string]string{
		"status":  "ok",
		"message": "project creation triggered",
	}, http.StatusAccepted)
}

// handleUpdateProject updates a project via AppleScript
func (s *Server) handleUpdateProject(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if uuid == "" {
		s.jsonError(w, "uuid is required", http.StatusBadRequest)
		return
	}

	var req updateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Check if at least one field is provided
	if req.Title == "" && req.Notes == "" && req.Deadline == "" && req.Tags == "" {
		s.jsonError(w, "at least one field (title, notes, deadline, tags) is required", http.StatusBadRequest)
		return
	}

	params := things.ProjectUpdateParams{
		UUID:     uuid,
		Name:     req.Title,
		Notes:    req.Notes,
		DueDate:  req.Deadline,
		TagNames: req.Tags,
	}

	if err := things.UpdateProject(params); err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "Can't get project") {
			s.jsonError(w, "project not found", http.StatusNotFound)
			return
		}
		s.jsonError(w, "failed to update project: "+errStr, http.StatusInternalServerError)
		return
	}

	s.jsonResponse(w, map[string]string{
		"status":  "ok",
		"message": "project updated",
		"uuid":    uuid,
	}, http.StatusOK)
}

// handleCompleteProject marks a project as complete
func (s *Server) handleCompleteProject(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if uuid == "" {
		s.jsonError(w, "uuid is required", http.StatusBadRequest)
		return
	}

	if err := things.CompleteProject(uuid); err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "Can't get project") {
			s.jsonError(w, "project not found", http.StatusNotFound)
			return
		}
		s.jsonError(w, "failed to complete project: "+errStr, http.StatusInternalServerError)
		return
	}

	s.jsonResponse(w, map[string]string{
		"status":  "ok",
		"message": "project marked complete",
		"uuid":    uuid,
	}, http.StatusOK)
}

// handleCancelProject marks a project as canceled
func (s *Server) handleCancelProject(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if uuid == "" {
		s.jsonError(w, "uuid is required", http.StatusBadRequest)
		return
	}

	if err := things.CancelProject(uuid); err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "Can't get project") {
			s.jsonError(w, "project not found", http.StatusNotFound)
			return
		}
		s.jsonError(w, "failed to cancel project: "+errStr, http.StatusInternalServerError)
		return
	}

	s.jsonResponse(w, map[string]string{
		"status":  "ok",
		"message": "project marked canceled",
		"uuid":    uuid,
	}, http.StatusOK)
}

// handleDeleteProject deletes a project (moves to trash)
func (s *Server) handleDeleteProject(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if uuid == "" {
		s.jsonError(w, "uuid is required", http.StatusBadRequest)
		return
	}

	if err := things.DeleteProject(uuid); err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "Can't get project") {
			s.jsonError(w, "project not found", http.StatusNotFound)
			return
		}
		s.jsonError(w, "failed to delete project: "+errStr, http.StatusInternalServerError)
		return
	}

	s.jsonResponse(w, map[string]string{
		"status":  "ok",
		"message": "project moved to trash",
		"uuid":    uuid,
	}, http.StatusOK)
}

// jsonResponse writes a JSON response with the given status code
func (s *Server) jsonResponse(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// jsonError writes a JSON error response
func (s *Server) jsonError(w http.ResponseWriter, message string, status int) {
	s.jsonResponse(w, map[string]string{"error": message}, status)
}
