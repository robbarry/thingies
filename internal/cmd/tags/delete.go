package tags

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
	Short: "Delete a tag",
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

	fullUUID, err := thingsDB.ResolveTagUUID(uuid)
	if err != nil {
		return err
	}

	// Get tag details for confirmation
	tag, err := thingsDB.GetTag(fullUUID)
	if err != nil {
		return err
	}

	if !forceDelete {
		fmt.Printf("Delete tag '%s'? This will remove the tag from all tasks. [y/N] ", tag.Title)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Canceled")
			return nil
		}
	}

	if err := things.DeleteTag(fullUUID); err != nil {
		return err
	}

	fmt.Printf("Deleted tag: %s\n", tag.Title)
	return nil
}
