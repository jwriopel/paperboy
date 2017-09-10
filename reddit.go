package paperboy

import (
	"golang.org/x/net/html"
	"strings"
)

// RedditConverter converts matched nodes to an absolute URL.
func RedditConverter(matches []*html.Node) []Item {
	cleanedItems := make([]Item, 0)
	items := AnchorConverter(matches)
	for _, item := range items {
		if !strings.HasPrefix(item.URL, "http") {
			item.URL = "https://reddit.com" + item.URL
		}
		cleanedItems = append(cleanedItems, item)
	}
	return cleanedItems
}
