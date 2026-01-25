package output

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"thingies/internal/models"
)

var (
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	cyan        = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	green       = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	blue        = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
	magenta     = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	red         = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	yellow      = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	dim         = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

// TableFormatter formats output as tables
type TableFormatter struct {
	noColor bool
}

// NewTableFormatter creates a new TableFormatter
func NewTableFormatter(noColor bool) *TableFormatter {
	return &TableFormatter{noColor: noColor}
}

func (f *TableFormatter) style(s lipgloss.Style, text string) string {
	if f.noColor {
		return text
	}
	return s.Render(text)
}

// FormatTasks formats a list of tasks as a table
func (f *TableFormatter) FormatTasks(tasks []models.Task) error {
	if len(tasks) == 0 {
		fmt.Println(f.style(yellow, "No tasks found"))
		return nil
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("8"))).
		Headers("Title", "Status", "Project", "Area", "Due", "Tags")

	for _, task := range tasks {
		status := models.TaskStatus(task.Status).Icon()
		if task.IsRepeating {
			status += " 🔁"
		}

		due := ""
		if task.Deadline.Valid {
			due = task.Deadline.Time.Format("2006-01-02")
		}

		projectName := ""
		if task.ProjectName.Valid {
			projectName = task.ProjectName.String
		}

		areaName := ""
		if task.AreaName.Valid {
			areaName = task.AreaName.String
		}

		tags := ""
		if task.Tags.Valid {
			tags = task.Tags.String
		}

		t.Row(
			f.style(cyan, task.Title),
			f.style(green, status),
			f.style(blue, projectName),
			f.style(magenta, areaName),
			f.style(red, due),
			f.style(yellow, tags),
		)
	}

	fmt.Println(t)
	fmt.Println(f.style(dim, fmt.Sprintf("\nFound %d task(s)", len(tasks))))
	return nil
}

// FormatTask formats a single task with full details
func (f *TableFormatter) FormatTask(task *models.Task) error {
	fmt.Println(f.style(headerStyle, "Task Details"))
	fmt.Println(strings.Repeat("─", 40))

	fmt.Printf("%s: %s\n", f.style(dim, "UUID"), task.UUID)
	fmt.Printf("%s: %s\n", f.style(dim, "Title"), f.style(cyan, task.Title))
	fmt.Printf("%s: %s %s\n", f.style(dim, "Status"), models.TaskStatus(task.Status).Icon(), models.TaskStatus(task.Status).String())

	if task.Notes.Valid && task.Notes.String != "" {
		fmt.Printf("%s:\n%s\n", f.style(dim, "Notes"), task.Notes.String)
	}

	if task.ProjectName.Valid {
		fmt.Printf("%s: %s\n", f.style(dim, "Project"), f.style(blue, task.ProjectName.String))
	}

	if task.AreaName.Valid {
		fmt.Printf("%s: %s\n", f.style(dim, "Area"), f.style(magenta, task.AreaName.String))
	}

	if task.Scheduled.Valid {
		fmt.Printf("%s: %s\n", f.style(dim, "Scheduled"), task.Scheduled.Time.Format("2006-01-02"))
	}

	if task.Deadline.Valid {
		fmt.Printf("%s: %s\n", f.style(dim, "Due"), f.style(red, task.Deadline.Time.Format("2006-01-02")))
	}

	if task.Tags.Valid && task.Tags.String != "" {
		fmt.Printf("%s: %s\n", f.style(dim, "Tags"), f.style(yellow, task.Tags.String))
	}

	if task.IsRepeating {
		fmt.Printf("%s: %s\n", f.style(dim, "Repeating"), "Yes 🔁")
	}

	if task.Created.Valid {
		fmt.Printf("%s: %s\n", f.style(dim, "Created"), task.Created.Time.Format("2006-01-02 15:04"))
	}

	if task.Modified.Valid {
		fmt.Printf("%s: %s\n", f.style(dim, "Modified"), task.Modified.Time.Format("2006-01-02 15:04"))
	}

	return nil
}

// FormatProjects formats a list of projects as a table
func (f *TableFormatter) FormatProjects(projects []models.Project) error {
	if len(projects) == 0 {
		fmt.Println(f.style(yellow, "No projects found"))
		return nil
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("8"))).
		Headers("Project", "Area", "Tasks", "Status")

	for _, proj := range projects {
		status := "Active"
		if models.TaskStatus(proj.Status) == models.StatusCompleted {
			status = "Completed"
		}

		areaName := ""
		if proj.AreaName.Valid {
			areaName = proj.AreaName.String
		}

		taskCount := fmt.Sprintf("%d/%d", proj.OpenTasks, proj.TotalTasks)

		t.Row(
			f.style(cyan, proj.Title),
			f.style(magenta, areaName),
			f.style(green, taskCount),
			f.style(yellow, status),
		)
	}

	fmt.Println(t)
	fmt.Println(f.style(dim, fmt.Sprintf("\nFound %d project(s)", len(projects))))
	return nil
}

