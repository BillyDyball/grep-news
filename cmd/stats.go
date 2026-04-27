package cmd

import (
	"fmt"
	"os"

	"github.com/BillyDyball/grep-news/internal/db"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Print database summary statistics",
	RunE: func(cmd *cobra.Command, args []string) error {
		stats, err := database.GetStats()
		if err != nil {
			return fmt.Errorf("getting stats: %w", err)
		}

		fmt.Printf("Feeds:     %d\n", stats.FeedCount)
		fmt.Printf("Articles:  %d\n", stats.ArticleCount)
		fmt.Println()

		// DB file size
		dbFile := dbPath
		if dbFile == "" {
			dbFile = db.DefaultPath()
		}
		if info, err := os.Stat(dbFile); err == nil {
			fmt.Printf("Database:  %s (%s)\n", dbFile, formatBytes(info.Size()))
		}

		fmt.Println()
		fmt.Printf("Fetch history:\n")
		fmt.Printf("  Successes: %d\n", stats.FetchSuccesses)
		fmt.Printf("  Failures:  %d\n", stats.FetchFailures)
		if stats.LastFetchAt != nil {
			fmt.Printf("  Last run:  %s\n", stats.LastFetchAt.Format("2006-01-02 15:04:05"))
		} else {
			fmt.Printf("  Last run:  never\n")
		}

		// Article count per feed
		counts, err := database.GetArticleCountPerFeed()
		if err != nil {
			return err
		}

		if len(counts) > 0 {
			fmt.Println()
			fmt.Println("Articles per feed:")
			for _, c := range counts {
				title := c.FeedTitle
				if title == "" {
					title = c.FeedURL
				}
				fmt.Printf("  %4d  %s\n", c.ArticleCount, title)
			}
		}

		return nil
	},
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
