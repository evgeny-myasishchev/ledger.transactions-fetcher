package ledger

import (
	"context"
	"math/rand"
	"testing"

	"github.com/bxcodec/faker/v3"
	tst "github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/internal/testing"
	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/types"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func Test_API_ListAccounts(t *testing.T) {
	defer gock.Clean()
	type fields struct {
		baseURL   string
		session   string
		csrfToken string
	}
	type args struct {
		ctx context.Context
	}
	type testCase struct {
		fields fields
		want   []AccountDTO
		after  func()
	}
	type tcFn func() (string, func(*testing.T) *testCase)
	tests := []tcFn{
		func() (string, func(*testing.T) *testCase) {
			return "get accounts", func(t *testing.T) *testCase {
				want := []AccountDTO{
					AccountDTO{ID: "acc-1-" + faker.Word(), Name: "Acc 1 " + faker.Word()},
					AccountDTO{ID: "acc-2-" + faker.Word(), Name: "Acc 2 " + faker.Word()},
					AccountDTO{ID: "acc-3-" + faker.Word(), Name: "Acc 3 " + faker.Word()},
				}
				body, ok := tst.JSONMarshalToReader(t, want)
				if !ok {
					return nil
				}
				fields := fields{
					baseURL:   "https://my-ledger." + faker.Word() + ".com",
					session:   "sess-" + faker.Word(),
					csrfToken: "csrf-token-" + faker.Word(),
				}
				gock.New(fields.baseURL).
					Get("/accounts").
					MatchHeaders(map[string]string{
						"Cookie": sessionCookieName + "=" + fields.session,
					}).
					Reply(200).
					Body(body)
				return &testCase{
					fields: fields,
					want:   want,
					after: func() {
						assert.True(t, gock.IsDone())
					},
				}
			}
		},
	}
	for _, tt := range tests {
		name, tt := tt()
		t.Run(name, func(t *testing.T) {
			tt := tt(t)
			if t.Failed() {
				return
			}
			a := API(&api{
				baseURL:   tt.fields.baseURL,
				session:   tt.fields.session,
				csrfToken: tt.fields.csrfToken,
			})
			got, err := a.ListAccounts(context.TODO())
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, tt.want, got)
			if tt.after != nil {
				tt.after()
			}
		})
	}
}

func Test_API_ReportPendingTransaction(t *testing.T) {
	defer gock.Clean()
	type fields struct {
		baseURL   string
		session   string
		csrfToken string
	}
	type args struct {
		trx PendingTransactionDTO
	}
	type testCase struct {
		fields fields
		args   args
		assert func()
	}
	randTrx := func() PendingTransactionDTO {
		return PendingTransactionDTO{
			ID:        faker.Word(),
			Amount:    faker.Word(),
			Date:      faker.Word(),
			Comment:   faker.Word(),
			AccountID: faker.Word(),
			TypeID:    uint8(rand.Intn(20)),
		}
	}
	type tcFn func() (string, func(*testing.T) *testCase)
	tests := []tcFn{
		func() (string, func(*testing.T) *testCase) {
			return "report valid data", func(t *testing.T) *testCase {
				trx := randTrx()
				body, ok := tst.JSONMarshalToReader(t, trx)
				if !ok {
					return nil
				}
				fields := fields{
					baseURL:   "https://my-ledger." + faker.Word() + ".com",
					session:   "sess-" + faker.Word(),
					csrfToken: "csrf-token-" + faker.Word(),
				}
				gock.New(fields.baseURL).
					Post("/pending-transactions").
					MatchHeaders(map[string]string{
						"Cookie":       sessionCookieName + "=" + fields.session,
						csrfHeaderName: fields.csrfToken,
					}).
					Reply(200).
					Body(body)
				return &testCase{
					fields: fields,
					args:   args{trx: trx},
					assert: func() {
						assert.True(t, gock.IsDone())
					},
				}
			}
		},
	}
	for _, tt := range tests {
		name, tt := tt()
		t.Run(name, func(t *testing.T) {
			tt := tt(t)
			if t.Failed() {
				return
			}
			a := API(&api{
				baseURL:   tt.fields.baseURL,
				session:   tt.fields.session,
				csrfToken: tt.fields.csrfToken,
			})
			err := a.ReportPendingTransaction(context.TODO(), tt.args.trx)
			if !assert.NoError(t, err) {
				return
			}
			tt.assert()
		})
	}
}

func TestNewAPI(t *testing.T) {
	defer gock.Clean()
	type args struct {
		baseURL string
		idToken types.IDToken
	}
	type testCase struct {
		args  args
		want  API
		after func()
	}
	type tcFn func(*testing.T) testCase
	tests := []func() (string, tcFn){
		func() (string, tcFn) {
			return "start session and return new api", func(t *testing.T) testCase {
				args := args{
					baseURL: "https://my-ledger." + faker.Word() + ".com",
					idToken: types.IDToken("id-token-" + faker.Word()),
				}
				session := "sess-" + faker.Word()
				csrf := "csrf-" + faker.Word()
				gock.New(args.baseURL).
					Post("/api/sessions").
					JSON(map[string]string{
						"google_id_token": args.idToken.Value(),
					}).
					Reply(200).
					AddHeader("Set-Cookie", sessionCookieName+"="+session).
					JSON(map[string]string{
						csrfTokenName: csrf,
					})
				return testCase{
					args: args,
					want: API(&api{
						baseURL:   args.baseURL,
						csrfToken: csrf,
						session:   session,
					}),
					after: func() {
						assert.True(t, gock.IsDone())
					},
				}
			}
		},
	}
	for _, tt := range tests {
		name, tt := tt()
		t.Run(name, func(t *testing.T) {
			tt := tt(t)
			got, err := APIFactory(NewAPI)(context.TODO(), tt.args.baseURL, tt.args.idToken)
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, tt.want, got)
			tt.after()
		})
	}
}
