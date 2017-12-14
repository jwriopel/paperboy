package main

// Define the commands available.

import (
	"fmt"
	"github.com/jwriopel/commands"
	"github.com/jwriopel/paperboy"
	"io"
	"strings"
)

func startCommand(b *paperboy.Bot, w io.Writer, stopper chan bool) *commands.Command {
	return &commands.Command{
		Name:  "start",
		Short: "Start polling sources for items.",
		Usage: "start",
		Run: func(*commands.Command, []string) {
			if !b.IsRunning() {
				b.Start(stopper)
				fmt.Fprintln(w, "Bot started")
			}
		},
	}
}

func stopCommand(b *paperboy.Bot, w io.Writer, stopper chan bool) *commands.Command {
	return &commands.Command{
		Name:  "stop",
		Short: "Stop polling sources for items",
		Usage: "stop",
		Run: func(*commands.Command, []string) {
			stopper <- true
			fmt.Fprintf(w, "Bot stopped")
		},
	}
}

func statusCommand(b *paperboy.Bot, w io.Writer) *commands.Command {
	return &commands.Command{
		Name:  "status",
		Short: "Get the status of the bot.",
		Usage: "status",
		Run: func(*commands.Command, []string) {
			fmt.Fprintf(w, "%d sent items\n%d pending items.\n", b.CacheSize(), b.NPending())
			if b.IsRunning() {
				fmt.Fprintf(w, "Bot is running\n")
			}
		},
	}
}

func showCommand(b *paperboy.Bot, w io.Writer) *commands.Command {
	return &commands.Command{
		Name:  "show",
		Short: "Show unread items.",
		Usage: "show",
		Run: func(*commands.Command, []string) {
			for _, item := range b.Unread() {
				writeItem(w, item)
			}
		},
	}
}

func searchCommand(b *paperboy.Bot, w io.Writer) *commands.Command {
	c := &commands.Command{
		Name:  "search",
		Short: "Search for items in the cache.",
		Usage: "search <search term>",
	}
	c.Run = func(cmd *commands.Command, args []string) {
		c.Flags.Parse(args)
		args = c.Flags.Args()

		if len(args) == 0 {
			fmt.Fprintf(w, "usage: `%s`\n", c.Usage)
			return
		}

		sterm := strings.Join(args, " ")
		results := b.Search(sterm)
		fmt.Fprintf(w, "Showing results for `%s`\n", sterm)
		for _, result := range results {
			writeItem(w, result)
		}
	}
	return c
}
