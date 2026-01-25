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
	Use:   "snapshot",
	Short: "Show hierarchical view",
	Long:  `Show a hierarchical view of all areas, projects, and tasks.`,
	RunE:  runSnapshot,
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
	Areas   []snapshotArea    `json:"areas"`
	Inbox   []models.TaskJSON `json:"inbox"`
	Today   []models.TaskJSON `json:"today"`
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
		headerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
		areaStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5"))
		projStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
		taskStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
		countStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	)

	if noColor {
		headerStyle = lipgloss.NewStyle()
		areaStyle = lipgloss.NewStyle()
		projStyle = lipgloss.NewStyle()
		taskStyle = lipgloss.NewStyle()
		countStyle = lipgloss.NewStyle()
	}

	// Today
	if len(output.Today) > 0 {
		fmt.Println(headerStyle.Render("📅 Today"))
		for _, t := range output.Today {
			fmt.Printf("  %s %s\n", models.StatusIncomplete.Icon(), taskStyle.Render(t.Title))
		}
		fmt.Println()
	}

	// Inbox
	if len(output.Inbox) > 0 {
		fmt.Println(headerStyle.Render("📥 Inbox"))
		for _, t := range output.Inbox {
			fmt.Printf("  %s %s\n", models.StatusIncomplete.Icon(), taskStyle.Render(t.Title))
		}
		fmt.Println()
	}

	// Areas
	for _, area := range output.Areas {
		areaCount := area.OpenTasks
		for _, p := range area.Projects {
			areaCount += len(p.Tasks)
		}

		fmt.Printf("%s %s\n", areaStyle.Render("■ "+area.Title), countStyle.Render(fmt.Sprintf("(%d)", areaCount)))

		// Projects
		for _, proj := range area.Projects {
			taskCount := len(proj.Tasks)
			fmt.Printf("  %s %s\n", projStyle.Render("▸ "+proj.Title), countStyle.Render(fmt.Sprintf("(%d)", taskCount)))

			for _, t := range proj.Tasks {
				icon := models.StatusIncomplete.Icon()
				fmt.Printf("    %s %s\n", icon, taskStyle.Render(t.Title))
			}
		}

		// Direct tasks
		for _, t := range area.Tasks {
			icon := models.StatusIncomplete.Icon()
			fmt.Printf("  %s %s\n", icon, taskStyle.Render(t.Title))
		}

		fmt.Println()
	}

	// Summary
	totalTasks := len(output.Inbox) + len(output.Today)
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
