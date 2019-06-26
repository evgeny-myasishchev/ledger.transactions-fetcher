package pbanua2x

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/banks"
)

type userConfig struct {
	UserID string

	// Merchants is a map where key is BankAccountID and value is a merchant config
	// that is configured for reading from that account
	Merchants map[string]*merchantConfig
}

type merchantConfig struct {
	ID       string
	Password string
}

type pbanua2xFetcher struct {
	apiURL  string
	userCfg *userConfig
}

func pbTimeForamt(t time.Time) string {
	return fmt.Sprint(t.Day(), ".", t.Month(), ".", t.Year())
}

func (f *pbanua2xFetcher) Fetch(ctx context.Context, params *banks.FetchParams) ([]banks.BankTransaction, error) {

	fmt.Println(params.BankAccountID)

	// TODO: Err if no such merchant
	merchant := f.userCfg.Merchants[params.BankAccountID]

	var data strings.Builder
	data.WriteString(`<oper>cmt</oper>`)
	data.WriteString(`<wait>0</wait>`)
	data.WriteString(`<test>0</test>`)
	data.WriteString(`<payment id="">`)
	data.WriteString(`<prop name="sd" value="` + pbTimeForamt(params.From) + `" />`)
	data.WriteString(`<prop name="ed" value="` + pbTimeForamt(params.From) + `" />`)
	data.WriteString(`<prop name="card" value="` + params.BankAccountID + `" />`)
	data.WriteString(`</payment>`)

	dataHash := md5.Sum([]byte(data.String() + merchant.Password))
	signature := sha1.Sum(dataHash[:])

	var payload strings.Builder
	payload.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	payload.WriteString(`<request version="1.0">`)
	payload.WriteString(`<merchant>`)
	payload.WriteString(`<id>` + merchant.ID + `</id>`)
	payload.WriteString(`<signature>`)
	payload.Write(signature[:])
	payload.WriteString(`</signature>`)
	payload.WriteString(`</merchant>`)
	payload.WriteString(`<data>`)
	payload.WriteString(data.String())
	payload.WriteString(`</data>`)
	payload.WriteString(`</request>`)

	_, err := http.Post(f.apiURL, "application/xml", strings.NewReader(payload.String()))

	return nil, err
}

// NewFetcher creates an instance of a pbanua2x fetcher
func NewFetcher(ctx context.Context, userID string, cfg banks.FetcherConfig) (banks.Fetcher, error) {
	var userCfg userConfig
	if err := cfg.GetUserConfig(ctx, userID, &userCfg); err != nil {
		return nil, err
	}
	return &pbanua2xFetcher{
		userCfg: &userCfg,
	}, nil
}
