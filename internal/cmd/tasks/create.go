package tasks

import (
	"fmt"

	"github.com/spf13/cobra"
	"thingies/internal/things"
)

var (
	createNotes     string
	createWhen      string
	createDeadline  string
	createTags      string
	createList      string
	createHeading   string
	createCompleted bool
	createCanceled  bool
)

var createCmd = &cobra.Command{
	Use:   "create <title>",
	Short: "Create a new task",
	Long:  `Create a new task using the Things URL scheme.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runCreate,
}

func init() {
	createCmd.Flags().StringVar(&createNotes, "notes", "", "Task notes")
	createCmd.Flags().StringVar(&createWhen, "when", "", "When to schedule (today, tomorrow, evening, someday, YYYY-MM-DD)")
	createCmd.Flags().StringVar(&createDeadline, "deadline", "", "Due date (YYYY-MM-DD)")
	createCmd.Flags().StringVar(&createTags, "tags", "", "Comma-separated tags")
	createCmd.Flags().StringVar(&createList, "list", "", "Project or area name")
	createCmd.Flags().StringVar(&createHeading, "heading", "", "Heading within project")
	createCmd.Flags().BoolVar(&createCompleted, "completed", false, "Mark as completed")
	createCmd.Flags().BoolVar(&createCanceled, "canceled", false, "Mark as canceled")
}

func runCreate(cmd *cobra.Command, args []string) error {
	params := things.AddParams{
		Title:     args[0],
		Notes:     createNotes,
		When:      createWhen,
		Deadline:  createDeadline,
		Tags:      createTags,
		List:      createList,
		Heading:   createHeading,
		Completed: createCompleted,
		Canceled:  createCanceled,
	}

	url := things.BuildAddURL(params)

	if err := things.OpenURL(url); err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	fmt.Printf("Created task: %s\n", args[0])
	return nil
}
