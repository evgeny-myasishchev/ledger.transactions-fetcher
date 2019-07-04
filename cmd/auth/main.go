package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/dal"

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

	svcCfg := config.Load()

	diag.SetupLoggingSystem(func(setup diag.LoggingSystemSetup) {
		setup.SetLogLevel(svcCfg.StringParam(config.LogLevel).Value())
	})

	oauthClient := auth.NewGoogleOAuthClient(auth.WithClientSecrets(
		svcCfg.StringParam(config.GoogleClientID).Value(),
		svcCfg.StringParam(config.GoogleClientSecret).Value(),
	))

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

	authSvc := auth.NewService(
		auth.WithOAuthClient(oauthClient),
		auth.WithStorage(storage),
	)

	switch cliArgs.cmd {
	case "auth-url":
		codeGrantURL := oauthClient.BuildCodeGrantURL()
		fmt.Println("Paste url below to browser and follow instructions")
		fmt.Println("Then use exchange-code action")
		fmt.Println(codeGrantURL)
	case "register-user":
		if cliArgs.authorizationCode == "" {
			showHelpAndExit()
		}
		if err := authSvc.RegisterUser(ctx, cliArgs.authorizationCode); err != nil {
			panic(err)
		}
		logger.Info(ctx, "User registered")
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}
}
