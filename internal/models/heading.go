package models

// Heading represents a Things 3 heading (section within a project)
type Heading struct {
	UUID  string `json:"uuid"`
	Title string `json:"title"`
	Index int    `json:"index"`
}
