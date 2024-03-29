package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/dal"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/banks"
	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/banks/monoua"
	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/banks/pbanua2x"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/app"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/lib-core-golang/diag"
)

var logger = diag.CreateLogger()

var cliArgs struct {
	user            string
	ledgerAccountID string
	daysToFetch     int64
	bank            string
}

func showHelpAndExit() {
	flag.PrintDefaults()
	os.Exit(1)
}

func init() {
	flag.StringVar(&cliArgs.user, "user", "", "User to fetch transactions for (email)")
	flag.StringVar(&cliArgs.ledgerAccountID, "acc", "", "Ledger account ID to fetch for")
	flag.Int64Var(&cliArgs.daysToFetch, "days", 2, "Number of days to fetch transactions for")
	flag.StringVar(&cliArgs.bank, "bank", "", "Bank code to fetch transactions for")

	flag.Parse()
}

func main() {
	if cliArgs.user == "" || cliArgs.ledgerAccountID == "" || cliArgs.bank == "" {
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

	err = injector(func(fetcherConfig banks.FetcherConfig, storage dal.Storage) error {
		var fetcher banks.Fetcher
		var err error
		switch cliArgs.bank {
		case "pbanua2x":
			fetcher, err = pbanua2x.NewFetcher(ctx, cliArgs.user, fetcherConfig)
			if err != nil {
				return err
			}
		case "monoua":
			fetcher, err = monoua.NewFetcher(ctx, cliArgs.user, fetcherConfig)
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
			trxDto, err := trx.ToDTO()
			if err != nil {
				return err
			}
			exists, err := storage.PendingTransactionExist(ctx, trxDto.ID)
			if err != nil {
				return err
			}
			if exists {
				logger.Debug(ctx, "Ignoring previously fetched transaction: {id=%v; amount=%v}", trxDto.ID, trxDto.Amount)
				continue
			} else {
				logger.Debug(ctx, "Saving transaction: {id=%v; amount=%v}", trxDto.ID, trxDto.Amount)
				if err := storage.SavePendingTransaction(ctx, trxDto); err != nil {
					return err
				}
			}
		}
		logger.Info(ctx, "Processed %v transactions", len(transactions))
		return nil
	})

	if err != nil {
		logger.WithError(err).Error(ctx, "Failed to fetch transactions")
		os.Exit(1)
	}
}
