package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/ledger"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/config"
	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/app"
	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/auth"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/lib-core-golang/diag"
)

var logger = diag.CreateLogger()

var cliArgs struct {
	cmd  string
	user string
}

func init() {
	flag.StringVar(&cliArgs.cmd, "cmd", "", "Command to run. Available commands: accounts")
	flag.StringVar(&cliArgs.user, "user", "", "Valid ledger user email")

	flag.Parse()
}

func showHelpAndExit() {
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	if cliArgs.user == "" || cliArgs.cmd == "" {
		showHelpAndExit()
	}

	appCfg := config.LoadAppConfig()

	diag.SetupLoggingSystem(func(setup diag.LoggingSystemSetup) {
		setup.SetLogLevel(appCfg.Log.Level.Value())
	})

	injector := app.BootstrapServices(appCfg)
	ctx := context.Background()

	switch cliArgs.cmd {
	case "accounts":
		if err := injector(func(authSvc auth.Service) error {
			idToken, err := authSvc.FetchAuthToken(ctx, cliArgs.user)
			if err != nil {
				return err
			}
			api, err := ledger.NewAPI(ctx, appCfg.Ledger.API.Value(), idToken)
			if err != nil {
				return err
			}
			accounts, err := api.ListAccounts(ctx)
			if err != nil {
				return err
			}
			for _, acc := range accounts {
				fmt.Println(acc)
			}
			return nil
		}); err != nil {
			logger.WithError(err).Error(ctx, "Failed to list accounts")
			os.Exit(1)
		}

	default:
		flag.PrintDefaults()
		os.Exit(1)
	}
}
