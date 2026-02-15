package tasks

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

var deleteForce bool

var deleteCmd = &cobra.Command{
	Use:   "delete <uuid>",
	Short: "Delete a task",
	Long:  `Delete (trash) a task using AppleScript.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runDelete,
}

func init() {
	deleteCmd.Flags().BoolVarP(&deleteForce, "force", "f", false, "Skip confirmation")
}

func runDelete(cmd *cobra.Command, args []string) error {
	thingsDB, err := db.Open(shared.GetDBPath(cmd))
	if err != nil {
		return err
	}
	defer thingsDB.Close()

	uuid, err := thingsDB.ResolveTaskUUID(args[0])
	if err != nil {
		return err
	}

	// Get task info for confirmation
	if !deleteForce {
		task, err := thingsDB.GetTask(uuid)
		if err != nil {
			return err
		}

		fmt.Printf("Delete task: %s\n", task.Title)
		fmt.Print("Are you sure? [y/N] ")

		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "y" && response != "yes" {
			fmt.Println("Canceled")
			return nil
		}
	}

	if err := things.DeleteTask(uuid); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	fmt.Printf("Deleted task: %s\n", uuid)
	return nil
}
