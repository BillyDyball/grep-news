package cmd

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

var (
	pruneOlderThan string
	pruneDryRun    bool
)

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove old articles from the database",
	Long:  "Removes articles older than the specified duration. Bookmarked articles are never pruned.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if pruneOlderThan == "" {
			return fmt.Errorf("--older-than is required (e.g., 90d, 30d)")
		}

		duration, err := parseDuration(pruneOlderThan)
		if err != nil {
			return err
		}

		cutoff := time.Now().Add(-duration)

		count, err := database.PruneArticles(cutoff, pruneDryRun)
		if err != nil {
			return fmt.Errorf("pruning articles: %w", err)
		}

		if pruneDryRun {
			fmt.Printf("Would remove %d articles older than %s\n", count, pruneOlderThan)
		} else {
			fmt.Printf("Removed %d articles older than %s\n", count, pruneOlderThan)
		}
		return nil
	},
}

// parseDuration parses a duration like "90d", "30d", "2w"
func parseDuration(s string) (time.Duration, error) {
	re := regexp.MustCompile(`^(\d+)([dwmDWM])$`)
	matches := re.FindStringSubmatch(s)
	if matches == nil {
		return 0, fmt.Errorf("invalid duration '%s'. Use format like 90d (days), 2w (weeks), 3m (months)", s)
	}

	n, _ := strconv.Atoi(matches[1])
	switch matches[2] {
	case "d", "D":
		return time.Duration(n) * 24 * time.Hour, nil
	case "w", "W":
		return time.Duration(n) * 7 * 24 * time.Hour, nil
	case "m", "M":
		return time.Duration(n) * 30 * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("unknown unit '%s'", matches[2])
	}
}

func init() {
	pruneCmd.Flags().StringVar(&pruneOlderThan, "older-than", "", "remove articles older than duration (e.g., 90d)")
	pruneCmd.Flags().BoolVar(&pruneDryRun, "dry-run", false, "preview what would be removed")
	rootCmd.AddCommand(pruneCmd)
}
