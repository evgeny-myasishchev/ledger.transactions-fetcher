package pbanua2x

import (
	"encoding/xml"
	"math/rand"
	"testing"
	"time"

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
		want   *dal.PendingTransactionDTO
	}
	type tcFn func() (string, testCase)

	randStmt := func(tranTime time.Time) *apiStatement {
		localTime := tranTime.Local()
		stmt := apiStatement{
			XMLName:         xml.Name{Local: "statement"},
			Terminal:        "term: " + faker.Word(),
			Description:     faker.Sentence(),
			Trandate:        localTime.Format("2006-01-02"),
			Trantime:        localTime.Format("15:04:05"),
			ledgerAccountID: "acc-" + faker.Word(),
		}
		return &stmt
	}

	tests := []tcFn{
		func() (string, testCase) {
			tranTime := time.Unix(faker.UnixTime(), 0)
			stmt := randStmt(tranTime)
			return "map standard properties", testCase{
				fields: fields{stmt: stmt},
				want: &dal.PendingTransactionDTO{
					Comment:   stmt.Description + " (" + stmt.Terminal + ")",
					AccountID: stmt.ledgerAccountID,
					Date:      tranTime.Local().Format(time.RFC3339),
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
			assert.Equal(t, tt.want, got)
		})
	}
}
