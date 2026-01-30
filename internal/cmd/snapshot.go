package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"thingies/internal/cmd/shared"
	"thingies/internal/db"
	"thingies/internal/models"
)

var snapshotCmd = &cobra.Command{
	Use:     "snapshot",
	Aliases: []string{"all"},
	Short:   "Show hierarchical view",
	Long:    `Show a hierarchical view of all areas, projects, and tasks.`,
	RunE:    runSnapshot,
}

type snapshotArea struct {
	models.Area
	Projects []snapshotProject `json:"projects,omitempty"`
	Tasks    []models.TaskJSON `json:"tasks,omitempty"`
}

type snapshotProject struct {
	models.ProjectJSON
	Tasks []models.TaskJSON `json:"tasks,omitempty"`
}

type snapshotOutput struct {
	Today    []models.TaskJSON `json:"today"`
	Inbox    []models.TaskJSON `json:"inbox"`
	Upcoming []models.TaskJSON `json:"upcoming"`
	Someday  []models.TaskJSON `json:"someday"`
	Areas    []snapshotArea    `json:"areas"`
}

func runSnapshot(cmd *cobra.Command, args []string) error {
	thingsDB, err := db.Open(shared.GetDBPath(cmd))
	if err != nil {
		return err
	}
	defer thingsDB.Close()

	// Get all data
	areas, err := thingsDB.ListAreas()
	if err != nil {
		return err
	}

	output := snapshotOutput{
		Areas: make([]snapshotArea, 0),
	}

	// Process each area
	for _, area := range areas {
		sa := snapshotArea{Area: area}

		projects, err := thingsDB.GetAreaProjects(area.UUID, false)
		if err != nil {
			return err
		}

		for _, proj := range projects {
			sp := snapshotProject{ProjectJSON: proj.ToJSON()}

			tasks, err := thingsDB.GetProjectTasks(proj.UUID, false)
			if err != nil {
				return err
			}

			for _, t := range tasks {
				sp.Tasks = append(sp.Tasks, t.ToJSON())
			}
			sa.Projects = append(sa.Projects, sp)
		}

		// Get tasks directly under area
		areaTasks, err := thingsDB.GetAreaTasks(area.UUID, false)
		if err != nil {
			return err
		}
		for _, t := range areaTasks {
			sa.Tasks = append(sa.Tasks, t.ToJSON())
		}

		output.Areas = append(output.Areas, sa)
	}

	// Get inbox
	inboxTasks, err := thingsDB.GetInboxTasks()
	if err != nil {
		return err
	}
	for _, t := range inboxTasks {
		output.Inbox = append(output.Inbox, t.ToJSON())
	}

	// Get today
	todayTasks, err := thingsDB.ListTasks(db.TaskFilter{Status: "incomplete", Today: true})
	if err != nil {
		return err
	}
	for _, t := range todayTasks {
		output.Today = append(output.Today, t.ToJSON())
	}

	// Get upcoming
	upcomingTasks, err := thingsDB.GetUpcomingTasks()
	if err != nil {
		return err
	}
	for _, t := range upcomingTasks {
		output.Upcoming = append(output.Upcoming, t.ToJSON())
	}

	// Get someday
	somedayTasks, err := thingsDB.GetSomedayTasks()
	if err != nil {
		return err
	}
	for _, t := range somedayTasks {
		output.Someday = append(output.Someday, t.ToJSON())
	}

	if shared.IsJSON(cmd) {
		data, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	// Pretty print
	return printSnapshot(output, shared.IsNoColor(cmd))
}

func printSnapshot(output snapshotOutput, noColor bool) error {
	var (
		headerStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
		areaStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5"))
		projStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
		headingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
		taskStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
		idStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
		countStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
		dateStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	)

	if noColor {
		headerStyle = lipgloss.NewStyle()
		areaStyle = lipgloss.NewStyle()
		projStyle = lipgloss.NewStyle()
		headingStyle = lipgloss.NewStyle()
		taskStyle = lipgloss.NewStyle()
		idStyle = lipgloss.NewStyle()
		countStyle = lipgloss.NewStyle()
		dateStyle = lipgloss.NewStyle()
	}

	shortID := func(uuid string) string {
		if len(uuid) > 8 {
			return uuid[:8]
		}
		return uuid
	}

	taskContext := func(t models.TaskJSON) string {
		var hierarchy []string
		if t.AreaName != "" {
			hierarchy = append(hierarchy, areaStyle.Render(t.AreaName))
		}
		if t.ProjectName != "" {
			hierarchy = append(hierarchy, projStyle.Render(t.ProjectName))
		}
		if t.HeadingName != "" {
			hierarchy = append(hierarchy, taskStyle.Render(t.HeadingName))
		}
		if len(hierarchy) > 0 {
			return " " + countStyle.Render("(") + strings.Join(hierarchy, countStyle.Render(" > ")) + countStyle.Render(")")
		}
		return ""
	}

	// Today
	if len(output.Today) > 0 {
		fmt.Printf("%s %s\n", headerStyle.Render("📅 Today"), countStyle.Render(fmt.Sprintf("(%d)", len(output.Today))))
		for _, t := range output.Today {
			fmt.Printf("  %s %s %s%s\n", idStyle.Render(shortID(t.UUID)), models.StatusIncomplete.Icon(), taskStyle.Render(t.Title), taskContext(t))
		}
		fmt.Println()
	}

	// Inbox
	if len(output.Inbox) > 0 {
		fmt.Printf("%s %s\n", headerStyle.Render("📥 Inbox"), countStyle.Render(fmt.Sprintf("(%d)", len(output.Inbox))))
		for _, t := range output.Inbox {
			fmt.Printf("  %s %s %s%s\n", idStyle.Render(shortID(t.UUID)), models.StatusIncomplete.Icon(), taskStyle.Render(t.Title), taskContext(t))
		}
		fmt.Println()
	}

	// Upcoming
	if len(output.Upcoming) > 0 {
		fmt.Printf("%s %s\n", headerStyle.Render("📆 Upcoming"), countStyle.Render(fmt.Sprintf("(%d)", len(output.Upcoming))))
		for _, t := range output.Upcoming {
			scheduled := ""
			if t.Scheduled != "" {
				// Extract just the date part (YYYY-MM-DD) from ISO format
				if len(t.Scheduled) >= 10 {
					scheduled = " " + dateStyle.Render(t.Scheduled[:10])
				}
			}
			fmt.Printf("  %s %s %s%s\n", idStyle.Render(shortID(t.UUID)), models.StatusIncomplete.Icon(), taskStyle.Render(t.Title), scheduled)
		}
		fmt.Println()
	}

	// Someday
	if len(output.Someday) > 0 {
		fmt.Printf("%s %s\n", headerStyle.Render("💭 Someday"), countStyle.Render(fmt.Sprintf("(%d)", len(output.Someday))))
		for _, t := range output.Someday {
			fmt.Printf("  %s %s %s%s\n", idStyle.Render(shortID(t.UUID)), models.StatusIncomplete.Icon(), taskStyle.Render(t.Title), taskContext(t))
		}
		fmt.Println()
	}

	// Areas
	for _, area := range output.Areas {
		areaCount := area.OpenTasks
		for _, p := range area.Projects {
			areaCount += len(p.Tasks)
		}

		fmt.Printf("%s %s %s\n", idStyle.Render(shortID(area.UUID)), areaStyle.Render("■ "+area.Title), countStyle.Render(fmt.Sprintf("(%d)", areaCount)))

		// Projects
		for _, proj := range area.Projects {
			taskCount := len(proj.Tasks)
			fmt.Printf("  %s %s %s\n", idStyle.Render(shortID(proj.UUID)), projStyle.Render("▸ "+proj.Title), countStyle.Render(fmt.Sprintf("(%d)", taskCount)))

			// Group tasks by heading
			tasksByHeading := make(map[string][]models.TaskJSON)
			var headingOrder []string
			for _, t := range proj.Tasks {
				heading := t.HeadingName
				if _, exists := tasksByHeading[heading]; !exists {
					headingOrder = append(headingOrder, heading)
				}
				tasksByHeading[heading] = append(tasksByHeading[heading], t)
			}

			// Print tasks grouped by heading
			for _, heading := range headingOrder {
				tasks := tasksByHeading[heading]
				if heading != "" {
					fmt.Printf("    %s\n", headingStyle.Render("› "+heading))
					for _, t := range tasks {
						fmt.Printf("      %s %s %s\n", idStyle.Render(shortID(t.UUID)), models.StatusIncomplete.Icon(), taskStyle.Render(t.Title))
					}
				} else {
					for _, t := range tasks {
						fmt.Printf("    %s %s %s\n", idStyle.Render(shortID(t.UUID)), models.StatusIncomplete.Icon(), taskStyle.Render(t.Title))
					}
				}
			}
		}

		// Direct tasks
		for _, t := range area.Tasks {
			fmt.Printf("  %s %s %s\n", idStyle.Render(shortID(t.UUID)), models.StatusIncomplete.Icon(), taskStyle.Render(t.Title))
		}

		fmt.Println()
	}

	// Summary
	totalTasks := len(output.Inbox) + len(output.Today) + len(output.Upcoming) + len(output.Someday)
	for _, a := range output.Areas {
		totalTasks += len(a.Tasks)
		for _, p := range a.Projects {
			totalTasks += len(p.Tasks)
		}
	}

	fmt.Println(strings.Repeat("─", 40))
	fmt.Printf("Total: %d open tasks across %d areas\n", totalTasks, len(output.Areas))

	return nil
}
