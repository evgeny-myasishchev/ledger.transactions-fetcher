package main

import (
	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/lib-core-golang/diag"
)

var logger = diag.CreateLogger()

func main() {
	svcCfg := config.Load()

	diag.SetupLoggingSystem(func(setup diag.LoggingSystemSetup) {
		setup.SetLogLevel(svcCfg.StringParam(config.LogLevel).Value())
	})
}
