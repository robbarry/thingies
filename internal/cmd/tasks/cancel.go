package tasks

import (
	"fmt"

	"github.com/spf13/cobra"
	"thingies/internal/cmd/shared"
	"thingies/internal/db"
	"thingies/internal/things"
)

var cancelCmd = &cobra.Command{
	Use:   "cancel <uuid>",
	Short: "Cancel a task",
	Long:  `Mark a task as canceled using AppleScript.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runCancel,
}

func runCancel(cmd *cobra.Command, args []string) error {
	thingsDB, err := db.Open(shared.GetDBPath(cmd))
	if err != nil {
		return err
	}
	defer thingsDB.Close()

	uuid, err := thingsDB.ResolveTaskUUID(args[0])
	if err != nil {
		return err
	}

	if err := things.CancelTask(uuid); err != nil {
		return fmt.Errorf("failed to cancel task: %w", err)
	}

	fmt.Printf("Canceled task: %s\n", uuid)
	return nil
}
