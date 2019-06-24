package router

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	tst "github.com/evgeny-myasishchev//pkg/internal/testing"
	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/assert"
)

func TestHandlerToolkit(t *testing.T) {
	newReq := func(method string, pattern string) *http.Request {
		req, err := http.NewRequest("GET", pattern, nil)
		if err != nil {
			panic(err)
		}
		return req
	}

	router := CreateRouter()

	jsonPayload := map[string]interface{}{
		"key1": faker.Word(),
		"key2": faker.Word(),
		"key3": faker.Word(),
	}

	handlerCalled := false
	paramValue := faker.Word()
	router.Handle("GET", "/v1/route/:param",
		ToolkitHandlerFunc(func(w http.ResponseWriter, req *http.Request, h HandlerToolkit) error {
			assert.NotNil(t, h, "handler toolkit should have been provided")

			paramsBinder := h.BindParams()
			assert.NotNil(t, paramsBinder)
			assert.Equal(t, req, paramsBinder.req)
			assert.NotNil(t, paramsBinder.pathParamValue)
			assert.Equal(t, paramValue, paramsBinder.pathParamValue(req, "param"))

			handlerCalled = true

			return h.WriteJSON(jsonPayload)
		}))

	req := newReq("GET", fmt.Sprintf("/v1/route/%v", paramValue))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.True(t, handlerCalled, "handler should have been called")

	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	actualPayload := map[string]interface{}{}
	tst.JSONUnmarshalBuffer(w.Body, &actualPayload)
	assert.Equal(t, jsonPayload, actualPayload)
}

func Test_HandlerToolkit_BindPayload(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	type fields struct {
		body io.Reader
	}
	type testCase struct {
		name   string
		fields fields
		assert func(t *testing.T, h HandlerToolkit)
	}
	tests := []func() testCase{
		func() testCase {
			want := map[string]interface{}{
				"key1": faker.Word(),
				"key2": rand.Float64(),
			}
			return testCase{
				name: "bind json payload map",
				assert: func(t *testing.T, h HandlerToolkit) {
					var got map[string]interface{}
					err := h.BindPayload(&got)
					if assert.NoError(t, err) {
						assert.Equal(t, want, got)
					}
				},
				fields: fields{
					body: tst.JSONMarshalToReader(&want),
				},
			}
		},
		func() testCase {
			type payload struct {
				Key1 string  `json:"key1"`
				Key2 float64 `json:"key2"`
			}

			want := payload{
				Key1: faker.Word(),
				Key2: rand.Float64(),
			}
			return testCase{
				name: "bind json payload struct",
				assert: func(t *testing.T, h HandlerToolkit) {
					var got payload
					err := h.BindPayload(&got)
					if assert.NoError(t, err) {
						assert.Equal(t, want, got)
					}
				},
				fields: fields{
					body: tst.JSONMarshalToReader(&want),
				},
			}
		},
		func() testCase {
			return testCase{
				name: "fail if bad json",
				assert: func(t *testing.T, h HandlerToolkit) {
					var p interface{}
					err := h.BindPayload(&p)
					if assert.Error(t, err) {
						assert.Contains(t, err.Error(), "invalid character")
					}
				},
				fields: fields{
					body: strings.NewReader(faker.Word()),
				},
			}
		},
		func() testCase {
			type payload struct {
				Key1 string `json:"key1" validate:"required"`
				Key2 string `json:"key2" validate:"email"`
			}
			p := payload{
				Key2: faker.Word(),
			}
			return testCase{
				name: "fail if invalid json",
				fields: fields{
					body: tst.JSONMarshalToReader(p),
				},
				assert: func(t *testing.T, h HandlerToolkit) {
					var p payload
					err := h.BindPayload(&p)
					if assert.Error(t, err) {
						httpErr, ok := err.(HTTPError)
						if !assert.True(t, ok, "Unexpected error type: ", err.Error()) {
							return
						}
						assert.Equal(t, httpErr.StatusCode, http.StatusBadRequest)

						// TODO: Should be "payload properties"
						assert.Contains(t, httpErr.Message, "params [Key1 Key2] are invalid")
					}
				},
			}
		},
	}
	for _, ttFn := range tests {
		tt := ttFn()
		t.Run(tt.name, func(t *testing.T) {
			h := &handlerToolkit{
				request:   httptest.NewRequest("POST", "/", tt.fields.body),
				validator: newStructValidator(),
			}
			tt.assert(t, h)
		})
	}
}

