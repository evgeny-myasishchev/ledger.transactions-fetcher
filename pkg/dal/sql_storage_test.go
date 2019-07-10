package dal

import (
	"context"
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/types"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func randomAccessToken() *AuthTokenDTO {
	return &AuthTokenDTO{
		Email:        faker.Email(),
		IDToken:      types.IDToken("idt-" + faker.Word()),
		RefreshToken: "rt-" + faker.Word(),
	}
}

type trxOpt func(*PendingTransactionDTO)

func withAccount(accountID string) trxOpt {
	return func(dto *PendingTransactionDTO) {
		dto.AccountID = accountID
	}
}

func withCreatedAt(createdAt time.Time) trxOpt {
	return func(dto *PendingTransactionDTO) {
		dto.CreatedAt = createdAt
	}
}

func withSyncedAt(syncedAt time.Time) trxOpt {
	return func(dto *PendingTransactionDTO) {
		dto.SyncedAt = &syncedAt
	}
}

func randTrx(opts ...trxOpt) *PendingTransactionDTO {
	dto := &PendingTransactionDTO{
		ID:        faker.Word(),
		Amount:    faker.Word(),
		Date:      faker.Word(),
		Comment:   faker.Word(),
		AccountID: faker.Word(),
		TypeID:    uint8(rand.Intn(20)),
	}
	for _, opt := range opts {
		opt(dto)
	}
	return dto
}

func setupMemoryDB(t *testing.T) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", ":memory:")
	if !assert.NoError(t, err) {
		return nil, err
	}
	s := Storage(&sqlStorage{db: db})
	if err := s.Setup(context.TODO()); !assert.NoError(t, err) {
		return nil, err
	}
	return db, nil
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
			db, err := setupMemoryDB(t)
			if err != nil {
				return
			}
			defer db.Close()
			s := Storage(&sqlStorage{db: db})
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
			db, err := setupMemoryDB(t)
			if err != nil {
				return
			}
			defer db.Close()
			s := Storage(&sqlStorage{db: db})
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

func Test_sqlStorage_SavePendingTransaction(t *testing.T) {
	rand.Seed(time.Now().Unix())
	type args struct {
		trx *PendingTransactionDTO
	}
	type testCase struct {
		args   args
		assert func()
	}

	type ttFn func(*testing.T, *sql.DB) testCase
	tests := []func() (string, ttFn){
		func() (string, ttFn) {
			return "insert new", func(t *testing.T, db *sql.DB) testCase {
				trx := randTrx()
				return testCase{
					args: args{trx: trx},
					assert: func() {
						row := db.QueryRow(`
						SELECT 
							id, amount, date, comment, account_id, type_id, created_at 
						FROM transactions 
						WHERE id=$1
						`, trx.ID)
						var got PendingTransactionDTO
						var gotCreatedAt time.Time
						if err := row.Scan(
							&got.ID,
							&got.Amount,
							&got.Date,
							&got.Comment,
							&got.AccountID,
							&got.TypeID,
							&gotCreatedAt,
						); !assert.NoError(t, err) {
							return
						}
						assert.Equal(t, trx, &got)
						assert.InDelta(t, time.Now().Unix(), gotCreatedAt.Unix(), 1)
					},
				}
			}
		},
		func() (string, ttFn) {
			return "update existing", func(t *testing.T, db *sql.DB) testCase {
				newTrx := randTrx()
				s := Storage(&sqlStorage{db: db, nowFn: defaultNowFn})
				if err := s.SavePendingTransaction(context.TODO(), newTrx); !assert.NoError(t, err) {
					return testCase{}
				}
				updatedTrx := randTrx()
				updatedTrx.ID = newTrx.ID
				syncedAt := time.Unix(faker.RandomUnixTime(), 0)
				updatedTrx.SyncedAt = &syncedAt

				return testCase{
					args: args{trx: updatedTrx},
					assert: func() {
						row := db.QueryRow(`
						SELECT 
							id, amount, date, comment, account_id, type_id, created_at, synced_at
						FROM transactions 
						WHERE id=$1
						`, updatedTrx.ID)
						var gotCreatedAt time.Time
						var gotSyncedAt time.Time
						var got PendingTransactionDTO
						if err := row.Scan(
							&got.ID,
							&got.Amount,
							&got.Date,
							&got.Comment,
							&got.AccountID,
							&got.TypeID,
							&gotCreatedAt,
							&gotSyncedAt,
						); !assert.NoError(t, err) {
							return
						}
						assert.InDelta(t, gotCreatedAt.Unix(), time.Now().Unix(), 1)
						assert.InDelta(t, gotSyncedAt.Unix(), syncedAt.Unix(), 1)
						got.SyncedAt = updatedTrx.SyncedAt
						assert.Equal(t, updatedTrx, &got)
					},
				}
			}
		},
	}
	for _, tt := range tests {
		name, ttFn := tt()
		t.Run(name, func(t *testing.T) {
			db, err := setupMemoryDB(t)
			if err != nil {
				return
			}
			defer db.Close()
			s := Storage(&sqlStorage{db: db, nowFn: defaultNowFn})
			tt := ttFn(t, db)
			if t.Failed() {
				return
			}
			err = s.SavePendingTransaction(context.Background(), tt.args.trx)
			if !assert.NoError(t, err) {
				return
			}
			tt.assert()
		})
	}
}

func Test_sqlStorage_FindNotSyncedTransactions(t *testing.T) {
	type args struct {
		accountID string
	}
	type fields struct {
		now time.Time
	}
	type testCase struct {
		args args
		want []PendingTransactionDTO
	}
	type tcFn func(*testing.T, Storage) *testCase
	tests := []func() (string, fields, tcFn){
		func() (string, fields, tcFn) {
			now := time.Unix(faker.UnixTime(), 0).UTC()
			return "get not synced transactions", fields{now: now}, func(t *testing.T, s Storage) *testCase {
				accountID := "acc-" + faker.Word()
				notSyncedTrxs := []PendingTransactionDTO{
					*randTrx(withCreatedAt(now), withAccount(accountID), withSyncedAt(time.Unix(faker.UnixTime(), 0).UTC())),
					*randTrx(withCreatedAt(now), withAccount(accountID), withSyncedAt(time.Unix(faker.UnixTime(), 0).UTC())),
					*randTrx(withCreatedAt(now), withAccount(accountID), withSyncedAt(time.Unix(faker.UnixTime(), 0).UTC())),
				}
				allTrxs := append(notSyncedTrxs, *randTrx(withCreatedAt(now), withAccount(accountID)))
				allTrxs = append(allTrxs, *randTrx(withCreatedAt(now), withAccount(accountID)))
				for _, trx := range allTrxs {
					if err := s.SavePendingTransaction(context.TODO(), &trx); !assert.NoError(t, err) {
						return nil
					}
				}
				return &testCase{
					args: args{accountID: accountID},
					want: notSyncedTrxs,
				}
			}
		},
	}
	for _, tt := range tests {
		name, fields, tt := tt()
		t.Run(name, func(t *testing.T) {
			db, err := setupMemoryDB(t)
			if err != nil {
				return
			}
			defer db.Close()
			s := Storage(&sqlStorage{
				db: db,
				nowFn: func() time.Time {
					return fields.now
				},
			})
			tt := tt(t, s)
			if tt == nil {
				return
			}
			got, err := s.FindNotSyncedTransactions(context.TODO(), tt.args.accountID)
			if !assert.NoError(t, err) {
				return
			}
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}
