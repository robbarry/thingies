package models

import "database/sql"

// Tag represents a Things 3 tag
type Tag struct {
	UUID      string         `json:"uuid"`
	Title     string         `json:"title"`
	Shortcut  sql.NullString `json:"shortcut,omitempty"`
	TaskCount int            `json:"task_count"`
}

// TagJSON is the JSON-serializable version of Tag
type TagJSON struct {
	UUID      string `json:"uuid"`
	Title     string `json:"title"`
	Shortcut  string `json:"shortcut,omitempty"`
	TaskCount int    `json:"task_count"`
}

// ToJSON converts Tag to its JSON-serializable form
func (t *Tag) ToJSON() TagJSON {
	shortcut := ""
	if t.Shortcut.Valid {
		shortcut = t.Shortcut.String
	}
	return TagJSON{
		UUID:      t.UUID,
		Title:     t.Title,
		Shortcut:  shortcut,
		TaskCount: t.TaskCount,
	}
}
