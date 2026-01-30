package cmd

import (
	"github.com/spf13/cobra"
	"thingies/internal/cmd/shared"
	"thingies/internal/db"
)

var upcomingCmd = &cobra.Command{
	Use:   "upcoming",
	Short: "Show upcoming tasks",
	Long:  `Show tasks scheduled for the future.`,
	RunE:  runUpcoming,
}

func runUpcoming(cmd *cobra.Command, args []string) error {
	thingsDB, err := db.Open(shared.GetDBPath(cmd))
	if err != nil {
		return err
	}
	defer thingsDB.Close()

	tasks, err := thingsDB.GetUpcomingTasks()
	if err != nil {
		return err
	}

	formatter := shared.GetFormatter(cmd)
	return formatter.FormatTasks(tasks)
}
