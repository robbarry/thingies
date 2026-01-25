package tasks

import (
	"fmt"

	"github.com/spf13/cobra"
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
	if err := things.CompleteTask(args[0]); err != nil {
		return fmt.Errorf("failed to complete task: %w", err)
	}

	fmt.Printf("Completed task: %s\n", args[0])
	return nil
}
