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
	"thingies/internal/models"
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

	// Heading routes
	mux.HandleFunc("DELETE /headings/{uuid}", s.handleDeleteHeading)
	mux.HandleFunc("PATCH /headings/{uuid}", s.handleUpdateHeading)
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
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
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

// handleSnapshot returns a hierarchical text snapshot of all Things data
func (s *Server) handleSnapshot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	snapshot, err := s.buildSnapshot()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"snapshot": snapshot})
}

// snapshotArea holds area data with projects and tasks for snapshot building
type snapshotArea struct {
	models.Area
	Projects []snapshotProject
	Tasks    []models.TaskJSON
}

// snapshotProject holds project data with tasks for snapshot building
type snapshotProject struct {
	models.ProjectJSON
	Tasks []models.TaskJSON
}

// buildSnapshot creates a hierarchical text representation of all Things data
func (s *Server) buildSnapshot() (string, error) {
	var sb strings.Builder

	// Track tasks we've already output to avoid duplicates
	seenTasks := make(map[string]bool)

	// Get today tasks
	todayTasks, err := s.db.ListTasks(db.TaskFilter{Status: "incomplete", Today: true})
	if err != nil {
		return "", err
	}

	// Get inbox tasks
	inboxTasks, err := s.db.GetInboxTasks()
	if err != nil {
		return "", err
	}

	// Get upcoming tasks
	upcomingTasks, err := s.db.GetUpcomingTasks()
	if err != nil {
		return "", err
	}

	// Get someday tasks
	somedayTasks, err := s.db.GetSomedayTasks()
	if err != nil {
		return "", err
	}

	// Get areas with projects and tasks
	areas, err := s.db.ListAreas()
	if err != nil {
		return "", err
	}

	var snapshotAreas []snapshotArea
	for _, area := range areas {
		sa := snapshotArea{Area: area}

		projects, err := s.db.GetAreaProjects(area.UUID, false)
		if err != nil {
			return "", err
		}

		for _, proj := range projects {
			sp := snapshotProject{ProjectJSON: proj.ToJSON()}

			tasks, err := s.db.GetProjectTasks(proj.UUID, false)
			if err != nil {
				return "", err
			}

			for _, t := range tasks {
				sp.Tasks = append(sp.Tasks, t.ToJSON())
			}
			sa.Projects = append(sa.Projects, sp)
		}

		// Get tasks directly under area
		areaTasks, err := s.db.GetAreaTasks(area.UUID, false)
		if err != nil {
			return "", err
		}
		for _, t := range areaTasks {
			sa.Tasks = append(sa.Tasks, t.ToJSON())
		}

		snapshotAreas = append(snapshotAreas, sa)
	}

	// Build text output

	// TODAY section
	if len(todayTasks) > 0 {
		sb.WriteString("# TODAY\n")
		for _, t := range todayTasks {
			seenTasks[t.UUID] = true
			writeTaskLine(&sb, t.ToJSON(), 0)
		}
		sb.WriteString("\n")
	}

	// ANYTIME section - tasks organized by area/project/heading (not in today/upcoming/someday/inbox)
	anytimeOutput := buildAnytimeSection(snapshotAreas, seenTasks)
	if anytimeOutput != "" {
		sb.WriteString("# ANYTIME\n")
		sb.WriteString(anytimeOutput)
		sb.WriteString("\n")
	}

	// UPCOMING section
	if len(upcomingTasks) > 0 {
		sb.WriteString("# UPCOMING\n")
		for _, t := range upcomingTasks {
			if !seenTasks[t.UUID] {
				seenTasks[t.UUID] = true
				writeTaskLine(&sb, t.ToJSON(), 0)
			}
		}
		sb.WriteString("\n")
	}

	// SOMEDAY section
	if len(somedayTasks) > 0 {
		sb.WriteString("# SOMEDAY\n")
		for _, t := range somedayTasks {
			if !seenTasks[t.UUID] {
				seenTasks[t.UUID] = true
				writeTaskLine(&sb, t.ToJSON(), 0)
			}
		}
		sb.WriteString("\n")
	}

	// INBOX section
	if len(inboxTasks) > 0 {
		sb.WriteString("# INBOX\n")
		for _, t := range inboxTasks {
			if !seenTasks[t.UUID] {
				seenTasks[t.UUID] = true
				writeTaskLine(&sb, t.ToJSON(), 0)
			}
		}
		sb.WriteString("\n")
	}

	return strings.TrimSpace(sb.String()), nil
}

// buildAnytimeSection builds the ANYTIME section with area/project/heading hierarchy
func buildAnytimeSection(areas []snapshotArea, seenTasks map[string]bool) string {
	var sb strings.Builder

	for _, area := range areas {
		areaHasOutput := false
		var areaSb strings.Builder

		for _, proj := range area.Projects {
			projHasOutput := false
			var projSb strings.Builder

			// Group tasks by heading
			tasksByHeading := make(map[string][]models.TaskJSON)
			var headingOrder []string
			for _, t := range proj.Tasks {
				if seenTasks[t.UUID] {
					continue
				}
				heading := t.HeadingName
				if _, exists := tasksByHeading[heading]; !exists {
					headingOrder = append(headingOrder, heading)
				}
				tasksByHeading[heading] = append(tasksByHeading[heading], t)
			}

			// Output tasks by heading
			for _, heading := range headingOrder {
				tasks := tasksByHeading[heading]
				if len(tasks) == 0 {
					continue
				}

				if heading != "" {
					projSb.WriteString(fmt.Sprintf("      %s:\n", heading))
					for _, t := range tasks {
						seenTasks[t.UUID] = true
						writeTaskLine(&projSb, t, 4)
						projHasOutput = true
					}
				} else {
					for _, t := range tasks {
						seenTasks[t.UUID] = true
						writeTaskLine(&projSb, t, 2)
						projHasOutput = true
					}
				}
			}

			if projHasOutput {
				areaSb.WriteString(fmt.Sprintf("    %s:\n", proj.Title))
				areaSb.WriteString(projSb.String())
				areaHasOutput = true
			}
		}

		// Direct area tasks
		for _, t := range area.Tasks {
			if seenTasks[t.UUID] {
				continue
			}
			if !areaHasOutput {
				areaHasOutput = true
			}
			seenTasks[t.UUID] = true
			areaSb.WriteString(fmt.Sprintf("    - %s (ID: %s", t.Title, shortID(t.UUID)))
			if t.Due != "" {
				areaSb.WriteString(fmt.Sprintf(", deadline: %s", t.Due[:10]))
			}
			areaSb.WriteString(")\n")
		}

		if areaHasOutput {
			sb.WriteString(fmt.Sprintf("  %s:\n", area.Title))
			sb.WriteString(areaSb.String())
		}
	}

	return sb.String()
}

// writeTaskLine writes a formatted task line to the builder
func writeTaskLine(sb *strings.Builder, t models.TaskJSON, extraIndent int) {
	indent := strings.Repeat("  ", extraIndent)
	sb.WriteString(fmt.Sprintf("%s  - %s (ID: %s", indent, t.Title, shortID(t.UUID)))
	if t.Due != "" && len(t.Due) >= 10 {
		sb.WriteString(fmt.Sprintf(", deadline: %s", t.Due[:10]))
	}
	sb.WriteString(")\n")
}

// shortID returns first 8 characters of UUID
func shortID(uuid string) string {
	if len(uuid) > 8 {
		return uuid[:8]
	}
	return uuid
}
