package main

import (
	"fmt"
	"github.com/google/logger"
	"github.com/jwriopel/paperboy"
	"golang.org/x/net/websocket"
	"io"
	"strings"
)

var sentItems map[string]paperboy.Item
var cmdMap map[string]func([]string) string

func main() {
	sentItems = make(map[string]paperboy.Item)
	cmdMap := make(map[string]func([]string) string)

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

	wsurl, botId := paperboy.StartRTM()
	ws, err := websocket.Dial(wsurl, "", "https://api.slack.com")
	if err != nil {
		logger.Fatal(err)
	}

	cmdMap["help"] = func(args []string) string {
		knownCmds := make([]string, 0)
		for key, _ := range cmdMap {
			knownCmds = append(knownCmds, key)
		}
		return fmt.Sprintf("Available commands: %s\n", strings.Join(knownCmds, ", "))
	}

	cmdMap["status"] = func(args []string) string {
		return fmt.Sprintf("I have %d stories in memory.\n", len(sentItems))
	}

	cmdMap["search"] = func(args []string) string {
		var rmsg string
		if len(args) == 1 {
			return "Missing search term."
		}

		sterm := strings.ToLower(strings.Join(args[1:], " "))
		results := make([]paperboy.Item, 0)
		for _, item := range sentItems {
			if strings.Contains(strings.ToLower(item.Title), sterm) {
				results = append(results, item)
			}
		}

		if len(results) == 0 {
			return fmt.Sprintf("No results for \"%s\"\n", sterm)
		}

		for _, result := range results {
			rmsg += fmt.Sprintf("%s\n%s\n", result.Title, result.URL)
		}
		return rmsg
	}

	cmdMap["flush"] = func(args []string) string {
		msg := fmt.Sprintf("Flushing %d items.\n", len(sentItems))
		sentItems = make(map[string]paperboy.Item)
		return msg
	}

	sourceHandler := func(args []string) string {
		var sourceMsg string
		var source paperboy.Source
		for _, s := range sources {
			if strings.ToLower(args[0]) == strings.ToLower(s.Name) {
				source = s
				break
			}
		}

		if source.Name == "" {
			return fmt.Sprintf("Unknown source: %s\n", args[0])
		}

		items, err := paperboy.GetItems(source)
		if err != nil {
			return fmt.Sprintf("Error: %s\n", err)
		}

		for _, item := range items {
			if _, seen := sentItems[item.URL]; !seen {
				sourceMsg += fmt.Sprintf("%s\n%s\n", item.Title, item.URL)
				sentItems[item.URL] = item
			}
		}
		return sourceMsg
	}

	for _, source := range sources {
		if _, exists := cmdMap[strings.ToLower(source.Name)]; exists {
			logger.Infof("Can't overwrite command with source %s\n", source.Name)
			continue
		}
		cmdMap[strings.ToLower(source.Name)] = sourceHandler
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
			var messageText string

			cmdParts := strings.Split(m.Text, " ")
			if len(cmdParts) == 1 {
				m.Text = cmdMap["help"](nil)
				paperboy.PostMessage(ws, m)
				continue
			}

			cmdParts = cmdParts[1:]
			cmd, ok := cmdMap[strings.Trim(cmdParts[0], " \n")]
			if !ok {
				messageText = fmt.Sprintf("Unknown command: %s\n", cmdParts)
			} else {
				messageText = cmd(cmdParts)
			}
			m.Text = messageText
			paperboy.PostMessage(ws, m)
		}
	}
}
