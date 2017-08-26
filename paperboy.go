package paperboy

// The paperboy package is can be used to get news Items from various news Sources.

import (
	"errors"
	"fmt"
	"github.com/andybalholm/cascadia"
	"github.com/google/logger"
	"golang.org/x/net/html"
	"net/http"
	"strings"
	"sync"
)

// Represents a news item.
type Item struct {
	Title      string
	Url        string
	SourceName string
}

// Source is a web site that paperboy will get news Items from.
type Source struct {
	Name        string
	Url         string
	Selector    string
	ConvertFunc func(matches []*html.Node) []Item
}

// attributeMap will build a map from the attributes defined in an
// html element.
func attributeMap(node *html.Node) (attrs map[string]string) {
	attrs = make(map[string]string)
	for _, attr := range node.Attr {
		attrs[strings.ToLower(attr.Key)] = strings.ToLower(attr.Val)
	}
	return
}

// AnchorConvereter converts matched anchor tag Nodes to a slice of Item.
func AnchorConverter(matches []*html.Node) []Item {
	items := make([]Item, 0)
	for _, match := range matches {
		attrs := attributeMap(match)
		item := Item{}
		item.Title = match.FirstChild.Data
		item.Url = attrs["href"]
		items = append(items, item)
	}
	return items
}

// GetItems will make the http request and run a CSS selector on the
// response's body, if the response code is 200.
func GetItems(source Source) ([]Item, error) {

	req, err := http.NewRequest("GET", source.Url, nil)
	if err != nil {
		panic(err)
	}
	// some sources will block based on User-Agent.
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.101 Safari/537.36")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Unexpected response code (%d) from: %s\n", resp.StatusCode, source.Url))
	}

	docNode, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	cssSelector := cascadia.MustCompile(source.Selector)
	items := source.ConvertFunc(cssSelector.MatchAll(docNode))

	return items, nil
}

// GetAll concurrently requests items from multiple sources.
func GetAll(sources []Source) (out chan Item) {
	var wg sync.WaitGroup

	sourceSink := func(source Source) {
		items, err := GetItems(source)
		if err != nil {
			logger.Errorf("Error getting items from %s: %s\n", source, err)
		}
		for _, item := range items {
			item.SourceName = source.Name
			out <- item
		}
		wg.Done()
	}

	wg.Add(len(sources))
	for _, source := range sources {
		go sourceSink(source)
	}

	go func() {
		wg.Wait()
		logger.Infof("Closing out chan.")
		close(out)
	}()

	return out
}
