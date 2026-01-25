package output

import (
	"encoding/json"
	"fmt"

	"thingies/internal/models"
)

// JSONFormatter formats output as JSON
type JSONFormatter struct{}

// NewJSONFormatter creates a new JSONFormatter
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

func (f *JSONFormatter) output(v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

// FormatTasks formats tasks as JSON
func (f *JSONFormatter) FormatTasks(tasks []models.Task) error {
	result := make([]models.TaskJSON, len(tasks))
	for i, t := range tasks {
		result[i] = t.ToJSON()
	}
	return f.output(result)
}

// FormatTask formats a single task as JSON
func (f *JSONFormatter) FormatTask(task *models.Task) error {
	return f.output(task.ToJSON())
}

// FormatProjects formats projects as JSON
func (f *JSONFormatter) FormatProjects(projects []models.Project) error {
	result := make([]models.ProjectJSON, len(projects))
	for i, p := range projects {
		result[i] = p.ToJSON()
	}
	return f.output(result)
}

// FormatProject formats a single project with tasks as JSON
func (f *JSONFormatter) FormatProject(project *models.Project, tasks []models.Task) error {
	taskResults := make([]models.TaskJSON, len(tasks))
	for i, t := range tasks {
		taskResults[i] = t.ToJSON()
	}

	result := struct {
		models.ProjectJSON
		Tasks []models.TaskJSON `json:"tasks"`
	}{
		ProjectJSON: project.ToJSON(),
		Tasks:       taskResults,
	}
	return f.output(result)
}

// FormatAreas formats areas as JSON
func (f *JSONFormatter) FormatAreas(areas []models.Area) error {
	return f.output(areas)
}

// FormatArea formats a single area with projects and tasks as JSON
func (f *JSONFormatter) FormatArea(area *models.Area, projects []models.Project, tasks []models.Task) error {
	projectResults := make([]models.ProjectJSON, len(projects))
	for i, p := range projects {
		projectResults[i] = p.ToJSON()
	}

	taskResults := make([]models.TaskJSON, len(tasks))
	for i, t := range tasks {
		taskResults[i] = t.ToJSON()
	}

	result := struct {
		*models.Area
		Projects []models.ProjectJSON `json:"projects"`
		Tasks    []models.TaskJSON    `json:"tasks"`
	}{
		Area:     area,
		Projects: projectResults,
		Tasks:    taskResults,
	}
	return f.output(result)
}

// FormatTags formats tags as JSON
func (f *JSONFormatter) FormatTags(tags []models.Tag) error {
	result := make([]models.TagJSON, len(tags))
	for i, t := range tags {
		result[i] = t.ToJSON()
	}
	return f.output(result)
}

// FormatSearchResults formats search results as JSON
func (f *JSONFormatter) FormatSearchResults(tasks []models.Task, term string) error {
	result := make([]models.TaskJSON, len(tasks))
	for i, t := range tasks {
		result[i] = t.ToJSON()
	}
	return f.output(result)
}
