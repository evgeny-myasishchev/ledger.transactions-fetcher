package banks

import (
	"context"
	"time"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/lib-core-golang/diag"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/ledger"
)

var logger = diag.CreateLogger()

// BankTransaction is a generic type that represents bank transaction
type BankTransaction interface {
	ToPendingTransaction() (*ledger.PendingTransaction, error)
}

// FetchParams represents what to fetch from bank
type FetchParams struct {
	From            time.Time
	To              time.Time
	LedgerAccountID string
}

// Fetcher can fetch transaction for particular bank accountID
type Fetcher interface {
	Fetch(ctx context.Context, params *FetchParams) ([]BankTransaction, error)
}
