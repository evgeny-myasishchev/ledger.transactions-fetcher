package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/app"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/auth"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/config"
	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/lib-core-golang/diag"
)

var logger = diag.CreateLogger()

var cliArgs struct {
	cmd               string
	authorizationCode string
}

func init() {
	flag.StringVar(&cliArgs.cmd, "cmd", "", "Command to run. Available commands: auth-url, register-user")
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

	ctx := context.Background()

	appCfg := config.LoadAppConfig()

	diag.SetupLoggingSystem(func(setup diag.LoggingSystemSetup) {
		setup.SetLogLevel(appCfg.Log.Level.Value())
	})

	injector := app.BootstrapServices(appCfg)

	switch cliArgs.cmd {
	case "auth-url":
		if err := injector(func(oauthClient auth.OAuthClient) {
			codeGrantURL := oauthClient.BuildCodeGrantURL()
			fmt.Println("Paste url below to browser and follow instructions")
			fmt.Println("Then use exchange-code action")
			fmt.Println(codeGrantURL)
		}); err != nil {
			panic(err)
		}
	case "register-user":
		if cliArgs.authorizationCode == "" {
			showHelpAndExit()
		}
		err := injector(func(authSvc auth.Service) error {
			return authSvc.RegisterUser(ctx, cliArgs.authorizationCode)
		})
		if err != nil {
			panic(err)
		}
		logger.Info(ctx, "User registered")
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}
}
