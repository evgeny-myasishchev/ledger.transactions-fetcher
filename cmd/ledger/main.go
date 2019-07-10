package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/dal"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/ledger"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/config"
	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/app"
	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/auth"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/lib-core-golang/diag"
)

var logger = diag.CreateLogger()

var cliArgs struct {
	cmd       string
	user      string
	accountID string
}

func init() {
	flag.StringVar(&cliArgs.cmd, "cmd", "", "Command to run. Available commands: accounts, sync")
	flag.StringVar(&cliArgs.user, "user", "", "Valid ledger user email")
	flag.StringVar(&cliArgs.accountID, "account", "", "Valid ledger account, used for sync")

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
	case "sync":
		if cliArgs.accountID == "" {
			showHelpAndExit()
		}
		if err := injector(func(authSvc auth.Service, storage dal.Storage) error {
			notSyncedTrxs, err := storage.FindNotSyncedTransactions(ctx, cliArgs.accountID)
			if err != nil {
				return err
			}
			if len(notSyncedTrxs) == 0 {
				logger.Info(ctx, "No pending transactions to sync")
				return nil
			}

			logger.Info(ctx, "Got %v pending transactions to sync", len(notSyncedTrxs))

			idToken, err := authSvc.FetchAuthToken(ctx, cliArgs.user)
			if err != nil {
				return err
			}
			api, err := ledger.NewAPI(ctx, appCfg.Ledger.API.Value(), idToken)
			if err != nil {
				return err
			}

			for _, trx := range notSyncedTrxs {
				err := api.ReportPendingTransaction(ctx, ledger.PendingTransactionDTO{
					ID:        trx.ID,
					Amount:    trx.Amount,
					Date:      trx.Date,
					Comment:   trx.Comment,
					AccountID: trx.AccountID,
					TypeID:    trx.TypeID,
				})
				if err != nil {
					return errors.Wrapf(err, "Failed to report pending trx: %v", trx.ID)
				}
				syncedAt := time.Now().UTC()
				trx.SyncedAt = &syncedAt
				if err := storage.SavePendingTransaction(ctx, &trx); err != nil {
					return errors.Wrapf(err, "Failed to mark pending transaction '%v' as synced", trx.ID)
				}
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
