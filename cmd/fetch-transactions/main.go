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
	user            string
	ledgerAccountID string
	daysToFetch     int64
}

func showHelpAndExit() {
	flag.PrintDefaults()
	os.Exit(1)
}

func init() {
	flag.StringVar(&cliArgs.user, "user", "", "User to fetch transactions for (email)")
	flag.StringVar(&cliArgs.ledgerAccountID, "acc", "", "Ledger account ID to fetch for")
	flag.Int64Var(&cliArgs.daysToFetch, "days", 2, "Number of days to fetch transactions for")

	flag.Parse()
}

func main() {
	if cliArgs.user == "" || cliArgs.ledgerAccountID == "" {
		showHelpAndExit()
	}

	appCfg := config.LoadAppConfig()

	diag.SetupLoggingSystem(func(setup diag.LoggingSystemSetup) {
		setup.SetLogLevel(appCfg.Log.Level.Value())
	})

	injector := app.BootstrapServices(appCfg)

	ctx := context.Background()

	err := injector(func(fetcherConfig banks.FetcherConfig) error {
		fetcher, err := pbanua2x.NewFetcher(ctx, cliArgs.user, fetcherConfig)
		if err != nil {
			return err
		}
		to := time.Now()
		from := to.Add(time.Duration(-24*cliArgs.daysToFetch) * time.Hour)
		logger.Info(ctx, "Fetching transactions from %v to %v", from, to)
		transactions, err := fetcher.Fetch(ctx, &banks.FetchParams{
			LedgerAccountID: cliArgs.ledgerAccountID,
			From:            from,
			To:              to,
		})
		if err != nil {
			return err
		}
		for _, trx := range transactions {
			fmt.Println(trx)
		}
		return nil
	})

	if err != nil {
		logger.WithError(err).Info(ctx, "Failed to fetch transactions")
		os.Exit(1)
	}
}
