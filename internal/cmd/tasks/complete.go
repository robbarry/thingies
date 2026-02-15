package tasks

import (
	"fmt"

	"github.com/spf13/cobra"
	"thingies/internal/cmd/shared"
	"thingies/internal/db"
	"thingies/internal/things"
)

var completeCmd = &cobra.Command{
	Use:   "complete <uuid>",
	Short: "Mark a task as complete",
	Long:  `Mark a task as complete using AppleScript.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runComplete,
}

func runComplete(cmd *cobra.Command, args []string) error {
	thingsDB, err := db.Open(shared.GetDBPath(cmd))
	if err != nil {
		return err
	}
	defer thingsDB.Close()

	uuid, err := thingsDB.ResolveTaskUUID(args[0])
	if err != nil {
		return err
	}

	if err := things.CompleteTask(uuid); err != nil {
		return fmt.Errorf("failed to complete task: %w", err)
	}

	fmt.Printf("Completed task: %s\n", uuid)
	return nil
}
