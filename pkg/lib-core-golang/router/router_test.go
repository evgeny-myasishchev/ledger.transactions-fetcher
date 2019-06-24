package router

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	tst "github.com/evgeny-myasishchev//pkg/internal/testing"

	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/assert"
)

func TestParamsBinder(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	type testCase struct {
		name string
		run  func(t *testing.T)
	}

	tests := []func() testCase{
		func() testCase {
			return testCase{
				name: "valid path params",
				run: func(t *testing.T) {
					param1 := fmt.Sprintf("param-1-%v", faker.Word())
					param1Val := fmt.Sprintf("param-1-val-%v", faker.Word())
					param2 := fmt.Sprintf("param-2-%v", faker.Word())
					param2Val := rand.Int()
					param3 := fmt.Sprintf("param-2-%v", faker.Word())
					param3Val := []string{faker.Word(), faker.Word()}

					type ctxVar string

					ctx := context.WithValue(context.Background(), ctxVar(param1), param1Val)
					ctx = context.WithValue(ctx, ctxVar(param2), strconv.Itoa(param2Val))
					ctx = context.WithValue(ctx, ctxVar(param3), strings.Join(param3Val, ","))
					req := httptest.NewRequest("GET", "/v1/some/api", nil).WithContext(ctx)

					binder := newParamsBinder(
						req,
						func(req *http.Request, name string) string {
							return req.Context().Value(ctxVar(name)).(string)
						},
					)
					var params struct {
						param1 string
						param2 int
						param3 []string
					}
					customValue := func(rawValue string) (interface{}, error) {
						return strings.Split(rawValue, ","), nil
					}
					err := binder.
						PathParam(param1).String(&params.param1).
						PathParam(param2).Int(&params.param2).
						PathParam(param3).Custom(&params.param3, customValue).
						Validate(&params)
					assert.Nil(t, err)
					assert.Equal(t, param1Val, params.param1)
					assert.Equal(t, param2Val, params.param2)
					assert.Equal(t, param3Val, params.param3)
				},
			}
		},
		func() testCase {
			return testCase{
				name: "valid query params",
				run: func(t *testing.T) {
					param1 := fmt.Sprintf("param-1-%v", faker.Word())
					param1Val := fmt.Sprintf("param-1-val-%v", faker.Word())
					param2 := fmt.Sprintf("param-2-%v", faker.Word())
					param2Val := rand.Int()
					param3 := fmt.Sprintf("param-2-%v", faker.Word())
					param3Val := []string{faker.Word(), faker.Word()}

					queryValues := url.Values{}
					queryValues.Add(param1, param1Val)
					queryValues.Add(param2, strconv.Itoa(param2Val))
					queryValues.Add(param3, strings.Join(param3Val, ","))

					req := httptest.NewRequest("GET", fmt.Sprintf("/v1/some/api?%v", queryValues.Encode()), nil)

					binder := newParamsBinder(req, nil)
					var params struct {
						param1 string
						param2 int
						param3 []string
					}
					customValue := func(rawValue string) (interface{}, error) {
						return strings.Split(rawValue, ","), nil
					}
					err := binder.
						QueryParam(param1).String(&params.param1).
						QueryParam(param2).Int(&params.param2).
						QueryParam(param3).Custom(&params.param3, customValue).
						Validate(&params)
					assert.Nil(t, err)
					assert.Equal(t, param1Val, params.param1)
					assert.Equal(t, param2Val, params.param2)
					assert.Equal(t, param3Val, params.param3)
				},
			}
		},
		func() testCase {
			return testCase{
				name: "bad int",
				run: func(t *testing.T) {
					paramName := fmt.Sprint("param-", faker.Word())
					binder := newParamsBinder(httptest.NewRequest("GET", "/", nil), nil)
					var receiver int
					param := binder.newParamBinder(PathParam, paramName, "not int")
					param.Int(&receiver)
					assert.Equal(t, ParamValidationError(PathParam, paramName), binder.err)
				},
			}
		},
		func() testCase {
			return testCase{
				name: "custom error",
				run: func(t *testing.T) {
					paramName := fmt.Sprint("param-", faker.Word())
					binder := newParamsBinder(httptest.NewRequest("GET", "/", nil), nil)
					var receiver int
					param := binder.newParamBinder(PathParam, paramName, "not int")
					param.Custom(&receiver, func(rawValue string) (interface{}, error) {
						return nil, errors.New("some error")
					})
					assert.Equal(t, ParamValidationError(PathParam, paramName), binder.err)
				},
			}
		},
		func() testCase {
			return testCase{
				name: "struct invalid",
				run: func(t *testing.T) {
					var params struct {
						Prop1 string `validate:"required"`
						Prop2 string `validate:"required"`
					}
					binder := newParamsBinder(httptest.NewRequest("GET", "/", nil), nil)
					err := binder.Validate(&params)
					assert.NotNil(t, err)
					assert.Equal(t, BadRequestError(fmt.Sprint("ValidationFailed: params [Prop1 Prop2] are invalid")), err)
				},
			}
		},
		func() testCase {
			return testCase{
				name: "validation failed",
				run: func(t *testing.T) {
					binder := newParamsBinder(httptest.NewRequest("GET", "/", nil), nil)
					err := binder.Validate(nil)
					assert.NotNil(t, err)
					assert.Equal(t, BadRequestError(fmt.Sprint("ValidationFailed: failed to validate params")), err)
				},
			}
		},
		func() testCase {
			return testCase{
				name: "do not bind and return err",
				run: func(t *testing.T) {
					queryValues := url.Values{}
					queryValues.Add("param1", "param1")
					queryValues.Add("param2", "2")
					queryValues.Add("param3", "param3")
					var params struct {
						param1 string
						param2 int
						param3 string
					}

					req := httptest.NewRequest("GET", fmt.Sprintf("/v1/some/api?%v", queryValues.Encode()), nil)
					err := errors.New("binder error")
					binder := newParamsBinder(req, nil)
					binder.err = err

					customValue := func(rawValue string) (interface{}, error) {
						return rawValue, nil
					}
					gotErr := binder.
						QueryParam("param1").String(&params.param1).
						QueryParam("param2").Int(&params.param2).
						QueryParam("param3").Custom(&params.param3, customValue).
						Validate(nil)
					assert.Equal(t, err, gotErr)

					assert.Empty(t, params.param1)
					assert.Empty(t, params.param2)
					assert.Empty(t, params.param3)
				},
			}
		},
		func() testCase {
			return testCase{
				name: "default value",
				run: func(t *testing.T) {
					req := httptest.NewRequest("GET", "/", nil)
					binder := newParamsBinder(req, nil)

					defaultVal := faker.Word()
					param := binder.newParamBinder(PathParam, "param1", "").Default(defaultVal)
					assert.Equal(t, defaultVal, param.rawValue)

					nonDefaultVal := faker.Word()
					param = binder.newParamBinder(PathParam, "param1", nonDefaultVal).Default(defaultVal)
					assert.Equal(t, nonDefaultVal, param.rawValue)
				},
			}
		},
	}

	for _, ttFn := range tests {
		tt := ttFn()
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestToolkitHandlerFunc_ServeHTTP(t *testing.T) {
	type args struct {
		validator      *structValidator
		pathParamValue pathParamValueFunc
	}
	type testCase struct {
		name string
		args args
		run  func(*testing.T, *http.Request, *httptest.ResponseRecorder)
	}
	tests := []func() testCase{
		func() testCase {
			validator := newStructValidator()
			fakePathParamVal := faker.Word()
			pathParamValue := pathParamValueFunc(func(req *http.Request, name string) string {
				return fakePathParamVal
			})
			return testCase{
				name: "invoke handler with toolkit",
				args: args{validator: validator, pathParamValue: pathParamValue},
				run: func(t *testing.T, req *http.Request, recorder *httptest.ResponseRecorder) {
					fnCalled := false
					fn := ToolkitHandlerFunc(func(w http.ResponseWriter, r *http.Request, h HandlerToolkit) error {
						fnCalled = true
						assert.Equal(t, req, r)
						assert.Equal(t, w, recorder)

						toolkit := h.(*handlerToolkit)
						if !assert.NotNil(t, toolkit) {
							return nil
						}

						assert.Equal(t, r, toolkit.request)
						assert.Equal(t, w, toolkit.responseWriter)
						assert.Equal(t, validator, toolkit.validator)
						if !assert.NotNil(t, toolkit.pathParamValue) {
							return nil
						}
						assert.Equal(t, fakePathParamVal, toolkit.pathParamValue(r, "fake"))

						return nil
					})
					fn.ServeHTTP(recorder, req)
					assert.True(t, fnCalled)
				},
			}
		},
		func() testCase {
			validator := newStructValidator()
			return testCase{
				name: "handle errors",
				args: args{validator: validator},
				run: func(t *testing.T, req *http.Request, recorder *httptest.ResponseRecorder) {
					fnCalled := false
					err := errors.New(faker.Sentence())
					fn := ToolkitHandlerFunc(func(w http.ResponseWriter, r *http.Request, h HandlerToolkit) error {
						fnCalled = true
						return err
					})
					fn.ServeHTTP(recorder, req)
					assert.True(t, fnCalled)
					httpErr := newHTTPErrorFromError(err)
					tst.AssertHTTPErrorResponse(t, tst.NewHTTPErrorPayload(
						httpErr.StatusCode,
						httpErr.Status,
						httpErr.Message,
					), recorder)
				},
			}
		},
	}
	for _, tt := range tests {
		tt := tt()
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/some-path", nil)
			recorder := httptest.NewRecorder()

			nextContext := context.WithValue(req.Context(), validatorRequestKey, tt.args.validator)
			nextContext = context.WithValue(nextContext, pathParamValueFuncKey, tt.args.pathParamValue)
			req = req.WithContext(nextContext)

			tt.run(t, req, recorder)
		})
	}
}
