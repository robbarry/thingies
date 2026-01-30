package cmd

import (
	"github.com/spf13/cobra"
	"thingies/internal/cmd/shared"
	"thingies/internal/db"
)

var somedayCmd = &cobra.Command{
	Use:   "someday",
	Short: "Show someday tasks",
	Long:  `Show tasks deferred to someday (no scheduled date).`,
	RunE:  runSomeday,
}

func runSomeday(cmd *cobra.Command, args []string) error {
	thingsDB, err := db.Open(shared.GetDBPath(cmd))
	if err != nil {
		return err
	}
	defer thingsDB.Close()

	tasks, err := thingsDB.GetSomedayTasks()
	if err != nil {
		return err
	}

	formatter := shared.GetFormatter(cmd)
	return formatter.FormatTasks(tasks)
}
