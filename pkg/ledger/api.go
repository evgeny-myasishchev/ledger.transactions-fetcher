package ledger

import "context"

// API is an interface to communicate with ledger
type API interface {
	ReportPendingTransaction(ctx context.Context, trx PendingTransaction) error
}

type api struct {
	session   string
	csrfToken string
}

// NewAPI returns an instance of a new API initialized with given token
func NewAPI(ctx context.Context, idToken string) (API, error) {
	return nil, nil
}
