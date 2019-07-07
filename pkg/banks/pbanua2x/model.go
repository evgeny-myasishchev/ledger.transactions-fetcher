package pbanua2x

import (
	"encoding/xml"
	"time"

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

func (stmt *apiStatement) ToDTO() (*dal.PendingTransactionDTO, error) {
	tranTime, err := time.ParseInLocation(
		"2006-01-02 15:04:05", stmt.Trandate+" "+stmt.Trantime,
		time.Local,
	)
	if err != nil {
		return nil, errors.Wrapf(err,
			"Failed to parse date/time string: '%v %v'",
			stmt.Trandate,
			stmt.Trantime)
	}
	return &dal.PendingTransactionDTO{
		Comment:   stmt.Description + " (" + stmt.Terminal + ")",
		AccountID: stmt.ledgerAccountID,

		// TODO: pb zone should be configurable
		Date: tranTime.Format(time.RFC3339),
	}, nil
}
