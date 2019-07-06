package app

import (
	"database/sql"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/dal"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/auth"

	"go.uber.org/dig"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/config"
)

// Injector is a function that will inject desired services
// to a target function
type Injector func(function interface{}) error

// BootstrapServices setup di container with all app services
func BootstrapServices(appCfg config.AppConfig) Injector {
	c := dig.New()

	c.Provide(func() (*sql.DB, error) {
		return sql.Open(appCfg.Storage.Driver.Value(), appCfg.Storage.DSN.Value())
	})

	c.Provide(func(db *sql.DB) (dal.Storage, error) {
		return dal.NewSQLStorage(dal.WithSQLDb(db))
	})

	c.Provide(func() auth.OAuthClient {
		return auth.NewGoogleOAuthClient(auth.WithClientSecrets(
			appCfg.Google.ClientID.Value(),
			appCfg.Google.ClientID.Value(),
		))
	})

	c.Provide(func(storage dal.Storage, oauthClient auth.OAuthClient) auth.Service {
		return auth.NewService(
			auth.WithOAuthClient(oauthClient),
			auth.WithStorage(storage),
		)
	})

	return func(function interface{}) error {
		return c.Invoke(function)
	}
}
