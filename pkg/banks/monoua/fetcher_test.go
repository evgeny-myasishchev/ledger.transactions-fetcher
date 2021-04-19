package monoua

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/pkg/errors"

	"gopkg.in/h2non/gock.v1"

	"github.com/bxcodec/faker/v3"
	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/banks"
	"github.com/stretchr/testify/assert"
)

type mockConfig struct {
	userConfigs map[string]interface{}
}

func (cfg *mockConfig) GetUserConfig(ctx context.Context, userID string, receiver interface{}) error {
	if userCfg, ok := cfg.userConfigs[userID]; ok {
		reflect.ValueOf(receiver).Elem().Set(reflect.ValueOf(userCfg).Elem())
		return nil
	}
	return errors.New("Config not found, user: " + userID)
}

func TestNewFetcher(t *testing.T) {
	type args struct {
		userID string
	}

	existingConfig := &userConfig{
		UserID: "uid-" + faker.Word(),
	}

	fetcherCfg := &mockConfig{
		userConfigs: map[string]interface{}{
			existingConfig.UserID: existingConfig,
		},
	}

	notExistingUser := "user-id-" + faker.Word()

	tests := []struct {
		name   string
		args   args
		assert func(*testing.T, banks.Fetcher, error)
	}{
		{
			name: "existing user",
			args: args{userID: existingConfig.UserID},
			assert: func(t *testing.T, fetcher banks.Fetcher, err error) {
				if !assert.NoError(t, err) {
					return
				}
				if !assert.IsType(t, &monoFetcher{}, fetcher) {
					return
				}
				bpfetcher := fetcher.(*monoFetcher)
				assert.Equal(t, existingConfig, bpfetcher.userCfg)
			},
		},
		{
			name: "not existing user",
			args: args{userID: notExistingUser},
			assert: func(t *testing.T, fetcher banks.Fetcher, err error) {
				cause := errors.Cause(err)
				assert.EqualError(t, cause, "Config not found, user: "+notExistingUser)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewFetcher(context.Background(), tt.args.userID, fetcherCfg)
			tt.assert(t, got, err)
		})
	}
}

func Test_monoFetcher_Fetch(t *testing.T) {
	type testCase struct {
		name string
		run  func(t *testing.T, f banks.Fetcher)
	}

	ledgerAccountID := "acc-" + faker.Word()

	merchant := merchantConfig{
		XToken:      "mcpwd-" + faker.Word(),
		BankAccount: "ba-" + faker.Word(),
	}

	userCfg := &userConfig{
		UserID: "uid-" + faker.Word(),
		Merchants: map[string]*merchantConfig{
			ledgerAccountID: &merchant,
		},
	}

	timeVal := func(t time.Time, err error) time.Time {
		if err != nil {
			panic(err)
		}
		return t
	}

	apiURL, err := url.Parse(faker.URL())
	if !assert.NoError(t, err) {
		return
	}

	tests := []func() testCase{
		func() testCase {
			return testCase{
				name: "regular api call",
				run: func(t *testing.T, f banks.Fetcher) {
					fetchParams := banks.FetchParams{
						LedgerAccountID: ledgerAccountID,
						From:            timeVal(time.Parse(faker.BaseDateFormat, faker.Date())),
						To:              timeVal(time.Parse(faker.BaseDateFormat, faker.Date())),
					}

					statements := []*monoTransaction{
						{
							ID: faker.UUIDDigit(),
						},
						{
							ID: faker.UUIDDigit(),
						},
						{
							ID: faker.UUIDDigit(),
						},
					}

					wantStatements := make([]banks.FetchedTransaction, len(statements))
					for i, v := range statements {
						stmt := *v
						stmt.ledgerAccountID = fetchParams.LedgerAccountID
						wantStatements[i] = &stmt
					}

					wantPath := fmt.Sprintf("/personal/statement/%v/%v/%v", merchant.BankAccount, fetchParams.From.Unix(), fetchParams.To.Unix())
					gock.New(apiURL.Scheme + "://" + apiURL.Host).
						Get(wantPath).
						Reply(200).
						JSON(statements)

					trxs, err := f.Fetch(context.Background(), &fetchParams)
					if !assert.NoError(t, err) {
						return
					}

					if !assert.True(t, gock.IsDone()) {
						return
					}
					assert.Equal(t, wantStatements, trxs)
				},
			}
		},
		func() testCase {
			return testCase{
				name: "fail if no merchant configured",
				run: func(t *testing.T, f banks.Fetcher) {
					notConfiguredAcc := "unknown-acc-" + faker.Word()
					fetchParams := banks.FetchParams{
						LedgerAccountID: notConfiguredAcc,
						From:            timeVal(time.Parse(faker.BaseDateFormat, faker.Date())),
						To:              timeVal(time.Parse(faker.BaseDateFormat, faker.Date())),
					}

					_, err := f.Fetch(context.Background(), &fetchParams)
					if !assert.Error(t, err) {
						return
					}

					assert.EqualError(t, err, "No monoua merchant configured for account: "+notConfiguredAcc)
				},
			}
		},
		func() testCase {
			return testCase{
				name: "fail if non-200 response status",
				run: func(t *testing.T, f banks.Fetcher) {
					fetchParams := banks.FetchParams{
						LedgerAccountID: ledgerAccountID,
						From:            timeVal(time.Parse(faker.BaseDateFormat, faker.Date())),
						To:              timeVal(time.Parse(faker.BaseDateFormat, faker.Date())),
					}

					code := rand.Intn(100) + 300

					wantPath := fmt.Sprintf("/personal/statement/%v/%v/%v", merchant.BankAccount, fetchParams.From.Unix(), fetchParams.To.Unix())

					gock.New(apiURL.Scheme + "://" + apiURL.Host).
						Get(wantPath).
						Reply(code).
						BodyString("Something went wrong")

					_, err := f.Fetch(context.Background(), &fetchParams)
					if !assert.Error(t, err) {
						return
					}
				},
			}
		},
	}
	for _, tt := range tests {
		tt := tt()
		t.Run(tt.name, func(t *testing.T) {
			defer gock.Off()
			f := &monoFetcher{userCfg: userCfg, apiBaseURL: apiURL.Scheme + "://" + apiURL.Host}
			tt.run(t, f)
		})
	}
}
