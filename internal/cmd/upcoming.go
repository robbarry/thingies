package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"thingies/internal/cmd/shared"
	"thingies/internal/db"
	"thingies/internal/output"
)

var upcomingCmd = &cobra.Command{
	Use:   "upcoming",
	Short: "Show upcoming tasks",
	Long:  `Show tasks scheduled for the future.`,
	RunE:  runUpcoming,
}

func runUpcoming(cmd *cobra.Command, args []string) error {
	thingsDB, err := db.Open(shared.GetDBPath(cmd))
	if err != nil {
		return err
	}
	defer thingsDB.Close()

	tasks, err := thingsDB.GetUpcomingTasks()
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
	return formatter.FormatUpcoming(tasks)
}
