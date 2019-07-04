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

// Send will send the request. Will fail if response status is other than 2xx
func Send(ctx context.Context, req *http.Request, opts ...SendOpt) (*http.Response, error) {
	httpClient := &http.Client{
		Transport: http.DefaultTransport,
	}
	return httpClient.Do(req)
}
