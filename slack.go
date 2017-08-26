package paperboy

import (
	"encoding/json"
	"fmt"
	"github.com/google/logger"
	"golang.org/x/net/websocket"
	"io/ioutil"
	"net/http"
	"os"
	"sync/atomic"
)

type responseStatement struct {
	Ok    bool         `json:"ok"`
	Error string       `json:"error"`
	Url   string       `json:"url"`
	Self  responseSelf `json:"self"`
}

type responseSelf struct {
	Id string `json:"id"`
}

// StartRTM creates a session with Slack's Real Time Messaging API.
func StartRTM() (wsurl, id string) {
	token := os.Getenv("SLACK_API_TOKEN")
	url := fmt.Sprintf("https://slack.com/api/rtm.start?token=%s", token)

	client := &http.Client{}
	resp, err := client.Get(url)
	if err != nil {
		logger.Fatalf("Error starting RTM session: %s\n", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Fatalf("Error reading response from rmt.start: %s\n", err)
	}

	var respObj responseStatement
	err = json.Unmarshal(body, &respObj)
	if err != nil {
		panic(err)
	}

	if !respObj.Ok {
		logger.Fatalf("Slack error: %s", respObj.Error)
	}

	wsurl = respObj.Url
	id = respObj.Self.Id
	return
}

type Message struct {
	Id      uint64 `json:"id"`
	Type    string `json:"type"`
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

// GetMessage gets the next message from Slack.
func GetMessage(ws *websocket.Conn) (m Message, err error) {
	err = websocket.JSON.Receive(ws, &m)
	return
}

var counter uint64

// PostMessage sends a message back to Slack.
func PostMessage(ws *websocket.Conn, m Message) error {
	m.Id = atomic.AddUint64(&counter, 1)
	return websocket.JSON.Send(ws, m)
}
