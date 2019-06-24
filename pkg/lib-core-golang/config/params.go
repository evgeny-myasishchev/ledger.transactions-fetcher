package config

type param interface {
	key() string
	service() string
	emptyValue() paramValue
}

type paramImpl struct {
	paramKey string
	paramSvc string
}

func (p paramImpl) emptyValue() paramValue {
	panic("not supported")
}

func (p paramImpl) key() string {
	return p.paramKey
}

func (p paramImpl) service() string {
	return p.paramSvc
}

func (p paramImpl) String() string {
	return "{key: " + p.paramKey + "; service: " + p.paramSvc + "}"
}

// StringParam represents params of string type
type StringParam struct {
	paramImpl
}

func newStringParam(key string, service string) StringParam {
	return StringParam{paramImpl{paramKey: key, paramSvc: service}}
}

func (p StringParam) emptyValue() paramValue {
	return StringVal{val: new(string)}
}

// IntParam represents params of int type
type IntParam struct {
	paramImpl
}

func newIntParam(key string, service string) IntParam {
	return IntParam{paramImpl{paramKey: key, paramSvc: service}}
}

func (p IntParam) emptyValue() paramValue {
	return IntVal{val: new(int)}
}

// BoolParam represents params of bool type
type BoolParam struct {
	paramImpl
}

func newBoolParam(key string, service string) BoolParam {
	return BoolParam{paramImpl{paramKey: key, paramSvc: service}}
}

func (p BoolParam) emptyValue() paramValue {
	return BoolVal{val: new(bool)}
}
