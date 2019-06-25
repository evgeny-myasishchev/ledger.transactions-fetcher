package pbanua2x

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bxcodec/faker/v3"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/banks"
)

type mockConfig struct {
	userConfigs map[string]interface{}
}

func (cfg *mockConfig) GetUserConfig(ctx context.Context, userID string, receiver interface{}) error {
	if userCfg, ok := cfg.userConfigs[userID]; ok {
		reflect.ValueOf(receiver).Elem().Set(reflect.ValueOf(userCfg).Elem())
		return nil
	}
	return errors.New("Config not found, user: " + userID)
}

func TestNewFetcher(t *testing.T) {
	type args struct {
		userID string
	}

	existingConfig := &userConfig{
		UserID: "uid-" + faker.Word(),
	}

	fetcherCfg := &mockConfig{
		userConfigs: map[string]interface{}{
			existingConfig.UserID: existingConfig,
		},
	}

	notExistingUser := "user-id-" + faker.Word()

	tests := []struct {
		name   string
		args   args
		assert func(*testing.T, banks.Fetcher, error)
	}{
		{
			name: "existing user",
			args: args{userID: existingConfig.UserID},
			assert: func(t *testing.T, fetcher banks.Fetcher, err error) {
				if !assert.NoError(t, err) {
					return
				}
				if !assert.IsType(t, &pbanua2xFetcher{}, fetcher) {
					return
				}
				bpfetcher := fetcher.(*pbanua2xFetcher)
				assert.Equal(t, existingConfig, bpfetcher.userCfg)
			},
		},
		{
			name: "not existing user",
			args: args{userID: notExistingUser},
			assert: func(t *testing.T, fetcher banks.Fetcher, err error) {
				assert.EqualError(t, err, "Config not found, user: "+notExistingUser)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewFetcher(context.Background(), tt.args.userID, fetcherCfg)
			tt.assert(t, got, err)
		})
	}
}
