package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/lib-core-golang/diag"
)

var logger = diag.CreateLogger()

// AccessToken represents access token data
type AccessToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	ExpiresIn    uint32 `json:"expires_in"`
}

// Client is an oauth client abstraction
type Client interface {
	BuildCodeGrantURL() string
	PerformAuthCodeExchangeFlow(ctx context.Context, code string) (*AccessToken, error)
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
	res, err := http.Post(
		"https://www.googleapis.com/oauth2/v4/token",
		"application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var accessToken AccessToken
	if err := json.NewDecoder(res.Body).Decode(&accessToken); err != nil {
		return nil, err
	}
	return &accessToken, nil
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

// NewGoogleOAuth creates an instance of a google OAuth client
func NewGoogleOAuth(opts ...GoogleOAuthOpt) Client {
	client := &googleOAuthClient{}
	for _, opt := range opts {
		opt(client)
	}
	return client
}
