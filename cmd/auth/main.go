package main

import (
	"fmt"
	"net/url"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/config"
	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/lib-core-golang/diag"
)

var logger = diag.CreateLogger()

func main() {
	svcCfg := config.Load()

	diag.SetupLoggingSystem(func(setup diag.LoggingSystemSetup) {
		setup.SetLogLevel(svcCfg.StringParam(config.LogLevel).Value())
	})

	authURL, err := url.Parse("https://accounts.google.com/o/oauth2/v2/auth")
	if err != nil {
		panic(err)
	}
	query := authURL.Query()
	query.Add("response_type", "code")
	query.Add("client_id", svcCfg.StringParam(config.GoogleClientID).Value())
	query.Add("redirect_uri", "urn:ietf:wg:oauth:2.0:oob")
	query.Add("scope", "email")
	query.Add("access_type", "offline")
	authURL.RawQuery = query.Encode()

	fmt.Println("Copy url below and paste it into the browser.")
	fmt.Println("Then follow the instruction and use rake get-access-token[code] with the code displayed")
	fmt.Println("The url:")
	fmt.Println(authURL)
}
