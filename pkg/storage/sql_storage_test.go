package storage

import (
	"context"
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"

	"github.com/stretchr/testify/assert"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/oauth"
)

func init() {
	rand.Seed(time.Now().Unix())
}

type fakeAccessToken struct {
	oauth.AccessToken
}

func randomAccessToken() *AuthTokenDTO {
	return &AuthTokenDTO{
		Email:        faker.Email(),
		IDToken:      "idt-" + faker.Word(),
		RefreshToken: "rt-" + faker.Word(),
	}
}

func Test_sqlStorage_GetAccessTokenByEmail(t *testing.T) {
	type args struct {
		userEmail string
	}
	type testCase struct {
		name   string
		args   args
		setup  func(db *sql.DB) error
		assert func(*testing.T, *AuthTokenDTO, error)
	}
	tests := []func() testCase{
		func() testCase {
			token := randomAccessToken()

			return testCase{
				name: "get existing email",
				args: args{userEmail: token.Email},
				setup: func(db *sql.DB) error {
					if _, err := db.Exec(`
					INSERT INTO users(email, id_token, refresh_token)
					VALUES($1, $2, $3)`,
						token.Email, token.IDToken, token.RefreshToken,
					); err != nil {
						panic(err)
					}
					return nil
				},
				assert: func(t *testing.T, got *AuthTokenDTO, err error) {
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
				assert: func(t *testing.T, got *AuthTokenDTO, err error) {
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
			if err := s.Setup(context.Background()); !assert.NoError(t, err) {
				return
			}
			if tt.setup != nil {
				if err := tt.setup(db); !assert.NoError(t, err) {
					return
				}
			}
			got, err := s.GetAuthTokenByEmail(context.Background(), tt.args.userEmail)
			tt.assert(t, got, err)
		})
	}
}

func Test_sqlStorage_SaveAccessToken(t *testing.T) {
	type args struct {
		token *AuthTokenDTO
	}
	type testCase struct {
		name   string
		args   args
		setup  func(s Storage) error
		assert func(*testing.T, Storage, error)
	}
	tests := []func() testCase{
		func() testCase {
			token := randomAccessToken()
			return testCase{
				name: "write new token",
				args: args{token: token},
				assert: func(t *testing.T, storage Storage, err error) {
					if !assert.NoError(t, err) {
						return
					}
					got, err := storage.GetAuthTokenByEmail(context.Background(), token.Email)
					if !assert.NoError(t, err) {
						return
					}
					assert.Equal(t, token, got)
				},
			}
		},
		func() testCase {
			newToken := randomAccessToken()
			updatedToken := randomAccessToken()
			updatedToken.Email = newToken.Email
			return testCase{
				name: "update existing token",
				args: args{token: updatedToken},
				setup: func(s Storage) error {
					return s.SaveAuthToken(context.TODO(), newToken)
				},
				assert: func(t *testing.T, storage Storage, err error) {
					if !assert.NoError(t, err) {
						return
					}
					got, err := storage.GetAuthTokenByEmail(context.TODO(), newToken.Email)
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
			if err := s.Setup(context.TODO()); !assert.NoError(t, err) {
				return
			}
			if tt.setup != nil {
				if err := tt.setup(s); !assert.NoError(t, err) {
					return
				}
			}
			err = s.SaveAuthToken(context.TODO(), tt.args.token)
			tt.assert(t, s, err)
		})
	}
}
