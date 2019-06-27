package pbanua2x

import (
	"encoding/xml"
)

type apiResponse struct {
	XMLName xml.Name    `xml:"response"`
	Data    apiRespData `xml:"data"`
}

type apiRespData struct {
	Error *apiRespError `xml:"error"`
}

type apiRespError struct {
	Message string `xml:"message,attr"`
}
