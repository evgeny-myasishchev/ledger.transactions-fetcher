package monoua

import (
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/bxcodec/faker/v3"
	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/dal"
	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/ledger"
	"github.com/stretchr/testify/assert"
)

func Test_monoTransaction_ToDTO(t *testing.T) {
	type args struct {
		tx monoTransaction
	}
	type want struct {
		dto *dal.PendingTransactionDTO
		err error
	}
	type testCase struct {
		name string
		args
		want
	}

	tests := []func() testCase{
		func() testCase {
			tx := monoTransaction{
				ID:              faker.UUIDDigit(),
				Description:     faker.Sentence(),
				ledgerAccountID: faker.UUIDDigit(),
				Amount:          -int64(gofakeit.Number(1000, 2000)) * 100,
				Time:            gofakeit.Date().Unix(),
			}
			return testCase{
				name: "map expense tx",
				args: args{
					tx: tx,
				},
				want: want{
					dto: &dal.PendingTransactionDTO{
						ID:        tx.ID,
						Comment:   tx.Description,
						AccountID: tx.ledgerAccountID,
						Amount:    strconv.FormatFloat(float64(tx.Amount)/100, 'f', -1, 64),
						Date:      time.Unix(tx.Time, 0).Format(time.RFC3339),
						TypeID:    ledger.TransactionTypeExpense,
					},
				},
			}
		},
		func() testCase {
			tx := monoTransaction{
				ID:              faker.UUIDDigit(),
				Description:     faker.Sentence(),
				ledgerAccountID: faker.UUIDDigit(),
				Amount:          int64(gofakeit.Number(1000, 2000)) * 100,
				Time:            gofakeit.Date().Unix(),
			}
			return testCase{
				name: "map income tx",
				args: args{
					tx: tx,
				},
				want: want{
					dto: &dal.PendingTransactionDTO{
						ID:        tx.ID,
						Comment:   tx.Description,
						AccountID: tx.ledgerAccountID,
						Amount:    strconv.FormatFloat(float64(tx.Amount)/100, 'f', -1, 64),
						Date:      time.Unix(tx.Time, 0).Format(time.RFC3339),
						TypeID:    ledger.TransactionTypeIncome,
					},
				},
			}
		},
	}
	for _, tt := range tests {
		tt := tt()
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.args.tx.ToDTO()
			if err != nil || tt.want.err != nil {
				assert.True(t,
					errors.Is(err, tt.want.err),
					"Unexpected err %v, expected %v", err, tt.want.err,
				)
				return
			}
			assert.Equal(t, tt.want.dto, got)
		})
	}
}
