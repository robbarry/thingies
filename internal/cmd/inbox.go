package cmd

import (
	"github.com/spf13/cobra"
	"thingies/internal/cmd/shared"
	"thingies/internal/db"
)

var inboxCmd = &cobra.Command{
	Use:   "inbox",
	Short: "Show inbox tasks",
	Long:  `Show tasks in the inbox (not assigned to any area or project).`,
	RunE:  runInbox,
}

func runInbox(cmd *cobra.Command, args []string) error {
	thingsDB, err := db.Open(shared.GetDBPath(cmd))
	if err != nil {
		return err
	}
	defer thingsDB.Close()

	tasks, err := thingsDB.GetInboxTasks()
	if err != nil {
		return err
	}

	formatter := shared.GetFormatter(cmd)
	return formatter.FormatTasks(tasks)
}
