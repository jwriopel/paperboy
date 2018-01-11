package main

// This file contains factory functions that build command.Commands
// that have access to the bot instance from main.go.

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jwriopel/commands"
	"github.com/jwriopel/paperboy"
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

			switch b.IsRunning() {
			case true:
				fmt.Print("Stopping...")
				stopper <- true
				fmt.Println("done.")
			case false:
				fmt.Println("bot isn't running.")
			}
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

func validSource(name string, b *paperboy.Bot) bool {
	for _, source := range b.Sources {
		if source.Name == name {
			return true
		}
	}
	return false
}

// streamCommand is a command that will stream items to the stdout, until
// the user presses any key.
func streamCommand(b *paperboy.Bot) *commands.Command {

	c := &commands.Command{
		Name:  "stream",
		Short: "Stream new items to stdout, like tail -f.",
		Usage: "stream",
	}

	var sourceFilter string
	c.Flags.StringVar(&sourceFilter, "source", "", "Only show items from a single source.")

	c.Run = func(c *commands.Command, args []string) {
		quit := make(chan bool)
		poller := time.Tick(time.Duration(10) * time.Second)
		reader := bufio.NewReader(os.Stdin)

		if sourceFilter != "" && !validSource(sourceFilter, b) {
			fmt.Fprintf(os.Stderr, "Invalid source: %s\n", sourceFilter)
			return
		}

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
				// reset
				sourceFilter = ""
				return
			case <-poller:
				for _, item := range b.Unread() {
					if sourceFilter == "" {
						printItem(item)
					}

					if sourceFilter != "" && item.SourceName == sourceFilter {
						printItem(item)
					}

				}
			}
		}

	}

	return c
}

func searchCommand(b *paperboy.Bot) *commands.Command {

	c := &commands.Command{
		Name:  "search",
		Short: "Search cache for items by title.",
		Usage: "search <search term>",
	}

	c.Run = func(cmd *commands.Command, args []string) {
		c.Flags.Parse(args)
		args = c.Flags.Args()

		if len(args) == 0 {
			c.Flags.Usage()
			return
		}

		sterm := strings.Join(args, " ")
		results := b.Search(sterm)
		for _, result := range results {
			printItem(result)
		}
	}
	return c
}

// saveCommand will create a commands.Command that is used to save the read
// items to a json file. This can be loaded and added to the Bot's memory.
func saveCommand(b *paperboy.Bot) *commands.Command {
	c := &commands.Command{
		Name:  "save",
		Short: "Save read items to a disk.",
		Usage: "save <path>",
	}

	c.Run = func(command *commands.Command, args []string) {
		if len(args) == 0 {
			c.Flags.Usage()
			return
		}

		dfile, err := os.Create(args[0])
		if err != nil {
			panic(err)
		}
		defer dfile.Close()

		err = b.Dump(dfile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
		}
	}

	return c
}

func loadCommand(b *paperboy.Bot) *commands.Command {

	c := &commands.Command{
		Name:  "load",
		Short: "Load items saved items from a file.",
		Usage: "load <path>",
	}

	c.Run = func(command *commands.Command, args []string) {
		if len(args) == 0 {
			c.Flags.Usage()
			return
		}

		ifile, err := os.Open(args[0])
		if err != nil {
			panic(err)
		}
		defer ifile.Close()

		err = b.Load(ifile)
		if err != nil {
			panic(err)
		}
	}

	return c
}