func Test_HandlerToolkit_WriteJSON(t *testing.T) {
	type args struct {
		payload    interface{}
		decorators []ResponseDecorator
	}
	type testCase struct {
		name   string
		args   args
		assert func(t *testing.T, recorder *httptest.ResponseRecorder, err error)
	}
	tests := []func() testCase{
		func() testCase {
			payload := map[string]interface{}{
				"key1": faker.Word(),
				"key2": faker.Word(),
			}
			return testCase{
				name: "write",
				args: args{payload: payload},
				assert: func(t *testing.T, recorder *httptest.ResponseRecorder, err error) {
					if !assert.NoError(t, err) {
						return
					}
					var actualPayload map[string]interface{}
					tst.JSONUnmarshalBuffer(recorder.Body, &actualPayload)
					assert.Equal(t, payload, actualPayload)
				},
			}
		},
		func() testCase {
			payload := map[string]interface{}{
				"key1": faker.Word(),
				"key2": faker.Word(),
			}
			header1Val := faker.Word()
			header2Val := faker.Word()
			return testCase{
				name: "write with decorator",
				args: args{
					payload: payload,
					decorators: []ResponseDecorator{
						func(w http.ResponseWriter) error {
							w.Header().Add("header1", header1Val)
							return nil
						},
						func(w http.ResponseWriter) error {
							w.Header().Add("header2", header2Val)
							return nil
						},
					},
				},
				assert: func(t *testing.T, recorder *httptest.ResponseRecorder, err error) {
					if !assert.NoError(t, err) {
						return
					}
					var actualPayload map[string]interface{}
					tst.JSONUnmarshalBuffer(recorder.Body, &actualPayload)
					assert.Equal(t, payload, actualPayload)

					assert.Equal(t, header1Val, recorder.Header().Get("header1"))
					assert.Equal(t, header2Val, recorder.Header().Get("header2"))
				},
			}
		},
		func() testCase {
			decoratorErr := errors.New(faker.Sentence())
			return testCase{
				name: "write with decorator error",
				args: args{
					payload: map[string]interface{}{},
					decorators: []ResponseDecorator{
						func(w http.ResponseWriter) error {
							return decoratorErr
						},
					},
				},
				assert: func(t *testing.T, recorder *httptest.ResponseRecorder, err error) {
					assert.EqualError(t, err, decoratorErr.Error())
				},
			}
		},
	}
	for _, ttFn := range tests {
		tt := ttFn()
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			h := HandlerToolkit(&handlerToolkit{
				request:        httptest.NewRequest("GET", "/", nil),
				responseWriter: recorder,
			})
			err := h.WriteJSON(tt.args.payload, tt.args.decorators...)
			tt.assert(t, recorder, err)
		})
	}
}

func Test_HandlerToolkit_Decorator(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	type args struct {
		decorator func(h HandlerToolkit) ResponseDecorator
	}

	type testCase struct {
		name   string
		args   args
		assert func(recorder *httptest.ResponseRecorder, err error)
	}

	tests := []func() testCase{
		func() testCase {
			status := rand.Int()
			return testCase{
				name: "status",
				args: args{
					decorator: func(h HandlerToolkit) ResponseDecorator {
						return h.WithStatus(status)
					},
				},
				assert: func(recorder *httptest.ResponseRecorder, err error) {
					assert.NoError(t, err)
					assert.Equal(t, status, recorder.Code)
				},
			}
		},
	}
	for _, ttFn := range tests {
		tt := ttFn()
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			h := HandlerToolkit(&handlerToolkit{
				request:        httptest.NewRequest("GET", "/", nil),
				responseWriter: recorder,
			})
			decorator := tt.args.decorator(h)
			err := decorator(recorder)
			tt.assert(recorder, err)
		})
	}
}
