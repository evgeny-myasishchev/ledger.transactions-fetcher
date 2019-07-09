package auth

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/types"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func Test_googleOAuthClient_BuildCodeGrantURL(t *testing.T) {
	type fields struct {
		clientID string
	}
	clientID := "client-id-" + faker.Word()
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "build code flow url",
			fields: fields{clientID: clientID},
			want: fmt.Sprint(
				"https://accounts.google.com/o/oauth2/v2/auth?",
				"response_type=code",
				"&client_id="+clientID,
				"&redirect_uri=urn:ietf:wg:oauth:2.0:oob",
				"&scope=email",
				"&access_type=offline",
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewGoogleOAuthClient(WithClientSecrets(tt.fields.clientID, ""))
			got := c.BuildCodeGrantURL()
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_googleOAuthClient_GetAccessTokenByCode(t *testing.T) {
	rand.Seed(time.Now().Unix())
	type fields struct {
		clientID     string
		clientSecret string
	}
	type args struct {
		code string
	}
	type testCase struct {
		name   string
		fields fields
		args   args
		assert func(t *testing.T, got *AccessToken)
	}
	tests := []func() testCase{
		func() testCase {
			code := faker.Word()
			clientID := faker.Word()
			clientSecret := faker.Word()
			want := AccessToken{
				RefreshToken: "rt-" + faker.Word(),
				IDToken:      types.IDToken("id-" + faker.Word()),
			}
			form := url.Values{}
			form.Add("code", code)
			form.Add("grant_type", "authorization_code")
			form.Add("redirect_uri", "urn:ietf:wg:oauth:2.0:oob")
			form.Add("client_id", clientID)
			form.Add("client_secret", clientSecret)
			gock.New("https://www.googleapis.com").
				Post("/oauth2/v4/token").
				MatchHeader("content-type", "application/x-www-form-urlencoded").
				BodyString(form.Encode()).
				Reply(200).
				BodyString(fmt.Sprintf(`{
					"access_token": "not-important",
					"expires_in": 123,
					"refresh_token": "%v",
					"id_token": "%v"
				}`, want.RefreshToken, want.IDToken))
			return testCase{
				name:   "get access token",
				fields: fields{clientID: clientID, clientSecret: clientSecret},
				args:   args{code: code},
				assert: func(t *testing.T, got *AccessToken) {
					if !assert.Equal(t, got, &want) {
						return
					}
					if !assert.True(t, gock.IsDone()) {
						return
					}
				},
			}
		},
	}
	for _, tt := range tests {
		tt := tt()
		t.Run(tt.name, func(t *testing.T) {
			c := OAuthClient(&googleOAuthClient{
				clientID:     tt.fields.clientID,
				clientSecret: tt.fields.clientSecret,
			})
			got, err := c.PerformAuthCodeExchangeFlow(context.Background(), tt.args.code)
			if !assert.NoError(t, err) {
				return
			}
			tt.assert(t, got)
		})
	}
}

func Test_googleOAuthClient_PerformRefreshFlow(t *testing.T) {
	rand.Seed(time.Now().Unix())
	type fields struct {
		clientID     string
		clientSecret string
	}
	type args struct {
		refreshToken string
	}
	type testCase struct {
		name   string
		fields fields
		args   args
		assert func(t *testing.T, got *RefreshedToken)
	}
	tests := []func() testCase{
		func() testCase {
			refreshToken := faker.Word()
			clientID := faker.Word()
			clientSecret := faker.Word()
			want := RefreshedToken{
				IDToken: types.IDToken("id-" + faker.Word()),
			}
			form := url.Values{}
			form.Add("refresh_token", refreshToken)
			form.Add("grant_type", "refresh_token")
			form.Add("client_id", clientID)
			form.Add("client_secret", clientSecret)
			gock.New("https://www.googleapis.com").
				Post("/oauth2/v4/token").
				MatchHeader("content-type", "application/x-www-form-urlencoded").
				BodyString(form.Encode()).
				Reply(200).
				BodyString(fmt.Sprintf(`{
					"access_token": "not-important",
					"expires_in": 123,
					"refresh_token": "not-important",
					"id_token": "%v"
				}`, want.IDToken))
			return testCase{
				name:   "get access token",
				fields: fields{clientID: clientID, clientSecret: clientSecret},
				args:   args{refreshToken: refreshToken},
				assert: func(t *testing.T, got *RefreshedToken) {
					if !assert.Equal(t, &want, got) {
						return
					}
					if !assert.True(t, gock.IsDone()) {
						return
					}
				},
			}
		},
	}
	for _, tt := range tests {
		tt := tt()
		t.Run(tt.name, func(t *testing.T) {
			c := OAuthClient(&googleOAuthClient{
				clientID:     tt.fields.clientID,
				clientSecret: tt.fields.clientSecret,
			})
			got, err := c.PerformRefreshFlow(context.Background(), tt.args.refreshToken)
			if !assert.NoError(t, err) {
				return
			}
			tt.assert(t, got)
		})
	}
}
