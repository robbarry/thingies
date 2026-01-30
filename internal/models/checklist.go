package models

// ChecklistItem represents a Things 3 checklist item within a task
type ChecklistItem struct {
	UUID      string `json:"uuid"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
	Index     int    `json:"index"`
}
