package models

// Area represents a Things 3 area
type Area struct {
	UUID           string `json:"uuid"`
	Title          string `json:"title"`
	OpenTasks      int    `json:"open_tasks"`
	ActiveProjects int    `json:"active_projects"`
}
