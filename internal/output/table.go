package output

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
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

// FormatTasks formats a list of tasks
func (f *TableFormatter) FormatTasks(tasks []models.Task) error {
	if len(tasks) == 0 {
		fmt.Println(f.style(yellow, "No tasks found"))
		return nil
	}

	for _, task := range tasks {
		// Show short ID (first 8 chars of UUID)
		shortID := task.UUID
		if len(shortID) > 8 {
			shortID = shortID[:8]
		}

		status := models.TaskStatus(task.Status).Icon()
		if task.IsRepeating {
			status += " 🔁"
		}

		// Build context parts
		var context []string
		// Show Area > Project > Heading hierarchy
		var hierarchy []string
		if task.AreaName.Valid && task.AreaName.String != "" {
			hierarchy = append(hierarchy, f.style(magenta, task.AreaName.String))
		}
		if task.ProjectName.Valid && task.ProjectName.String != "" {
			hierarchy = append(hierarchy, f.style(blue, task.ProjectName.String))
		}
		if task.HeadingName.Valid && task.HeadingName.String != "" {
			hierarchy = append(hierarchy, f.style(cyan, task.HeadingName.String))
		}
		if len(hierarchy) > 0 {
			context = append(context, strings.Join(hierarchy, f.style(dim, " > ")))
		}
		if task.Deadline.Valid {
			context = append(context, f.style(red, "due ")+f.style(red, task.Deadline.Time.Format("2006-01-02")))
		}
		if task.Tags.Valid && task.Tags.String != "" {
			context = append(context, f.style(yellow, task.Tags.String))
		}

		line := fmt.Sprintf("%s %s %s", f.style(dim, shortID), f.style(green, status), f.style(cyan, task.Title))
		if len(context) > 0 {
			line += " " + f.style(dim, "(") + strings.Join(context, f.style(dim, ", ")) + f.style(dim, ")")
		}
		fmt.Println(line)
	}

	fmt.Println(f.style(dim, fmt.Sprintf("\n%d task(s)", len(tasks))))
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

// FormatProjects formats a list of projects
func (f *TableFormatter) FormatProjects(projects []models.Project) error {
	if len(projects) == 0 {
		fmt.Println(f.style(yellow, "No projects found"))
		return nil
	}

	for _, proj := range projects {
		// Show short ID (first 8 chars of UUID)
		shortID := proj.UUID
		if len(shortID) > 8 {
			shortID = shortID[:8]
		}

		taskCount := fmt.Sprintf("%d/%d", proj.OpenTasks, proj.TotalTasks)

		// Build context parts
		var context []string
		if proj.AreaName.Valid && proj.AreaName.String != "" {
			context = append(context, f.style(magenta, proj.AreaName.String))
		}
		context = append(context, f.style(green, taskCount))
		if models.TaskStatus(proj.Status) == models.StatusCompleted {
			context = append(context, f.style(yellow, "Completed"))
		}

		line := fmt.Sprintf("%s %s", f.style(dim, shortID), f.style(cyan, proj.Title))
		if len(context) > 0 {
			line += " " + f.style(dim, "(") + strings.Join(context, f.style(dim, ", ")) + f.style(dim, ")")
		}
		fmt.Println(line)
	}

	fmt.Println(f.style(dim, fmt.Sprintf("\n%d project(s)", len(projects))))
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

// FormatAreas formats a list of areas
func (f *TableFormatter) FormatAreas(areas []models.Area) error {
	if len(areas) == 0 {
		fmt.Println(f.style(yellow, "No areas found"))
		return nil
	}

	for _, area := range areas {
		// Show short ID (first 8 chars of UUID)
		shortID := area.UUID
		if len(shortID) > 8 {
			shortID = shortID[:8]
		}

		context := fmt.Sprintf("%s tasks, %s projects",
			f.style(green, fmt.Sprintf("%d", area.OpenTasks)),
			f.style(blue, fmt.Sprintf("%d", area.ActiveProjects)))

		line := fmt.Sprintf("%s %s %s",
			f.style(dim, shortID),
			f.style(cyan, area.Title),
			f.style(dim, "("+context+")"))
		fmt.Println(line)
	}

	fmt.Println(f.style(dim, fmt.Sprintf("\n%d area(s)", len(areas))))
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

// FormatTags formats a list of tags
func (f *TableFormatter) FormatTags(tags []models.Tag) error {
	if len(tags) == 0 {
		fmt.Println(f.style(yellow, "No tags found"))
		return nil
	}

	for _, tag := range tags {
		// Show short ID (first 8 chars of UUID)
		shortID := tag.UUID
		if len(shortID) > 8 {
			shortID = shortID[:8]
		}

		var context []string
		if tag.Shortcut.Valid && tag.Shortcut.String != "" {
			context = append(context, f.style(yellow, tag.Shortcut.String))
		}
		context = append(context, f.style(green, fmt.Sprintf("%d tasks", tag.TaskCount)))

		line := fmt.Sprintf("%s %s %s",
			f.style(dim, shortID),
			f.style(cyan, tag.Title),
			f.style(dim, "("+strings.Join(context, ", ")+")"))
		fmt.Println(line)
	}

	fmt.Println(f.style(dim, fmt.Sprintf("\n%d tag(s)", len(tags))))
	return nil
}

// FormatSearchResults formats search results
func (f *TableFormatter) FormatSearchResults(tasks []models.Task, term string) error {
	if len(tasks) == 0 {
		fmt.Println(f.style(yellow, fmt.Sprintf("No results found for '%s'", term)))
		return nil
	}

	for _, task := range tasks {
		// Show short ID (first 8 chars of UUID)
		shortID := task.UUID
		if len(shortID) > 8 {
			shortID = shortID[:8]
		}

		typeName := task.Type.String()
		status := task.Status.String()
		if task.Type == models.TypeTask && task.IsRepeating {
			status += " 🔁"
		}

		// Build context parts
		var context []string
		context = append(context, f.style(yellow, typeName))
		// Show Area > Project > Heading hierarchy
		var hierarchy []string
		if task.AreaName.Valid && task.AreaName.String != "" {
			hierarchy = append(hierarchy, f.style(magenta, task.AreaName.String))
		}
		if task.ProjectName.Valid && task.ProjectName.String != "" {
			hierarchy = append(hierarchy, f.style(blue, task.ProjectName.String))
		}
		if task.HeadingName.Valid && task.HeadingName.String != "" {
			hierarchy = append(hierarchy, f.style(cyan, task.HeadingName.String))
		}
		if len(hierarchy) > 0 {
			context = append(context, strings.Join(hierarchy, f.style(dim, " > ")))
		}
		context = append(context, f.style(green, status))

		line := fmt.Sprintf("%s %s %s",
			f.style(dim, shortID),
			f.style(cyan, task.Title),
			f.style(dim, "("+strings.Join(context, ", ")+")"))
		fmt.Println(line)
	}

	fmt.Println(f.style(dim, fmt.Sprintf("\n%d result(s)", len(tasks))))
	return nil
}

// PrintError prints an error message
func PrintError(err error) {
	fmt.Fprintln(os.Stderr, lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render("Error: "+err.Error()))
}
