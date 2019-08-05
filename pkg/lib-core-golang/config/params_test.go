package config

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"

	"github.com/stretchr/testify/assert"
)

func Test_bindParamsToReceiver(t *testing.T) {
	type args struct {
		receiver interface{}
	}
	type ttFn func(*testing.T)

	tests := []func() (string, ttFn){
		func() (string, ttFn) {
			return "valid tags", func(t *testing.T) {
				defaultService := "default-svc-" + faker.Word()
				type params1 struct {
					StrParam1 string `config:"key=str-param11"`
					StrParam2 string `config:"key=str-param12,service=svc1"`
				}
				type params2 struct {
					StrParam1 string `config:"key=str-param21"`
					StrParam2 string `config:"key=str-param22,service=svc2"`
				}
				type receiver struct {
					Params1 *params1 `config:"source=source1"`
					Params2 *params2 `config:"source=source2"`
				}

				var rec receiver

				params, err := bindParamsToReceiver(&rec, defaultService)
				if !assert.NoError(t, err) {
					return
				}
				if !assert.Len(t, params, 4) {
					return
				}
				paramsByKey := map[string]param{}
				for _, param := range params {
					paramsByKey[param.key] = param
				}
				assert.Equal(t, "source1", paramsByKey["str-param11"].source)
				assert.Equal(t, defaultService, paramsByKey["str-param11"].service)
				assert.Equal(t, "svc1", paramsByKey["str-param12"].service)
				assert.Equal(t, defaultService, paramsByKey["str-param21"].service)
				assert.Equal(t, "source2", paramsByKey["str-param21"].source)
				assert.Equal(t, "svc2", paramsByKey["str-param22"].service)
			}
		},
		func() (string, ttFn) {
			return "source overrides", func(t *testing.T) {
				defaultService := "default-svc-" + faker.Word()
				type params1 struct {
					StrParam1 string `config:"key=str-param11,source=source3"`
					StrParam2 string `config:"key=str-param12"`
				}
				type params2 struct {
					StrParam1 string `config:"key=str-param21"`
					StrParam2 string `config:"key=str-param22,source=source4"`
				}
				type receiver struct {
					Params1 *params1 `config:"source=source1"`
					Params2 *params2 `config:"source=source2"`
				}

				var rec receiver

				params, err := bindParamsToReceiver(&rec, defaultService)
				if !assert.NoError(t, err) {
					return
				}
				if !assert.Len(t, params, 4) {
					return
				}
				paramsByKey := map[string]param{}
				for _, param := range params {
					paramsByKey[param.key] = param
				}
				assert.Equal(t, "source3", paramsByKey["str-param11"].source)
				assert.Equal(t, "source1", paramsByKey["str-param12"].source)
				assert.Equal(t, "source2", paramsByKey["str-param21"].source)
				assert.Equal(t, "source4", paramsByKey["str-param22"].source)
			}
		},
		func() (string, ttFn) {
			return "initialize source bound fields", func(t *testing.T) {
				type params1 struct {
					StrParam1 string `config:"key=str-param11"`
					StrParam2 string `config:"key=str-param12,service=svc1"`
				}
				type receiver struct {
					Params1 *params1 `config:"source=source1"`
					Params2 *struct {
					} `config:"source=source2"`
				}

				rec := receiver{}

				_, err := bindParamsToReceiver(&rec, "default-svc")
				if !assert.NoError(t, err) {
					return
				}
				assert.NotNil(t, rec.Params1)
				assert.NotNil(t, rec.Params2)
			}
		},
		func() (string, ttFn) {
			return "receiver is not struct ptr", func(t *testing.T) {
				type somethign struct{}
				_, err := bindParamsToReceiver(somethign{}, "default-svc")
				if !assert.Error(t, err) {
					return
				}
				assert.EqualError(t, err, "Expected receiver to be a struct pointer, got struct")

				rec := 10
				_, err = bindParamsToReceiver(&rec, "default-svc")
				if !assert.Error(t, err) {
					return
				}
				assert.EqualError(t, err, fmt.Sprintf("Expected receiver to be a struct, got %T", rec))
			}
		},
		func() (string, ttFn) {
			return "receiver has fields that are not struct pointers", func(t *testing.T) {
				type someCfg struct {
					Field11 *struct{} `config:"source=xxx"`
					Field12 int
				}
				_, err := bindParamsToReceiver(&someCfg{}, "default-svc")
				if !assert.Error(t, err) {
					return
				}
				assert.EqualError(t, err, "Expected Field12 to be a struct pointer, got int")

				type someCfg2 struct {
					Field21 *struct{} `config:"source=xxx"`
					Field22 *bool
				}
				_, err = bindParamsToReceiver(&someCfg2{}, "default-svc")
				if !assert.Error(t, err) {
					return
				}
				assert.EqualError(t, err, "Expected Field22 to be a struct, got bool")
			}
		},
		func() (string, ttFn) {
			return "receiver fields don't have source bindings", func(t *testing.T) {
				type someCfg struct {
					Field11 *struct{} `config:"source=xxx"`
					Field12 *struct{} `config:"not-source=xxx,key=yyy"`
				}
				_, err := bindParamsToReceiver(&someCfg{}, "default-svc")
				if !assert.Error(t, err) {
					return
				}
				assert.EqualError(t, err, "Expected Field12 to have source binding tag, got: not-source=xxx,key=yyy")
			}
		},
		func() (string, ttFn) {
			return "no param binding tags", func(t *testing.T) {
				type someCfg struct {
					Field11 *struct {
						Prop11 string `config:"key=prop1"`
						Prop12 string `config:"not-key=prop2"`
					} `config:"source=xxx"`
				}
				_, err := bindParamsToReceiver(&someCfg{}, "default-svc")
				if !assert.Error(t, err) {
					return
				}
				assert.EqualError(t, err, "Expected Field11.Prop12 to have at least key tag, got: not-key=prop2")

				type someCfg2 struct {
					Field21 *struct {
						Prop21 string
					} `config:"source=xxx"`
				}
				_, err = bindParamsToReceiver(&someCfg2{}, "default-svc")
				if !assert.Error(t, err) {
					return
				}
				assert.EqualError(t, err, "Expected Field21.Prop21 to have at least key tag, got: ")
			}
		},
		func() (string, ttFn) {
			return "fail if fields are not exported", func(t *testing.T) {
				type someCfg struct {
					field11 *struct {
					} `config:"source=xxx"`
				}
				_, err := bindParamsToReceiver(&someCfg{}, "default-svc")
				if !assert.Error(t, err) {
					return
				}
				assert.EqualError(t, err, "Expected field11 to be exported")

				type someCfg2 struct {
					Field21 *struct {
						prop21 string `config:"key=xxx"`
					} `config:"source=xxx"`
				}
				_, err = bindParamsToReceiver(&someCfg2{}, "default-svc")
				if !assert.Error(t, err) {
					return
				}
				assert.EqualError(t, err, "Expected Field21.prop21 to be exported")
			}
		},
	}
	for _, tt := range tests {
		name, tt := tt()
		t.Run(name, tt)
	}
}

