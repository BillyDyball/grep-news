package ui

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"golang.org/x/term"
)

type Format string

const (
	FormatPlain Format = "plain"
	FormatTable Format = "table"
	FormatCSV   Format = "csv"
	FormatJSON  Format = "json"
)

func ParseFormat(s string) Format {
	switch strings.ToLower(s) {
	case "table":
		return FormatTable
	case "csv":
		return FormatCSV
	case "json":
		return FormatJSON
	default:
		return FormatPlain
	}
}

func IsTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

type ArticleRow struct {
	ID          int64      `json:"id"`
	Date        string     `json:"date"`
	Link        string     `json:"link"`
	Title       string     `json:"title"`
	Author      string     `json:"author,omitempty"`
	FeedURL     string     `json:"feed_url"`
	Unread      bool       `json:"unread"`
	Bookmarked  bool       `json:"bookmarked"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
}

func PrintArticles(w io.Writer, articles []ArticleRow, format Format) error {
	switch format {
	case FormatJSON:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(articles)
	case FormatCSV:
		cw := csv.NewWriter(w)
		cw.Write([]string{"ID", "Date", "Link", "Title", "Author", "Feed"})
		for _, a := range articles {
			cw.Write([]string{
				fmt.Sprintf("%d", a.ID),
				a.Date,
				a.Link,
				a.Title,
				a.Author,
				a.FeedURL,
			})
		}
		cw.Flush()
		return cw.Error()
	case FormatTable:
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintf(tw, "ID\tDATE\tTITLE\tLINK\n")
		for _, a := range articles {
			fmt.Fprintf(tw, "%d\t%s\t%s\t%s\n", a.ID, a.Date, truncate(a.Title, 60), a.Link)
		}
		return tw.Flush()
	default:
		for _, a := range articles {
			fmt.Fprintf(w, "%s  %s  %s\n", a.Date, a.Link, a.Title)
		}
		return nil
	}
}

type FeedRow struct {
	URL           string     `json:"url"`
	Title         string     `json:"title"`
	LastFetchedAt *time.Time `json:"last_fetched_at,omitempty"`
	ArticleCount  int        `json:"article_count"`
	ErrorCount    int        `json:"error_count"`
}

func PrintFeeds(w io.Writer, feeds []FeedRow, format Format) error {
	switch format {
	case FormatJSON:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(feeds)
	case FormatCSV:
		cw := csv.NewWriter(w)
		cw.Write([]string{"URL", "Title", "Last Fetched", "Articles", "Errors"})
		for _, f := range feeds {
			lastFetched := ""
			if f.LastFetchedAt != nil {
				lastFetched = f.LastFetchedAt.Format("2006-01-02 15:04")
			}
			cw.Write([]string{f.URL, f.Title, lastFetched, fmt.Sprintf("%d", f.ArticleCount), fmt.Sprintf("%d", f.ErrorCount)})
		}
		cw.Flush()
		return cw.Error()
	case FormatTable:
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintf(tw, "URL\tTITLE\tLAST FETCHED\tARTICLES\tERRORS\n")
		for _, f := range feeds {
			lastFetched := "never"
			if f.LastFetchedAt != nil {
				lastFetched = f.LastFetchedAt.Format("2006-01-02 15:04")
			}
			fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%d\n", f.URL, truncate(f.Title, 40), lastFetched, f.ArticleCount, f.ErrorCount)
		}
		return tw.Flush()
	default:
		for _, f := range feeds {
			lastFetched := "never"
			if f.LastFetchedAt != nil {
				lastFetched = f.LastFetchedAt.Format("2006-01-02 15:04")
			}
			fmt.Fprintf(w, "%s  %s  %s  %d  %d\n", f.URL, f.Title, lastFetched, f.ArticleCount, f.ErrorCount)
		}
		return nil
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
