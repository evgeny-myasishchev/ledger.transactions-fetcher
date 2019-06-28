package config

import (
	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/lib-core-golang/config"
	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/version"
)

var appEnv = config.NewAppEnv(version.AppName)
var configBuilder = config.NewBuilder(appEnv)

var localParams = configBuilder.NewParamsBuilder(configBuilder.WithLocalSource())

// Do not change vars below at runtime
var (
	LogLevel = localParams.NewParam("log/logLevel").String()

	RuntimeEnv = localParams.NewParam("runtimeEnv").String()

	Pbanua2xAPI = localParams.NewParam("pbanua2x/api").String()

	GoogleClientID     = localParams.NewParam("google/client-id").String()
	GoogleClientSecret = localParams.NewParam("google/client-secret").String()
)

// Load will load and initialize config
func Load() config.ServiceConfig {
	cfg, err := configBuilder.LoadConfig()
	if err != nil {
		panic(err)
	}
	return cfg
}
