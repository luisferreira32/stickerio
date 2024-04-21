package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
)

const (
	getcmd    = "get"
	listcmd   = "list"
	createcmd = "create"
	deletecmd = "delete"
	infocmd   = "info"

	city              = "city"
	movement          = "movement"
	buildingqueueitem = "buildingqueueitem"
	unitqueueitem     = "unitqueueitem"
)

func genericHelpAndExit() {
	fmt.Printf(`The CLI %s will allow you to take over the Stickerio world! Godspeed.

To proeficiently control your empire keep the following commands in mind:
  get
  list
  create
  delete
  info

These can be done over the resources of your empire, namely: cities, buildings, units, and others.
In order to obtain a more detailed explanation of the power that is in your hands, try to ask for
help on specific commands.

Additionally, the following flags are always available:
`, os.Args[0])
	flag.PrintDefaults()
	os.Exit(0)
}

func getHelpAndExit() {
	fmt.Printf(`To get something you must specify what you want. One of the following is a valid object:
	city
	movement (or mov for short)
	buildingqueueitem (or bqi for short)
	unitqueueitem (or uqi for short)

You will also have to specify the city from which you want to query.
// TODO: more explanation on how you can query other cities limited info, and bqi/uqi/mov are list 
// with limit 1 unless specific id is given
`)
	os.Exit(0)
}

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)

	help := flag.Bool("h", false, "display the help command")
	flag.Parse()

	args := flag.Args()

	if *help || len(args) == 0 {
		genericHelpAndExit()
	}

	// TODO: make it check some sort of local login token cache
	// find out how normally CLIs store credentials

	switch args[0] {
	case getcmd:
		if *help || len(args[1:]) == 0 {
			getHelpAndExit()
		}
		processGet(ctx, args[1:])
	case listcmd:
	case createcmd:
	case deletecmd:
	case infocmd:
	default:
		fmt.Printf("Unexpected command %s. Check -h for more info.\n", args[0])
	}
}

func processGet(ctx context.Context, args []string) {
	switch args[0] {
	case city:
		// WIP...
	}
}
