package feed

import (
	"encoding/xml"
	"fmt"
	"io"
	"time"
)

type OPML struct {
	XMLName xml.Name `xml:"opml"`
	Version string   `xml:"version,attr"`
	Head    OPMLHead `xml:"head"`
	Body    OPMLBody `xml:"body"`
}

type OPMLHead struct {
	Title       string `xml:"title"`
	DateCreated string `xml:"dateCreated,omitempty"`
}

type OPMLBody struct {
	Outlines []OPMLOutline `xml:"outline"`
}

type OPMLOutline struct {
	Text    string        `xml:"text,attr"`
	Title   string        `xml:"title,attr,omitempty"`
	Type    string        `xml:"type,attr,omitempty"`
	XMLURL  string        `xml:"xmlUrl,attr,omitempty"`
	HTMLURL string        `xml:"htmlUrl,attr,omitempty"`
	Outlines []OPMLOutline `xml:"outline,omitempty"`
}

func ParseOPML(r io.Reader) ([]OPMLOutline, error) {
	var opml OPML
	if err := xml.NewDecoder(r).Decode(&opml); err != nil {
		return nil, fmt.Errorf("parsing OPML: %w", err)
	}

	var feeds []OPMLOutline
	collectFeeds(&feeds, opml.Body.Outlines)
	return feeds, nil
}

func collectFeeds(result *[]OPMLOutline, outlines []OPMLOutline) {
	for _, o := range outlines {
		if o.XMLURL != "" {
			*result = append(*result, o)
		}
		if len(o.Outlines) > 0 {
			collectFeeds(result, o.Outlines)
		}
	}
}

type ExportFeed struct {
	URL     string
	Title   string
	SiteURL string
}

func WriteOPML(w io.Writer, feeds []ExportFeed) error {
	opml := OPML{
		Version: "2.0",
		Head: OPMLHead{
			Title:       "grep-news subscriptions",
			DateCreated: time.Now().UTC().Format(time.RFC1123Z),
		},
	}

	for _, f := range feeds {
		opml.Body.Outlines = append(opml.Body.Outlines, OPMLOutline{
			Text:    f.Title,
			Title:   f.Title,
			Type:    "rss",
			XMLURL:  f.URL,
			HTMLURL: f.SiteURL,
		})
	}

	fmt.Fprint(w, xml.Header)
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	return enc.Encode(opml)
}
