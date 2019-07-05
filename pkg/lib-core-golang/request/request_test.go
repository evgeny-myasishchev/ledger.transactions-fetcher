package request

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/lib-core-golang/diag"

	"gopkg.in/h2non/gock.v1"

	"github.com/stretchr/testify/assert"

	"github.com/bxcodec/faker/v3"
)

type mockLogger struct {
	messages       []string
	msgDataEntries []diag.MsgData
}

func newMockLogger() *mockLogger {
	return &mockLogger{
		messages:       []string{},
		msgDataEntries: []diag.MsgData{},
	}
}

func (l *mockLogger) Error(ctx context.Context, msg string, args ...interface{}) {
	l.messages = append(l.messages, fmt.Sprintf(msg, args...))
}
func (l *mockLogger) Warn(ctx context.Context, msg string, args ...interface{}) {
	l.messages = append(l.messages, fmt.Sprintf(msg, args...))
}
func (l *mockLogger) Info(ctx context.Context, msg string, args ...interface{}) {
	l.messages = append(l.messages, fmt.Sprintf(msg, args...))
}
func (l *mockLogger) Debug(ctx context.Context, msg string, args ...interface{}) {
	l.messages = append(l.messages, fmt.Sprintf(msg, args...))
}
func (l *mockLogger) WithError(err error) diag.Logger {
	return l
}
func (l *mockLogger) WithData(data diag.MsgData) diag.Logger {
	l.msgDataEntries = append(l.msgDataEntries, data)
	return l
}

func TestDo(t *testing.T) {
	rand.Seed(time.Now().Unix())

	type args struct {
		ctx  context.Context
		req  *http.Request
		opts []SendOpt
	}
	type tcFn func(*testing.T)
	tests := []func() (string, tcFn){
		func() (string, tcFn) {
			return "should send the request and return response", func(t *testing.T) {
				url := faker.URL()
				expectedBody := faker.Sentence()

				gock.New(url).
					Get("/").
					Reply(200).
					BodyString(expectedBody)

				resp := Do(context.TODO(), Get(url))
				if !assert.True(t, gock.IsDone(), "No request performed") {
					return
				}

				respVal, err := resp()
				if !assert.NoError(t, err) {
					return
				}
				assert.Equal(t, 200, respVal.StatusCode)

				actualBody, err := resp.ReadAll()
				if !assert.NoError(t, err) {
					return
				}
				assert.Equal(t, expectedBody, string(actualBody))
			}
		},

		func() (string, tcFn) {
			return "should fail with http err if status code is > 299", func(t *testing.T) {
				url := faker.URL()
				expectedBody := faker.Sentence()

				status := 299 + rand.Intn(300)

				gock.New(url).
					Get("/").
					Reply(status).
					BodyString(expectedBody)

				resp := Do(context.TODO(), Get(url))
				if !assert.True(t, gock.IsDone(), "No request performed") {
					return
				}

				_, err := resp()
				if !assert.Error(t, err) {
					return
				}

				httpErr, ok := err.(HTTPError)
				if !assert.True(t, ok, "Expected http err but got something else:", err) {
					return
				}
				assert.Equal(t, status, httpErr.StatusCode)
			}
		},

		func() (string, tcFn) {
			return "should log req start/end", func(t *testing.T) {
				reqURLString := faker.URL() + fmt.Sprintf("?qs1=%v&qs2=%v", faker.Word(), faker.Word())
				expectedBody := faker.Sentence()

				status := 199 + rand.Intn(100)
				log := newMockLogger()

				gock.New(reqURLString).
					Get("/").
					Reply(status).
					BodyString(expectedBody)

				reqFactory := Get(reqURLString)
				resp := Do(context.TODO(), reqFactory, withLogger(log))

				req, err := reqFactory()
				if !assert.NoError(t, err) {
					return
				}

				_, err = resp()
				if !assert.NoError(t, err) {
					return
				}

				reqURL, err := url.Parse(reqURLString)
				if !assert.NoError(t, err) {
					return
				}

				assert.Len(t, log.messages, 2)
				assert.Len(t, log.msgDataEntries, 2)
				assert.Equal(t, "SEND REQUEST START", log.messages[0])
				assert.Equal(t, diag.MsgData{
					"protocol": reqURL.Scheme,
					"url":      reqURLString,
					"qs":       flattenAndObfuscate(reqURL.Query()),
					"headers":  flattenAndObfuscate(req.Header),
					"method":   "GET",
				}, log.msgDataEntries[0])

				assert.Equal(t, "SEND REQUEST COMPLETE", log.messages[1])
				assert.Equal(t, diag.MsgData{
					"url":        reqURLString,
					"method":     "GET",
					"httpStatus": status,
				}, log.msgDataEntries[1])
			}
		},

		// TODO: Fail with http err if no 200
	}
	for _, tt := range tests {
		t.Run(tt())
	}
}
