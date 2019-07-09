package auth

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/dal"
	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/types"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/lib-core-golang/diag"
)

var logger = diag.CreateLogger()

// Service is an auth service abstraction
type Service interface {
	RegisterUser(ctx context.Context, oauthCode string) error
	FetchAuthToken(ctx context.Context, email string) (types.IDToken, error)
}

type service struct {
	oauthClient OAuthClient
	storage     dal.Storage
}

func (svc *service) RegisterUser(ctx context.Context, oauthCode string) error {
	logger.Debug(ctx, "Submitting oauth code and exchange it for token")
	accessToken, err := svc.oauthClient.PerformAuthCodeExchangeFlow(ctx, oauthCode)
	if err != nil {
		return errors.Wrap(err, "Failed to perform oauth code flow")
	}
	idTokenDetails, err := accessToken.IDToken.ExtractIDTokenDetails()
	if err != nil {
		return errors.Wrap(err, "Failed to extract ID token")
	}
	logger.Debug(ctx, "Got token for user %v, saving", idTokenDetails.Email)
	return svc.storage.SaveAuthToken(ctx, &dal.AuthTokenDTO{
		Email:        idTokenDetails.Email,
		IDToken:      accessToken.IDToken,
		RefreshToken: accessToken.RefreshToken,
	})
}

func (svc *service) FetchAuthToken(ctx context.Context, email string) (types.IDToken, error) {
	logger.Debug(ctx, "Fetching auth token for user %v", email)
	token, err := svc.storage.GetAuthTokenByEmail(ctx, email)
	if err != nil {
		return "", errors.Wrap(err, "Failed to get auth token from storage")
	}

	tokenDetails, err := token.IDToken.ExtractIDTokenDetails()
	if err != nil {
		return "", errors.Wrap(err, "Failed to extract token details")
	}

	// Refreshing 15 seconds before expiry
	now := time.Now()
	if tokenDetails.Expires+15 <= now.Unix() {
		logger.Debug(ctx, "User token expired. Refreshing")
		refreshedToken, err := svc.oauthClient.PerformRefreshFlow(ctx, token.RefreshToken)
		if err != nil {
			return "", errors.Wrap(err, "Failed to refresh token")
		}
		if err := svc.storage.SaveAuthToken(ctx, &dal.AuthTokenDTO{
			Email:        token.Email,
			IDToken:      refreshedToken.IDToken,
			RefreshToken: token.RefreshToken,
		}); err != nil {
			return "", errors.Wrap(err, "Failed to save refreshed token")
		}
		return refreshedToken.IDToken, nil
	}

	return token.IDToken, nil
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
