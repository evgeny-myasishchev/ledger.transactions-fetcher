package request

import (
	"context"
	"net/http"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/lib-core-golang/diag"
)

var defaultLogger = diag.CreateLogger()

type sendCfg struct {
	logger diag.Logger
}

// SendOpt is a send specific option
type SendOpt func(cfg *sendCfg)

func withLogger(logger diag.Logger) SendOpt {
	return func(cfg *sendCfg) {
		cfg.logger = logger
	}
}

// ReqFactory is a function that creates an instance of a request
type ReqFactory func() (*http.Request, error)

// Get creates a new req factory that creates a get request for given url
func Get(url string) ReqFactory {
	return func() (*http.Request, error) {
		return http.NewRequest("GET", url, nil)
	}
}

// Do will send the request. Will fail if response status is other than 2xx
func Do(ctx context.Context, factory ReqFactory, opts ...SendOpt) (*http.Response, error) {
	httpClient := &http.Client{
		Transport: http.DefaultTransport,
	}
	req, err := factory()
	if err != nil {
		return nil, err
	}
	return httpClient.Do(req)
}
