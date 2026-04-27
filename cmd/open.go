package cmd

import (
	"database/sql"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"

	"github.com/spf13/cobra"
)

var openCmd = &cobra.Command{
	Use:   "open [article-id]",
	Short: "Open an article in the default browser",
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

		if article.Link == "" {
			return fmt.Errorf("article %d has no link", id)
		}

		// Mark as read
		database.MarkArticleRead(id)

		return openBrowser(article.Link)
	},
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform %s", runtime.GOOS)
	}
	return cmd.Start()
}

func init() {
	rootCmd.AddCommand(openCmd)
}
