package projects

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"thingies/internal/things"
)

var (
	createNotes    string
	createWhen     string
	createDeadline string
	createTags     string
	createArea     string
	createToDos    string
)

var createCmd = &cobra.Command{
	Use:   "create <title>",
	Short: "Create a new project",
	Long:  `Create a new project using the Things URL scheme.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runCreate,
}

func init() {
	createCmd.Flags().StringVar(&createNotes, "notes", "", "Project notes")
	createCmd.Flags().StringVar(&createWhen, "when", "", "When to schedule")
	createCmd.Flags().StringVar(&createDeadline, "deadline", "", "Due date")
	createCmd.Flags().StringVar(&createTags, "tags", "", "Comma-separated tags")
	createCmd.Flags().StringVar(&createArea, "area", "", "Area name")
	createCmd.Flags().StringVar(&createToDos, "todos", "", "Newline-separated task titles")
}

func runCreate(cmd *cobra.Command, args []string) error {
	var todos []string
	if createToDos != "" {
		todos = strings.Split(createToDos, "\n")
	}

	params := things.AddProjectParams{
		Title:    args[0],
		Notes:    createNotes,
		When:     createWhen,
		Deadline: createDeadline,
		Tags:     createTags,
		Area:     createArea,
		ToDos:    todos,
	}

	url := things.BuildAddProjectURL(params)

	if err := things.OpenURL(url); err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	fmt.Printf("Created project: %s\n", args[0])
	return nil
}
