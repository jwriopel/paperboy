package main

import (
	"github.com/fatih/color"
	"github.com/jwriopel/paperboy"
	"log"
	"time"
)

func main() {
	sources := []paperboy.Source{
		paperboy.Source{
			Name:        "HackerNews",
			Url:         "https://news.ycombinator.com",
			Selector:    ".storylink",
			ConvertFunc: paperboy.AnchorConverter,
		},
		paperboy.Source{
			Name:        "Reddit",
			Url:         "https://www.reddit.com",
			Selector:    "a.title",
			ConvertFunc: paperboy.RedditConverter,
		},
	}

	colors := make(map[string]color.Attribute)
	for i, source := range sources {
		colors[source.Name] = color.Attribute((i + 1) + 30)
	}

	seenItems := make(map[string]paperboy.Item)
	ichan := make(chan paperboy.Item, 100)

	getItems := func() {
		for item := range paperboy.GetAll(sources) {
			if _, seen := seenItems[item.Url]; !seen {
				ichan <- item
				seenItems[item.Url] = item
			}
		}

	}

	go func() {
		pollTimer := time.Tick(time.Duration(120) * time.Second)
		for {
			select {
			case <-pollTimer:
				getItems()
			}
		}
	}()

	getItems()
	for {
		for item := range ichan {
			color.Set(colors[item.SourceName])
			log.Printf("%s - [%s] %s\n", item.Title, item.SourceName, item.Url)
			color.Unset()
		}
	}
}
