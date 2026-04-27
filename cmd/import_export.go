package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/BillyDyball/grep-news/internal/db"
	"github.com/BillyDyball/grep-news/internal/feed"
	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import [file]",
	Short: "Import feeds from an OPML file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f, err := os.Open(args[0])
		if err != nil {
			return fmt.Errorf("opening file: %w", err)
		}
		defer f.Close()

		outlines, err := feed.ParseOPML(f)
		if err != nil {
			return err
		}

		fmt.Printf("Found %d feeds in OPML file\n", len(outlines))

		var added, skipped, failed int
		for _, o := range outlines {
			exists, err := database.FeedExists(o.XMLURL)
			if err != nil {
				return err
			}
			if exists {
				fmt.Printf("  Skipping %s (already exists)\n", o.XMLURL)
				skipped++
				continue
			}

			// Validate the feed
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			result, err := feed.Fetch(ctx, o.XMLURL, "", "", 30*time.Second)
			cancel()

			if err != nil {
				fmt.Fprintf(os.Stderr, "  Error validating %s: %v\n", o.XMLURL, err)
				failed++
				continue
			}

			finalURL := result.FinalURL
			if finalURL == "" {
				finalURL = o.XMLURL
			}

			// Check again with final URL after redirects
			exists, _ = database.FeedExists(finalURL)
			if exists {
				fmt.Printf("  Skipping %s (already exists after redirect)\n", finalURL)
				skipped++
				continue
			}

			title := result.FeedTitle
			if title == "" {
				title = o.Text
			}

			if err := database.InsertFeed(&db.Feed{
				URL:     finalURL,
				Title:   title,
				SiteURL: result.SiteURL,
			}); err != nil {
				fmt.Fprintf(os.Stderr, "  Error storing %s: %v\n", finalURL, err)
				failed++
				continue
			}

			fmt.Printf("  Added %s (%s)\n", finalURL, title)
			added++
		}

		fmt.Printf("\nImport complete: %d added, %d skipped, %d failed\n", added, skipped, failed)
		return nil
	},
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export feeds as OPML to stdout",
	RunE: func(cmd *cobra.Command, args []string) error {
		feeds, err := database.ListFeeds()
		if err != nil {
			return fmt.Errorf("listing feeds: %w", err)
		}

		exportFeeds := make([]feed.ExportFeed, len(feeds))
		for i, f := range feeds {
			exportFeeds[i] = feed.ExportFeed{
				URL:     f.URL,
				Title:   f.Title,
				SiteURL: f.SiteURL,
			}
		}

		return feed.WriteOPML(os.Stdout, exportFeeds)
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(exportCmd)
}
