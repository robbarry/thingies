package tags

import (
	"fmt"

	"github.com/spf13/cobra"
	"thingies/internal/cmd/shared"
	"thingies/internal/db"
	"thingies/internal/things"
)

var updateName string

var updateCmd = &cobra.Command{
	Use:   "update <uuid>",
	Short: "Update a tag",
	Args:  cobra.ExactArgs(1),
	RunE:  runUpdate,
}

func init() {
	updateCmd.Flags().StringVar(&updateName, "name", "", "New name for the tag")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	uuid := args[0]

	if updateName == "" {
		return fmt.Errorf("no update parameters provided; use --name")
	}

	// Resolve short UUID if needed
	thingsDB, err := db.Open(shared.GetDBPath(cmd))
	if err != nil {
		return err
	}
	defer thingsDB.Close()

	fullUUID, err := thingsDB.ResolveTagUUID(uuid)
	if err != nil {
		return err
	}

	if err := things.UpdateTag(fullUUID, updateName); err != nil {
		return err
	}

	fmt.Printf("Updated tag: %s\n", fullUUID)
	return nil
}
