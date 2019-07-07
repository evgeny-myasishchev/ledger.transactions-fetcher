package ledger

const (
	// TransactionTypeIncome is holds type of income transactions
	TransactionTypeIncome uint8 = 1

	// TransactionTypeExpense is holds type of expense transactions
	TransactionTypeExpense uint8 = 2
)

// PendingTransaction represents ledger pending transaction
type PendingTransaction struct {
	ID        string
	Amount    string
	Date      string
	Comment   string
	AccountID string
	TypeID    string
}
