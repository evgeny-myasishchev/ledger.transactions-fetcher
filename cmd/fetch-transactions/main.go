package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/banks"
	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/banks/pbanua2x"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/app"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/config"
	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/lib-core-golang/diag"
)

var logger = diag.CreateLogger()

var cliArgs struct {
	user          string
	bankAccountID string
}

func showHelpAndExit() {
	flag.PrintDefaults()
	os.Exit(1)
}

func init() {
	flag.StringVar(&cliArgs.user, "user", "", "User to fetch transactions for (email)")
	flag.StringVar(&cliArgs.bankAccountID, "bic", "", "Bank account ID to fetch for (e.g card number)")

	flag.Parse()
}

func main() {
	if cliArgs.user == "" || cliArgs.bankAccountID == "" {
		showHelpAndExit()
	}

	appCfg := config.LoadAppConfig()

	diag.SetupLoggingSystem(func(setup diag.LoggingSystemSetup) {
		setup.SetLogLevel(appCfg.Log.Level.Value())
	})

	injector := app.BootstrapServices(appCfg)

	ctx := context.Background()

	if err := injector(func(fetcherConfig banks.FetcherConfig) error {
		fetcher, err := pbanua2x.NewFetcher(ctx, cliArgs.user, fetcherConfig)
		if err != nil {
			return err
		}
		transactions, err := fetcher.Fetch(ctx, &banks.FetchParams{
			BankAccountID: cliArgs.bankAccountID,
			From:          time.Now().Add(-24 * time.Hour),
			To:            time.Now(),
		})
		if err != nil {
			return err
		}
		for _, trx := range transactions {
			fmt.Println(trx)
		}
		return nil
	}); err != nil {
		panic(err)
	}
}
