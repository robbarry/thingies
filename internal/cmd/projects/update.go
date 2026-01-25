package projects

import (
	"fmt"

	"github.com/spf13/cobra"
	"thingies/internal/things"
)

var (
	updateTitle    string
	updateNotes    string
	updateDeadline string
	updateTags     string
)

var updateCmd = &cobra.Command{
	Use:   "update <uuid>",
	Short: "Update a project",
	Long:  `Update a project's properties using AppleScript.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runUpdate,
}

func init() {
	updateCmd.Flags().StringVar(&updateTitle, "title", "", "New title")
	updateCmd.Flags().StringVar(&updateNotes, "notes", "", "New notes (replaces existing)")
	updateCmd.Flags().StringVar(&updateDeadline, "deadline", "", "Due date (YYYY-MM-DD)")
	updateCmd.Flags().StringVar(&updateTags, "tags", "", "Tags (comma-separated, replaces existing)")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	params := things.ProjectUpdateParams{
		UUID:     args[0],
		Name:     updateTitle,
		Notes:    updateNotes,
		DueDate:  updateDeadline,
		TagNames: updateTags,
	}

	if err := things.UpdateProject(params); err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	fmt.Printf("Updated project: %s\n", args[0])
	return nil
}
