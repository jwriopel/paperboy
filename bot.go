package paperboy

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"strings"
	"sync"
	"time"
)

// Bot is used to collect items, store items, and search for items.
type Bot struct {
	unreadItems   map[string]Item
	sentItems     map[string]Item
	PollFrequency time.Duration
	Sources       []Source
	mux           sync.Mutex
	running       bool
}

// NewBot creates a Bot instance with the default settings.
func NewBot(sources []Source) *Bot {
	return &Bot{
		unreadItems:   make(map[string]Item),
		sentItems:     make(map[string]Item),
		PollFrequency: time.Duration(10) * time.Second,
		Sources:       sources,
	}
}

// Start causes the Bot to start polling all sources for items.
func (b *Bot) Start(stop chan bool) {

	getItems := func() {
		for item := range GetAll(b.Sources) {
			if _, seen := b.sentItems[item.URL]; !seen {
				b.mux.Lock()
				b.unreadItems[item.URL] = item
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

// Unread will return a list of unread items and flush the pending items
// cache.
func (b *Bot) Unread() []Item {
	b.mux.Lock()

	items := make([]Item, 0)
	for _, item := range b.unreadItems {
		b.sentItems[item.URL] = item
		items = append(items, item)
	}

	b.unreadItems = make(map[string]Item)
	b.mux.Unlock()

	return items
}

// Search will look through the bots cache of read items for items with a
// Title that contain sterm.
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

// NPending returns the number of unread items.
func (b *Bot) NPending() int {
	b.mux.Lock()
	defer b.mux.Unlock()
	return len(b.unreadItems)
}

// IsRunning is used to determine if the bot has been started and in its
// polling loop.
func (b *Bot) IsRunning() bool {
	return b.running
}

// Dump all items to w encoded as json.
func (b *Bot) Dump(w io.Writer) error {
	b.mux.Lock()
	defer b.mux.Unlock()

	encodedItems, err := json.Marshal(b.sentItems)
	if err != nil {
		return err
	}

	_, err = w.Write(encodedItems)
	if err != nil {
		return err
	}
	return nil
}

// Load will import items from r into the bot's memory.
func (b *Bot) Load(r io.Reader) error {
	b.mux.Lock()
	defer b.mux.Unlock()

	ibytes, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	tmpMap := make(map[string]Item)
	err = json.Unmarshal(ibytes, &tmpMap)
	if err != nil {
		return err
	}

	for key, val := range tmpMap {
		b.sentItems[key] = val
	}
	return nil
}

func (b *Bot) DumpAll(w io.Writer) error {
	items := make([]Item, 0, 20)
	for _, item := range b.sentItems {
		items = append(items, item)
	}

	for _, item := range b.unreadItems {
		items = append(items, item)
	}

	encoded, err := json.Marshal(items)
	if err != nil {
		return err
	}

	w.Write(encoded)
	return nil
}
