package main

import (
	"fmt"
	"joe/commands"
	"strings"
)

var toUpper bool

var echoCommand = &commands.Command{
	Name:  "echo",
	Short: "Echo non flags back to the user.",
}

func runEcho(c *commands.Command, args []string) {

	c.Flags.Parse(args)
	args = c.Flags.Args()
	toEcho := strings.Join(args, " ")
	if toUpper {
		toEcho = strings.ToUpper(toEcho)
	}

	fmt.Println(toEcho)
}

func init() {
	echoCommand.Flags.BoolVar(&toUpper, "U", false, "echo in upper-case letters.")
	echoCommand.Run = runEcho
	echoCommand.Usage = "echo [-U] string"
}
