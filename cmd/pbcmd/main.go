package main

// paperboy REPL

import (
	"bufio"
	"fmt"
	"github.com/fatih/color"
	"github.com/jwriopel/commands"
	"github.com/jwriopel/paperboy"
	"io"
	"os"
	"strings"
)

func printItem(item paperboy.Item) {
	green := color.New(color.FgGreen).SprintfFunc()
	yellow := color.New(color.FgCyan).SprintfFunc()

	fmt.Printf("[%s] %s - %s\n", item.SourceName, green(item.Title), yellow(item.Url))
}

func buildBot() *paperboy.Bot {
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
	return paperboy.NewBot(sources)
}

func main() {
	stopper := make(chan bool)
	bot := buildBot()

	commands.Add(startCommand(bot, stopper))
	commands.Add(stopCommand(bot, stopper))
	commands.Add(sourcesCommand(bot))
	commands.Add(statusCommand(bot))
	commands.Add(showCommand(bot))
	commands.Add(streamCommand(bot))
	commands.Add(searchCommand(bot))
	commands.Add(saveCommand(bot))
	commands.Add(loadCommand(bot))

	cmdReader := bufio.NewReader(os.Stdin)
	for {
		if bot.NPending() > 0 {
			fmt.Print("* ")
		}
		fmt.Print("pbcmd> ")
		commandLine, err := cmdReader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("bye")
				return
			}
			panic(err)
		}

		if commandLine == "" || commandLine == "\n" {
			continue
		}

		if strings.HasPrefix(commandLine, "exit ") {
			break
		}

		err = commands.Run(commandLine)
		if err != nil {
			fmt.Printf("error: %s\n", err.Error())
		}
	}
}
