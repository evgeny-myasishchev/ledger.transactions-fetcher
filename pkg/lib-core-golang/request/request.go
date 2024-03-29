package request

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

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

func flattenAndObfuscate(values map[string][]string, obfuscateKeys ...string) map[string]string {
	flattened := make(map[string]string, len(values))
	for key, val := range values {
		flattened[key] = strings.Join(val, ", ")
	}
	for _, obfuscateKey := range obfuscateKeys {
		if val, ok := flattened[obfuscateKey]; ok {
			flattened[obfuscateKey] = fmt.Sprint("*obfuscated, length=", len(val), "*")
		}
	}
	return flattened
}

// ReqFactory is a function that creates an instance of a request
type ReqFactory func() (*http.Request, error)

// WithHeader injects request header
func (f ReqFactory) WithHeader(key string, value string) ReqFactory {
	return func() (*http.Request, error) {
		req, err := f()
		req.Header.Add(key, value)
		return req, err
	}
}

func newRequest(method string, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "*/*")
	return req, nil
}

// Get creates a new req factory that creates a get request for given url
func Get(url string) ReqFactory {
	return func() (*http.Request, error) {
		return newRequest("GET", url, nil)
	}
}

// Post creates a new req factory that creates a post request body
func Post(reqURL string, contentType string, body io.Reader) ReqFactory {
	return func() (*http.Request, error) {
		req, err := newRequest("POST", reqURL, body)
		if req != nil {
			req.Header.Set("Content-Type", contentType)
		}
		return req, err
	}
}

// PostForm creates a new req factory that creates a post request with form data
func PostForm(reqURL string, data url.Values) ReqFactory {
	return func() (*http.Request, error) {
		req, err := newRequest("POST", reqURL, strings.NewReader(data.Encode()))
		if req != nil {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
		return req, err
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

// DecodeJSON will decode response body to a given receiver
func (f ResFactory) DecodeJSON(receiver interface{}) error {
	res, err := f()
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return json.NewDecoder(res.Body).Decode(receiver)
}

func newResFactory(res *http.Response, err error) ResFactory {
	return func() (*http.Response, error) {
		if res != nil && res.StatusCode >= 300 {
			return nil, NewHTTPErrorFromResponse(res)
		}
		return res, err
	}
}

type requestLogger func(req *http.Request) (*http.Response, error)

func (rl requestLogger) RoundTrip(req *http.Request) (*http.Response, error) {
	return rl(req)
}

// Do will send the request. Will fail if response status is other than 2xx
func Do(ctx context.Context, factory ReqFactory, opts ...SendOpt) ResFactory {
	// TODO: Include requestID header from context

	cfg := &sendCfg{
		logger: defaultLogger,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	httpClient := &http.Client{
		Transport: requestLogger(func(req *http.Request) (*http.Response, error) {
			cfg.logger.
				WithData(diag.MsgData{
					"protocol": req.URL.Scheme,
					"url":      req.URL.String(),
					"qs":       flattenAndObfuscate(req.URL.Query()),

					// TODO: Obfuscate some headers
					"headers": flattenAndObfuscate(req.Header),

					"method":        req.Method,
					"contentLength": req.ContentLength,
				}).
				Info(ctx, "SEND REQUEST START")
			res, err := http.DefaultTransport.RoundTrip(req)
			if res != nil {
				msgData := diag.MsgData{
					"url":           req.URL.String(),
					"httpStatus":    res.StatusCode,
					"method":        req.Method,
					"headers":       flattenAndObfuscate(res.Header),
					"contentLength": res.ContentLength,
				}
				if res.StatusCode >= 300 {
					defer res.Body.Close()
					if body, err := ioutil.ReadAll(res.Body); err != nil {
						cfg.logger.WithError(err).Error(ctx, "Failed to read response body")
					} else {
						msgData["body"] = string(body)
					}

				}
				cfg.logger.
					WithData(msgData).
					Info(ctx, "SEND REQUEST COMPLETE")
			}
			return res, err
		}),
	}
	req, err := factory()
	if err != nil {
		return newResFactory(nil, err)
	}
	return newResFactory(httpClient.Do(req))
}
