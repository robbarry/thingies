package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"thingies/internal/cmd/areas"
	"thingies/internal/cmd/projects"
	"thingies/internal/cmd/tags"
	"thingies/internal/cmd/tasks"
)

var (
	dbPath  string
	jsonOut bool
	noColor bool
	verbose bool
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "thingies",
	Short: "CLI for Things 3 task management",
	Long:  `Thingies provides command-line access to Things 3 for listing, creating, updating, and deleting tasks, projects, and more.`,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&dbPath, "db", "d", "", "Path to Things database (default: auto-detect)")
	rootCmd.PersistentFlags().BoolVarP(&jsonOut, "json", "j", false, "Output as JSON")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	// Add subcommands
	rootCmd.AddCommand(tasks.TasksCmd)
	rootCmd.AddCommand(projects.ProjectsCmd)
	rootCmd.AddCommand(areas.AreasCmd)
	rootCmd.AddCommand(tags.TagsCmd)
	rootCmd.AddCommand(todayCmd)
	rootCmd.AddCommand(inboxCmd)
	rootCmd.AddCommand(upcomingCmd)
	rootCmd.AddCommand(somedayCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(snapshotCmd)
	rootCmd.AddCommand(logbookCmd)
}

// GetDBPath returns the database path flag value
func GetDBPath() string {
	return dbPath
}

// IsJSON returns whether JSON output is enabled
func IsJSON() bool {
	return jsonOut
}

// IsNoColor returns whether color is disabled
func IsNoColor() bool {
	return noColor
}

// IsVerbose returns whether verbose output is enabled
func IsVerbose() bool {
	return verbose
}
