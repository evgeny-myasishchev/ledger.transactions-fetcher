package ledger

import (
	"context"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/lib-core-golang/request"

	"github.com/pkg/errors"
)

const (
	csrfTokenName     = "form_authenticity_token"
	csrfHeaderName    = "X-CSRF-Token"
	sessionCookieName = "_ledger_session_v1"
)

// API is an interface to communicate with ledger
type API interface {
	ListAccounts(ctx context.Context) ([]AccountDTO, error)
}

type api struct {
	baseURL   string
	session   string
	csrfToken string
}

func (a *api) ListAccounts(ctx context.Context) ([]AccountDTO, error) {
	req := request.Get(a.baseURL+"/accounts").
		WithHeader("Cookie", sessionCookieName+"="+a.session)
	res := request.Do(ctx, req)
	var accounts []AccountDTO
	if err := res.DecodeJSON(&accounts); err != nil {
		return nil, errors.Wrap(err, "Failed to fetch accounts")
	}
	return accounts, nil
}

// Factory is a function that creates ledger API instance for given idToken
type Factory func(ctx context.Context, baseURL string, idToken string) (API, error)

// NewAPI returns an instance of a new API initialized with given token
func NewAPI(ctx context.Context, baseURL string, idToken string) (API, error) {
	return nil, nil
}
