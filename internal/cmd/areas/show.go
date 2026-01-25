package areas

import (
	"github.com/spf13/cobra"
	"thingies/internal/cmd/shared"
	"thingies/internal/db"
)

var showIncludeCompleted bool

var showCmd = &cobra.Command{
	Use:   "show <uuid>",
	Short: "Show area details",
	Long:  `Show detailed information about a specific area including its projects and tasks.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runShow,
}

func init() {
	showCmd.Flags().BoolVar(&showIncludeCompleted, "include-completed", false, "Include completed projects and tasks")
}

func runShow(cmd *cobra.Command, args []string) error {
	thingsDB, err := db.Open(shared.GetDBPath(cmd))
	if err != nil {
		return err
	}
	defer thingsDB.Close()

	area, err := thingsDB.GetArea(args[0])
	if err != nil {
		return err
	}

	projects, err := thingsDB.GetAreaProjects(args[0], showIncludeCompleted)
	if err != nil {
		return err
	}

	tasks, err := thingsDB.GetAreaTasks(args[0], showIncludeCompleted)
	if err != nil {
		return err
	}

	formatter := shared.GetFormatter(cmd)
	return formatter.FormatArea(area, projects, tasks)
}
