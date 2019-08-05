package config

// Log represents logger specific options
type Log struct {
	Level string `config:"key=log/logLevel"`
}

// Google represents google settings
type Google struct {
	ClientID     string `config:"key=google/client-id"`
	ClientSecret string `config:"key=google/client-secret"`
}

// Storage represents storage settings
type Storage struct {
	Driver string `config:"key=storage/driver"`
	DSN    string `config:"key=storage/data-source-name"`
}

// FetcherConfig represents settings of a fetcher-config service
type FetcherConfig struct {
	ConfigDir string `config:"key=fetcher-config/config-dir"`
}

// Ledger ledger config
type Ledger struct {
	API string `config:"key=ledger/api"`
}

// Config is a toplevel config structure
type Config struct {
	Log           *Log           `config:"source=local"`
	Google        *Google        `config:"source=local"`
	Storage       *Storage       `config:"source=local"`
	FetcherConfig *FetcherConfig `config:"source=local"`
	Ledger        *Ledger        `config:"source=local"`
}
