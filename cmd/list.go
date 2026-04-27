package cmd

import (
	"fmt"
	"os"

	"github.com/BillyDyball/grep-news/internal/db"
	"github.com/BillyDyball/grep-news/internal/ui"
	"github.com/spf13/cobra"
)

var (
	listSize       int
	listPage       int
	listSort       string
	listFeed       string
	listSince      string
	listSearch     string
	listUnread     bool
	listBookmarked bool
	listAsc        bool
	listFormat     string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List articles",
	Long:  "Lists articles from the database with filtering, sorting, and pagination.",
	RunE: func(cmd *cobra.Command, args []string) error {
		sortDir := "desc"
		if listAsc {
			sortDir = "asc"
		}

		params := db.ListArticlesParams{
			PageSize:   listSize,
			Page:       listPage,
			SortField:  listSort,
			SortDir:    sortDir,
			FeedURL:    listFeed,
			Since:      listSince,
			Search:     listSearch,
			Unread:     listUnread,
			Bookmarked: listBookmarked,
		}

		articles, err := database.ListArticles(params)
		if err != nil {
			return fmt.Errorf("listing articles: %w", err)
		}

		if len(articles) == 0 {
			fmt.Println("No articles found.")
			return nil
		}

		rows := make([]ui.ArticleRow, len(articles))
		for i, a := range articles {
			date := a.FetchedAt.Format("2006-01-02")
			if a.PublishedAt != nil {
				date = a.PublishedAt.Format("2006-01-02")
			}

			var author string
			if a.Author != nil {
				author = *a.Author
			}

			rows[i] = ui.ArticleRow{
				ID:          a.ID,
				Date:        date,
				Link:        a.Link,
				Title:       a.Title,
				Author:      author,
				FeedURL:     a.FeedURL,
				Unread:      a.ReadAt == nil,
				Bookmarked:  a.BookmarkedAt != nil,
				PublishedAt: a.PublishedAt,
			}
		}

		format := ui.ParseFormat(listFormat)
		return ui.PrintArticles(os.Stdout, rows, format)
	},
}

func init() {
	listCmd.Flags().IntVar(&listSize, "size", 10, "number of articles per page")
	listCmd.Flags().IntVar(&listPage, "page", 1, "page number")
	listCmd.Flags().StringVar(&listSort, "sort", "published_at", "sort by: published_at, title, author, feed")
	listCmd.Flags().StringVar(&listFeed, "feed", "", "filter by feed URL")
	listCmd.Flags().StringVar(&listSince, "since", "", "show articles published after date")
	listCmd.Flags().StringVar(&listSearch, "search", "", "search article titles")
	listCmd.Flags().BoolVar(&listUnread, "unread", false, "show only unread articles")
	listCmd.Flags().BoolVar(&listBookmarked, "bookmarked", false, "show only bookmarked articles")
	listCmd.Flags().BoolVar(&listAsc, "asc", false, "sort ascending")
	listCmd.Flags().StringVar(&listFormat, "format", "plain", "output format: plain, table, csv, json")

	// Keep compatibility aliases
	listCmd.Flags().Bool("table", false, "shorthand for --format table")
	listCmd.Flags().Bool("csv", false, "shorthand for --format csv")
	listCmd.Flags().Bool("json", false, "shorthand for --format json")
	listCmd.Flags().Bool("desc", false, "sort descending (default)")

	rootCmd.AddCommand(listCmd)
}
