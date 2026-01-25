package output

import (
	"thingies/internal/models"
)

// Formatter defines the interface for output formatting
type Formatter interface {
	FormatTasks(tasks []models.Task) error
	FormatTask(task *models.Task) error
	FormatProjects(projects []models.Project) error
	FormatProject(project *models.Project, tasks []models.Task) error
	FormatAreas(areas []models.Area) error
	FormatArea(area *models.Area, projects []models.Project, tasks []models.Task) error
	FormatTags(tags []models.Tag) error
	FormatSearchResults(tasks []models.Task, term string) error
}