func Test_param_setValue(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	type field1Cfg struct {
		StrVal  string `config:"key=strVal"`
		IntVal  int    `config:"key=intVal"`
		BoolVal bool   `config:"key=boolVal"`

		// We don't support complex128 for now, but if we do at some point
		// then use some different type here
		UnsupportedTypeVal complex128 `config:"key=unsupportedType"`
	}

	type configStruct struct {
		Field1 *field1Cfg `config:"source=src1"`
	}

	type fields struct {
		receiver interface{}
		paramKey string
	}
	type args struct {
		newVal interface{}
	}
	type testCase struct {
		name      string
		fields    fields
		args      args
		afterBind func()
		assert    func(t *testing.T, err error)
	}

	tests := []func() testCase{
		func() testCase {
			cfg := configStruct{}
			newVal := "new-val-" + faker.Word()

			return testCase{
				name:   "string value",
				fields: fields{receiver: &cfg, paramKey: "strVal"},
				args:   args{newVal: newVal},
				assert: func(t *testing.T, err error) {
					if !assert.NoError(t, err) {
						return
					}
					assert.Equal(t, newVal, cfg.Field1.StrVal)
				},
			}
		},
		func() testCase {
			initialVal := "val-" + faker.Word()
			cfg := configStruct{
				Field1: &field1Cfg{StrVal: initialVal},
			}

			return testCase{
				name:   "not string value",
				fields: fields{receiver: &cfg, paramKey: "strVal"},
				args:   args{newVal: 10},
				afterBind: func() {
					cfg.Field1.StrVal = initialVal
				},
				assert: func(t *testing.T, err error) {
					assert.EqualError(t, err, fmt.Sprintf("Expected string value but got: %v(%[1]T)", 10))
					assert.Equal(t, initialVal, cfg.Field1.StrVal)
				},
			}
		},

		func() testCase {
			cfg := configStruct{}
			newVal := rand.Int()

			return testCase{
				name:   "int value",
				fields: fields{receiver: &cfg, paramKey: "intVal"},
				args:   args{newVal: newVal},
				assert: func(t *testing.T, err error) {
					if !assert.NoError(t, err) {
						return
					}
					assert.Equal(t, newVal, cfg.Field1.IntVal)
				},
			}
		},
		func() testCase {
			cfg := configStruct{}
			newVal := rand.Intn(50000)

			return testCase{
				name:   "int as float32",
				fields: fields{receiver: &cfg, paramKey: "intVal"},
				args:   args{newVal: float32(newVal)},
				assert: func(t *testing.T, err error) {
					if !assert.NoError(t, err) {
						return
					}
					assert.Equal(t, newVal, cfg.Field1.IntVal)
				},
			}
		},
		func() testCase {
			cfg := configStruct{}
			newVal := rand.Intn(50000)

			return testCase{
				name:   "int as float64",
				fields: fields{receiver: &cfg, paramKey: "intVal"},
				args:   args{newVal: float64(newVal)},
				assert: func(t *testing.T, err error) {
					if !assert.NoError(t, err) {
						return
					}
					assert.Equal(t, newVal, cfg.Field1.IntVal)
				},
			}
		},
		func() testCase {
			cfg := configStruct{}
			newVal := rand.Int()

			return testCase{
				name:   "int as string",
				fields: fields{receiver: &cfg, paramKey: "intVal"},
				args:   args{newVal: strconv.Itoa(newVal)},
				assert: func(t *testing.T, err error) {
					if !assert.NoError(t, err) {
						return
					}
					assert.Equal(t, newVal, cfg.Field1.IntVal)
				},
			}
		},

		func() testCase {
			cfg := configStruct{}
			newVal := "aaa123321"

			return testCase{
				name:   "not int string",
				fields: fields{receiver: &cfg, paramKey: "intVal"},
				args:   args{newVal: newVal},
				assert: func(t *testing.T, err error) {
					if !assert.Error(t, err) {
						return
					}
					assert.EqualError(t, err, fmt.Sprintf("Expected int value but got: %v(%[1]T)", newVal))
				},
			}
		},

		func() testCase {
			cfg := configStruct{}
			newVal := time.Now()

			return testCase{
				name:   "not int value",
				fields: fields{receiver: &cfg, paramKey: "intVal"},
				args:   args{newVal: newVal},
				assert: func(t *testing.T, err error) {
					if !assert.Error(t, err) {
						return
					}
					assert.EqualError(t, err, fmt.Sprintf("Expected int value but got: %v(%[1]T)", newVal))
				},
			}
		},

		func() testCase {
			cfg := configStruct{}
			newVal := rand.Intn(2) == 1

			return testCase{
				name:   "bool value",
				fields: fields{receiver: &cfg, paramKey: "boolVal"},
				args:   args{newVal: newVal},
				assert: func(t *testing.T, err error) {
					if !assert.NoError(t, err) {
						return
					}
					assert.Equal(t, newVal, cfg.Field1.BoolVal)
				},
			}
		},

		func() testCase {
			cfg := configStruct{}
			newVal := "not bool"

			return testCase{
				name:   "not bool value",
				fields: fields{receiver: &cfg, paramKey: "boolVal"},
				args:   args{newVal: newVal},
				assert: func(t *testing.T, err error) {
					if !assert.Error(t, err) {
						return
					}
					assert.EqualError(t, err, fmt.Sprintf("Expected bool value but got: %v(%[1]T)", newVal))
				},
			}
		},

		func() testCase {
			cfg := configStruct{}
			newVal := rand.Intn(2) == 1

			return testCase{
				name:   "bool as string",
				fields: fields{receiver: &cfg, paramKey: "boolVal"},
				args:   args{newVal: strconv.FormatBool(newVal)},
				assert: func(t *testing.T, err error) {
					if !assert.NoError(t, err) {
						return
					}
					assert.Equal(t, newVal, cfg.Field1.BoolVal)
				},
			}
		},

		func() testCase {
			cfg := configStruct{}

			return testCase{
				name:   "not supported type",
				fields: fields{receiver: &cfg, paramKey: "unsupportedType"},
				args:   args{newVal: "not important"},
				assert: func(t *testing.T, err error) {
					if !assert.Error(t, err) {
						return
					}
					assert.EqualError(t, err, "Type complex128 is not supported")
				},
			}
		},
	}

	for _, tt := range tests {
		tt := tt()
		t.Run(tt.name, func(t *testing.T) {
			params, err := bindParamsToReceiver(tt.fields.receiver, "default-svc")
			if !assert.NoError(t, err) {
				return
			}
			if tt.afterBind != nil {
				tt.afterBind()
			}
			for _, param := range params {
				if param.key == tt.fields.paramKey {
					tt.assert(t, param.setValue(tt.args.newVal))
					return
				}
			}
			assert.Fail(t, "Unexpected paramKey: "+tt.fields.paramKey)
		})
	}
}
