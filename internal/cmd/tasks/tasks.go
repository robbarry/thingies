package tasks

import (
	"github.com/spf13/cobra"
)

// TasksCmd is the parent command for task operations
var TasksCmd = &cobra.Command{
	Use:     "tasks",
	Aliases: []string{"task", "t"},
	Short:   "Manage tasks",
	Long:    `List, show, create, update, complete, and delete tasks.`,
}

func init() {
	TasksCmd.AddCommand(listCmd)
	TasksCmd.AddCommand(showCmd)
	TasksCmd.AddCommand(createCmd)
	TasksCmd.AddCommand(updateCmd)
	TasksCmd.AddCommand(completeCmd)
	TasksCmd.AddCommand(cancelCmd)
	TasksCmd.AddCommand(deleteCmd)
}
