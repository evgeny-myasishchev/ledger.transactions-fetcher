package ledger

import (
	"context"
	"testing"

	"gopkg.in/h2non/gock.v1"

	tst "github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/internal/testing"

	"github.com/bxcodec/faker/v3"

	"github.com/stretchr/testify/assert"
)

func Test_API_ListAccounts(t *testing.T) {
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
