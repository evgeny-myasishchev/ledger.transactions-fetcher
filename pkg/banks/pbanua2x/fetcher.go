package pbanua2x

import (
	"context"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/banks"
)

type userConfig struct {
	UserID    string
	Merchants *map[string]merchantConfig
}

type merchantConfig struct {
	ID       string
	Password string
}

type pbanua2xFetcher struct {
	userCfg *userConfig
}

func (f *pbanua2xFetcher) Fetch(ctx context.Context, params *banks.FetchParams) ([]banks.BankTransaction, error) {
	return nil, nil
}

// NewFetcher creates an instance of a pbanua2x fetcher
func NewFetcher(ctx context.Context, userID string, cfg banks.FetcherConfig) (banks.Fetcher, error) {
	return nil, nil
}