// FormatProject formats a single project with its tasks
func (f *TableFormatter) FormatProject(project *models.Project, tasks []models.Task) error {
	fmt.Println(f.style(headerStyle, "Project: "+project.Title))
	fmt.Println(strings.Repeat("─", 40))

	fmt.Printf("%s: %s\n", f.style(dim, "UUID"), project.UUID)

	status := "Active"
	if models.TaskStatus(project.Status) == models.StatusCompleted {
		status = "Completed"
	}
	fmt.Printf("%s: %s\n", f.style(dim, "Status"), status)

	if project.AreaName.Valid {
		fmt.Printf("%s: %s\n", f.style(dim, "Area"), f.style(magenta, project.AreaName.String))
	}

	fmt.Printf("%s: %d open / %d total\n", f.style(dim, "Tasks"), project.OpenTasks, project.TotalTasks)

	if project.Notes.Valid && project.Notes.String != "" {
		fmt.Printf("%s:\n%s\n", f.style(dim, "Notes"), project.Notes.String)
	}

	if len(tasks) > 0 {
		fmt.Println()
		fmt.Println(f.style(headerStyle, "Tasks"))
		return f.FormatTasks(tasks)
	}

	return nil
}

// FormatAreas formats a list of areas as a table
func (f *TableFormatter) FormatAreas(areas []models.Area) error {
	if len(areas) == 0 {
		fmt.Println(f.style(yellow, "No areas found"))
		return nil
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("8"))).
		Headers("Area", "Open Tasks", "Active Projects")

	for _, area := range areas {
		t.Row(
			f.style(cyan, area.Title),
			f.style(green, fmt.Sprintf("%d", area.OpenTasks)),
			f.style(blue, fmt.Sprintf("%d", area.ActiveProjects)),
		)
	}

	fmt.Println(t)
	fmt.Println(f.style(dim, fmt.Sprintf("\nFound %d area(s)", len(areas))))
	return nil
}

// FormatArea formats a single area with its projects and tasks
func (f *TableFormatter) FormatArea(area *models.Area, projects []models.Project, tasks []models.Task) error {
	fmt.Println(f.style(headerStyle, "Area: "+area.Title))
	fmt.Println(strings.Repeat("─", 40))

	fmt.Printf("%s: %s\n", f.style(dim, "UUID"), area.UUID)
	fmt.Printf("%s: %d\n", f.style(dim, "Open Tasks"), area.OpenTasks)
	fmt.Printf("%s: %d\n", f.style(dim, "Active Projects"), area.ActiveProjects)

	if len(projects) > 0 {
		fmt.Println()
		fmt.Println(f.style(headerStyle, "Projects"))
		f.FormatProjects(projects)
	}

	if len(tasks) > 0 {
		fmt.Println()
		fmt.Println(f.style(headerStyle, "Tasks (not in projects)"))
		f.FormatTasks(tasks)
	}

	return nil
}

// FormatTags formats a list of tags as a table
func (f *TableFormatter) FormatTags(tags []models.Tag) error {
	if len(tags) == 0 {
		fmt.Println(f.style(yellow, "No tags found"))
		return nil
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("8"))).
		Headers("Tag", "Shortcut", "Tasks")

	for _, tag := range tags {
		shortcut := ""
		if tag.Shortcut.Valid {
			shortcut = tag.Shortcut.String
		}

		t.Row(
			f.style(cyan, tag.Title),
			f.style(yellow, shortcut),
			f.style(green, fmt.Sprintf("%d", tag.TaskCount)),
		)
	}

	fmt.Println(t)
	fmt.Println(f.style(dim, fmt.Sprintf("\nFound %d tag(s)", len(tags))))
	return nil
}

// FormatSearchResults formats search results as a table
func (f *TableFormatter) FormatSearchResults(tasks []models.Task, term string) error {
	if len(tasks) == 0 {
		fmt.Println(f.style(yellow, fmt.Sprintf("No results found for '%s'", term)))
		return nil
	}

	fmt.Println(f.style(headerStyle, fmt.Sprintf("Search Results for '%s'", term)))

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("8"))).
		Headers("Type", "Title", "Project", "Area", "Status")

	for _, task := range tasks {
		typeName := task.Type.String()
		status := task.Status.String()
		if task.Type == models.TypeTask && task.IsRepeating {
			status += " 🔁"
		}

		projectName := ""
		if task.ProjectName.Valid {
			projectName = task.ProjectName.String
		}

		areaName := ""
		if task.AreaName.Valid {
			areaName = task.AreaName.String
		}

		t.Row(
			f.style(yellow, typeName),
			f.style(cyan, task.Title),
			f.style(blue, projectName),
			f.style(magenta, areaName),
			f.style(green, status),
		)
	}

	fmt.Println(t)
	fmt.Println(f.style(dim, fmt.Sprintf("\nFound %d result(s)", len(tasks))))
	return nil
}

// PrintError prints an error message
func PrintError(err error) {
	fmt.Fprintln(os.Stderr, lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render("Error: "+err.Error()))
}
