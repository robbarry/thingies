package projects

import (
	"github.com/spf13/cobra"
	"thingies/internal/cmd/shared"
	"thingies/internal/db"
)

var includeCompleted bool

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List projects",
	Long:  `List all projects with task counts.`,
	RunE:  runList,
}

func init() {
	listCmd.Flags().BoolVar(&includeCompleted, "include-completed", false, "Include completed projects")
}

func runList(cmd *cobra.Command, args []string) error {
	thingsDB, err := db.Open(shared.GetDBPath(cmd))
	if err != nil {
		return err
	}
	defer thingsDB.Close()

	projects, err := thingsDB.ListProjects(includeCompleted)
	if err != nil {
		return err
	}

	formatter := shared.GetFormatter(cmd)
	return formatter.FormatProjects(projects)
}
