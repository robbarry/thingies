package areas

import (
	"fmt"

	"github.com/spf13/cobra"
	"thingies/internal/cmd/shared"
	"thingies/internal/db"
	"thingies/internal/things"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <uuid>",
	Short: "Delete an area",
	Args:  cobra.ExactArgs(1),
	RunE:  runDelete,
}

func runDelete(cmd *cobra.Command, args []string) error {
	uuid := args[0]

	thingsDB, err := db.Open(shared.GetDBPath(cmd))
	if err != nil {
		return err
	}
	defer thingsDB.Close()

	fullUUID, err := thingsDB.ResolveAreaUUID(uuid)
	if err != nil {
		return err
	}

	area, err := thingsDB.GetArea(fullUUID)
	if err != nil {
		return err
	}

	if err := things.DeleteArea(fullUUID); err != nil {
		return err
	}

	fmt.Printf("Deleted area: %s\n", area.Title)
	return nil
}
