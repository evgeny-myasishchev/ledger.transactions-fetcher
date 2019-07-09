package ledger

import (
	"bytes"
	"context"
	"encoding/json"

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
	startSessionPayload, err := json.Marshal(map[string]string{
		"google_id_token": idToken,
	})
	if err != nil {
		return nil, err
	}
	req := request.Post(
		baseURL+"/api/sessions",
		"application/json",
		bytes.NewReader(startSessionPayload))

	var sessionData map[string]string
	res := request.Do(ctx, req)
	if err := res.DecodeJSON(&sessionData); err != nil {
		return nil, err
	}
	resVal, err := res()
	if err != nil {
		return nil, err
	}
	cookies := resVal.Cookies()
	var session string
	for _, cookie := range cookies {
		if cookie.Name == sessionCookieName {
			session = cookie.Value
			break
		}
	}

	// TODO: Fail if no session

	return &api{
		baseURL:   baseURL,
		csrfToken: sessionData[csrfTokenName],
		session:   session,
	}, nil
}
