package models

import "database/sql"

// Project represents a Things 3 project
type Project struct {
	UUID         string         `json:"uuid"`
	Title        string         `json:"title"`
	Notes        sql.NullString `json:"notes,omitempty"`
	Status       TaskStatus     `json:"status"`
	AreaName     sql.NullString `json:"area_name,omitempty"`
	OpenTasks    int            `json:"open_tasks"`
	TotalTasks   int            `json:"total_tasks"`
}

// ProjectJSON is the JSON-serializable version of Project
type ProjectJSON struct {
	UUID       string `json:"uuid"`
	Title      string `json:"title"`
	Notes      string `json:"notes,omitempty"`
	Status     string `json:"status"`
	AreaName   string `json:"area_name,omitempty"`
	OpenTasks  int    `json:"open_tasks"`
	TotalTasks int    `json:"total_tasks"`
}

// ToJSON converts Project to its JSON-serializable form
func (p *Project) ToJSON() ProjectJSON {
	return ProjectJSON{
		UUID:       p.UUID,
		Title:      p.Title,
		Notes:      nullString(p.Notes),
		Status:     p.Status.String(),
		AreaName:   nullString(p.AreaName),
		OpenTasks:  p.OpenTasks,
		TotalTasks: p.TotalTasks,
	}
}
