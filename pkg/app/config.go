package app

import (
	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/config"
	coreCfg "github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/lib-core-golang/config"
	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/version"
)

// LoadConfig will load and initialize config
func LoadConfig() (*config.Config, error) {
	appEnv := coreCfg.NewAppEnv(version.AppName)

	var cfg config.Config

	localSource := coreCfg.NewLocalSource(
		coreCfg.LocalOpts.WithAppEnv(appEnv),
		coreCfg.LocalOpts.WithIgnoreDefaultService(),
	)

	var remoteSource coreCfg.SourceFactory

	// When running locally using local source for remote params
	// Make sure to define sensible defaults
	if appEnv.Name == "dev" {
		remoteSource = coreCfg.NewLocalSource(
			coreCfg.LocalOpts.WithAppEnv(appEnv),
			coreCfg.LocalOpts.WithDir("config"),
		)
	} else {
		remoteSource = coreCfg.NewAWSSSMSource(coreCfg.AwsSSMOpts.WithAppEnv(appEnv))
	}

	if err := coreCfg.Bind(&cfg, appEnv,
		coreCfg.WithSource("local", localSource),
		coreCfg.WithSource("remote", remoteSource),
	); err != nil {
		return nil, err
	}
	return &cfg, nil
}
