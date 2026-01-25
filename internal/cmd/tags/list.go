package tags

import (
	"github.com/spf13/cobra"
	"thingies/internal/cmd/shared"
	"thingies/internal/db"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List tags",
	Long:  `List all tags with usage counts.`,
	RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
	thingsDB, err := db.Open(shared.GetDBPath(cmd))
	if err != nil {
		return err
	}
	defer thingsDB.Close()

	tags, err := thingsDB.ListTags()
	if err != nil {
		return err
	}

	formatter := shared.GetFormatter(cmd)
	return formatter.FormatTags(tags)
}
