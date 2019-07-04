package request

import (
	"context"
	"io/ioutil"
	"net/http"
	"testing"

	"gopkg.in/h2non/gock.v1"

	"github.com/stretchr/testify/assert"

	"github.com/bxcodec/faker/v3"
)

func TestDo(t *testing.T) {
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

				resp, err := Do(context.TODO(), Get(url))
				if !assert.NoError(t, err) {
					return
				}
				if !assert.True(t, gock.IsDone(), "No request performed") {
					return
				}
				defer resp.Body.Close()

				assert.Equal(t, 200, resp.StatusCode)

				actualBody, err := ioutil.ReadAll(resp.Body)
				if !assert.NoError(t, err) {
					return
				}
				assert.Equal(t, expectedBody, string(actualBody))
			}
		},

		// TODO: Fail with http err if no 200
	}
	for _, tt := range tests {
		t.Run(tt())
	}
}
