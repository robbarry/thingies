package cmd

import (
	"github.com/spf13/cobra"
	"thingies/internal/cmd/shared"
	"thingies/internal/db"
)

var todayCmd = &cobra.Command{
	Use:   "today",
	Short: "Show today's tasks",
	Long:  `Show tasks scheduled for today.`,
	RunE:  runToday,
}

func runToday(cmd *cobra.Command, args []string) error {
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

	formatter := shared.GetFormatter(cmd)
	return formatter.FormatTasks(tasks)
}
