package request

import (
	"context"
	"io/ioutil"
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

// ResFactory is a function that holds a request result with a response or error
type ResFactory func() (*http.Response, error)

// ReadAll will read entire body as a byte array
func (f ResFactory) ReadAll() ([]byte, error) {
	res, err := f()
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return ioutil.ReadAll(res.Body)
}

func newResFactory(res *http.Response, err error) ResFactory {
	return func() (*http.Response, error) {
		if res.StatusCode >= 300 {
			return nil, NewHTTPErrorFromResponse(res)
		}
		return res, err
	}
}

// Do will send the request. Will fail if response status is other than 2xx
func Do(ctx context.Context, factory ReqFactory, opts ...SendOpt) ResFactory {
	httpClient := &http.Client{
		Transport: http.DefaultTransport,
	}
	req, err := factory()
	if err != nil {
		return newResFactory(nil, err)
	}
	return newResFactory(httpClient.Do(req))
}
