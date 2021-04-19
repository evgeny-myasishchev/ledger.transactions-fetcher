package monoua

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/banks"
	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/lib-core-golang/diag"
	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/lib-core-golang/request"
	"github.com/pkg/errors"
)

var logger = diag.CreateLogger()

type monoFetcher struct {
	apiBaseURL string
	userCfg    *userConfig
}

func pbTimeForamt(t time.Time) string {
	return fmt.Sprint(t.Day(), ".", int(t.Month()), ".", t.Year())
}

func (f *monoFetcher) Fetch(ctx context.Context, params *banks.FetchParams) ([]banks.FetchedTransaction, error) {
	merchant, ok := f.userCfg.Merchants[params.LedgerAccountID]
	if !ok {
		return nil, fmt.Errorf("No monoua merchant configured for account: %v", params.LedgerAccountID)
	}
	reqPath := fmt.Sprintf("/personal/statement/%v/%v/%v", merchant.BankAccount, params.From.Unix(), params.To.Unix())
	req := request.Get(f.apiBaseURL + reqPath)
	req = req.WithHeader("X-Token", merchant.XToken)
	res := request.Do(ctx, req)
	body, err := res.ReadAll()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to fetch transactions")
	}

	var statements []monoTransaction
	if err := json.Unmarshal(body, &statements); err != nil {
		logger.
			WithData(diag.MsgData{"body": body}).
			WithError(err).
			Error(ctx, "Failed to unmarshal response")
		return nil, err
	}

	logger.Info(ctx, "Fetched %v statements for account: %v", len(statements), params.LedgerAccountID)

	trxs := make([]banks.FetchedTransaction, len(statements))
	for i, stmt := range statements {
		stmt := stmt
		stmt.ledgerAccountID = params.LedgerAccountID
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
	return &monoFetcher{
		apiBaseURL: "https://api.monobank.ua",
		userCfg:    &userCfg,
	}, nil
}
