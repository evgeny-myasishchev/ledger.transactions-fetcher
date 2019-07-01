package storage

import (
	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/oauth"
)

// Storage is a persistance layer
type Storage interface {
	Setup() error
	GetAccessTokenByEmail(userEmail string) (*oauth.AccessToken, error)
	SaveAccessToken(token *oauth.AccessToken) error
}
