package projects

import (
	"github.com/spf13/cobra"
	"thingies/internal/cmd/shared"
	"thingies/internal/db"
)

var showIncludeCompleted bool

var showCmd = &cobra.Command{
	Use:   "show <name-or-uuid>",
	Short: "Show project details",
	Long:  `Show detailed information about a specific project including its tasks.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runShow,
}

func init() {
	showCmd.Flags().BoolVar(&showIncludeCompleted, "include-completed", false, "Include completed tasks")
}

func runShow(cmd *cobra.Command, args []string) error {
	thingsDB, err := db.Open(shared.GetDBPath(cmd))
	if err != nil {
		return err
	}
	defer thingsDB.Close()

	// Resolve name to UUID if needed
	uuid, err := thingsDB.ResolveProjectID(args[0])
	if err != nil {
		return err
	}

	project, err := thingsDB.GetProject(uuid)
	if err != nil {
		return err
	}

	tasks, err := thingsDB.GetProjectTasks(uuid, showIncludeCompleted)
	if err != nil {
		return err
	}

	formatter := shared.GetFormatter(cmd)
	return formatter.FormatProject(project, tasks)
}
