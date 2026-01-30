package tasks

import (
	"github.com/spf13/cobra"
	"thingies/internal/cmd/shared"
	"thingies/internal/db"
)

var showCmd = &cobra.Command{
	Use:   "show <uuid>",
	Short: "Show task details",
	Long:  `Show detailed information about a specific task.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runShow,
}

func runShow(cmd *cobra.Command, args []string) error {
	thingsDB, err := db.Open(shared.GetDBPath(cmd))
	if err != nil {
		return err
	}
	defer thingsDB.Close()

	// Resolve short UUID if needed
	fullUUID, err := thingsDB.ResolveTaskUUID(args[0])
	if err != nil {
		return err
	}

	task, err := thingsDB.GetTask(fullUUID)
	if err != nil {
		return err
	}

	formatter := shared.GetFormatter(cmd)
	return formatter.FormatTask(task)
}
