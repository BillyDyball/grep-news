package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/BillyDyball/grep-news/internal/feed"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check health of the database and feeds",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Running health checks...")
		fmt.Println()

		// 1. Database integrity
		fmt.Print("Database integrity... ")
		result, err := database.IntegrityCheck()
		if err != nil {
			fmt.Println("ERROR:", err)
		} else if result == "ok" {
			fmt.Println("OK")
		} else {
			fmt.Println("ISSUES FOUND:", result)
		}

		// 2. Stale feeds (not fetched in 7 days)
		staleSince := time.Now().AddDate(0, 0, -7)
		staleFeeds, err := database.GetStaleFeedsSince(staleSince)
		if err != nil {
			return fmt.Errorf("checking stale feeds: %w", err)
		}

		if len(staleFeeds) > 0 {
			fmt.Printf("\nStale feeds (not fetched in 7+ days): %d\n", len(staleFeeds))
			for _, f := range staleFeeds {
				lastFetched := "never"
				if f.LastFetchedAt != nil {
					lastFetched = f.LastFetchedAt.Format("2006-01-02 15:04")
				}
				fmt.Printf("  %s (last fetched: %s)\n", f.URL, lastFetched)
			}
		} else {
			fmt.Println("\nNo stale feeds.")
		}

		// 3. Reachability check
		feeds, err := database.ListFeeds()
		if err != nil {
			return fmt.Errorf("listing feeds: %w", err)
		}

		fmt.Printf("\nChecking feed reachability (%d feeds)...\n", len(feeds))
		var unreachable int
		for _, f := range feeds {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			_, err := feed.Fetch(ctx, f.URL, "", "", 10*time.Second)
			cancel()

			if err != nil {
				fmt.Printf("  ✗ %s - %v\n", f.URL, err)
				unreachable++
			} else {
				fmt.Printf("  ✓ %s\n", f.URL)
			}
		}

		fmt.Println()
		if unreachable > 0 {
			fmt.Printf("Done. %d/%d feeds unreachable.\n", unreachable, len(feeds))
		} else {
			fmt.Println("Done. All feeds healthy.")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
