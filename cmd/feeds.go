package cmd

import (
	"fmt"
	"os"

	"github.com/BillyDyball/grep-news/internal/ui"
	"github.com/spf13/cobra"
)

var feedsFormat string

var feedsCmd = &cobra.Command{
	Use:   "feeds",
	Short: "List subscribed feeds",
	Long:  "Lists all subscribed feeds with health stats: last successful fetch time, article count, and error count.",
	RunE: func(cmd *cobra.Command, args []string) error {
		feeds, err := database.ListFeedsWithStats()
		if err != nil {
			return fmt.Errorf("listing feeds: %w", err)
		}

		if len(feeds) == 0 {
			fmt.Println("No feeds subscribed. Use 'grep-news add <url>' to add a feed.")
			return nil
		}

		rows := make([]ui.FeedRow, len(feeds))
		for i, f := range feeds {
			rows[i] = ui.FeedRow{
				URL:           f.URL,
				Title:         f.Title,
				LastFetchedAt: f.LastFetchedAt,
				ArticleCount:  f.ArticleCount,
				ErrorCount:    f.ErrorCount,
			}
		}

		format := ui.ParseFormat(feedsFormat)
		return ui.PrintFeeds(os.Stdout, rows, format)
	},
}

func init() {
	feedsCmd.Flags().StringVar(&feedsFormat, "format", "plain", "output format: plain, table, csv, json")
	rootCmd.AddCommand(feedsCmd)
}
