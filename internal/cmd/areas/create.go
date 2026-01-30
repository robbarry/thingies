package areas

import (
	"fmt"

	"github.com/spf13/cobra"
	"thingies/internal/things"
)

var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new area",
	Args:  cobra.ExactArgs(1),
	RunE:  runCreate,
}

func runCreate(cmd *cobra.Command, args []string) error {
	name := args[0]

	uuid, err := things.CreateArea(name)
	if err != nil {
		return err
	}

	fmt.Printf("Created area: %s (%s)\n", name, uuid)
	return nil
}
