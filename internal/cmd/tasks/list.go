package tasks

import (
	"github.com/spf13/cobra"
	"thingies/internal/cmd/shared"
	"thingies/internal/db"
	"thingies/internal/output"
)

var (
	listStatus        string
	listArea          string
	listProject       string
	listTag           string
	listToday         bool
	listIncludeFuture bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	Long:  `List tasks with optional filters for status, area, project, tag, and today view.`,
	RunE:  runList,
}

func init() {
	listCmd.Flags().StringVar(&listStatus, "status", "incomplete", "Filter by status (all, incomplete, completed, canceled)")
	listCmd.Flags().StringVar(&listArea, "area", "", "Filter by area name")
	listCmd.Flags().StringVar(&listProject, "project", "", "Filter by project name")
	listCmd.Flags().StringVar(&listTag, "tag", "", "Filter by tag name")
	listCmd.Flags().BoolVar(&listToday, "today", false, "Show only Today items")
	listCmd.Flags().BoolVar(&listIncludeFuture, "include-future", false, "Include future instances of repeating tasks")
}

func runList(cmd *cobra.Command, args []string) error {
	thingsDB, err := db.Open(shared.GetDBPath(cmd))
	if err != nil {
		return err
	}
	defer thingsDB.Close()

	filter := db.TaskFilter{
		Status:        listStatus,
		Area:          listArea,
		Project:       listProject,
		Tag:           listTag,
		Today:         listToday,
		IncludeFuture: listIncludeFuture,
	}

	tasks, err := thingsDB.ListTasks(filter)
	if err != nil {
		return err
	}

	formatter := shared.GetFormatter(cmd)
	return formatter.FormatTasks(tasks)
}

// RunListToday runs the list command with today filter
func RunListToday(cmd *cobra.Command, args []string) error {
	thingsDB, err := db.Open(shared.GetDBPath(cmd))
	if err != nil {
		return err
	}
	defer thingsDB.Close()

	filter := db.TaskFilter{
		Status: "incomplete",
		Today:  true,
	}

	tasks, err := thingsDB.ListTasks(filter)
	if err != nil {
		return err
	}

	if shared.IsJSON(cmd) {
		return output.NewJSONFormatter().FormatTasks(tasks)
	}
	return output.NewTableFormatter(shared.IsNoColor(cmd)).FormatTasks(tasks)
}

// RunListInbox runs the list command for inbox
func RunListInbox(cmd *cobra.Command, args []string) error {
	thingsDB, err := db.Open(shared.GetDBPath(cmd))
	if err != nil {
		return err
	}
	defer thingsDB.Close()

	tasks, err := thingsDB.GetInboxTasks()
	if err != nil {
		return err
	}

	if shared.IsJSON(cmd) {
		return output.NewJSONFormatter().FormatTasks(tasks)
	}
	return output.NewTableFormatter(shared.IsNoColor(cmd)).FormatTasks(tasks)
}
