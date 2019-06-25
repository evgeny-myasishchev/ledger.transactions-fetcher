package pbanua2x

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

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
				assert.EqualError(t, err, "Config not found, user: "+notExistingUser)
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

	bankAccountID := "acc-" + faker.Word()

	merchant := merchantConfig{
		ID:       "mc1-" + faker.Word(),
		Password: "mcpwd-" + faker.Word(),
	}

	userCfg := &userConfig{
		UserID: "uid-" + faker.Word(),
		Merchants: map[string]*merchantConfig{
			bankAccountID: &merchant,
		},
	}

	timeVal := func(t time.Time, err error) time.Time {
		if err != nil {
			panic(err)
		}
		return t
	}
	pbTimeForamt := func(t time.Time) string {
		return fmt.Sprint(t.Day(), ".", t.Month(), ".", t.Year())
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
						BankAccountID: bankAccountID,
						From:          timeVal(time.Parse(faker.BaseDateFormat, faker.Date())),
						To:            timeVal(time.Parse(faker.BaseDateFormat, faker.Date())),
					}

					var expectedXML strings.Builder
					expectedXML.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
					expectedXML.WriteString(`<request version="1.0">`)
					expectedXML.WriteString(`<merchant>`)
					expectedXML.WriteString(`<id>` + merchant.ID + `</id>`)
					expectedXML.WriteString(`<signature></signature>`)
					expectedXML.WriteString(`</merchant>`)
					expectedXML.WriteString(`<data><oper>cmt</oper>`)
					expectedXML.WriteString(`<wait>0</wait>`)
					expectedXML.WriteString(`<test>0</test>`)
					expectedXML.WriteString(`<payment id="">`)
					expectedXML.WriteString(`<prop name="sd" value="` + pbTimeForamt(fetchParams.From) + `" />`)
					expectedXML.WriteString(`<prop name="ed" value="` + pbTimeForamt(fetchParams.From) + `" />`)
					expectedXML.WriteString(`<prop name="card" value="` + bankAccountID + `" />`)
					expectedXML.WriteString(`</payment></data>`)
					expectedXML.WriteString(`</request>`)

					gock.New(apiURL.Scheme+"://"+apiURL.Host).
						Post(apiURL.Path).
						MatchHeader("content-type", "application/xml").
						BodyString(expectedXML.String()).
						Reply(200)

					_, err := f.Fetch(context.Background(), &fetchParams)
					if !assert.NoError(t, err) {
						return
					}

					if !assert.True(t, gock.IsDone()) {
						fmt.Println(gock.GetUnmatchedRequests())
					}
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
