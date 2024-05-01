package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
)

type (
	commandType       string
	resourceType      string
	resourceTypeShort string
)

const (
	getcmd    commandType = "get"
	listcmd   commandType = "list"
	createcmd commandType = "create"
	deletecmd commandType = "delete"
	infocmd   commandType = "info"

	city              resourceType = "city"
	movement          resourceType = "movement"
	buildingqueueitem resourceType = "buildingqueueitem"
	unitqueueitem     resourceType = "unitqueueitem"

	cityShort              resourceTypeShort = "cit"
	movementShort          resourceTypeShort = "mov"
	buildingqueueitemShort resourceTypeShort = "bqi"
	unitqueueitemShort     resourceTypeShort = "uqi"
)

var (
	validCmd = map[commandType]struct{}{
		getcmd:    {},
		listcmd:   {},
		createcmd: {},
		deletecmd: {},
		infocmd:   {},
	}
	validResources = map[resourceType]struct{}{
		city:              {},
		movement:          {},
		buildingqueueitem: {},
		unitqueueitem:     {},
	}
	fromShortResourceType = map[resourceTypeShort]resourceType{
		cityShort:              city,
		movementShort:          movement,
		buildingqueueitemShort: buildingqueueitem,
		unitqueueitemShort:     unitqueueitem,
	}
)

// TODO: generate this from reading the api.yaml and actually make it available
// as a package? Stickerio CLI is optional - the server is th one that rules.
// though new binaries if the API is changed should be used (as we're using
// typed structures to make the prints more reasonable).

var (
	urlFmtsFromResource = map[resourceType]string{
		city:              "/v1/cities",
		movement:          "/v1/movements",
		buildingqueueitem: "/v1/cities/%s/buildingqitems",
		unitqueueitem:     "/v1/cities/%s/unitqitems",
	}
	methodFromCmd = map[commandType]string{
		getcmd:    "GET",
		listcmd:   "GET",
		createcmd: "POST",
		deletecmd: "DELETE",
		infocmd:   "GET",
	}
	addIDToURLFromCmd = map[commandType]struct{}{
		getcmd:    {},
		deletecmd: {},
		infocmd:   {},
	}
	helpForCmd = map[commandType]func() error{
		getcmd:    getCmdHelp,
		listcmd:   noOpHelp, // TODO
		createcmd: noOpHelp, // TODO
		deletecmd: noOpHelp, // TODO
		infocmd:   noOpHelp, // TODO
		"":        noOpHelp, // TODO
	}
)

var (
	errHelpRequested  = fmt.Errorf("help requested")
	errInvalidCommand = fmt.Errorf("invalid command")
	errRequestIssue   = fmt.Errorf("issue with request")
)

