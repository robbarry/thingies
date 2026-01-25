package things

import (
	"net/url"
	"strings"
)

// AddParams contains parameters for creating a task
type AddParams struct {
	Title     string
	Notes     string
	When      string
	Deadline  string
	Tags      string
	List      string
	Heading   string
	Completed bool
	Canceled  bool
	ChecklistItems []string
}

// BuildAddURL builds a things:///add URL
func BuildAddURL(params AddParams) string {
	u := url.URL{
		Scheme: "things",
		Host:   "",
		Path:   "/add",
	}

	q := u.Query()
	q.Set("title", params.Title)

	if params.Notes != "" {
		q.Set("notes", params.Notes)
	}
	if params.When != "" {
		q.Set("when", params.When)
	}
	if params.Deadline != "" {
		q.Set("deadline", params.Deadline)
	}
	if params.Tags != "" {
		q.Set("tags", params.Tags)
	}
	if params.List != "" {
		q.Set("list", params.List)
	}
	if params.Heading != "" {
		q.Set("heading", params.Heading)
	}
	if params.Completed {
		q.Set("completed", "true")
	}
	if params.Canceled {
		q.Set("canceled", "true")
	}
	if len(params.ChecklistItems) > 0 {
		q.Set("checklist-items", strings.Join(params.ChecklistItems, "\n"))
	}

	// Use %20 for spaces instead of + (Things doesn't decode + as space)
	u.RawQuery = strings.ReplaceAll(q.Encode(), "+", "%20")
	return u.String()
}

// AddProjectParams contains parameters for creating a project
type AddProjectParams struct {
	Title     string
	Notes     string
	When      string
	Deadline  string
	Tags      string
	Area      string
	ToDos     []string
	Completed bool
	Canceled  bool
}

// BuildAddProjectURL builds a things:///add-project URL
func BuildAddProjectURL(params AddProjectParams) string {
	u := url.URL{
		Scheme: "things",
		Host:   "",
		Path:   "/add-project",
	}

	q := u.Query()
	q.Set("title", params.Title)

	if params.Notes != "" {
		q.Set("notes", params.Notes)
	}
	if params.When != "" {
		q.Set("when", params.When)
	}
	if params.Deadline != "" {
		q.Set("deadline", params.Deadline)
	}
	if params.Tags != "" {
		q.Set("tags", params.Tags)
	}
	if params.Area != "" {
		q.Set("area", params.Area)
	}
	if len(params.ToDos) > 0 {
		q.Set("to-dos", strings.Join(params.ToDos, "\n"))
	}
	if params.Completed {
		q.Set("completed", "true")
	}
	if params.Canceled {
		q.Set("canceled", "true")
	}

	// Use %20 for spaces instead of + (Things doesn't decode + as space)
	u.RawQuery = strings.ReplaceAll(q.Encode(), "+", "%20")
	return u.String()
}

// UpdateParams contains parameters for updating a task
type UpdateParams struct {
	ID           string
	AuthToken    string
	Title        string
	Notes        string
	PrependNotes string
	AppendNotes  string
	When         string
	Deadline     string
	Tags         string
	AddTags      string
	Completed    bool
	Canceled     bool
}

// BuildUpdateURL builds a things:///update URL
func BuildUpdateURL(params UpdateParams) string {
	u := url.URL{
		Scheme: "things",
		Host:   "",
		Path:   "/update",
	}

	q := u.Query()
	q.Set("id", params.ID)
	q.Set("auth-token", params.AuthToken)

	if params.Title != "" {
		q.Set("title", params.Title)
	}
	if params.Notes != "" {
		q.Set("notes", params.Notes)
	}
	if params.PrependNotes != "" {
		q.Set("prepend-notes", params.PrependNotes)
	}
	if params.AppendNotes != "" {
		q.Set("append-notes", params.AppendNotes)
	}
	if params.When != "" {
		q.Set("when", params.When)
	}
	if params.Deadline != "" {
		q.Set("deadline", params.Deadline)
	}
	if params.Tags != "" {
		q.Set("tags", params.Tags)
	}
	if params.AddTags != "" {
		q.Set("add-tags", params.AddTags)
	}
	if params.Completed {
		q.Set("completed", "true")
	}
	if params.Canceled {
		q.Set("canceled", "true")
	}

	// Use %20 for spaces instead of + (Things doesn't decode + as space)
	u.RawQuery = strings.ReplaceAll(q.Encode(), "+", "%20")
	return u.String()
}
