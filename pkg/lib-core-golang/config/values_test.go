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

func TestParamValue_setValue(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	type fields struct {
		target paramValue
	}
	type args struct {
		newVal interface{}
	}
	type testCase struct {
		name   string
		fields fields
		args   args
		assert func(t *testing.T, err error)
	}
	tests := []func() testCase{
		func() testCase {
			var (
				initialVal = "initial-val-" + faker.Word()
				newVal     = "new-val-" + faker.Word()
			)
			target := StringVal{val: &initialVal}

			return testCase{
				name:   "string value",
				fields: fields{target: target},
				args:   args{newVal: newVal},
				assert: func(t *testing.T, err error) {
					if !assert.NoError(t, err) {
						return
					}
					assert.Equal(t, newVal, target.Value())
				},
			}
		},
		func() testCase {
			var (
				initialVal = "initial-val-" + faker.Word()
				newVal     = 100
			)
			target := StringVal{val: &initialVal}

			return testCase{
				name:   "not string value",
				fields: fields{target: target},
				args:   args{newVal: newVal},
				assert: func(t *testing.T, err error) {
					if !assert.Error(t, err) {
						return
					}
					assert.EqualError(t, err, fmt.Sprintf("Expected string value but got: %v(%[1]T)", newVal))
					assert.Equal(t, initialVal, target.Value())
				},
			}
		},
		func() testCase {
			var (
				initialVal = rand.Int()
				newVal     = rand.Int()
			)
			target := IntVal{val: &initialVal}

			return testCase{
				name:   "int value",
				fields: fields{target: target},
				args:   args{newVal: newVal},
				assert: func(t *testing.T, err error) {
					if !assert.NoError(t, err) {
						return
					}
					assert.Equal(t, newVal, target.Value())
				},
			}
		},
		func() testCase {
			var (
				initialVal = rand.Int()
				newVal     = rand.Intn(50000)
			)
			target := IntVal{val: &initialVal}

			return testCase{
				name:   "int as float32",
				fields: fields{target: target},
				args:   args{newVal: float32(newVal)},
				assert: func(t *testing.T, err error) {
					if !assert.NoError(t, err) {
						return
					}
					assert.Equal(t, newVal, target.Value())
				},
			}
		},
		func() testCase {
			var (
				initialVal = rand.Int()
				newVal     = rand.Intn(50000)
			)
			target := IntVal{val: &initialVal}

			return testCase{
				name:   "int as float64",
				fields: fields{target: target},
				args:   args{newVal: float64(newVal)},
				assert: func(t *testing.T, err error) {
					if !assert.NoError(t, err) {
						return
					}
					assert.Equal(t, newVal, target.Value())
				},
			}
		},
		func() testCase {
			var (
				initialVal = rand.Int()
				newVal     = "not int"
			)
			target := IntVal{val: &initialVal}

			return testCase{
				name:   "not int value",
				fields: fields{target: target},
				args:   args{newVal: newVal},
				assert: func(t *testing.T, err error) {
					if !assert.Error(t, err) {
						return
					}
					assert.EqualError(t, err, fmt.Sprintf("Expected int value but got: %v(%[1]T)", newVal))
					assert.Equal(t, initialVal, target.Value())
				},
			}
		},
		func() testCase {
			var (
				initialVal = rand.Int()
				newVal     = rand.Int()
			)
			target := IntVal{val: &initialVal}

			return testCase{
				name:   "int as string",
				fields: fields{target: target},
				args:   args{newVal: strconv.Itoa(newVal)},
				assert: func(t *testing.T, err error) {
					if !assert.NoError(t, err) {
						return
					}
					assert.Equal(t, newVal, target.Value())
				},
			}
		},
		func() testCase {
			var (
				initialVal = rand.Intn(2) == 1
				newVal     = rand.Intn(2) == 1
			)
			target := BoolVal{val: &initialVal}

			return testCase{
				name:   "bool value",
				fields: fields{target: target},
				args:   args{newVal: newVal},
				assert: func(t *testing.T, err error) {
					if !assert.NoError(t, err) {
						return
					}
					assert.Equal(t, newVal, target.Value())
				},
			}
		},
		func() testCase {
			var (
				initialVal = rand.Intn(2) == 1
				newVal     = "not bool"
			)
			target := BoolVal{val: &initialVal}

			return testCase{
				name:   "not bool value",
				fields: fields{target: target},
				args:   args{newVal: newVal},
				assert: func(t *testing.T, err error) {
					if !assert.Error(t, err) {
						return
					}
					assert.EqualError(t, err, fmt.Sprintf("Expected bool value but got: %v(%[1]T)", newVal))
					assert.Equal(t, initialVal, target.Value())
				},
			}
		},
		func() testCase {
			var (
				initialVal = rand.Intn(2) == 1
				newVal     = rand.Intn(2) == 1
			)
			target := BoolVal{val: &initialVal}

			return testCase{
				name:   "bool as string",
				fields: fields{target: target},
				args:   args{newVal: strconv.FormatBool(newVal)},
				assert: func(t *testing.T, err error) {
					if !assert.NoError(t, err) {
						return
					}
					assert.Equal(t, newVal, target.Value())
				},
			}
		},
	}
	for _, tt := range tests {
		tt := tt()
		t.Run(tt.name, func(t *testing.T) {
			target := tt.fields.target
			err := target.setValue(tt.args.newVal)
			tt.assert(t, err)
		})
	}
}
