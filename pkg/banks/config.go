package banks

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"path"
)

// FetcherConfig is a storage where config of user is stored
type FetcherConfig interface {
	GetUserConfig(ctx context.Context, userID string, receiver interface{}) error
}

type fsFetcherConfig struct {
	dir string
}

func (cfg *fsFetcherConfig) GetUserConfig(ctx context.Context, userID string, receiver interface{}) error {
	logger.Debug(ctx, "Reading user config: %v", userID)
	buffer, err := ioutil.ReadFile(path.Join(cfg.dir, userID+".json"))
	if err != nil {
		return err
	}
	return json.Unmarshal(buffer, receiver)
}

// NewFSFetcherConfig creates an instance of a fetcher config
// that is reading from local file system
func NewFSFetcherConfig(configDir string) FetcherConfig {
	logger.Info(nil, "Initializing fetcher config: %v", configDir)
	return &fsFetcherConfig{configDir}
}
