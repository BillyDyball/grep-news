package cmd

import (
	"fmt"
	"os"

	"github.com/BillyDyball/grep-news/internal/db"
	"github.com/spf13/cobra"
)

var (
	dbPath   string
	colorOpt string
	database *db.DB
)

var rootCmd = &cobra.Command{
	Use:   "grep-news",
	Short: "A terminal RSS/Atom feed reader",
	Long:  "grep-news is a terminal RSS/Atom feed reader that stores feeds and articles in a local SQLite database.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Commands that don't need the database
		switch cmd.Name() {
		case "completions", "man", "help":
			return nil
		}

		path := dbPath
		if path == "" {
			path = db.DefaultPath()
		}

		var err error
		database, err = db.Open(path)
		if err != nil {
			return fmt.Errorf("opening database: %w", err)
		}

		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if database != nil {
			database.Close()
		}
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "", "path to SQLite database (default: ~/.grep-news/grep-news.db)")
	rootCmd.PersistentFlags().StringVar(&colorOpt, "color", "auto", "color output: auto, always, never")
}

func exitWithError(msg string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
}
