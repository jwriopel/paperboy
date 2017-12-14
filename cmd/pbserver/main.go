package main

import (
	"encoding/json"
	"fmt"
	"github.com/jwriopel/paperboy"
	"log"
	"net/http"
	"os"
)

var bot *paperboy.Bot
var pollStop chan bool

type botStatus struct {
	Running     bool `json:"running"`
	ReadCount   int  `json:"readcount"`
	UnreadCount int  `json:"unreadCount"`
}

func currentStatus() botStatus {
	return botStatus{
		Running:     bot.IsRunning(),
		ReadCount:   bot.CacheSize(),
		UnreadCount: bot.NPending(),
	}
}

func writeStatus(w http.ResponseWriter) {
	enc, err := json.Marshal(currentStatus())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding status")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(enc)
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	writeStatus(w)
}

func startHandler(w http.ResponseWriter, r *http.Request) {
	if !bot.IsRunning() {
		go bot.Start(pollStop)
	}
}

func stopHandler(w http.ResponseWriter, r *http.Request) {
	if bot.IsRunning() {
		go func() {
			pollStop <- true
		}()
	}
}

func itemsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	bot.DumpAll(w)
}

func main() {

	sources := []paperboy.Source{
		paperboy.Source{
			Name:        "HackerNews",
			URL:         "https://news.ycombinator.com",
			Selector:    ".storylink",
			ConvertFunc: paperboy.AnchorConverter,
		},
	}

	bot = paperboy.NewBot(sources)
	pollStop = make(chan bool)

	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/status", statusHandler)
	http.HandleFunc("/start", startHandler)
	http.HandleFunc("/stop", stopHandler)
	http.HandleFunc("/items", itemsHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
