package monoua

import (
	"strconv"
	"time"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/dal"
	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/ledger"
)

// https://api.monobank.ua/docs/#/definitions/StatementItems
type monoTransaction struct {
	ID              string `json:"id"`
	Time            int64  `json:"time"`
	Description     string `json:"description"`
	Mcc             int32  `json:"mcc"`
	OriginalMcc     int32  `json:"originalMcc"`
	Amount          int64  `json:"amount"`
	OperationAmount int64  `json:"operationAmount"`
	CurrencyCode    int32  `json:"currencyCode"`
	CommissionRate  int64  `json:"commissionRate"`
	CashbackAmount  int64  `json:"cashbackAmount"`
	Balance         int64  `json:"balance"`
	Hold            bool   `json:"hold"`

	ledgerAccountID string
}

func (stmt *monoTransaction) ToDTO() (*dal.PendingTransactionDTO, error) {
	return &dal.PendingTransactionDTO{
		ID:        stmt.ID,
		Comment:   stmt.Description,
		AccountID: stmt.ledgerAccountID,
		Amount:    strconv.FormatFloat(float64(stmt.Amount)/100, 'f', -1, 64),
		TypeID:    ledger.TransactionTypeExpense,

		Date: time.Unix(stmt.Time, 0).Format(time.RFC3339),
	}, nil
}
