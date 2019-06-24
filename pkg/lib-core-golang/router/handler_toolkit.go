package router

import (
	"encoding/json"
	"net/http"
)

type handlerToolkit struct {
	request        *http.Request
	responseWriter http.ResponseWriter
	validator      *structValidator
	pathParamValue pathParamValueFunc
}

func (h *handlerToolkit) BindParams() *ParamsBinder {
	return &ParamsBinder{
		req:            h.request,
		validator:      h.validator,
		pathParamValue: h.pathParamValue,
	}
}

func (h *handlerToolkit) BindPayload(receiver interface{}) error {
	if err := json.NewDecoder(h.request.Body).Decode(&receiver); err != nil {
		return err
	}

	// Validator is failing to validate maps so have to ignore explicitly
	_, isMap := receiver.(*map[string]interface{})
	if isMap {
		return nil
	}

	return h.validator.validateStruct(h.request.Context(), receiver)
}

func (h *handlerToolkit) WriteJSON(payload interface{}, decorators ...ResponseDecorator) error {

	// This should go first. If we use WithStatus decorator then it will send the header
	// and adding new headers will make no difference
	h.responseWriter.Header().Add("content-type", "application/json")

	for _, decorator := range decorators {
		if err := decorator(h.responseWriter); err != nil {
			return err
		}
	}
	return json.NewEncoder(h.responseWriter).Encode(payload)
}

// WithStatus decorate response with particular http status
func (h *handlerToolkit) WithStatus(status int) ResponseDecorator {
	return func(w http.ResponseWriter) error {
		w.WriteHeader(status)
		return nil
	}
}
