package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/config"
	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/lib-core-golang/diag"
)

var logger = diag.CreateLogger()

var cliArgs struct {
	cmd string
}

func init() {
	const (
		defaultGopher = "pocket"
		usage         = "the variety of gopher"
	)
	flag.StringVar(&cliArgs.cmd, "cmd", "", "Command to run. Available commands: auth-url")

	flag.Parse()
}

func printAuthUrl(googleClientID string) {
	authURL, err := url.Parse("https://accounts.google.com/o/oauth2/v2/auth")
	if err != nil {
		panic(err)
	}
	query := authURL.Query()
	query.Add("response_type", "code")
	query.Add("client_id", googleClientID)
	query.Add("redirect_uri", "urn:ietf:wg:oauth:2.0:oob")
	query.Add("scope", "email")
	query.Add("access_type", "offline")
	authURL.RawQuery = query.Encode()

	fmt.Println("Copy url below and paste it into the browser.")
	fmt.Println("Then follow the instruction and use rake get-access-token[code] with the code displayed")
	fmt.Println("The url:")
	fmt.Println(authURL)
}

func main() {
	fmt.Println(cliArgs)

	if cliArgs.cmd == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	svcCfg := config.Load()

	diag.SetupLoggingSystem(func(setup diag.LoggingSystemSetup) {
		setup.SetLogLevel(svcCfg.StringParam(config.LogLevel).Value())
	})

	switch cliArgs.cmd {
	case "auth-url":
		printAuthUrl(svcCfg.StringParam(config.GoogleClientID).Value())
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}
}
