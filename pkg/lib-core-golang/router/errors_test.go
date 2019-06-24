package router

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	tst "github.com/evgeny-myasishchev//pkg/internal/testing"

	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/assert"
)

func Test_newHTTPErrorFromError(t *testing.T) {
	type args struct {
		err error
	}
	type testCase struct {
		name string
		args args
		want HTTPError
	}
	tests := []func() testCase{
		func() testCase {
			err := errors.New(faker.Sentence())
			return testCase{
				name: "generic server error",
				args: args{err: err},
				want: HTTPError{
					StatusCode: http.StatusInternalServerError,
					Status:     http.StatusText(http.StatusInternalServerError),
					Message:    err.Error(),
				},
			}
		},
		func() testCase {
			message := faker.Sentence()
			err := ResourceNotFoundError(message)
			return testCase{
				name: "http error",
				args: args{err: err},
				want: err.(HTTPError),
			}
		},
	}
	for _, ttFn := range tests {
		tt := ttFn()
		t.Run(tt.name, func(t *testing.T) {
			got := newHTTPErrorFromError(tt.args.err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestErrorConstructors(t *testing.T) {
	type testCase struct {
		name  string
		build func() error
		want  HTTPError
	}
	tests := []func() testCase{
		func() testCase {
			message := faker.Sentence()
			statusCode := 400 + rand.Intn(10)
			return testCase{
				name: "generic",
				build: func() error {
					return NewHTTPError(statusCode, message)
				},
				want: HTTPError{
					StatusCode: statusCode,
					Status:     http.StatusText(statusCode),
					Message:    message,
				},
			}
		},
		func() testCase {
			message := faker.Sentence()
			return testCase{
				name: "not found",
				build: func() error {
					return ResourceNotFoundError(message)
				},
				want: HTTPError{
					StatusCode: http.StatusNotFound,
					Status:     http.StatusText(http.StatusNotFound),
					Message:    message,
				},
			}
		},
		func() testCase {
			message := faker.Sentence()
			return testCase{
				name: "bad request",
				build: func() error {
					return BadRequestError(message)
				},
				want: HTTPError{
					StatusCode: http.StatusBadRequest,
					Status:     http.StatusText(http.StatusBadRequest),
					Message:    message,
				},
			}
		},
		func() testCase {
			paramName := faker.Word()
			paramType := faker.Word()
			return testCase{
				name: "param validation error",
				build: func() error {
					return ParamValidationError(RequestParamType(paramType), paramName)
				},
				want: HTTPError{
					StatusCode: http.StatusBadRequest,
					Status:     http.StatusText(http.StatusBadRequest),
					Message:    fmt.Sprint("ValidationFailed: ", paramType, " parameter '", paramName, "' is invalid"),
				},
			}
		},
	}
	for _, ttFn := range tests {
		tt := ttFn()
		t.Run(tt.name, func(t *testing.T) {
			got := tt.build()
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_errorResponse_Error(t *testing.T) {
	message := faker.Sentence()
	statusCode := 400 + rand.Intn(10)
	statusText := http.StatusText(statusCode)
	err := HTTPError{
		StatusCode: statusCode,
		Status:     statusText,
		Message:    message,
	}

	actual := err.Error()
	assert.Equal(t, fmt.Sprintf("[%v](%v): %v", statusCode, statusText, message), actual)
}

func TestHTTPError_Send(t *testing.T) {
	type args struct {
		err HTTPError
	}
	type testCase struct {
		name string
		args args
		want func(t *testing.T, recorder *httptest.ResponseRecorder)
	}
	tests := []func() testCase{
		func() testCase {
			statusCode := 400 + rand.Intn(10)
			err := HTTPError{
				StatusCode: statusCode,
				Status:     http.StatusText(statusCode),
				Message:    faker.Sentence(),
			}
			return testCase{
				name: "write http error",
				args: args{err: err},
				want: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					assert.Equal(t, statusCode, recorder.Code)
					assert.Equal(t, "application/json", recorder.Header().Get("content-type"))

					var got HTTPError
					if !tst.JSONUnmarshalReader(t, recorder.Body, &got) {
						return
					}
					assert.Equal(t, err, got)
				},
			}
		},
	}
	for _, tt := range tests {
		tt := tt()
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			tt.args.err.Send(w)
			tt.want(t, w)
		})
	}
}
