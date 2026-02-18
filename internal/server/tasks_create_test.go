package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestCreateTaskRejectsUnknownFields verifies that POST /tasks returns 400
// when the request body contains fields not in TaskCreateRequest.
// Currently FAILS because json.Decoder silently ignores unknown fields.
func TestCreateTaskRejectsUnknownFields(t *testing.T) {
	s := &Server{}

	tests := []struct {
		name string
		body string
	}{
		{
			name: "unknown field 'scheduled'",
			body: `{"title": "test task", "scheduled": "today"}`,
		},
		{
			name: "misspelled 'titel'",
			body: `{"titel": "test task"}`,
		},
		{
			name: "misspelled 'tag' instead of 'tags'",
			body: `{"title": "test task", "tag": "urgent"}`,
		},
		{
			name: "unknown field 'priority'",
			body: `{"title": "test task", "priority": "high"}`,
		},
		{
			name: "multiple unknown fields",
			body: `{"title": "test task", "foo": "bar", "baz": 123}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			s.handleCreateTask(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("expected status 400, got %d; body: %s", w.Code, w.Body.String())
			}

			// Response should mention the unknown field
			body := w.Body.String()
			if !strings.Contains(body, "unknown") && !strings.Contains(body, "unrecognized") {
				t.Errorf("expected error message about unknown/unrecognized field, got: %s", body)
			}
		})
	}
}

// TestCreateTaskAcceptsValidFields verifies that POST /tasks does NOT reject
// a request that only contains known fields. We don't need Things to actually
// create the task — we just verify the handler doesn't return 400 for valid input.
func TestCreateTaskAcceptsValidFields(t *testing.T) {
	s := &Server{}

	tests := []struct {
		name string
		body string
	}{
		{
			name: "title only",
			body: `{"title": "test task"}`,
		},
		{
			name: "all valid fields",
			body: `{"title": "test", "notes": "some notes", "when": "today", "deadline": "2026-03-01", "tags": "urgent", "list": "Work", "heading": "Section"}`,
		},
		{
			name: "subset of optional fields",
			body: `{"title": "test", "when": "tomorrow", "tags": "quick"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			s.handleCreateTask(w, req)

			// Should NOT be 400 — any other status means the body was accepted.
			// It may fail later (e.g., 500 because Things isn't running), but
			// the point is it shouldn't be rejected as invalid input.
			if w.Code == http.StatusBadRequest {
				t.Errorf("valid fields should not be rejected, got 400; body: %s", w.Body.String())
			}
		})
	}
}
