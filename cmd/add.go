package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/BillyDyball/grep-news/internal/db"
	"github.com/BillyDyball/grep-news/internal/feed"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [urls...]",
	Short: "Add RSS/Atom feed subscriptions",
	Long:  "Feeds are validated and added to a local SQLite database. Multiple URLs can be added at once.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var hasError bool
		for _, rawURL := range args {
			if err := addFeed(rawURL); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error adding %s: %v\n", rawURL, err)
				hasError = true
			}
		}
		if hasError {
			return fmt.Errorf("some feeds could not be added")
		}
		return nil
	},
}

func addFeed(rawURL string) error {
	// Fetch to validate the feed is reachable and parseable
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := feed.Fetch(ctx, rawURL, "", "", 30*time.Second)
	if err != nil {
		return err
	}

	// Use the final URL after redirects
	finalURL := result.FinalURL
	if finalURL == "" {
		finalURL = rawURL
	}

	exists, err := database.FeedExists(finalURL)
	if err != nil {
		return err
	}
	if exists {
		fmt.Printf("Warning: feed %s already exists, skipping\n", finalURL)
		return nil
	}

	f := &db.Feed{
		URL:     finalURL,
		Title:   result.FeedTitle,
		SiteURL: result.SiteURL,
	}

	if err := database.InsertFeed(f); err != nil {
		return fmt.Errorf("storing feed: %w", err)
	}

	fmt.Printf("Added %s (%s)\n", finalURL, result.FeedTitle)
	return nil
}

func init() {
	rootCmd.AddCommand(addCmd)
}
