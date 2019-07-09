package types

import (
	"testing"

	"github.com/bxcodec/faker/v3"
	tst "github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/internal/testing"
	"github.com/stretchr/testify/assert"
)

func Test_IDToken_ExtractIDTokenDetails(t *testing.T) {
	type fields struct {
		IDToken IDToken
	}
	type testCase struct {
		name   string
		fields fields
		want   *IDTokenDetails
	}

	tests := []func() testCase{
		func() testCase {
			email := faker.Email()
			expires := faker.UnixTime()
			idToken := map[string]interface{}{
				"email": email,
				"exp":   expires,
			}

			tokenString, err := tst.EncodeUnsignedJWT(t, idToken)
			if err != nil {
				panic(err)
			}

			return testCase{
				name:   "correct jwt token",
				fields: fields{IDToken: IDToken(tokenString)},
				want: &IDTokenDetails{
					Email:   email,
					Expires: expires,
				},
			}
		},
		func() testCase {
			tokenString := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Imtmb3ZrZG8zMmkzOTBAbnZvZGtxLmdqbCIsImV4cCI6MzA0ODEzMzA3NX0.HrYUdClKf3NnMhWOuILHf41nDd-1emuSH2ayNBEY3K4"

			return testCase{
				name:   "real token",
				fields: fields{IDToken: IDToken(tokenString)},
				want: &IDTokenDetails{
					Email:   "kfovkdo32i390@nvodkq.gjl",
					Expires: 3048133075,
				},
			}
		},
	}
	for _, tt := range tests {
		tt := tt()
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.fields.IDToken.ExtractIDTokenDetails()
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
