package tags

import (
	"fmt"

	"github.com/spf13/cobra"
	"thingies/internal/cmd/shared"
	"thingies/internal/db"
	"thingies/internal/things"
)

var parentTag string

var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new tag",
	Args:  cobra.ExactArgs(1),
	RunE:  runCreate,
}

func init() {
	createCmd.Flags().StringVar(&parentTag, "parent", "", "Parent tag UUID (for nested tags)")
}

func runCreate(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Resolve parent tag UUID if provided
	if parentTag != "" {
		thingsDB, err := db.Open(shared.GetDBPath(cmd))
		if err != nil {
			return err
		}
		defer thingsDB.Close()

		resolved, err := thingsDB.ResolveTagUUID(parentTag)
		if err != nil {
			return err
		}
		parentTag = resolved
	}

	uuid, err := things.CreateTag(name, parentTag)
	if err != nil {
		return err
	}

	fmt.Printf("Created tag: %s (%s)\n", name, uuid)
	return nil
}
