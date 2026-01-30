package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

	// Projects
	mux.HandleFunc("GET /projects", s.handleListProjects)
	mux.HandleFunc("GET /projects/{uuid}", s.handleGetProject)
	mux.HandleFunc("GET /projects/{uuid}/tasks", s.handleGetProjectTasks)
	mux.HandleFunc("GET /projects/{uuid}/headings", s.handleGetProjectHeadings)
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

// handleListProjects returns all projects
func (s *Server) handleListProjects(w http.ResponseWriter, r *http.Request) {
	includeCompleted := r.URL.Query().Get("include-completed") == "true"

	projects, err := s.db.ListProjects(includeCompleted)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert to JSON-serializable form
	result := make([]interface{}, len(projects))
	for i, p := range projects {
		result[i] = p.ToJSON()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleGetProject returns a single project
func (s *Server) handleGetProject(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if uuid == "" {
		http.Error(w, "uuid required", http.StatusBadRequest)
		return
	}

	project, err := s.db.GetProject(uuid)
	if err != nil {
		if err.Error() == fmt.Sprintf("project not found: %s", uuid) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project.ToJSON())
}

// handleGetProjectTasks returns tasks in a project
func (s *Server) handleGetProjectTasks(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if uuid == "" {
		http.Error(w, "uuid required", http.StatusBadRequest)
		return
	}

	includeCompleted := r.URL.Query().Get("include-completed") == "true"

	tasks, err := s.db.GetProjectTasks(uuid, includeCompleted)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert to JSON-serializable form
	result := make([]interface{}, len(tasks))
	for i, t := range tasks {
		result[i] = t.ToJSON()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleGetProjectHeadings returns headings in a project
func (s *Server) handleGetProjectHeadings(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if uuid == "" {
		http.Error(w, "uuid required", http.StatusBadRequest)
		return
	}

	headings, err := s.db.GetProjectHeadings(uuid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(headings)
}

// Addr returns the server address
func (s *Server) Addr() string {
	return s.httpServer.Addr
}
