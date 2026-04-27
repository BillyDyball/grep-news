package cmd

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/BillyDyball/grep-news/internal/db"
	"github.com/BillyDyball/grep-news/internal/feed"
	"github.com/BillyDyball/grep-news/internal/ui"
	"github.com/spf13/cobra"
)

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch all feeds and store new articles",
	Long:  "Fetches all subscribed feeds concurrently and upserts articles into the database.",
	RunE: func(cmd *cobra.Command, args []string) error {
		feeds, err := database.ListFeeds()
		if err != nil {
			return fmt.Errorf("listing feeds: %w", err)
		}

		if len(feeds) == 0 {
			fmt.Println("No feeds subscribed. Use 'grep-news add <url>' to add a feed.")
			return nil
		}

		interactive := ui.IsTerminal()
		progress := ui.NewFetchProgress(os.Stdout, interactive)

		type fetchResult struct {
			feedURL  string
			result   *feed.FetchResult
			err      error
		}

		results := make(chan fetchResult, len(feeds))
		var wg sync.WaitGroup
		sem := make(chan struct{}, 5) // max 5 concurrent fetches

		for _, f := range feeds {
			wg.Add(1)
			go func(f db.Feed) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				progress.Send(ui.FetchEvent{Type: ui.EventFetching, FeedURL: f.URL})

				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				result, err := feed.Fetch(ctx, f.URL, f.ETag, f.LastModified, 30*time.Second)
				results <- fetchResult{feedURL: f.URL, result: result, err: err}
			}(f)
		}

		go func() {
			wg.Wait()
			close(results)
		}()

		var hasError bool
		for r := range results {
			if r.err != nil {
				hasError = true
				errMsg := r.err.Error()
				progress.Send(ui.FetchEvent{Type: ui.EventError, FeedURL: r.feedURL, Error: errMsg})

				category := feed.ErrorCategory(r.err)
				database.IncrementFeedErrorCount(r.feedURL)
				database.InsertFetchLog(&db.FetchLogEntry{
					FeedURL: r.feedURL,
					Status:  category,
					Error:   &errMsg,
				})
				continue
			}

			if r.result.NotModified {
				progress.Send(ui.FetchEvent{Type: ui.EventSkipped, FeedURL: r.feedURL})
				database.InsertFetchLog(&db.FetchLogEntry{
					FeedURL: r.feedURL,
					Status:  "success",
				})
				continue
			}

			// Upsert articles (serialized through single DB connection)
			newCount := 0
			for _, item := range r.result.Items {
				var author *string
				if item.Author != "" {
					author = &item.Author
				}

				a := &db.Article{
					FeedURL:     r.feedURL,
					GUID:        item.GUID,
					Link:        item.Link,
					Title:       item.Title,
					Author:      author,
					ContentHTML: item.ContentHTML,
					PublishedAt: item.PublishedAt,
				}

				isNew, err := database.UpsertArticle(a)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: error inserting article '%s': %v\n", item.Title, err)
					continue
				}
				if isNew {
					newCount++
				}
			}

			database.UpdateFeedAfterFetch(r.feedURL, r.result.ETag, r.result.LastModified)
			database.InsertFetchLog(&db.FetchLogEntry{
				FeedURL:  r.feedURL,
				Status:   "success",
				NewCount: newCount,
			})

			progress.Send(ui.FetchEvent{Type: ui.EventFetched, FeedURL: r.feedURL, NewCount: newCount})
		}

		if hasError {
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(fetchCmd)
}
