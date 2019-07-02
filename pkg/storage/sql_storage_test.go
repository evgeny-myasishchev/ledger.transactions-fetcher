package storage

import (
	"database/sql"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"

	"github.com/stretchr/testify/assert"

	tst "github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/internal/testing"
	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/oauth"
)

func init() {
	rand.Seed(time.Now().Unix())
}

type fakeAccessToken struct {
	oauth.AccessToken
}

func randomAccessToken(t *testing.T) (*oauth.AccessToken, error) {
	idToken := oauth.IDTokenDetails{
		Email:   faker.Email(),
		Expires: faker.UnixTime(),
	}
	tokenData, err := tst.EncodeUnsignedJWT(t, &idToken)
	if !assert.NoError(t, err) {
		return nil, err
	}
	return &oauth.AccessToken{
		AccessToken:  "at-" + faker.Word(),
		RefreshToken: "rt-" + faker.Word(),
		IDToken:      tokenData,
		ExpiresIn:    rand.Uint32(),
	}, nil
}

func Test_sqlStorage_GetAccessTokenByEmail(t *testing.T) {
	type args struct {
		userEmail string
	}
	type testCase struct {
		name   string
		args   args
		setup  func(db *sql.DB) error
		assert func(*testing.T, *oauth.AccessToken, error)
	}
	tests := []func() testCase{
		func() testCase {
			token, err := randomAccessToken(t)
			if !assert.NoError(t, err) {
				panic(err)
			}
			email := faker.Email()

			return testCase{
				name: "get existing email",
				args: args{userEmail: email},
				setup: func(db *sql.DB) error {
					if _, err := db.Exec(`
					INSERT INTO users(email, access_token, refresh_token, id_token, expires_in)
					VALUES($1, $2, $3, $4, $5)`,
						email, token.AccessToken, token.RefreshToken, token.IDToken, token.ExpiresIn,
					); err != nil {
						panic(err)
					}
					return nil
				},
				assert: func(t *testing.T, got *oauth.AccessToken, err error) {
					if !assert.NoError(t, err) {
						return
					}
					assert.Equal(t, token, got)
				},
			}
		},
		func() testCase {
			email := faker.Email()
			return testCase{
				name: "get not existing email",
				args: args{userEmail: email},
				assert: func(t *testing.T, got *oauth.AccessToken, err error) {
					if !assert.Error(t, err) {
						return
					}
					assert.EqualError(t, err, "Unknown user: "+email)
				},
			}
		},
	}
	for _, tt := range tests {
		tt := tt()
		t.Run(tt.name, func(t *testing.T) {
			db, err := sql.Open("sqlite3", ":memory:")
			if !assert.NoError(t, err) {
				panic(err)
			}
			defer db.Close()
			s := Storage(&sqlStorage{db: db})
			if err := s.Setup(); !assert.NoError(t, err) {
				return
			}
			if tt.setup != nil {
				if err := tt.setup(db); !assert.NoError(t, err) {
					return
				}
			}
			got, err := s.GetAccessTokenByEmail(tt.args.userEmail)
			tt.assert(t, got, err)
		})
	}
}

func Test_sqlStorage_SaveAccessToken(t *testing.T) {
	type args struct {
		token *oauth.AccessToken
	}
	type testCase struct {
		name   string
		args   args
		setup  func(s Storage) error
		assert func(*testing.T, Storage, error)
	}
	tests := []func() testCase{
		func() testCase {
			token, err := randomAccessToken(t)
			if !assert.NoError(t, err) {
				panic(err)
			}
			return testCase{
				name: "write new token",
				args: args{token: token},
				assert: func(t *testing.T, storage Storage, err error) {
					if !assert.NoError(t, err) {
						return
					}
					details, err := token.ExtractIDTokenDetails()
					if !assert.NoError(t, err) {
						return
					}
					got, err := storage.GetAccessTokenByEmail(details.Email)
					if !assert.NoError(t, err) {
						return
					}
					assert.Equal(t, token, got)
				},
			}
		},
		func() testCase {
			newToken, err := randomAccessToken(t)
			if !assert.NoError(t, err) {
				panic(err)
			}
			updatedToken, err := randomAccessToken(t)
			if !assert.NoError(t, err) {
				panic(err)
			}
			updatedToken.IDToken = newToken.IDToken
			return testCase{
				name: "update existing token",
				args: args{token: updatedToken},
				setup: func(s Storage) error {
					return s.SaveAccessToken(newToken)
				},
				assert: func(t *testing.T, storage Storage, err error) {
					if !assert.NoError(t, err) {
						return
					}
					details, err := updatedToken.ExtractIDTokenDetails()
					if !assert.NoError(t, err) {
						fmt.Print(updatedToken.IDToken)
						return
					}
					got, err := storage.GetAccessTokenByEmail(details.Email)
					if !assert.NoError(t, err) {
						return
					}
					assert.Equal(t, updatedToken, got)
				},
			}
		},
	}
	for _, tt := range tests {
		tt := tt()
		t.Run(tt.name, func(t *testing.T) {
			db, err := sql.Open("sqlite3", ":memory:")
			if !assert.NoError(t, err) {
				panic(err)
			}
			defer db.Close()
			s := Storage(&sqlStorage{db: db})
			if err := s.Setup(); !assert.NoError(t, err) {
				return
			}
			if tt.setup != nil {
				if err := tt.setup(s); !assert.NoError(t, err) {
					return
				}
			}
			err = s.SaveAccessToken(tt.args.token)
			tt.assert(t, s, err)
		})
	}
}
