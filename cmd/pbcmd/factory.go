package main

// This file contains factory functions that build command.Commands
// that have access to the bot instance from main.go.

import (
	"bufio"
	"fmt"
	"github.com/jwriopel/commands"
	"github.com/jwriopel/paperboy"
	"os"
	"time"
)

func startCommand(b *paperboy.Bot, stopper chan bool) *commands.Command {
	return &commands.Command{
		Name:  "start",
		Short: "Start the paperboy Bot instance.",
		Usage: "start",
		Run: func(*commands.Command, []string) {
			if b.IsRunning() {
				fmt.Println("Bot already running.")
				return
			}
			b.Start(stopper)
			fmt.Println("Bot started.")
		},
	}
}

func stopCommand(b *paperboy.Bot, stopper chan bool) *commands.Command {
	return &commands.Command{
		Name:  "stop",
		Short: "Command the bot to stop polling.",
		Usage: "stop",
		Run: func(*commands.Command, []string) {
			fmt.Print("Stopping...")
			stopper <- true
			fmt.Println("done.")
		},
	}
}

func sourcesCommand(b *paperboy.Bot) (c *commands.Command) {
	c = &commands.Command{
		Name:  "sources",
		Short: "Print the list of sources.",
		Usage: "sources",
	}

	c.Run = func(*commands.Command, []string) {
		var slist string
		for _, source := range b.Sources {
			slist += fmt.Sprintf("%s\n", source.Name)
		}
		fmt.Println(slist)
	}
	return
}

func statusCommand(b *paperboy.Bot) (c *commands.Command) {

	c = &commands.Command{
		Name:  "status",
		Short: "Show how many items are cached (read) and unread.",
		Usage: "status",
		Run: func(*commands.Command, []string) {
			fmt.Printf("%d read items.\n", b.CacheSize())
			if b.NPending() > 0 {
				fmt.Printf("%d unread items.\n", b.NPending())
			}
			if b.IsRunning() {
				fmt.Println("Bot is running.")
			}
		},
	}
	return
}

func showCommand(b *paperboy.Bot) *commands.Command {
	return &commands.Command{
		Name:  "show",
		Short: "Show new items.",
		Usage: "show",
		Run: func(*commands.Command, []string) {
			for _, item := range b.Unread() {
				printItem(item)
			}
		},
	}
}

// streamCommand is a command that will stream items to the stdout, until
// the user presses any key.
func streamCommand(b *paperboy.Bot) *commands.Command {
	c := &commands.Command{
		Name:  "stream",
		Short: "Stream new items to stdout, like tail -f.",
		Usage: "stream",
	}

	c.Run = func(*commands.Command, []string) {
		quit := make(chan bool)
		poller := time.Tick(time.Duration(10) * time.Second)
		reader := bufio.NewReader(os.Stdin)

		go func() {
			if !b.IsRunning() {
				b.Start(quit)
			}
			fmt.Println("Press any key to stop streaming.")
			_, _ = reader.ReadString('\n')
			quit <- true
		}()

		for {
			select {
			case <-quit:
				close(quit)
				return
			case <-poller:
				for _, item := range b.Unread() {
					printItem(item)
				}
			}
		}

	}

	return c
}
