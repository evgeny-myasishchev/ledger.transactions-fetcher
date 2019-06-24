package router

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// HTTPError represents a generic http error structure
type HTTPError struct {
	StatusCode int    `json:"statusCode"`
	Status     string `json:"error"`
	Message    string `json:"message"`
}

func (e HTTPError) Error() string {
	return fmt.Sprintf("[%v](%v): %v", e.StatusCode, e.Status, e.Message)
}

// Send will marshal and send the error response to the client
// panic if failed to send
func (e HTTPError) Send(w http.ResponseWriter) {
	errorData, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(e.StatusCode)
	if _, err := w.Write(errorData); err != nil {
		panic(err)
	}
}

// NewHTTPError - creates a generic http error
func NewHTTPError(statusCode int, message string) error {
	return HTTPError{
		StatusCode: statusCode,
		Status:     http.StatusText(statusCode),
		Message:    message,
	}
}

// ResourceNotFoundError a standard 404 error
func ResourceNotFoundError(message string) error {
	return NewHTTPError(http.StatusNotFound, message)
}

// BadRequestError a standard 400 error
func BadRequestError(message string) error {
	return NewHTTPError(http.StatusBadRequest, message)
}

// ParamValidationError a bad request error related to params validation
func ParamValidationError(paramType RequestParamType, paramName string) error {
	return BadRequestError(fmt.Sprint("ValidationFailed: ", paramType, " parameter '", paramName, "' is invalid"))
}

func newHTTPErrorFromError(err error) HTTPError {
	if errResp, ok := err.(HTTPError); ok {
		return errResp
	}
	return HTTPError{
		StatusCode: http.StatusInternalServerError,
		Status:     http.StatusText(http.StatusInternalServerError),

		// TODO: Potentially do not expose this, or at least not for prod
		// A config could be passed via context
		Message: err.Error(),
	}
}
