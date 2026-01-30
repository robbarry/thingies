package areas

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"thingies/internal/cmd/shared"
	"thingies/internal/db"
	"thingies/internal/things"
)

var forceDelete bool

var deleteCmd = &cobra.Command{
	Use:   "delete <uuid>",
	Short: "Delete an area",
	Args:  cobra.ExactArgs(1),
	RunE:  runDelete,
}

func init() {
	deleteCmd.Flags().BoolVarP(&forceDelete, "force", "f", false, "Skip confirmation")
}

func runDelete(cmd *cobra.Command, args []string) error {
	uuid := args[0]

	// Resolve short UUID if needed
	thingsDB, err := db.Open(shared.GetDBPath(cmd))
	if err != nil {
		return err
	}
	defer thingsDB.Close()

	fullUUID, err := thingsDB.ResolveAreaUUID(uuid)
	if err != nil {
		return err
	}

	// Get area details for confirmation
	area, err := thingsDB.GetArea(fullUUID)
	if err != nil {
		return err
	}

	if !forceDelete {
		fmt.Printf("Delete area '%s'? This will NOT delete tasks/projects in the area. [y/N] ", area.Title)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Canceled")
			return nil
		}
	}

	if err := things.DeleteArea(fullUUID); err != nil {
		return err
	}

	fmt.Printf("Deleted area: %s\n", area.Title)
	return nil
}
