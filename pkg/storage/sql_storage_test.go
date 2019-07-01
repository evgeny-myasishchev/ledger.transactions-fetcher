package storage

import (
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

func randomAccessToken() *oauth.AccessToken {
	return &oauth.AccessToken{
		AccessToken:  "at-" + faker.Word(),
		RefreshToken: "rt-" + faker.Word(),
		IDToken:      "id-" + faker.Word(),
		ExpiresIn:    rand.Uint32(),
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
		assert func(*testing.T, *oauth.AccessToken, error)
	}
	tests := []func() testCase{
		func() testCase {
			token := randomAccessToken()
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
