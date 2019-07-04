package request

import (
	"context"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"gopkg.in/h2non/gock.v1"

	"github.com/stretchr/testify/assert"

	"github.com/bxcodec/faker/v3"
)

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

		// TODO: Fail with http err if no 200
	}
	for _, tt := range tests {
		t.Run(tt())
	}
}
