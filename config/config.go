package config

import (
	"cmd/go/internal/version"
	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/lib-core-golang/config"
)

var appEnv = config.NewAppEnv(version.AppName)
var configBuilder = config.NewBuilder(appEnv)

var localParams = configBuilder.NewParamsBuilder(configBuilder.WithLocalSource())

// Do not change vars below at runtime
var (
	LogLevel = localParams.NewParam("log/logLevel").String()

	RuntimeEnv = localParams.NewParam("runtimeEnv").String()
)

// Load will load and initialize config
func Load() config.ServiceConfig {
	cfg, err := configBuilder.LoadConfig()
	if err != nil {
		panic(err)
	}
	return cfg
}
