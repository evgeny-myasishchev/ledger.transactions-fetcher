package banks

import "context"

// FetcherConfig is a storage where config of user is stored
type FetcherConfig interface {
	GetUserConfig(ctx context.Context, userID string, receiver interface{}) error
}
