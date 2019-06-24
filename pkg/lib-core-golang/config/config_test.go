package config

import (
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockSource struct {
	err        error
	parameters map[param]interface{}
	mock.Mock
}

func (s *mockSource) GetParameters(params []param) (map[param]interface{}, error) {
	if len(s.ExpectedCalls) > 0 {
		args := s.Called(params)
		return args.Get(0).(map[param]interface{}), args.Error(1)
	}
	return s.parameters, s.err
}

func TestNewAppEnv(t *testing.T) {
	type args struct {
		serviceName string
		opts        []appEnvOpt
	}
	type testCase struct {
		name string
		args args
		want AppEnv
	}
	serviceName := "svc-" + faker.Word()
	tests := []func() testCase{
		func() testCase {
			return testCase{
				name: "default",
				args: args{
					serviceName: serviceName,
					opts: []appEnvOpt{withLookupFlag(func(name string) *flag.Flag {
						return nil
					})},
				},
				want: AppEnv{Name: "dev", ServiceName: serviceName},
			}
		},
		func() testCase {
			return testCase{
				name: "test",
				args: args{serviceName: serviceName},
				want: AppEnv{Name: "test", ServiceName: serviceName},
			}
		},
		func() testCase {
			appEnv := fmt.Sprint("app-env-", faker.Word())
			appEnvFacet := fmt.Sprint("app-env-facet", faker.Word())
			if err := os.Setenv(appEnvVar, appEnv); err != nil {
				panic(err)
			}
			if err := os.Setenv(facetVar, appEnvFacet); err != nil {
				panic(err)
			}
			return testCase{
				name: "from env",
				args: args{serviceName: serviceName},
				want: AppEnv{Name: appEnv, ServiceName: serviceName, Facet: appEnvFacet},
			}
		},
		func() testCase {
			appEnv := fmt.Sprint("app-env-", faker.Word())
			appEnvFacet := fmt.Sprint("app-env-facet", faker.Word())
			clusterName := fmt.Sprint("cluster-name-", faker.Word())
			if err := os.Setenv(appEnvVar, appEnv); err != nil {
				panic(err)
			}
			if err := os.Setenv(facetVar, appEnvFacet); err != nil {
				panic(err)
			}
			if err := os.Setenv(clusterNameVar, clusterName); err != nil {
				panic(err)
			}
			return testCase{
				name: "from env with cluster",
				args: args{serviceName: serviceName},
				want: AppEnv{Name: appEnv, ServiceName: serviceName, Facet: appEnvFacet, ClusterName: clusterName},
			}
		},
	}
	for _, ttFn := range tests {
		tt := ttFn()
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if err := os.Unsetenv(appEnvVar); !assert.NoError(t, err) {
					return
				}
				if err := os.Unsetenv(facetVar); !assert.NoError(t, err) {
					return
				}
				if err := os.Unsetenv(clusterNameVar); !assert.NoError(t, err) {
					return
				}
			}()
			got := NewAppEnv(tt.args.serviceName, tt.args.opts...)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLoadInitialValues(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	type args struct {
		sources []sourceBinding
	}
	type testCase struct {
		name   string
		args   args
		assert func(t *testing.T, cfg ServiceConfig, err error)
	}
	tests := []func() testCase{
		func() testCase {
			intParam1 := newIntParam("int1-param-"+faker.Word(), "")
			intParam2 := newIntParam("int2-param-"+faker.Word(), "")
			strParam1 := newStringParam("str1-param-"+faker.Word(), "")
			strParam2 := newStringParam("str2-param-"+faker.Word(), "")
			boolParam1 := newBoolParam("bool1-param-"+faker.Word(), "")
			boolParam2 := newBoolParam("bool2-param-"+faker.Word(), "")

			initialParams1 := map[param]interface{}{
				intParam1:  rand.Int(),
				strParam1:  faker.Word(),
				boolParam1: rand.Intn(2) == 1,
			}
			initialParams2 := map[param]interface{}{
				intParam2:  rand.Int(),
				strParam2:  faker.Word(),
				boolParam2: rand.Intn(2) == 1,
			}
			source1 := sourceBinding{
				params: []param{intParam1, strParam1, boolParam1},
				source: &mockSource{parameters: initialParams1},
			}
			source2 := sourceBinding{
				params: []param{intParam2, strParam2, boolParam2},
				source: &mockSource{parameters: initialParams2},
			}

			return testCase{
				name: "load and init params",
				args: args{
					sources: []sourceBinding{source1, source2},
				},
				assert: func(t *testing.T, cfg ServiceConfig, err error) {
					if !assert.NoError(t, err) {
						return
					}
					assert.Equal(t, initialParams1[intParam1], cfg.IntParam(intParam1).Value())
					assert.Equal(t, initialParams1[strParam1], cfg.StringParam(strParam1).Value())
					assert.Equal(t, initialParams1[boolParam1], cfg.BoolParam(boolParam1).Value())

					assert.Equal(t, initialParams2[intParam2], cfg.IntParam(intParam2).Value())
					assert.Equal(t, initialParams2[strParam2], cfg.StringParam(strParam2).Value())
					assert.Equal(t, initialParams2[boolParam2], cfg.BoolParam(boolParam2).Value())
				},
			}
		},
		func() testCase {
			intParam1 := newIntParam("int1-param-"+faker.Word(), "")
			strParam1 := newStringParam("str1-param-"+faker.Word(), "")

			initialParams1 := map[param]interface{}{
				intParam1: rand.Int(),
			}
			source1 := sourceBinding{
				params: []param{intParam1, strParam1},
				source: &mockSource{parameters: initialParams1},
			}

			return testCase{
				name: "fail if requested params are missing",
				args: args{sources: []sourceBinding{source1}},
				assert: func(t *testing.T, cfg ServiceConfig, err error) {
					if !assert.Error(t, err) {
						return
					}
					assert.EqualError(t, err, fmt.Sprintf("Parameter %v not found", strParam1))
				},
			}
		},
		func() testCase {
			intParam1 := newIntParam("int1-param-"+faker.Word(), "")
			strParam1 := newStringParam("str1-param-"+faker.Word(), "")

			badInt := faker.Word()
			initialParams1 := map[param]interface{}{
				intParam1: badInt,
				strParam1: faker.Word(),
			}
			source1 := sourceBinding{
				params: []param{intParam1, strParam1},
				source: &mockSource{parameters: initialParams1},
			}

			return testCase{
				name: "fail if some params are of a bad type",
				args: args{sources: []sourceBinding{source1}},
				assert: func(t *testing.T, cfg ServiceConfig, err error) {
					if !assert.Error(t, err) {
						return
					}
					assert.EqualError(t, err, fmt.Sprintf("Failed to set value for parameter %v: Expected int value but got: %v(%[2]T)", intParam1, badInt))
				},
			}
		},
		func() testCase {
			intParam1 := newIntParam("int1-param-"+faker.Word(), "")
			strParam1 := newStringParam("str1-param-"+faker.Word(), "")

			sourceErr := fmt.Errorf("Failed to get params: %v", faker.Word())
			source1 := sourceBinding{
				params: []param{intParam1, strParam1},
				source: &mockSource{err: sourceErr},
			}

			return testCase{
				name: "fail if source failed",
				args: args{sources: []sourceBinding{source1}},
				assert: func(t *testing.T, cfg ServiceConfig, err error) {
					if !assert.Error(t, err) {
						return
					}
					assert.EqualError(t, err, sourceErr.Error())
				},
			}
		},
		func() testCase {
			intParam1 := newIntParam("int1-param-"+faker.Word(), "")
			strParam1 := newStringParam("str1-param-"+faker.Word(), "")
			boolParam1 := newBoolParam("bool1-param-"+faker.Word(), "")

			source1 := sourceBinding{
				params: []param{},
				source: &mockSource{parameters: map[param]interface{}{}},
			}

			return testCase{
				name: "panic if getting not existing param",
				args: args{sources: []sourceBinding{source1}},
				assert: func(t *testing.T, cfg ServiceConfig, err error) {
					if !assert.NoError(t, err) {
						return
					}
					assert.PanicsWithValue(t, fmt.Sprintf("Unknown parameter: %v", intParam1), func() {
						cfg.IntParam(intParam1)
					})
					assert.PanicsWithValue(t, fmt.Sprintf("Unknown parameter: %v", strParam1), func() {
						cfg.StringParam(strParam1)
					})
					assert.PanicsWithValue(t, fmt.Sprintf("Unknown parameter: %v", boolParam1), func() {
						cfg.BoolParam(boolParam1)
					})
				},
			}
		},
	}
	for _, tt := range tests {
		tt := tt()
		t.Run(tt.name, func(t *testing.T) {
			opts := make([]ServiceConfigOpt, 0, len(tt.args.sources))
			for _, source := range tt.args.sources {
				opts = append(opts, WithSource(source))
			}
			cfg := newServiceConfig(opts...)
			err := loadInitialValues(cfg)
			tt.assert(t, cfg, err)
		})
	}
}

func TestRefresh(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	type args struct {
		sources []sourceBinding
	}
	type testCase struct {
		name   string
		args   args
		assert func(t *testing.T, refresh func(), cfg ServiceConfig)
	}

	tests := []func() testCase{
		func() testCase {
			param11 := newStringParam("param11-"+faker.Word(), "")
			param12 := newStringParam("param12-"+faker.Word(), "")
			param21 := newStringParam("param21-"+faker.Word(), "")
			param22 := newStringParam("param22-"+faker.Word(), "")

			params1 := map[param]interface{}{
				param11: "initial-val11-" + faker.Word(),
				param12: "initial-val12-" + faker.Word(),
			}
			params2 := map[param]interface{}{
				param21: "initial-val21-" + faker.Word(),
				param22: "initial-val22-" + faker.Word(),
			}
			source1 := sourceBinding{
				params: []param{param11, param12},
				source: &mockSource{parameters: params1},
			}
			source2 := sourceBinding{
				params: []param{param21, param22},
				source: &mockSource{parameters: params2},
			}
			return testCase{
				name: "refresh values from bound sources",
				args: args{
					sources: []sourceBinding{source1, source2},
				},
				assert: func(t *testing.T, refresh func(), cfg ServiceConfig) {
					params1[param11] = "new-val11-" + faker.Word()
					params1[param12] = "new-val12-" + faker.Word()
					params2[param21] = "new-val21-" + faker.Word()
					params2[param22] = "new-val22-" + faker.Word()
					refresh()
					assert.Equal(t, params1[param11], cfg.StringParam(param11).Value())
					assert.Equal(t, params1[param12], cfg.StringParam(param12).Value())
					assert.Equal(t, params2[param21], cfg.StringParam(param21).Value())
					assert.Equal(t, params2[param22], cfg.StringParam(param22).Value())
				},
			}
		},
		func() testCase {
			param11 := newStringParam("param11-"+faker.Word(), "")
			param12 := newIntParam("param12-"+faker.Word(), "")
			param21 := newStringParam("param21-"+faker.Word(), "")
			param22 := newStringParam("param22-"+faker.Word(), "")

			param12InitialVal := rand.Int()
			params1 := map[param]interface{}{
				param11: "initial-val11-" + faker.Word(),
				param12: param12InitialVal,
			}
			params2 := map[param]interface{}{
				param21: "initial-val21-" + faker.Word(),
				param22: "initial-val22-" + faker.Word(),
			}
			source1 := sourceBinding{
				params: []param{param11, param12},
				source: &mockSource{parameters: params1},
			}
			source2 := sourceBinding{
				params: []param{param21, param22},
				source: &mockSource{parameters: params2},
			}
			return testCase{
				name: "ignore refresh errors of individual params",
				args: args{
					sources: []sourceBinding{source1, source2},
				},
				assert: func(t *testing.T, refresh func(), cfg ServiceConfig) {
					params1[param11] = "new-val11-" + faker.Word()
					params1[param12] = "new-val12-" + faker.Word()
					params2[param21] = "new-val21-" + faker.Word()
					params2[param22] = "new-val22-" + faker.Word()
					refresh()
					assert.Equal(t, params1[param11], cfg.StringParam(param11).Value())
					assert.Equal(t, param12InitialVal, cfg.IntParam(param12).Value())
					assert.Equal(t, params2[param21], cfg.StringParam(param21).Value())
					assert.Equal(t, params2[param22], cfg.StringParam(param22).Value())
				},
			}
		},
		func() testCase {
			param11 := newStringParam("param11-"+faker.Word(), "")
			param21 := newStringParam("param21-"+faker.Word(), "")
			param22 := newStringParam("param22-"+faker.Word(), "")

			params1 := map[param]interface{}{
				param11: "initial-val11-" + faker.Word(),
			}
			params2 := map[param]interface{}{
				param21: "initial-val21-" + faker.Word(),
				param22: "initial-val22-" + faker.Word(),
			}
			mockSrc1 := &mockSource{parameters: params1}
			source1 := sourceBinding{
				params: []param{param11},
				source: mockSrc1,
			}
			source2 := sourceBinding{
				params: []param{param21, param22},
				source: &mockSource{parameters: params2},
			}
			return testCase{
				name: "should try next source if refreshing failed",
				args: args{
					sources: []sourceBinding{source1, source2},
				},
				assert: func(t *testing.T, refresh func(), cfg ServiceConfig) {
					params2[param21] = "new-val21-" + faker.Word()
					params2[param22] = "new-val22-" + faker.Word()

					mockSrc1.On("GetParameters", mock.Anything).
						Return(map[param]interface{}{}, errors.New(faker.Sentence()))
					refresh()
					assert.Equal(t, params1[param11], cfg.StringParam(param11).Value())
					assert.Equal(t, params2[param21], cfg.StringParam(param21).Value())
					assert.Equal(t, params2[param22], cfg.StringParam(param22).Value())
				},
			}
		},
		func() testCase {
			param11 := newStringParam("param11-"+faker.Word(), "")
			param21 := newStringParam("param21-"+faker.Word(), "")

			param11Val := "initial-val11-" + faker.Word()
			params1 := map[param]interface{}{param11: param11Val}
			params2 := map[param]interface{}{
				param21: "initial-val21-" + faker.Word(),
			}
			mockSrc1 := &mockSource{parameters: params1}
			source1 := sourceBinding{
				params: []param{param11},
				source: mockSrc1,
			}
			source2 := sourceBinding{
				params: []param{param21},
				source: &mockSource{parameters: params2},
			}
			return testCase{
				name: "ignore missing params",
				args: args{
					sources: []sourceBinding{source1, source2},
				},
				assert: func(t *testing.T, refresh func(), cfg ServiceConfig) {
					delete(params1, param11)
					params2[param21] = "new-val21-" + faker.Word()
					refresh()
					assert.Equal(t, param11Val, cfg.StringParam(param11).Value())
					assert.Equal(t, params2[param21], cfg.StringParam(param21).Value())
				},
			}
		},
	}

	for _, tt := range tests {
		tt := tt()
		t.Run(tt.name, func(t *testing.T) {
			refreshChannel := make(chan time.Time)
			refreshed := make(chan bool)
			stop := make(chan bool)
			opts := make([]ServiceConfigOpt, 0, len(tt.args.sources)+1)
			for _, source := range tt.args.sources {
				opts = append(opts, WithSource(source))
			}
			opts = append(opts,
				withTicker(&time.Ticker{C: refreshChannel}),
				withRefreshed(refreshed),
				withStop(stop),
			)
			defer func() {
				stop <- true
			}()
			cfg, err := Load(opts...)
			if !assert.NoError(t, err) {
				return
			}
			refreshFn := func() {
				refreshChannel <- time.Now()
				<-refreshed
			}
			tt.assert(t, refreshFn, cfg)
		})
	}
}
