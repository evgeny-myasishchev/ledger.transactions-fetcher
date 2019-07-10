package dal

import (
	"context"
	"time"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/types"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/lib-core-golang/diag"
)

var logger = diag.CreateLogger()

// AuthTokenDTO is a DTO to store user auth token
type AuthTokenDTO struct {
	Email        string
	IDToken      types.IDToken
	RefreshToken string
}

// PendingTransactionDTO is a DTO to store pending transactions
type PendingTransactionDTO struct {
	ID        string
	Amount    string
	Date      string
	Comment   string
	AccountID string
	TypeID    uint8

	CreatedAt time.Time
	SyncedAt  *time.Time
}

// Storage is a persistance layer
type Storage interface {
	Setup(ctx context.Context) error
	GetAuthTokenByEmail(ctx context.Context, email string) (*AuthTokenDTO, error)
	SaveAuthToken(ctx context.Context, token *AuthTokenDTO) error

	SavePendingTransaction(ctx context.Context, trx *PendingTransactionDTO) error
	PendingTransactionExist(ctx context.Context, id string) (bool, error)

	FindNotSyncedTransactions(ctx context.Context, accountID string) ([]PendingTransactionDTO, error)
}
