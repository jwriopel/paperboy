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
		if !strings.HasPrefix(item.Url, "http") {
			item.Url = "https://reddit.com" + item.Url
		}
		cleanedItems = append(cleanedItems, item)
	}
	return cleanedItems
}
