package pbanua2x

import (
	"encoding/xml"
	"errors"

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
}

func (pt *apiStatement) ToDTO() (*dal.PendingTransactionDTO, error) {
	return nil, errors.New("Not implemented")
}
