package cmd

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var bookmarkRemove bool

var bookmarkCmd = &cobra.Command{
	Use:   "bookmark [article-id]",
	Short: "Bookmark or unbookmark an article",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid article ID: %s", args[0])
		}

		// Verify article exists
		if _, err := database.GetArticle(id); err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("article %d not found", id)
			}
			return fmt.Errorf("fetching article: %w", err)
		}

		if bookmarkRemove {
			if err := database.SetBookmark(id, false); err != nil {
				return fmt.Errorf("removing bookmark: %w", err)
			}
			fmt.Printf("Removed bookmark from article %d\n", id)
		} else {
			if err := database.SetBookmark(id, true); err != nil {
				return fmt.Errorf("adding bookmark: %w", err)
			}
			fmt.Printf("Bookmarked article %d\n", id)
		}
		return nil
	},
}

func init() {
	bookmarkCmd.Flags().BoolVar(&bookmarkRemove, "remove", false, "remove bookmark")
	rootCmd.AddCommand(bookmarkCmd)
}
