package areas

import (
	"github.com/spf13/cobra"
)

// AreasCmd is the parent command for area operations
var AreasCmd = &cobra.Command{
	Use:     "areas",
	Aliases: []string{"area", "a"},
	Short:   "Manage areas",
	Long:    `List, show, create, update, and delete areas.`,
}

func init() {
	AreasCmd.AddCommand(listCmd)
	AreasCmd.AddCommand(showCmd)
	AreasCmd.AddCommand(createCmd)
	AreasCmd.AddCommand(updateCmd)
	AreasCmd.AddCommand(deleteCmd)
}
