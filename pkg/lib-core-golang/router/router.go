package router

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	"gopkg.in/go-playground/validator.v9"

	"github.com/evgeny-myasishchev//pkg/lib-core-golang/diag"
)

var logger = diag.CreateLogger()

type contextKey string

const (
	validatorRequestKey   contextKey = "validator"
	pathParamValueFuncKey contextKey = "path-param-value-func"
)

// RequestParamType represents type of a request parameter
type RequestParamType string

const (
	// PathParam is a request path parameter type
	PathParam RequestParamType = "path"

	// QueryParam is a request query parameter type
	QueryParam RequestParamType = "query"
)

type structValidator validator.Validate

func newStructValidator() *structValidator {
	return (*structValidator)(validator.New())
}

// TODO: Currently tested indirectly. Dedicated tests required
// TODO: Better messages are required
func (v *structValidator) validateStruct(ctx context.Context, target interface{}) error {
	vdt := (*validator.Validate)(v)
	if err := vdt.Struct(target); err != nil {
		logger.WithError(err).Info(ctx, "Failed to validate params")
		if err, ok := err.(validator.ValidationErrors); ok {
			badFields := make([]string, 0, len(err))
			for _, fieldErr := range err {
				badFields = append(badFields, fieldErr.Field())
			}
			return BadRequestError(fmt.Sprint("ValidationFailed: params ", badFields, " are invalid"))
		}
		return BadRequestError("ValidationFailed: failed to validate params")
	}
	return nil
}

type pathParamValueFunc func(req *http.Request, name string) string

// ParamsBinder binds request params to values
type ParamsBinder struct {
	req            *http.Request
	err            error
	validator      *structValidator
	pathParamValue pathParamValueFunc
}

func newParamsBinder(req *http.Request, pathParamValue pathParamValueFunc) *ParamsBinder {
	// TODO: Should take validator as well
	v := validator.New()
	return &ParamsBinder{req: req, validator: (*structValidator)(v), pathParamValue: pathParamValue}
}

func (b *ParamsBinder) newParamBinder(paramType RequestParamType, name string, rawValue string) *ParamBinder {
	return &ParamBinder{paramType: paramType, name: name, rawValue: rawValue, binder: b}
}

// PathParam binds param from request path
func (b *ParamsBinder) PathParam(name string) *ParamBinder {
	rawValue := b.pathParamValue(b.req, name)
	return b.newParamBinder(PathParam, name, rawValue)
}

// QueryParam binds param from request query
func (b *ParamsBinder) QueryParam(name string) *ParamBinder {
	rawValue := b.req.URL.Query().Get(name)
	return b.newParamBinder(QueryParam, name, rawValue)
}

// ParamBinder binds particular param
type ParamBinder struct {
	paramType RequestParamType
	name      string
	rawValue  string
	binder    *ParamsBinder
}

// Validate will validate exposed fields of a target structure.
// See https://godoc.org/gopkg.in/go-playground/validator.v9 for more details
func (b *ParamsBinder) Validate(target interface{}) error {
	if b.err != nil {
		return b.err
	}

	return b.validator.validateStruct(b.req.Context(), target)
}

// Default assign param default value
func (pb *ParamBinder) Default(value string) *ParamBinder {
	if pb.rawValue == "" {
		pb.rawValue = value
	}
	return pb
}

// Int bind param as int
func (pb *ParamBinder) Int(receiver *int) *ParamsBinder {
	if pb.binder.err != nil {
		return pb.binder
	}
	if value, err := strconv.Atoi(pb.rawValue); err != nil {
		logger.WithError(err).Info(pb.binder.req.Context(), "Failed to parse %v param %v", pb.paramType, pb.name)
		pb.binder.err = ParamValidationError(pb.paramType, pb.name)
	} else {
		*receiver = value
	}
	return pb.binder
}

// String bind param as string
func (pb *ParamBinder) String(receiver *string) *ParamsBinder {
	if pb.binder.err != nil {
		return pb.binder
	}
	*receiver = pb.rawValue
	return pb.binder
}

// CustomValue is a function that converts raw string to a target value
type CustomValue func(rawValue string) (interface{}, error)

// Custom binds custom valuse
func (pb *ParamBinder) Custom(receiver interface{}, valueFn CustomValue) *ParamsBinder {
	if pb.binder.err != nil {
		return pb.binder
	}
	if value, err := valueFn(pb.rawValue); err != nil {
		logger.WithError(err).Info(pb.binder.req.Context(), "Failed to bind custom %v param %v", pb.paramType, pb.name)
		pb.binder.err = ParamValidationError(pb.paramType, pb.name)
	} else {
		reflect.ValueOf(receiver).Elem().Set(reflect.ValueOf(value))
	}
	return pb.binder
}

// ResponseDecorator is a helper function to decorate response
type ResponseDecorator func(w http.ResponseWriter) error

// HandlerToolkit - Collection of various tools to help processing request and build a response
type HandlerToolkit interface {
	BindParams() *ParamsBinder
	BindPayload(receiver interface{}) error

	// WriteJSON will serialize the payload and write it to the response
	// Optionally use decorators, for example WithStatus
	WriteJSON(payload interface{}, decorators ...ResponseDecorator) error

	// WithStatus is a decorator function that will set particular http status
	// used togeather with WriteJSON
	WithStatus(status int) ResponseDecorator
}

// ToolkitHandlerFunc - a little extension of a builtin HandlerFunc
type ToolkitHandlerFunc func(w http.ResponseWriter, req *http.Request, h HandlerToolkit) error

// ServeHTTP is an implementation of http.Handler. This allows ToolkitHandlerFunc to be used
// in place of the http.Handler
func (f ToolkitHandlerFunc) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	toolkit := handlerToolkit{
		request:        req,
		responseWriter: w,
		validator:      req.Context().Value(validatorRequestKey).(*structValidator),
		pathParamValue: req.Context().Value(pathParamValueFuncKey).(pathParamValueFunc),
	}
	err := f(w, req, &toolkit)
	if err != nil {
		logger.WithError(err).Error(req.Context(), "Failed to process request")
		errorResponse := newHTTPErrorFromError(err)
		errorResponse.Send(w)
	}
}

// MiddlewareFunc is a function that can be injected into a request chain
type MiddlewareFunc func(next http.Handler) http.Handler

// Router is a layer to abstract underlying http router implementation
// so we could swap it with relatively low efforts
type Router interface {
	Handle(method string, pattern string, handler http.Handler)

	// TODO: Build 404 middleware
	// TODO: Build no-panic middleware (e.g respond with consistent 500 error)
	Use(mw MiddlewareFunc)

	/*
		pathParam returns the bound parameter with the given name.
		Suppose we have a route pattern:

			/v1/users/:id

		and the URL Path:

			/v1/users/100

		in this case pathParam will return 100
	*/
	pathParam(r *http.Request, name string) string

	ServeHTTP(http.ResponseWriter, *http.Request)
}

// CreateRouter returns default router implementation
func CreateRouter() Router {
	router := createGojiRouter()
	router.Use(MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCtx := context.WithValue(r.Context(), validatorRequestKey, newStructValidator())
			nextCtx = context.WithValue(nextCtx, pathParamValueFuncKey, pathParamValueFunc(router.pathParam))
			nextReq := r.WithContext(nextCtx)
			next.ServeHTTP(w, nextReq)
		})
	}))
	return router
}

// StartServer start the server with setup router function
func StartServer(port int, setup func(r Router)) error {
	router := CreateRouter()
	setup(router)
	err := http.ListenAndServe(fmt.Sprintf(":%v", port), router)
	return err
}
