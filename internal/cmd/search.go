package cmd

import (
	"github.com/spf13/cobra"
	"thingies/internal/cmd/shared"
	"thingies/internal/db"
)

var (
	searchInNotes      bool
	searchIncludeFuture bool
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for tasks",
	Long:  `Search for tasks by title (and optionally notes).`,
	Args:  cobra.ExactArgs(1),
	RunE:  runSearch,
}

func init() {
	searchCmd.Flags().BoolVar(&searchInNotes, "in-notes", false, "Search in notes as well")
	searchCmd.Flags().BoolVar(&searchIncludeFuture, "include-future", false, "Include future instances of repeating tasks")
}

func runSearch(cmd *cobra.Command, args []string) error {
	thingsDB, err := db.Open(shared.GetDBPath(cmd))
	if err != nil {
		return err
	}
	defer thingsDB.Close()

	tasks, err := thingsDB.Search(args[0], searchInNotes, searchIncludeFuture)
	if err != nil {
		return err
	}

	formatter := shared.GetFormatter(cmd)
	return formatter.FormatSearchResults(tasks, args[0])
}
