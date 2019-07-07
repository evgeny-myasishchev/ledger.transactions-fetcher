package dal

import (
	"context"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/lib-core-golang/diag"
)

var logger = diag.CreateLogger()

// AuthTokenDTO is a DTO to store user auth token
type AuthTokenDTO struct {
	Email        string
	IDToken      string
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
}

// Storage is a persistance layer
type Storage interface {
	Setup(ctx context.Context) error
	GetAuthTokenByEmail(ctx context.Context, email string) (*AuthTokenDTO, error)
	SaveAuthToken(ctx context.Context, token *AuthTokenDTO) error
}
