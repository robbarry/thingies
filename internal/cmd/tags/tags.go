package tags

import (
	"github.com/spf13/cobra"
)

// TagsCmd is the parent command for tag operations
var TagsCmd = &cobra.Command{
	Use:     "tags",
	Aliases: []string{"tag"},
	Short:   "Manage tags",
	Long:    `List tags.`,
}

func init() {
	TagsCmd.AddCommand(listCmd)
}
