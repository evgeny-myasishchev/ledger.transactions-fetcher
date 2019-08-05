package config

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type valueMapper map[string]func(val interface{}, target reflect.Value) error

func (m valueMapper) setValue(val interface{}, target reflect.Value) error {
	setter, ok := m[target.Type().Name()]
	if !ok {
		return fmt.Errorf("Type %v is not supported", target.Type())
	}
	return setter(val, target)
}

var supportedParamTypes = valueMapper{
	"string": func(val interface{}, target reflect.Value) error {
		strVal, ok := val.(string)
		if !ok {
			return fmt.Errorf("Expected string value but got: %v(%[1]T)", val)
		}
		target.Set(reflect.ValueOf(strVal))
		return nil
	},
	"int": func(val interface{}, target reflect.Value) error {
		var intVal int
		var err error
		switch actualVal := val.(type) {
		case int:
			intVal = actualVal
		case float32:
			intVal = int(actualVal)
		case float64:
			intVal = int(actualVal)
		case string:
			intVal, err = strconv.Atoi(actualVal)
		default:
			err = errors.New("Unexpected type")
		}
		if err != nil {
			return fmt.Errorf("Expected int value but got: %v(%[1]T)", val)
		}
		target.Set(reflect.ValueOf(intVal))
		return nil
	},
	"bool": func(val interface{}, target reflect.Value) error {
		var boolVal bool
		var err error
		switch newVal := val.(type) {
		case bool:
			boolVal = newVal
		case string:
			boolVal, err = strconv.ParseBool(newVal)
		}
		if err != nil {
			return fmt.Errorf("Expected bool value but got: %v(%[1]T)", val)
		}
		target.Set(reflect.ValueOf(boolVal))
		return nil
	},
}

type paramID struct {
	key     string
	service string
}

type param struct {
	paramID
	source string
	value  reflect.Value
}

// MarshalJSON is defined to have params logged properly
func (p param) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	buffer.WriteString(`"key":"`)
	buffer.WriteString(p.key)
	buffer.WriteString(`",`)

	buffer.WriteString(`"service":"`)
	buffer.WriteString(p.service)
	buffer.WriteString(`"`)

	buffer.WriteString("}")
	return buffer.Bytes(), nil
}

func (p param) setValue(val interface{}) error {
	return supportedParamTypes.setValue(val, p.value)
}

func parseTag(tag string, defaultService string) map[string]string {
	result := map[string]string{}
	if tag == "" {
		return result
	}
	for _, subTag := range strings.Split(tag, ",") {
		parts := strings.Split(subTag, "=")
		result[parts[0]] = parts[1]
	}
	if _, ok := result["service"]; !ok {
		result["service"] = defaultService
	}
	return result
}

func getStructPtrElem(t reflect.Type, fieldName string) (reflect.Type, error) {
	if t.Kind() != reflect.Ptr {
		return nil, errors.New("Expected " + fieldName + " to be a struct pointer, got " + t.Kind().String())
	}

	elem := t.Elem()
	if elem.Kind() != reflect.Struct {
		return nil, fmt.Errorf("Expected "+fieldName+" to be a struct, got %v", elem.Kind())
	}
	return elem, nil
}

func bindParamsToReceiver(receiver interface{}, defaultService string) ([]param, error) {
	receiverValPtr := reflect.ValueOf(receiver)
	receiverType, err := getStructPtrElem(receiverValPtr.Type(), "receiver")
	if err != nil {
		return nil, err
	}

	receiverVal := reflect.Indirect(receiverValPtr)

	params := []param{}

	for fieldIndex := 0; fieldIndex < receiverVal.NumField(); fieldIndex++ {
		sourceBoundField := receiverType.Field(fieldIndex)
		sourceBoundFieldType, err := getStructPtrElem(sourceBoundField.Type, sourceBoundField.Name)
		if err != nil {
			return nil, err
		}

		sourceBoundValPtr := receiverVal.Field(fieldIndex)

		if !sourceBoundValPtr.CanSet() {
			return nil, fmt.Errorf("Expected %v to be exported", sourceBoundField.Name)
		}
		sourceBoundValPtr.Set(reflect.New(sourceBoundFieldType))
		sourceBoundVal := reflect.Indirect(sourceBoundValPtr)

		recConfigTagStr := sourceBoundField.Tag.Get("config")
		recTags := parseTag(recConfigTagStr, defaultService)
		source, ok := recTags["source"]
		if !ok {
			return nil, fmt.Errorf("Expected %v to have source binding tag, got: %v", sourceBoundField.Name, recConfigTagStr)
		}

		for paramIndex := 0; paramIndex < sourceBoundFieldType.NumField(); paramIndex++ {
			paramField := sourceBoundFieldType.Field(paramIndex)
			configTag := paramField.Tag.Get("config")
			paramTags := parseTag(configTag, defaultService)
			key, ok := paramTags["key"]
			if !ok {
				return nil, fmt.Errorf("Expected %v.%v to have at least key tag, got: %v", sourceBoundField.Name, paramField.Name, configTag)
			}

			var paramSource string
			var hasSource bool
			if paramSource, hasSource = paramTags["source"]; !hasSource {
				paramSource = source
			}

			propVal := sourceBoundVal.Field(paramIndex)
			if !propVal.CanSet() {
				return nil, fmt.Errorf("Expected %v.%v to be exported", sourceBoundField.Name, paramField.Name)
			}

			params = append(params, param{
				paramID: paramID{
					key:     key,
					service: paramTags["service"],
				},
				source: paramSource,
				value:  propVal,
			})
		}
	}

	return params, nil
}
