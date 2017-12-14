package main

import (
	"bytes"
	"fmt"
	"github.com/google/logger"
	"github.com/jwriopel/commands"
	"github.com/jwriopel/paperboy"
	"golang.org/x/net/websocket"
	"io"
	"strings"
)

var sentItems map[string]paperboy.Item
var cmdMap map[string]func([]string) string

func writeItem(w io.Writer, item paperboy.Item) {
	fmt.Fprintf(w, "[%s] %s - %s\n", item.SourceName, item.Title, item.URL)
}

func main() {

	sources := []paperboy.Source{
		paperboy.Source{
			Name:        "HackerNews",
			URL:         "https://news.ycombinator.com",
			Selector:    ".storylink",
			ConvertFunc: paperboy.AnchorConverter,
		},
		paperboy.Source{
			Name:        "Reddit",
			URL:         "https://www.reddit.com",
			Selector:    "a.title",
			ConvertFunc: paperboy.RedditConverter,
		},
	}

	bot := paperboy.NewBot(sources)
	stopper := make(chan bool)
	cmdBuffer := new(bytes.Buffer)

	commands.Add(startCommand(bot, cmdBuffer, stopper))
	commands.Add(stopCommand(bot, cmdBuffer, stopper))
	commands.Add(statusCommand(bot, cmdBuffer))
	commands.Add(showCommand(bot, cmdBuffer))
	commands.Add(searchCommand(bot, cmdBuffer))

	wsurl, botId := paperboy.StartRTM()
	ws, err := websocket.Dial(wsurl, "", "https://api.slack.com")
	if err != nil {
		logger.Fatal(err)
	}

	for {
		m, err := paperboy.GetMessage(ws)
		if err != nil {
			if err != io.EOF {
				logger.Fatal(err)
			}
			continue
		}

		if m.Type == "message" && strings.HasPrefix(m.Text, "<@"+botId+">") {
			cmdLine := m.Text
			cmdParts := strings.Split(cmdLine, " ")
			cmdBuffer.Reset()
			err = commands.Run(strings.Join(cmdParts[1:], " "))
			if err != nil {
				m.Text = fmt.Sprintf("error running `%s`: %s\n", cmdLine, err)
				paperboy.PostMessage(ws, m)
				continue
			}
			m.Text = cmdBuffer.String()
			paperboy.PostMessage(ws, m)
		}
	}
}
