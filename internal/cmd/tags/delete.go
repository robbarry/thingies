package tags

import (
	"fmt"

	"github.com/spf13/cobra"
	"thingies/internal/cmd/shared"
	"thingies/internal/db"
	"thingies/internal/things"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <uuid>",
	Short: "Delete a tag",
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

	fullUUID, err := thingsDB.ResolveTagUUID(uuid)
	if err != nil {
		return err
	}

	tag, err := thingsDB.GetTag(fullUUID)
	if err != nil {
		return err
	}

	if err := things.DeleteTag(fullUUID); err != nil {
		return err
	}

	fmt.Printf("Deleted tag: %s\n", tag.Title)
	return nil
}
