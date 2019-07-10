package pbanua2x

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/ledger"

	"github.com/bxcodec/faker/v3"
	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/banks"

	"github.com/stretchr/testify/assert"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/dal"
)

func Test_apiStatement_ToDTO(t *testing.T) {
	rand.Seed(time.Now().Unix())

	type fields struct {
		stmt *apiStatement
	}
	type testCase struct {
		fields fields
		assert func(*testing.T, *dal.PendingTransactionDTO)
	}
	type tcFn func() (string, testCase)

	randStmt := func(tranTime time.Time, amount string) *apiStatement {
		localTime := tranTime.Local()
		stmt := apiStatement{
			XMLName:         xml.Name{Local: "statement"},
			Appcode:         "appcode-" + faker.Word(),
			Terminal:        "term-" + faker.Word(),
			Description:     faker.Sentence(),
			Cardamount:      amount,
			Trandate:        localTime.Format("2006-01-02"),
			Trantime:        localTime.Format("15:04:05"),
			ledgerAccountID: "acc-" + faker.Word(),
		}
		return &stmt
	}

	tests := []tcFn{
		func() (string, testCase) {
			tranTime := time.Unix(faker.UnixTime(), 0)
			amountStr := fmt.Sprintf("%.2f", 100+100*rand.Float32())
			stmt := randStmt(tranTime, amountStr)

			idSha1Hash := sha1.New()
			if _, err := idSha1Hash.Write([]byte(
				stmt.Appcode + ":" + stmt.Amount + ":" + stmt.Trandate + ":" + stmt.Trantime,
			)); !assert.NoError(t, err) {
				panic(err)
			}

			id := base64.RawURLEncoding.EncodeToString(sha1.New().Sum(nil))

			return "map standard properties", testCase{
				fields: fields{stmt: stmt},
				assert: func(t *testing.T, got *dal.PendingTransactionDTO) {
					assert.Equal(t, &dal.PendingTransactionDTO{
						ID:        id,
						Comment:   stmt.Description + " (" + stmt.Terminal + ")",
						AccountID: stmt.ledgerAccountID,
						Amount:    amountStr,
						TypeID:    ledger.TransactionTypeIncome,
						Date:      tranTime.Local().Format(time.RFC3339),
					}, got)
				},
			}
		},
		func() (string, testCase) {
			tranTime := time.Unix(faker.UnixTime(), 0)
			amountStr := fmt.Sprintf("%.2f", 100+100*rand.Float32())
			stmt := randStmt(tranTime, "-"+amountStr)
			return "map negative amount as expense", testCase{
				fields: fields{stmt: stmt},
				assert: func(t *testing.T, got *dal.PendingTransactionDTO) {
					assert.Equal(t, amountStr, got.Amount)
					assert.Equal(t, ledger.TransactionTypeExpense, got.TypeID)
				},
			}
		},
	}
	for _, tt := range tests {
		name, tt := tt()
		t.Run(name, func(t *testing.T) {
			got, err := banks.FetchedTransaction(tt.fields.stmt).ToDTO()
			if !assert.NoError(t, err) {
				return
			}
			tt.assert(t, got)
		})
	}
}
