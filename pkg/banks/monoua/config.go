package monoua

type userConfig struct {
	UserID string

	// Merchants is a map where key is LedgerAccountID and value is a merchant config
	// that is configured for reading from that account
	Merchants map[string]*merchantConfig
}

type merchantConfig struct {
	XToken      string
	BankAccount string
}
