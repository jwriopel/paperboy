package paperboy

import (
	"strings"
	"sync"
	"time"
)

type Bot struct {
	unreadItems   map[string]Item
	sentItems     map[string]Item
	PollFrequency time.Duration
	Sources       []Source
	mux           sync.Mutex
	running       bool
}

func NewBot(sources []Source) *Bot {
	return &Bot{
		unreadItems:   make(map[string]Item),
		sentItems:     make(map[string]Item),
		PollFrequency: time.Duration(10) * time.Second,
		Sources:       sources,
	}
}

// Start causes the Bot to start polling for news Items.
func (b *Bot) Start(stop chan bool) {

	getItems := func() {
		for item := range GetAll(b.Sources) {
			if _, seen := b.sentItems[item.Url]; !seen {
				b.mux.Lock()
				b.unreadItems[item.Url] = item
				b.mux.Unlock()
			}
		}

	}
	getItems()
	go func() {
		pollTimer := time.Tick(b.PollFrequency)
		for {
			select {
			case <-pollTimer:
				getItems()
			case <-stop:
				b.running = false
				return
			}
		}
	}()
	b.running = true
}

// CacheSize provides the number of items the Bot has in memory.
func (b *Bot) CacheSize() int {
	b.mux.Lock()
	size := len(b.sentItems)
	b.mux.Unlock()
	return size
}

// Flush will clear cached items from the Bot's memory.
func (b *Bot) Flush() {
	b.mux.Lock()
	b.sentItems = make(map[string]Item)
	b.mux.Unlock()
}

// GetUnread will return a list of unread items and flush the pending items
// cache.
func (b *Bot) Unread() []Item {
	b.mux.Lock()

	items := make([]Item, 0)
	for _, item := range b.unreadItems {
		b.sentItems[item.Url] = item
		items = append(items, item)
	}

	b.unreadItems = make(map[string]Item)
	b.mux.Unlock()

	return items
}

func (b *Bot) Search(sterm string) []Item {
	b.mux.Lock()

	matches := make([]Item, 0)
	for _, item := range b.sentItems {
		if strings.Contains(strings.ToLower(item.Title), strings.ToLower(sterm)) {
			matches = append(matches, item)
		}
	}
	b.mux.Unlock()
	return matches
}

func (b *Bot) NPending() int {
	b.mux.Lock()
	defer b.mux.Unlock()
	return len(b.unreadItems)
}

func (b *Bot) IsRunning() bool {
	return b.running
}
