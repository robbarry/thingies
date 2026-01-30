package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"thingies/internal/cmd/shared"
	"thingies/internal/db"
	"thingies/internal/output"
)

var logbookLimit int

var logbookCmd = &cobra.Command{
	Use:   "logbook",
	Short: "Show completed tasks",
	Long:  `Show tasks that have been completed, ordered by completion date.`,
	RunE:  runLogbook,
}

func init() {
	logbookCmd.Flags().IntVarP(&logbookLimit, "limit", "n", 50, "Maximum number of tasks to show")
}

func runLogbook(cmd *cobra.Command, args []string) error {
	thingsDB, err := db.Open(shared.GetDBPath(cmd))
	if err != nil {
		return err
	}
	defer thingsDB.Close()

	tasks, err := thingsDB.GetLogbook(logbookLimit)
	if err != nil {
		return err
	}

	if shared.IsJSON(cmd) {
		jsonTasks := make([]interface{}, len(tasks))
		for i, t := range tasks {
			jsonTasks[i] = t.ToJSON()
		}
		data, err := json.MarshalIndent(jsonTasks, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	formatter := output.NewTableFormatter(shared.IsNoColor(cmd))
	return formatter.FormatTasks(tasks)
}
