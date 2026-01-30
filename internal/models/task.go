package models

import (
	"database/sql"
	"time"
)

// Task represents a Things 3 task
type Task struct {
	UUID        string         `json:"uuid"`
	Title       string         `json:"title"`
	Notes       sql.NullString `json:"notes,omitempty"`
	Status      TaskStatus     `json:"status"`
	Type        TaskType       `json:"type"`
	Created     sql.NullTime   `json:"created,omitempty"`
	Modified    sql.NullTime   `json:"modified,omitempty"`
	Scheduled   sql.NullTime   `json:"scheduled,omitempty"`
	Deadline    sql.NullTime   `json:"due,omitempty"`
	Completed   sql.NullTime   `json:"completed,omitempty"`
	AreaName    sql.NullString `json:"area_name,omitempty"`
	ProjectName sql.NullString `json:"project_name,omitempty"`
	HeadingName sql.NullString `json:"heading_name,omitempty"`
	Tags        sql.NullString `json:"tags,omitempty"`
	IsRepeating bool           `json:"is_repeating"`
	TodayIndex  sql.NullInt64  `json:"today_index,omitempty"`
}

// TaskJSON is the JSON-serializable version of Task
type TaskJSON struct {
	UUID        string `json:"uuid"`
	Title       string `json:"title"`
	Notes       string `json:"notes,omitempty"`
	Status      string `json:"status"`
	Type        string `json:"type"`
	Created     string `json:"created,omitempty"`
	Modified    string `json:"modified,omitempty"`
	Scheduled   string `json:"scheduled,omitempty"`
	Due         string `json:"due,omitempty"`
	Completed   string `json:"completed,omitempty"`
	AreaName    string `json:"area_name,omitempty"`
	ProjectName string `json:"project_name,omitempty"`
	HeadingName string `json:"heading_name,omitempty"`
	Tags        string `json:"tags,omitempty"`
	IsRepeating bool   `json:"is_repeating"`
}

// ToJSON converts Task to its JSON-serializable form
func (t *Task) ToJSON() TaskJSON {
	return TaskJSON{
		UUID:        t.UUID,
		Title:       t.Title,
		Notes:       nullString(t.Notes),
		Status:      t.Status.String(),
		Type:        t.Type.String(),
		Created:     formatTime(t.Created),
		Modified:    formatTime(t.Modified),
		Scheduled:   formatTime(t.Scheduled),
		Due:         formatTime(t.Deadline),
		Completed:   formatTime(t.Completed),
		AreaName:    nullString(t.AreaName),
		ProjectName: nullString(t.ProjectName),
		HeadingName: nullString(t.HeadingName),
		Tags:        nullString(t.Tags),
		IsRepeating: t.IsRepeating,
	}
}

func nullString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func formatTime(t sql.NullTime) string {
	if t.Valid {
		return t.Time.Format(time.RFC3339)
	}
	return ""
}

func formatDate(t sql.NullTime) string {
	if t.Valid {
		return t.Time.Format("2006-01-02")
	}
	return ""
}
