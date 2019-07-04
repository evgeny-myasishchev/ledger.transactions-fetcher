package main

import (
	"context"
	"database/sql"
	"flag"
	"os"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/dal"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/config"
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

	svcCfg := config.Load()

	diag.SetupLoggingSystem(func(setup diag.LoggingSystemSetup) {
		setup.SetLogLevel(svcCfg.StringParam(config.LogLevel).Value())
	})

	driver := svcCfg.StringParam(config.StorageDriver).Value()
	dsn := svcCfg.StringParam(config.StorageDSN).Value()

	db, err := sql.Open(driver, dsn)
	if err != nil {
		panic(err)
	}

	storage, err := dal.NewSQLStorage(dal.WithSQLDb(db))
	if err != nil {
		panic(err)
	}

	switch cliArgs.cmd {
	case "setup":
		if err := storage.Setup(context.Background()); err != nil {
			panic(err)
		}
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}
}
