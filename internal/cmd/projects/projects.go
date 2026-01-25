package projects

import (
	"github.com/spf13/cobra"
)

// ProjectsCmd is the parent command for project operations
var ProjectsCmd = &cobra.Command{
	Use:     "projects",
	Aliases: []string{"project", "p"},
	Short:   "Manage projects",
	Long:    `List, show, create, update, complete, and delete projects.`,
}

func init() {
	ProjectsCmd.AddCommand(listCmd)
	ProjectsCmd.AddCommand(showCmd)
	ProjectsCmd.AddCommand(createCmd)
	ProjectsCmd.AddCommand(updateCmd)
	ProjectsCmd.AddCommand(completeCmd)
	ProjectsCmd.AddCommand(deleteCmd)
}
