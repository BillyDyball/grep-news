package cmd

import (
	"database/sql"
	"fmt"

	"github.com/spf13/cobra"
)

var purgeFlag bool

var removeCmd = &cobra.Command{
	Use:   "remove [url]",
	Short: "Remove a feed subscription",
	Long:  "Feeds are removed from the SQLite database. Use --purge to also delete all articles associated with the feed.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		feedURL := args[0]

		if purgeFlag {
			if err := database.DeleteArticlesByFeed(feedURL); err != nil {
				return fmt.Errorf("deleting articles: %w", err)
			}
		}

		if err := database.DeleteFeed(feedURL); err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("feed %s not found", feedURL)
			}
			return fmt.Errorf("removing feed: %w", err)
		}

		if purgeFlag {
			fmt.Printf("Removed %s and all its articles\n", feedURL)
		} else {
			fmt.Printf("Removed %s\n", feedURL)
		}
		return nil
	},
}

func init() {
	removeCmd.Flags().BoolVar(&purgeFlag, "purge", false, "also delete all articles from this feed")
	rootCmd.AddCommand(removeCmd)
}
