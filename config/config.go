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

	StorageDriver = localParams.NewParam("storage/driver").String()
	StorageDSN    = localParams.NewParam("storage/data-source-name").String()

	FetcherConfigConfigDir = localParams.NewParam("fetcher-config/config-dir").String()

	LedgerAPI = localParams.NewParam("ledger/api").String()
)

// Log represents logger specific options
type Log struct {
	Level config.StringVal
}

// Google represents google settings
type Google struct {
	ClientID     config.StringVal
	ClientSecret config.StringVal
}

// Storage represents storage settings
type Storage struct {
	Driver config.StringVal
	DSN    config.StringVal
}

// FetcherConfig represents settings of a fetcher-config service
type FetcherConfig struct {
	ConfigDir config.StringVal
}

// Ledger ledger config
type Ledger struct {
	API config.StringVal
}

// AppConfig is a toplevel config structure
type AppConfig struct {
	Log           Log
	Google        Google
	Storage       Storage
	FetcherConfig FetcherConfig
	Ledger        Ledger
}

// Load will load and initialize config
func Load() config.ServiceConfig {
	cfg, err := configBuilder.LoadConfig()
	if err != nil {
		panic(err)
	}
	return cfg
}

// LoadAppConfig will load and initialize app config structure
func LoadAppConfig() *AppConfig {
	cfg, err := configBuilder.LoadConfig()
	if err != nil {
		panic(err)
	}

	appCfg := AppConfig{
		Log: Log{
			Level: cfg.StringParam(LogLevel),
		},
		Storage: Storage{
			Driver: cfg.StringParam(StorageDriver),
			DSN:    cfg.StringParam(StorageDSN),
		},
		Google: Google{
			ClientID:     cfg.StringParam(GoogleClientID),
			ClientSecret: cfg.StringParam(GoogleClientSecret),
		},
		FetcherConfig: FetcherConfig{
			ConfigDir: cfg.StringParam(FetcherConfigConfigDir),
		},
		Ledger: Ledger{
			API: cfg.StringParam(LedgerAPI),
		},
	}

	return &appCfg
}
