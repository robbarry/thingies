package tasks

import (
	"fmt"

	"github.com/spf13/cobra"
	"thingies/internal/cmd/shared"
	"thingies/internal/db"
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
	Long:  `Update a task's properties. Uses AppleScript for most updates; specific date scheduling (YYYY-MM-DD) uses the Things URL scheme.`,
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

	// Specific dates need an auth token for the URL scheme
	if things.IsSpecificDate(updateWhen) {
		thingsDB, err := db.Open(shared.GetDBPath(cmd))
		if err != nil {
			return fmt.Errorf("failed to open database for auth token: %w", err)
		}
		defer thingsDB.Close()

		token, err := thingsDB.GetAuthToken()
		if err != nil {
			return fmt.Errorf("failed to get auth token: %w", err)
		}
		params.AuthToken = token
	}

	if err := things.UpdateTask(params); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	fmt.Printf("Updated task: %s\n", args[0])
	return nil
}
