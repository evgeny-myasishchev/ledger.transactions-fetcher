package pbanua2x

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/xml"
	"time"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/ledger"

	"github.com/pkg/errors"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/dal"
)

type apiResponse struct {
	XMLName xml.Name    `xml:"response"`
	Data    apiRespData `xml:"data"`
}

type apiRespData struct {
	Error *apiRespError `xml:"error"`
	Info  struct {
		XMLName    xml.Name `xml:"info"`
		Statements *struct {
			XMLName xml.Name       `xml:"statements"`
			Values  []apiStatement `xml:"statement"`
		}
		Value string `xml:",innerxml"`
	}
}

type apiRespError struct {
	Message string `xml:"message,attr"`
}

type apiStatement struct {
	XMLName     xml.Name `xml:"statement"`
	Card        string   `xml:"card,attr"`
	Appcode     string   `xml:"appcode,attr"`
	Trandate    string   `xml:"trandate,attr"`
	Trantime    string   `xml:"trantime,attr"`
	Amount      string   `xml:"amount,attr"`
	Cardamount  string   `xml:"cardamount,attr"`
	Rest        string   `xml:"rest,attr"`
	Terminal    string   `xml:"terminal,attr"`
	Description string   `xml:"description,attr"`

	// This is injected in order to be able to implement ToDTO
	ledgerAccountID string
}

type amountSpec struct {
	typeID uint8
	value  string
}

func parseAmount(amountStr string) amountSpec {
	typeID := ledger.TransactionTypeIncome
	amountStart := 0
	if amountStr[0] == '-' {
		typeID = ledger.TransactionTypeExpense
		amountStart = 1
	}
	amountEnd := amountStart
	for ; amountEnd < len(amountStr); amountEnd++ {
		if amountStr[amountEnd] == '.' {
			continue
		}
		if amountStr[amountEnd] >= '0' && amountStr[amountEnd] <= '9' {
			continue
		}
		break
	}

	return amountSpec{
		typeID: typeID,
		value:  amountStr[amountStart:amountEnd],
	}
}

func (stmt *apiStatement) ToDTO() (*dal.PendingTransactionDTO, error) {
	tranTime, err := time.ParseInLocation(
		"2006-01-02 15:04:05", stmt.Trandate+" "+stmt.Trantime,
		time.Local,
	)
	amount := parseAmount(stmt.Cardamount)
	if err != nil {
		return nil, errors.Wrapf(err,
			"Failed to parse date/time string: '%v %v'",
			stmt.Trandate,
			stmt.Trantime)
	}
	return &dal.PendingTransactionDTO{
		ID: base64.RawURLEncoding.
			EncodeToString(sha1.New().Sum([]byte(
				stmt.Appcode + ":" + stmt.Amount + ":" + stmt.Trandate + ":" + stmt.Trantime,
			))),
		Comment:   stmt.Description + " (" + stmt.Terminal + ")",
		AccountID: stmt.ledgerAccountID,
		Amount:    amount.value,
		TypeID:    amount.typeID,

		// TODO: pb zone should be configurable
		Date: tranTime.Format(time.RFC3339),
	}, nil
}
