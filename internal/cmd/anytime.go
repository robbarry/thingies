package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"thingies/internal/cmd/shared"
	"thingies/internal/db"
	"thingies/internal/output"
)

var anytimeCmd = &cobra.Command{
	Use:   "anytime",
	Short: "Show anytime tasks",
	Long:  `Show tasks in the Anytime list (available but not scheduled).`,
	RunE:  runAnytime,
}

func runAnytime(cmd *cobra.Command, args []string) error {
	thingsDB, err := db.Open(shared.GetDBPath(cmd))
	if err != nil {
		return err
	}
	defer thingsDB.Close()

	tasks, err := thingsDB.GetAnytimeTasks()
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
