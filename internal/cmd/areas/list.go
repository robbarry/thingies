package areas

import (
	"github.com/spf13/cobra"
	"thingies/internal/cmd/shared"
	"thingies/internal/db"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List areas",
	Long:  `List all visible areas with task and project counts.`,
	RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
	thingsDB, err := db.Open(shared.GetDBPath(cmd))
	if err != nil {
		return err
	}
	defer thingsDB.Close()

	areas, err := thingsDB.ListAreas()
	if err != nil {
		return err
	}

	formatter := shared.GetFormatter(cmd)
	return formatter.FormatAreas(areas)
}
