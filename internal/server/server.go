package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"thingies/internal/db"
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

	// Area endpoints
	mux.HandleFunc("GET /areas", s.handleListAreas)
	mux.HandleFunc("GET /areas/{uuid}", s.handleGetArea)
	mux.HandleFunc("GET /areas/{uuid}/tasks", s.handleGetAreaTasks)
	mux.HandleFunc("GET /areas/{uuid}/projects", s.handleGetAreaProjects)

	// Tag endpoints
	mux.HandleFunc("GET /tags", s.handleListTags)
	mux.HandleFunc("GET /tags/{name}/tasks", s.handleGetTagTasks)
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

// handleListAreas returns all visible areas
func (s *Server) handleListAreas(w http.ResponseWriter, r *http.Request) {
	areas, err := s.db.ListAreas()
	if err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.jsonResponse(w, areas)
}

// handleGetArea returns a single area by UUID
func (s *Server) handleGetArea(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if uuid == "" {
		s.jsonError(w, "uuid is required", http.StatusBadRequest)
		return
	}

	area, err := s.db.GetArea(uuid)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			s.jsonError(w, err.Error(), http.StatusNotFound)
			return
		}
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.jsonResponse(w, area)
}

// handleGetAreaTasks returns loose tasks in an area (not in projects)
func (s *Server) handleGetAreaTasks(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if uuid == "" {
		s.jsonError(w, "uuid is required", http.StatusBadRequest)
		return
	}

	// Check if area exists first
	_, err := s.db.GetArea(uuid)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			s.jsonError(w, err.Error(), http.StatusNotFound)
			return
		}
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	includeCompleted := r.URL.Query().Get("include_completed") == "true"
	tasks, err := s.db.GetAreaTasks(uuid, includeCompleted)
	if err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.jsonResponse(w, tasks)
}

// handleGetAreaProjects returns projects in an area
func (s *Server) handleGetAreaProjects(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if uuid == "" {
		s.jsonError(w, "uuid is required", http.StatusBadRequest)
		return
	}

	// Check if area exists first
	_, err := s.db.GetArea(uuid)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			s.jsonError(w, err.Error(), http.StatusNotFound)
			return
		}
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	includeCompleted := r.URL.Query().Get("include_completed") == "true"
	projects, err := s.db.GetAreaProjects(uuid, includeCompleted)
	if err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.jsonResponse(w, projects)
}

// handleListTags returns all tags with usage counts
func (s *Server) handleListTags(w http.ResponseWriter, r *http.Request) {
	tags, err := s.db.ListTags()
	if err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert to JSON-serializable form
	result := make([]interface{}, len(tags))
	for i, tag := range tags {
		result[i] = tag.ToJSON()
	}

	s.jsonResponse(w, result)
}

// handleGetTagTasks returns tasks with a specific tag
func (s *Server) handleGetTagTasks(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		s.jsonError(w, "tag name is required", http.StatusBadRequest)
		return
	}

	// URL-decode the tag name to handle spaces and special characters
	decodedName, err := url.PathUnescape(name)
	if err != nil {
		s.jsonError(w, "invalid tag name encoding", http.StatusBadRequest)
		return
	}

	tasks, err := s.db.GetTasksByTag(decodedName)
	if err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.jsonResponse(w, tasks)
}

// jsonResponse writes a JSON response
func (s *Server) jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// jsonError writes a JSON error response
func (s *Server) jsonError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
