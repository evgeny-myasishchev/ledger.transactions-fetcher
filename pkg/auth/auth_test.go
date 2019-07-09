package auth

import (
	"context"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/dal"
	tst "github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/internal/testing"
	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
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
					IDToken:      types.IDToken(idToken),
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

func Test_Service_FetchAuthToken(t *testing.T) {
	type fields struct {
		oauthClient OAuthClient
		storage     dal.Storage
	}
	type args struct {
		ctx   context.Context
		email string
	}
	type testCase struct {
		fields fields
		args   args
		assert func(token types.IDToken, err error)
	}

	randomTokenDto := func(t *testing.T, exp int64) *dal.AuthTokenDTO {
		email := faker.Email()
		idToken, err := tst.EncodeUnsignedJWT(t, map[string]interface{}{
			"email": email,
			"exp":   exp,
		})
		if !assert.NoError(t, err) {
			return nil
		}
		return &dal.AuthTokenDTO{
			Email:        email,
			IDToken:      types.IDToken(idToken),
			RefreshToken: "rt-" + faker.Word(),
		}
	}

	type tcFn func(*testing.T) *testCase
	tests := []func() (string, tcFn){
		func() (string, tcFn) {
			return "get the token", func(t *testing.T) *testCase {
				authToken := randomTokenDto(t, time.Now().Unix()+20)
				if authToken == nil {
					return nil
				}
				ctx := context.TODO()
				ctrl := gomock.NewController(t)
				storage := NewMockStorage(ctrl)
				storage.
					EXPECT().
					GetAuthTokenByEmail(ctx, authToken.Email).
					Return(authToken, nil)
				return &testCase{
					fields: fields{storage: storage},
					args:   args{ctx: ctx, email: authToken.Email},
					assert: func(token types.IDToken, err error) {
						defer ctrl.Finish()
						if !assert.NoError(t, err) {
							return
						}
						assert.Equal(t, authToken.IDToken, token)
					},
				}
			}
		},
		func() (string, tcFn) {
			return "refresh expired token", func(t *testing.T) *testCase {
				authToken := randomTokenDto(t, time.Now().Unix()-20)
				if authToken == nil {
					return nil
				}

				refreshedToken := &RefreshedToken{
					IDToken: types.IDToken("refreshed-token-" + faker.Word()),
				}

				ctx := context.TODO()
				ctrl := gomock.NewController(t)

				storage := NewMockStorage(ctrl)
				storage.
					EXPECT().
					GetAuthTokenByEmail(ctx, authToken.Email).
					Return(authToken, nil)

				oauthClient := NewMockOAuthClient(ctrl)
				oauthClient.
					EXPECT().
					PerformRefreshFlow(ctx, authToken.RefreshToken).
					Return(refreshedToken, nil)

				storage.
					EXPECT().
					SaveAuthToken(ctx, &dal.AuthTokenDTO{
						Email:        authToken.Email,
						IDToken:      refreshedToken.IDToken,
						RefreshToken: authToken.RefreshToken,
					}).
					Return(nil)

				return &testCase{
					fields: fields{storage: storage, oauthClient: oauthClient},
					args:   args{ctx: ctx, email: authToken.Email},
					assert: func(token types.IDToken, err error) {
						defer ctrl.Finish()
						if !assert.NoError(t, err) {
							return
						}
						assert.Equal(t, refreshedToken.IDToken, token)
					},
				}
			}
		},
	}
	for _, tt := range tests {
		name, tt := tt()
		t.Run(name, func(t *testing.T) {
			tt := tt(t)
			if !assert.NotNil(t, tt) {
				return
			}
			svc := Service(&service{
				oauthClient: tt.fields.oauthClient,
				storage:     tt.fields.storage,
			})
			got, err := svc.FetchAuthToken(tt.args.ctx, tt.args.email)
			tt.assert(got, err)
		})
	}
}
