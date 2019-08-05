package main

import (
	"context"
	"flag"
	"os"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/dal"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/app"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/lib-core-golang/diag"
)

var logger = diag.CreateLogger()

var cliArgs struct {
	cmd string
}

func init() {
	flag.StringVar(&cliArgs.cmd, "cmd", "", "Command to run. Available commands: setup")

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

	appCfg, err := app.LoadConfig()
	if err != nil {
		logger.WithError(err).Error(ctx, "Failed to load app config")
		os.Exit(1)
	}

	diag.SetupLoggingSystem(func(setup diag.LoggingSystemSetup) {
		setup.SetLogLevel(appCfg.Log.Level)
	})

	injector := app.BootstrapServices(appCfg)

	switch cliArgs.cmd {
	case "setup":
		if err := injector(func(storage dal.Storage) error {
			return storage.Setup(ctx)
		}); err != nil {
			panic(err)
		}

	default:
		flag.PrintDefaults()
		os.Exit(1)
	}
}
