package ledger

// PendingTransaction represents ledger pending transaction
type PendingTransaction struct {
	ID        string
	Amount    string
	Date      string
	Comment   string
	AccountID string
	TypeID    string
}
