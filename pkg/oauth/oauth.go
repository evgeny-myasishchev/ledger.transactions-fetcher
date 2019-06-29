package oauth

import "fmt"

// Client is an oauth client abstraction
type Client interface {
	BuildCodeGrantURL() string
	GetAccessTokenByCode(code string) (string, error)
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

func (c *googleOAuthClient) GetAccessTokenByCode(code string) (string, error) {
	return "", nil
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
