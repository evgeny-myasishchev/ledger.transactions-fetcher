package auth

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/bxcodec/faker/v3"
	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/dal"
	tst "github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/internal/testing"
)

func Test_Service_RegisterUser(t *testing.T) {
	type tcFn func(t *testing.T)
	tests := []func() (string, tcFn){
		func() (string, tcFn) {
			return "register new user", func(t *testing.T) {
				email := faker.Email()
				idToken, err := tst.EncodeUnsignedJWT(t, map[string]interface{}{
					"email": email,
				})
				if err != nil {
					return
				}
				accessToken := AccessToken{
					RefreshToken: "rt-" + faker.Word(),
					IDToken:      idToken,
				}
				code := "code-" + faker.Word()
				ctx := context.TODO()

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				oauthClient := NewMockOAuthClient(ctrl)
				oauthClient.EXPECT().
					PerformAuthCodeExchangeFlow(ctx, code).
					Return(&accessToken, nil)

				storage := NewMockStorage(ctrl)
				storage.
					EXPECT().
					SaveAuthToken(ctx, &dal.AuthTokenDTO{
						Email:        email,
						IDToken:      accessToken.IDToken,
						RefreshToken: accessToken.RefreshToken,
					}).
					Return(nil)
				svc := NewService(WithStorage(storage), WithOAuthClient(oauthClient))
				if err := svc.RegisterUser(context.TODO(), code); !assert.NoError(t, err) {
					return
				}
			}
		},
	}
	for _, tt := range tests {
		t.Run(tt())
	}
}
