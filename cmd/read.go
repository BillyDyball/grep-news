package cmd

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/BillyDyball/grep-news/internal/ui"
	"github.com/spf13/cobra"
)

var readCmd = &cobra.Command{
	Use:   "read [article-id]",
	Short: "Read an article in the terminal",
	Long:  "Renders an article's content as readable plain text. If no content is available, displays the link instead.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid article ID: %s", args[0])
		}

		article, err := database.GetArticle(id)
		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("article %d not found", id)
			}
			return fmt.Errorf("fetching article: %w", err)
		}

		// Mark as read
		database.MarkArticleRead(id)

		// Display header
		fmt.Printf("Title: %s\n", article.Title)
		if article.Author != nil {
			fmt.Printf("Author: %s\n", *article.Author)
		}
		if article.PublishedAt != nil {
			fmt.Printf("Date: %s\n", article.PublishedAt.Format("2006-01-02 15:04"))
		}
		fmt.Printf("Link: %s\n", article.Link)
		fmt.Println("---")

		if article.ContentHTML != "" {
			text := ui.HTMLToText(article.ContentHTML)
			fmt.Println(text)
		} else {
			fmt.Printf("\nNo content available. Visit: %s\n", article.Link)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(readCmd)
}
