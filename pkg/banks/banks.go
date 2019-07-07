package banks

import (
	"context"
	"time"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/dal"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/lib-core-golang/diag"
)

var logger = diag.CreateLogger()

// FetchedTransaction is a generic type that represents transaction fetched from bank
type FetchedTransaction interface {
	ToDTO() (*dal.PendingTransactionDTO, error)
}

// FetchParams represents what to fetch from bank
type FetchParams struct {
	From            time.Time
	To              time.Time
	LedgerAccountID string
}

// Fetcher can fetch transaction for particular bank accountID
type Fetcher interface {
	Fetch(ctx context.Context, params *FetchParams) ([]FetchedTransaction, error)
}
