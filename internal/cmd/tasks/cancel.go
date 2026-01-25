package tasks

import (
	"fmt"

	"github.com/spf13/cobra"
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
	uuid := args[0]

	if err := things.CancelTask(uuid); err != nil {
		return fmt.Errorf("failed to cancel task: %w", err)
	}

	fmt.Printf("Canceled task: %s\n", uuid)
	return nil
}
