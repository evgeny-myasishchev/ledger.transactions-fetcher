package ledger

const (
	// TransactionTypeIncome is a type of income transactions
	TransactionTypeIncome uint8 = 1

	// TransactionTypeExpense is a type of expense transactions
	TransactionTypeExpense uint8 = 2
)

// AccountDTO represents ledger account data
type AccountDTO struct {
	ID   string `json:"aggregate_id"`
	Name string `json:"name"`
}
