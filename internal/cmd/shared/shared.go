package shared

import (
	"github.com/spf13/cobra"
	"thingies/internal/output"
)

// GetDBPath returns the database path from flags
func GetDBPath(cmd *cobra.Command) string {
	dbPath, _ := cmd.Flags().GetString("db")
	if dbPath == "" {
		// Check parent
		if cmd.Parent() != nil {
			dbPath, _ = cmd.Parent().Flags().GetString("db")
		}
	}
	// Try root
	if dbPath == "" {
		root := cmd.Root()
		dbPath, _ = root.PersistentFlags().GetString("db")
	}
	return dbPath
}

// IsJSON returns whether JSON output is requested
func IsJSON(cmd *cobra.Command) bool {
	jsonOut, _ := cmd.Flags().GetBool("json")
	if !jsonOut {
		if cmd.Parent() != nil {
			jsonOut, _ = cmd.Parent().Flags().GetBool("json")
		}
	}
	if !jsonOut {
		root := cmd.Root()
		jsonOut, _ = root.PersistentFlags().GetBool("json")
	}
	return jsonOut
}

// IsNoColor returns whether color is disabled
func IsNoColor(cmd *cobra.Command) bool {
	noColor, _ := cmd.Flags().GetBool("no-color")
	if !noColor {
		if cmd.Parent() != nil {
			noColor, _ = cmd.Parent().Flags().GetBool("no-color")
		}
	}
	if !noColor {
		root := cmd.Root()
		noColor, _ = root.PersistentFlags().GetBool("no-color")
	}
	return noColor
}

// GetFormatter returns the appropriate formatter based on flags
func GetFormatter(cmd *cobra.Command) output.Formatter {
	if IsJSON(cmd) {
		return output.NewJSONFormatter()
	}
	return output.NewTableFormatter(IsNoColor(cmd))
}
