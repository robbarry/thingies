package projects

import (
	"fmt"

	"github.com/spf13/cobra"
	"thingies/internal/cmd/shared"
	"thingies/internal/db"
	"thingies/internal/things"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <uuid>",
	Short: "Delete a project",
	Long:  `Delete (trash) a project using AppleScript.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runDelete,
}

func runDelete(cmd *cobra.Command, args []string) error {
	thingsDB, err := db.Open(shared.GetDBPath(cmd))
	if err != nil {
		return err
	}
	defer thingsDB.Close()

	uuid, err := thingsDB.ResolveProjectUUID(args[0])
	if err != nil {
		return err
	}

	if err := things.DeleteProject(uuid); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	fmt.Printf("Deleted project: %s\n", uuid)
	return nil
}