func main() {
	defer func() {
		panicked := recover()
		if panicked == nil {
			os.Exit(0)
		}
		err, ok := panicked.(error)
		if !ok {
			fmt.Printf("An unxpected issue has happened.\n%v\n", panicked)
			os.Exit(1)
		}
		switch {
		case errors.Is(err, errHelpRequested):
		case errors.Is(err, errInvalidCommand):
			os.Exit(1)
		case errors.Is(err, errRequestIssue):
			os.Exit(1)
		default:
			fmt.Printf("An unxpected issue has happened.\n%v\n", panicked)
			os.Exit(1)
		}
		os.Exit(0)
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	err := process(ctx)
	if err != nil {
		panic(err)
	}
}

// TODO: make this documentation loadable from a config file that can come in various languages

func genericHelp() error {
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
	return errHelpRequested
}

func stickerioAPIInfo() error {
	fmt.Printf(`The CLI %s requires a known API URL to be able to reach the game servers.
That must be provided by you as either:
- An environment variable "STICKERIO_API";
- A variable in the configuration file "stickerio.env".
`, os.Args[0])
	return errInvalidCommand
}

func unexpectedCommand(v string) error {
	fmt.Printf("Unexpected command %s. Check -h for more info.\n", v)
	return errInvalidCommand
}

func unexpectedResource(v string) error {
	fmt.Printf("Unexpected resource %s. Check -h for more info.\n", v)
	return errInvalidCommand
}

func noOpHelp() error {
	fmt.Printf("This should not be here... Well... Good luck?\n")
	return errInvalidCommand
}

func apiResponseError(c int, v string) error {
	fmt.Printf(`The API does not like you, find anything wrong here?
	HTTP: %d
	Resp: %s
`, c, v)
	return errRequestIssue
}

func getCmdHelp() error {
	fmt.Printf(`To get something you must specify what you want. 
One of the following is a valid object (and their shortened version):
	city (cit)
	movement (mov)
	buildingqueueitem (bqi)
	unitqueueitem (uqi)

You will also have to either specify the movement ID or the city ID from which you want 
to query as well as the ID of the building queue item, or unit queue item. For example:

%s get mov mov-id-321
%s get bqi city-id-123 bqi-id-321
%s get city city-id-123

`, os.Args[0], os.Args[0], os.Args[0])
	return errHelpRequested
}

func process(ctx context.Context) error {
	helpFlag := flag.Bool("help", false, "display the help command")
	flag.Parse()
	help := *helpFlag

	debug, _ := strconv.ParseBool(os.Getenv("DEBUG"))

	// TODO: make it possible to pass create objects another way?
	// FIXME
	// body, err := io.ReadAll(os.Stdin)
	// if err != nil {
	// 	return err
	// }
	body := []byte{}

	// TODO: make the CLI save some local configurations in known places
	// so that we don't have to send it all by flags or env vars.
	// e.g., tell location of the env file

	args := flag.Args()

	if help || len(args) == 0 {
		return genericHelp()
	}

	apiURLEnv := os.Getenv("STICKERIO_API")
	apiURLFile, err := os.ReadFile("stickerio.env")
	if apiURLEnv == "" && (err != nil || len(apiURLFile) == 0) {
		return stickerioAPIInfo()
	}
	// TODO: actually read from the .env file
	apiURL := apiURLEnv

	// TODO: make it check some sort of local login token cache
	// find out how normally CLIs store credentials

	command := commandType(args[0])
	_, ok := validCmd[command]
	if !ok {
		return unexpectedCommand(args[0])
	}

	if help || len(args[1:]) == 0 {
		return helpForCmd[command]()
	}

	resourceShort := resourceTypeShort(args[1])
	resource := resourceType(args[1])
	if len(resourceShort) == 3 {
		resource, _ = fromShortResourceType[resourceShort]
	}
	_, ok = validResources[resource]
	if !ok {
		return unexpectedResource(string(resource))
	}
	urlFmt, ok := urlFmtsFromResource[resource]
	if !ok {
		return noOpHelp()
	}
	_, ok = addIDToURLFromCmd[command]
	if ok {
		urlFmt += "/%s"
	}
	if ok && len(args[2:]) < 1 {
		return helpForCmd[command]()
	}
	method, ok := methodFromCmd[command]
	if !ok {
		return noOpHelp()
	}

	// TODO: make the GET possible to do by city name
	// with an underlying list?

	arguments := make([]any, len(args[2:]))
	for i := 0; i < len(args[2:]); i++ {
		arguments[i] = url.PathEscape(args[2+i])
	}

	var httpBody io.Reader
	if len(body) == 0 {
		httpBody = http.NoBody
	} else {
		httpBody = bytes.NewReader(body)
	}

	completeUrl := fmt.Sprintf(apiURL+urlFmt, arguments...)
	if debug {
		fmt.Printf("%s: %s\nBody: %s\n", method, completeUrl, body)
	}
	req, err := http.NewRequestWithContext(ctx, method, completeUrl, httpBody)
	if err != nil {
		return err
	}

	// TODO: we might want to configure the client a different way
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode > 400 {
		return apiResponseError(res.StatusCode, string(resBody))
	}
	if method != "GET" {
		return nil
	}

	// TODO: utilize the typed structures to possibly organize the info better?
	// for now, just work based on the assumption we're working with JSON and can
	// do stuff
	fmt.Printf("%s\n", resBody)

	return nil
}
