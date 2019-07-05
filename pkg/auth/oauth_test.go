package auth

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"testing"
	"time"

	tst "github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/internal/testing"

	"github.com/bxcodec/faker/v3"
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
				IDToken:      "id-" + faker.Word(),
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

func TestAccessToken_ExtractIDTokenDetails(t *testing.T) {
	type fields struct {
		IDToken string
	}
	type testCase struct {
		name   string
		fields fields
		want   *IDTokenDetails
	}

	tests := []func() testCase{
		func() testCase {
			email := faker.Email()
			expires := faker.UnixTime()
			idToken := map[string]interface{}{
				"email": email,
				"exp":   expires,
			}

			tokenString, err := tst.EncodeUnsignedJWT(t, idToken)
			if err != nil {
				panic(err)
			}

			return testCase{
				name:   "correct jwt token",
				fields: fields{IDToken: tokenString},
				want: &IDTokenDetails{
					Email:   email,
					Expires: expires,
				},
			}
		},
		func() testCase {
			tokenString := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Imtmb3ZrZG8zMmkzOTBAbnZvZGtxLmdqbCIsImV4cCI6MzA0ODEzMzA3NX0.HrYUdClKf3NnMhWOuILHf41nDd-1emuSH2ayNBEY3K4"

			return testCase{
				name:   "real token",
				fields: fields{IDToken: tokenString},
				want: &IDTokenDetails{
					Email:   "kfovkdo32i390@nvodkq.gjl",
					Expires: 3048133075,
				},
			}
		},
	}
	for _, tt := range tests {
		tt := tt()
		t.Run(tt.name, func(t *testing.T) {
			at := &AccessToken{
				IDToken: tt.fields.IDToken,
			}
			got, err := at.ExtractIDTokenDetails()
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
