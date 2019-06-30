package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/oauth"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/config"
	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/lib-core-golang/diag"
)

var logger = diag.CreateLogger()

var cliArgs struct {
	cmd               string
	authorizationCode string
}

func init() {
	const (
		defaultGopher = "pocket"
		usage         = "the variety of gopher"
	)
	flag.StringVar(&cliArgs.cmd, "cmd", "", "Command to run. Available commands: auth-url, authorize-code")
	flag.StringVar(&cliArgs.authorizationCode, "code", "", "Authorization code obtained by following auth-url instruction")

	flag.Parse()
}

func showHelpAndExit() {
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	if cliArgs.cmd == "" {
		showHelpAndExit()
	}

	svcCfg := config.Load()

	diag.SetupLoggingSystem(func(setup diag.LoggingSystemSetup) {
		setup.SetLogLevel(svcCfg.StringParam(config.LogLevel).Value())
	})

	oauthClient := oauth.NewGoogleOAuth(oauth.WithClientSecrets(
		svcCfg.StringParam(config.GoogleClientID).Value(),
		svcCfg.StringParam(config.GoogleClientSecret).Value(),
	))

	switch cliArgs.cmd {
	case "auth-url":
		codeGrantURL := oauthClient.BuildCodeGrantURL()
		fmt.Println("Paste url below to browser and follow instructions")
		fmt.Println("Then use exchange-code action")
		fmt.Println(codeGrantURL)
	case "authorize-code":
		if cliArgs.authorizationCode == "" {
			showHelpAndExit()
		}
		accessToken, err := oauthClient.PerformAuthCodeExchangeFlow(
			context.Background(),
			cliArgs.authorizationCode)
		if err != nil {
			panic(err)
		}
		fmt.Println(accessToken)
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}
}
