package pbanua2x

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"math/rand"
	"net/url"
	"reflect"
	"strings"
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
				if !assert.IsType(t, &pbanua2xFetcher{}, fetcher) {
					return
				}
				bpfetcher := fetcher.(*pbanua2xFetcher)
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

func Test_pbanua2xFetcher_Fetch(t *testing.T) {
	type testCase struct {
		name string
		run  func(t *testing.T, f banks.Fetcher)
	}

	ledgerAccountID := "acc-" + faker.Word()

	merchant := merchantConfig{
		ID:          "mc1-" + faker.Word(),
		Password:    "mcpwd-" + faker.Word(),
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
	pbTimeForamt := func(t time.Time) string {
		return fmt.Sprint(t.Day(), ".", int(t.Month()), ".", t.Year())
	}

	apiURL, err := url.Parse(faker.URL())
	if !assert.NoError(t, err) {
		return
	}

	writeStatementString := func(builder *strings.Builder, index int) apiStatement {
		stmt := apiStatement{
			XMLName: xml.Name{Local: "statement"},
			Card:    fmt.Sprint(index, "-card-"+faker.Word()),
			Appcode: "app-code-" + faker.Word(),
		}
		res, err := xml.Marshal(&stmt)
		if err != nil {
			panic(err)
		}
		builder.Write(res)
		return stmt
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

					var expectedData strings.Builder
					expectedData.WriteString(`<oper>cmt</oper>`)
					expectedData.WriteString(`<wait>0</wait>`)
					expectedData.WriteString(`<test>0</test>`)
					expectedData.WriteString(`<payment id="">`)
					expectedData.WriteString(`<prop name="sd" value="` + pbTimeForamt(fetchParams.From) + `" />`)
					expectedData.WriteString(`<prop name="ed" value="` + pbTimeForamt(fetchParams.To) + `" />`)
					expectedData.WriteString(`<prop name="card" value="` + merchant.BankAccount + `" />`)
					expectedData.WriteString(`</payment>`)

					md5hash := md5.Sum([]byte(expectedData.String() + merchant.Password))
					md5hashHex := hex.EncodeToString(md5hash[:])
					signature := sha1.Sum([]byte(md5hashHex))

					var expectedXML strings.Builder
					expectedXML.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
					expectedXML.WriteString(`<request version="1.0">`)
					expectedXML.WriteString(`<merchant>`)
					expectedXML.WriteString(`<id>` + merchant.ID + `</id>`)
					expectedXML.WriteString(`<signature>`)
					expectedXML.WriteString(hex.EncodeToString(signature[:]))
					expectedXML.WriteString(`</signature>`)
					expectedXML.WriteString(`</merchant>`)
					expectedXML.WriteString(`<data>`)
					expectedXML.WriteString(expectedData.String())
					expectedXML.WriteString(`</data>`)
					expectedXML.WriteString(`</request>`)

					var resp strings.Builder
					resp.WriteString("<response>")
					resp.WriteString("<data><info><statements>")
					statements := []apiStatement{
						writeStatementString(&resp, 1),
						writeStatementString(&resp, 2),
						writeStatementString(&resp, 3),
					}
					resp.WriteString("</statements></info></data>")
					resp.WriteString("</response>")

					gock.New(apiURL.Scheme+"://"+apiURL.Host).
						Post(apiURL.Path).
						MatchHeader("content-type", "application/xml").
						BodyString(expectedXML.String()).
						Reply(200).
						BodyString(resp.String())

					trxs, err := f.Fetch(context.Background(), &fetchParams)
					if !assert.NoError(t, err) {
						return
					}

					if !assert.True(t, gock.IsDone()) {
						return
					}
					assert.Len(t, trxs, len(statements))
					for i, trx := range trxs {
						stmt := statements[i]
						stmt.ledgerAccountID = fetchParams.LedgerAccountID
						assert.Equal(t, &stmt, trx)
					}
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

					assert.EqualError(t, err, "No pbanua2x merchant configured for account: "+notConfiguredAcc)
				},
			}
		},
		func() testCase {
			return testCase{
				name: "fail if message with info",
				run: func(t *testing.T, f banks.Fetcher) {
					fetchParams := banks.FetchParams{
						LedgerAccountID: ledgerAccountID,
						From:            timeVal(time.Parse(faker.BaseDateFormat, faker.Date())),
						To:              timeVal(time.Parse(faker.BaseDateFormat, faker.Date())),
					}

					errorMessage := "Err: " + faker.Sentence()

					var expectedData strings.Builder
					expectedData.WriteString(`<oper>cmt</oper>`)
					expectedData.WriteString(`<wait>0</wait>`)
					expectedData.WriteString(`<test>0</test>`)
					expectedData.WriteString(`<payment id="">`)
					expectedData.WriteString(`<prop name="sd" value="` + pbTimeForamt(fetchParams.From) + `" />`)
					expectedData.WriteString(`<prop name="ed" value="` + pbTimeForamt(fetchParams.To) + `" />`)
					expectedData.WriteString(`<prop name="card" value="` + merchant.BankAccount + `" />`)
					expectedData.WriteString(`</payment>`)

					md5hash := md5.Sum([]byte(expectedData.String() + merchant.Password))
					md5hashHex := hex.EncodeToString(md5hash[:])
					signature := sha1.Sum([]byte(md5hashHex))

					var expectedXML strings.Builder
					expectedXML.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
					expectedXML.WriteString(`<request version="1.0">`)
					expectedXML.WriteString(`<merchant>`)
					expectedXML.WriteString(`<id>` + merchant.ID + `</id>`)
					expectedXML.WriteString(`<signature>`)
					expectedXML.WriteString(hex.EncodeToString(signature[:]))
					expectedXML.WriteString(`</signature>`)
					expectedXML.WriteString(`</merchant>`)
					expectedXML.WriteString(`<data>`)
					expectedXML.WriteString(expectedData.String())
					expectedXML.WriteString(`</data>`)
					expectedXML.WriteString(`</request>`)

					var resp strings.Builder
					resp.WriteString("<response>")
					resp.WriteString("<data><info>")
					resp.WriteString(errorMessage)
					resp.WriteString("</info></data>")
					resp.WriteString("</response>")

					gock.New(apiURL.Scheme+"://"+apiURL.Host).
						Post(apiURL.Path).
						MatchHeader("content-type", "application/xml").
						BodyString(expectedXML.String()).
						Reply(200).
						BodyString(resp.String())

					_, err := f.Fetch(context.Background(), &fetchParams)
					if !assert.Error(t, err) {
						return
					}
					assert.EqualError(t, err, errorMessage)
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

					gock.New(apiURL.Scheme + "://" + apiURL.Host).
						Post(apiURL.Path).
						Reply(code).
						BodyString("Something went wrong")

					_, err := f.Fetch(context.Background(), &fetchParams)
					if !assert.Error(t, err) {
						return
					}
				},
			}
		},
		func() testCase {
			return testCase{
				name: "fail if error response",
				run: func(t *testing.T, f banks.Fetcher) {
					fetchParams := banks.FetchParams{
						LedgerAccountID: ledgerAccountID,
						From:            timeVal(time.Parse(faker.BaseDateFormat, faker.Date())),
						To:              timeVal(time.Parse(faker.BaseDateFormat, faker.Date())),
					}

					errorMessage := faker.Sentence()
					respBody := `<response><data><error message="` + errorMessage + `"></error></data></response>`

					gock.New(apiURL.Scheme + "://" + apiURL.Host).
						Post(apiURL.Path).
						Reply(200).
						BodyString(respBody)

					_, err = f.Fetch(context.Background(), &fetchParams)
					if !assert.Error(t, err) {
						return
					}

					assert.EqualError(t, err, "PB api call failed: "+errorMessage)
				},
			}
		},
	}
	for _, tt := range tests {
		tt := tt()
		t.Run(tt.name, func(t *testing.T) {
			defer gock.Off()
			f := &pbanua2xFetcher{userCfg: userCfg, apiURL: apiURL.String()}
			tt.run(t, f)
		})
	}
}
