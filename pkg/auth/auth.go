package auth

import (
	"context"
	"errors"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/dal"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/lib-core-golang/diag"
)

var logger = diag.CreateLogger()

// Service is an auth service abstraction
type Service interface {
	RegisterUser(ctx context.Context, oauthCode string) error
	FetchAuthToken(ctx context.Context, email string) (string, error)
}

type service struct {
	oauthClient OAuthClient
	storage     dal.Storage
}

func (svc *service) RegisterUser(ctx context.Context, oauthCode string) error {
	return errors.New("Not implemented")
}

func (svc *service) FetchAuthToken(ctx context.Context, email string) (string, error) {
	panic("not implemented")
}

// ServiceOpt is an option for auth service
type ServiceOpt func(*service)

// WithOAuthClient will init the service with oauth client
func WithOAuthClient(client OAuthClient) ServiceOpt {
	return func(svc *service) {
		svc.oauthClient = client
	}
}

// WithStorage will init the service with storage
func WithStorage(storage dal.Storage) ServiceOpt {
	return func(svc *service) {
		svc.storage = storage
	}
}

// NewService returns an instance of an auth service
func NewService(opts ...ServiceOpt) Service {
	svc := &service{}
	for _, opt := range opts {
		opt(svc)
	}
	return Service(svc)
}
