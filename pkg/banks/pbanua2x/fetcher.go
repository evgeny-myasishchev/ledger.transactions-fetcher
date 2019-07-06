package pbanua2x

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/lib-core-golang/request"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/lib-core-golang/diag"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/banks"
)

var logger = diag.CreateLogger()

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
	return fmt.Sprint(t.Day(), ".", int(t.Month()), ".", t.Year())
}

func (f *pbanua2xFetcher) Fetch(ctx context.Context, params *banks.FetchParams) ([]banks.BankTransaction, error) {
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

	md5hash := md5.Sum([]byte(data.String() + merchant.Password))
	md5hashHex := hex.EncodeToString(md5hash[:])
	signature := sha1.Sum([]byte(md5hashHex))

	var payload strings.Builder
	payload.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	payload.WriteString(`<request version="1.0">`)
	payload.WriteString(`<merchant>`)
	payload.WriteString(`<id>` + merchant.ID + `</id>`)
	payload.WriteString(`<signature>`)
	payload.WriteString(hex.EncodeToString(signature[:]))
	payload.WriteString(`</signature>`)
	payload.WriteString(`</merchant>`)
	payload.WriteString(`<data>`)
	payload.WriteString(data.String())
	payload.WriteString(`</data>`)
	payload.WriteString(`</request>`)

	req := request.Post(f.apiURL, "application/xml", strings.NewReader(payload.String()))
	res := request.Do(ctx, req)
	body, err := res.ReadAll()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to fetch transactions")
	}

	var apiResp apiResponse
	if err := xml.Unmarshal(body, &apiResp); err != nil {
		logger.
			WithData(diag.MsgData{"body": body}).
			WithError(err).
			Error(ctx, "Failed to unmarshal response")
		return nil, err
	}

	if apiResp.Data.Error != nil {
		return nil, fmt.Errorf("PB api call failed: %v", apiResp.Data.Error.Message)
	}

	statements := apiResp.Data.Info.Statements.Values

	// TODO: Obfuscate BankAccountID
	logger.Info(ctx, "Fetched %v statements for account: %v", len(statements), params.BankAccountID)

	trxs := make([]banks.BankTransaction, len(statements))
	for i, stmt := range statements {
		trxs[i] = &stmt
	}

	return trxs, err
}

// NewFetcher creates an instance of a pbanua2x fetcher
func NewFetcher(ctx context.Context, userID string, cfg banks.FetcherConfig) (banks.Fetcher, error) {
	var userCfg userConfig
	if err := cfg.GetUserConfig(ctx, userID, &userCfg); err != nil {
		return nil, errors.Wrap(err, "Failed to fetch user config")
	}
	return &pbanua2xFetcher{
		// TODO: Parametrize
		apiURL:  "https://api.privatbank.ua/p24api/rest_fiz",
		userCfg: &userCfg,
	}, nil
}
