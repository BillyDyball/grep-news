package ui

import (
	"fmt"
	"io"
	"sync"
)

type FetchEventType int

const (
	EventFetching FetchEventType = iota
	EventFetched
	EventSkipped
	EventError
)

type FetchEvent struct {
	Type     FetchEventType
	FeedURL  string
	NewCount int
	Error    string
}

type FetchProgress struct {
	mu          sync.Mutex
	w           io.Writer
	interactive bool
	events      []FetchEvent
}

func NewFetchProgress(w io.Writer, interactive bool) *FetchProgress {
	return &FetchProgress{w: w, interactive: interactive}
}

func (fp *FetchProgress) Send(e FetchEvent) {
	fp.mu.Lock()
	defer fp.mu.Unlock()

	// Update existing event for same feed or append
	found := false
	for i, existing := range fp.events {
		if existing.FeedURL == e.FeedURL {
			fp.events[i] = e
			found = true
			break
		}
	}
	if !found {
		fp.events = append(fp.events, e)
	}

	if fp.interactive {
		fp.renderInteractive()
	} else {
		fp.renderPlain(e)
	}
}

func (fp *FetchProgress) renderPlain(e FetchEvent) {
	switch e.Type {
	case EventFetching:
		fmt.Fprintf(fp.w, "Fetching %s...\n", e.FeedURL)
	case EventFetched:
		fmt.Fprintf(fp.w, "Fetched %s (%d new articles)\n", e.FeedURL, e.NewCount)
	case EventSkipped:
		fmt.Fprintf(fp.w, "Skipped %s (not modified)\n", e.FeedURL)
	case EventError:
		fmt.Fprintf(fp.w, "Error %s - %s\n", e.FeedURL, e.Error)
	}
}

func (fp *FetchProgress) renderInteractive() {
	// Move cursor up to overwrite previous output
	if len(fp.events) > 1 {
		fmt.Fprintf(fp.w, "\033[%dA", len(fp.events)-1)
	}
	fmt.Fprintf(fp.w, "\r")

	for _, e := range fp.events {
		switch e.Type {
		case EventFetching:
			fmt.Fprintf(fp.w, "\033[K◌ [FETCHING]  %s\n", shortURL(e.FeedURL))
		case EventFetched:
			fmt.Fprintf(fp.w, "\033[K● [FETCHED]   %s  (%d new articles)\n", shortURL(e.FeedURL), e.NewCount)
		case EventSkipped:
			fmt.Fprintf(fp.w, "\033[K○ [SKIPPED]   %s  (not modified)\n", shortURL(e.FeedURL))
		case EventError:
			fmt.Fprintf(fp.w, "\033[K✗ [ERROR]     %s  - %s\n", shortURL(e.FeedURL), e.Error)
		}
	}
}

func shortURL(u string) string {
	// Strip scheme for display
	for _, prefix := range []string{"https://", "http://"} {
		if len(u) > len(prefix) && u[:len(prefix)] == prefix {
			u = u[len(prefix):]
			break
		}
	}
	if len(u) > 50 {
		u = u[:47] + "..."
	}
	return u
}
