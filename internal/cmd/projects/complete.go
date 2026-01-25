package projects

import (
	"fmt"

	"github.com/spf13/cobra"
	"thingies/internal/things"
)

var completeCmd = &cobra.Command{
	Use:   "complete <uuid>",
	Short: "Mark a project as complete",
	Long:  `Mark a project as complete using AppleScript.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runComplete,
}

func runComplete(cmd *cobra.Command, args []string) error {
	uuid := args[0]

	if err := things.CompleteProject(uuid); err != nil {
		return fmt.Errorf("failed to complete project: %w", err)
	}

	fmt.Printf("Completed project: %s\n", uuid)
	return nil
}
