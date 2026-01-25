package tasks

import (
	"fmt"

	"github.com/spf13/cobra"
	"thingies/internal/things"
)

var (
	updateTitle    string
	updateNotes    string
	updateWhen     string
	updateDeadline string
	updateTags     string
)

var updateCmd = &cobra.Command{
	Use:   "update <uuid>",
	Short: "Update a task",
	Long:  `Update a task's properties using AppleScript.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runUpdate,
}

func init() {
	updateCmd.Flags().StringVar(&updateTitle, "title", "", "New title")
	updateCmd.Flags().StringVar(&updateNotes, "notes", "", "New notes (replaces existing)")
	updateCmd.Flags().StringVar(&updateWhen, "when", "", "When to schedule (today, tomorrow, evening, anytime, someday, or YYYY-MM-DD)")
	updateCmd.Flags().StringVar(&updateDeadline, "deadline", "", "Due date (YYYY-MM-DD)")
	updateCmd.Flags().StringVar(&updateTags, "tags", "", "Tags (comma-separated, replaces existing)")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	params := things.TaskUpdateParams{
		UUID:     args[0],
		Name:     updateTitle,
		Notes:    updateNotes,
		When:     updateWhen,
		DueDate:  updateDeadline,
		TagNames: updateTags,
	}

	if err := things.UpdateTask(params); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	fmt.Printf("Updated task: %s\n", args[0])
	return nil
}
