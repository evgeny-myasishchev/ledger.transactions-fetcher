package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/dal"
)

func Test_Service_RegisterUser(t *testing.T) {
	type fields struct {
		oauthClient OAuthClient
		storage     dal.Storage
	}
	type args struct {
		oauthCode string
	}
	type testCase struct {
		name   string
		fields fields
		args   args
		assert func(t *testing.T)
	}
	tests := []func() testCase{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		tt := tt()
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(
				WithOAuthClient(tt.fields.oauthClient),
				WithStorage(tt.fields.storage),
			)
			if err := svc.RegisterUser(context.TODO(), tt.args.oauthCode); !assert.NoError(t, err) {
				return
			}
			tt.assert(t)
		})
	}
}
