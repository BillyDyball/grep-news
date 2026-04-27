package feed

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
)

type FetchResult struct {
	FeedTitle    string
	SiteURL      string
	Items        []Item
	ETag         string
	LastModified string
	NotModified  bool
	FinalURL     string
}

type Item struct {
	GUID        string
	Link        string
	Title       string
	Author      string
	ContentHTML string
	PublishedAt *time.Time
}

func Fetch(ctx context.Context, feedURL, etag, lastModified string, timeout time.Duration) (*FetchResult, error) {
	client := &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", "grep-news/1.0")
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}
	if lastModified != "" {
		req.Header.Set("If-Modified-Since", lastModified)
	}

	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, &TimeoutError{URL: feedURL, Err: err}
		}
		return nil, &NetworkError{URL: feedURL, Err: err}
	}
	defer resp.Body.Close()

	finalURL := resp.Request.URL.String()

	if resp.StatusCode == http.StatusNotModified {
		return &FetchResult{NotModified: true, FinalURL: finalURL}, nil
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &HTTPError{URL: feedURL, StatusCode: resp.StatusCode, Status: resp.Status}
	}

	parser := gofeed.NewParser()
	parsed, err := parser.Parse(resp.Body)
	if err != nil {
		return nil, &ParseError{URL: feedURL, Err: err}
	}

	result := &FetchResult{
		FeedTitle:    parsed.Title,
		SiteURL:      parsed.Link,
		ETag:         resp.Header.Get("ETag"),
		LastModified: resp.Header.Get("Last-Modified"),
		FinalURL:     finalURL,
	}

	for _, item := range parsed.Items {
		fi := Item{
			Title: strings.TrimSpace(item.Title),
		}

		if item.GUID != "" {
			fi.GUID = item.GUID
		}
		if item.Link != "" {
			fi.Link = CanonicalizeURL(item.Link)
		}
		if item.Author != nil {
			fi.Author = item.Author.Name
		}
		if item.Content != "" {
			fi.ContentHTML = item.Content
		} else if item.Description != "" {
			fi.ContentHTML = item.Description
		}
		if item.PublishedParsed != nil {
			t := item.PublishedParsed.UTC()
			fi.PublishedAt = &t
		} else if item.UpdatedParsed != nil {
			t := item.UpdatedParsed.UTC()
			fi.PublishedAt = &t
		}

		result.Items = append(result.Items, fi)
	}

	return result, nil
}

// CanonicalizeURL normalizes a URL by lowercasing the host, removing fragments,
// and stripping common tracking parameters.
func CanonicalizeURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	u.Host = strings.ToLower(u.Host)
	u.Fragment = ""

	// Strip common tracking parameters
	q := u.Query()
	trackingParams := []string{
		"utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content",
		"fbclid", "gclid", "mc_cid", "mc_eid",
	}
	for _, p := range trackingParams {
		q.Del(p)
	}

	// Sort query params for consistency
	keys := make([]string, 0, len(q))
	for k := range q {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	u.RawQuery = ""
	vals := url.Values{}
	for _, k := range keys {
		for _, v := range q[k] {
			vals.Add(k, v)
		}
	}
	u.RawQuery = vals.Encode()

	return u.String()
}

// Error types for categorization
type TimeoutError struct {
	URL string
	Err error
}

func (e *TimeoutError) Error() string { return fmt.Sprintf("timeout fetching %s: %v", e.URL, e.Err) }

type NetworkError struct {
	URL string
	Err error
}

func (e *NetworkError) Error() string { return fmt.Sprintf("network error fetching %s: %v", e.URL, e.Err) }

type HTTPError struct {
	URL        string
	StatusCode int
	Status     string
}

func (e *HTTPError) Error() string { return fmt.Sprintf("HTTP %s fetching %s", e.Status, e.URL) }

type ParseError struct {
	URL string
	Err error
}

func (e *ParseError) Error() string { return fmt.Sprintf("parse error for %s: %v", e.URL, e.Err) }

func ErrorCategory(err error) string {
	switch err.(type) {
	case *TimeoutError:
		return "timeout"
	case *HTTPError:
		return "http_error"
	case *ParseError:
		return "parse_error"
	default:
		return "timeout"
	}
}
