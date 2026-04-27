package ui

import (
	"io"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// HTMLToText converts HTML content to readable plain text.
func HTMLToText(htmlContent string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return stripTags(htmlContent)
	}

	var sb strings.Builder
	renderNode(&sb, doc)
	text := sb.String()

	// Collapse multiple blank lines
	re := regexp.MustCompile(`\n{3,}`)
	text = re.ReplaceAllString(text, "\n\n")
	return strings.TrimSpace(text)
}

func renderNode(w io.Writer, n *html.Node) {
	if n.Type == html.TextNode {
		text := strings.TrimSpace(n.Data)
		if text != "" {
			io.WriteString(w, text)
		}
		return
	}

	if n.Type == html.ElementNode {
		switch n.Data {
		case "br":
			io.WriteString(w, "\n")
		case "p", "div", "article", "section", "blockquote":
			io.WriteString(w, "\n\n")
		case "h1", "h2", "h3", "h4", "h5", "h6":
			io.WriteString(w, "\n\n")
		case "li":
			io.WriteString(w, "\n  • ")
		case "a":
			// Render link text, then URL in parens
			var linkText strings.Builder
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				renderNode(&linkText, c)
			}
			lt := strings.TrimSpace(linkText.String())
			href := getAttr(n, "href")
			if lt != "" && href != "" && lt != href {
				io.WriteString(w, lt+" ("+href+")")
			} else if lt != "" {
				io.WriteString(w, lt)
			} else if href != "" {
				io.WriteString(w, href)
			}
			return
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		renderNode(w, c)
	}

	if n.Type == html.ElementNode {
		switch n.Data {
		case "p", "div", "article", "section", "blockquote", "h1", "h2", "h3", "h4", "h5", "h6":
			io.WriteString(w, "\n")
		}
	}
}

func getAttr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func stripTags(s string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(s, "")
}
