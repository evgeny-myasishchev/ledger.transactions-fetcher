package auth

import (
	"context"
	"fmt"
	"net/url"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/types"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/lib-core-golang/request"
	"github.com/pkg/errors"
)

// AccessToken represents access token data
type AccessToken struct {
	RefreshToken string        `json:"refresh_token"`
	IDToken      types.IDToken `json:"id_token"`
}

// RefreshedToken represents refreshed token data
type RefreshedToken struct {
	IDToken types.IDToken `json:"id_token"`
}

// OAuthClient is an oauth client abstraction
type OAuthClient interface {
	BuildCodeGrantURL() string
	PerformAuthCodeExchangeFlow(ctx context.Context, code string) (*AccessToken, error)
	PerformRefreshFlow(ctx context.Context, refreshToken string) (*RefreshedToken, error)
}

type googleOAuthClient struct {
	clientID     string
	clientSecret string
}

func (c *googleOAuthClient) BuildCodeGrantURL() string {
	return fmt.Sprint(
		"https://accounts.google.com/o/oauth2/v2/auth?",
		"response_type=code",
		"&client_id="+c.clientID,
		"&redirect_uri=urn:ietf:wg:oauth:2.0:oob",
		"&scope=email",
		"&access_type=offline",
	)
}

func (c *googleOAuthClient) PerformAuthCodeExchangeFlow(ctx context.Context, code string) (*AccessToken, error) {
	form := url.Values{}
	form.Add("code", code)
	form.Add("grant_type", "authorization_code")
	form.Add("redirect_uri", "urn:ietf:wg:oauth:2.0:oob")
	form.Add("client_id", c.clientID)
	form.Add("client_secret", c.clientSecret)
	logger.Debug(ctx, "Exchaning auth code on access token")

	req := request.PostForm("https://www.googleapis.com/oauth2/v4/token", form)
	res := request.Do(ctx, req)
	var accessToken AccessToken
	if err := res.DecodeJSON(&accessToken); err != nil {
		return nil, errors.Wrap(err, "Failed to get access token")
	}
	return &accessToken, nil
}

func (c *googleOAuthClient) PerformRefreshFlow(ctx context.Context, refreshToken string) (*RefreshedToken, error) {
	return nil, nil
}

// GoogleOAuthOpt represents options for google oauth client
type GoogleOAuthOpt func(*googleOAuthClient)

// WithClientSecrets option will set google oauth client id/secret
func WithClientSecrets(clientID string, clientSecret string) GoogleOAuthOpt {
	return func(c *googleOAuthClient) {
		c.clientID = clientID
		c.clientSecret = clientSecret
	}
}

// NewGoogleOAuthClient creates an instance of a google OAuth client
func NewGoogleOAuthClient(opts ...GoogleOAuthOpt) OAuthClient {
	client := &googleOAuthClient{}
	for _, opt := range opts {
		opt(client)
	}
	return client
}
