package config

import (
	"errors"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/assert"
)

func TestParamsBuilder(t *testing.T) {
	type fields struct {
		serviceName string
	}
	type testCase struct {
		name   string
		fields fields
		run    func(t *testing.T, b *ParamsBuilder)
	}
	tests := []func() testCase{
		func() testCase {
			serviceName := "svc-" + faker.Word()
			return testCase{
				name:   "NewParam",
				fields: fields{serviceName: serviceName},
				run: func(t *testing.T, b *ParamsBuilder) {
					paramKey := "param-key-" + faker.Word()
					paramBuilder := b.NewParam(paramKey)
					assert.Equal(t, &ParamBuilder{
						paramKey: paramKey,
						paramSvc: serviceName,
						pb:       b,
					}, paramBuilder)
				},
			}
		},
	}
	for _, tt := range tests {
		tt := tt()
		t.Run(tt.name, func(t *testing.T) {
			b := &ParamsBuilder{
				params:      []param{},
				serviceName: tt.fields.serviceName,
			}
			tt.run(t, b)
		})
	}
}

func TestParamBuilder(t *testing.T) {
	type testCase struct {
		name string
		run  func(t *testing.T, b *ParamBuilder)
	}
	tests := []func() testCase{
		func() testCase {
			return testCase{
				name: "build int param",
				run: func(t *testing.T, b *ParamBuilder) {
					param := b.Int()
					assert.Equal(t, newIntParam(b.paramKey, b.paramSvc), param)
					assert.Contains(t, b.pb.params, param)
				},
			}
		},
		func() testCase {
			return testCase{
				name: "build string param",
				run: func(t *testing.T, b *ParamBuilder) {
					param := b.String()
					assert.Equal(t, newStringParam(b.paramKey, b.paramSvc), param)
					assert.Contains(t, b.pb.params, param)
				},
			}
		},
		func() testCase {
			return testCase{
				name: "build bool param",
				run: func(t *testing.T, b *ParamBuilder) {
					param := b.Bool()
					assert.Equal(t, newBoolParam(b.paramKey, b.paramSvc), param)
					assert.Contains(t, b.pb.params, param)
				},
			}
		},
		func() testCase {
			return testCase{
				name: "with service",
				run: func(t *testing.T, b *ParamBuilder) {
					fakeService := "fake-svc-" + faker.Word()
					assert.Equal(t, b, b.WithService(fakeService))
					assert.Equal(t, fakeService, b.paramSvc)
				},
			}
		},
	}
	for _, tt := range tests {
		tt := tt()
		t.Run(tt.name, func(t *testing.T) {
			b := &ParamBuilder{
				paramKey: "key-" + faker.Word(),
				paramSvc: "svc-" + faker.Word(),
				pb:       &ParamsBuilder{params: []param{}},
			}
			tt.run(t, b)
		})
	}
}

func TestBuilder(t *testing.T) {
	type args struct {
		appEnv AppEnv
	}
	type testCase struct {
		name string
		args args
		run  func(t *testing.T, b *Builder)
	}
	tests := []func() testCase{
		func() testCase {
			appEnv := AppEnv{
				ServiceName: "svc-" + faker.Word(),
				Name:        "app-env-" + faker.Word(),
			}
			return testCase{
				name: "NewBuilder",
				args: args{appEnv: appEnv},
				run: func(t *testing.T, b *Builder) {
					assert.Equal(t, appEnv, b.appEnv)
				},
			}
		},
		func() testCase {
			appEnv := AppEnv{
				ServiceName: "svc-" + faker.Word(),
				Name:        "app-env-" + faker.Word(),
			}
			return testCase{
				name: "NewParamsBuilder",
				args: args{appEnv: appEnv},
				run: func(t *testing.T, b *Builder) {
					sourceFactory := SourceFactory(func() (Source, error) {
						return nil, nil
					})
					paramsBuilder := b.NewParamsBuilder(sourceFactory)
					assert.Equal(t, []param{}, paramsBuilder.params)
					assert.Equal(t, appEnv.ServiceName, paramsBuilder.serviceName)
					assert.IsType(t, sourceFactory, paramsBuilder.sourceFactory)

					assert.Len(t, b.paramsBuilders, 1)
					assert.Contains(t, b.paramsBuilders, paramsBuilder)
				},
			}
		},
		func() testCase {
			appEnv := AppEnv{
				ServiceName: "svc-" + faker.Word(),
				Name:        "app-env-" + faker.Word(),
			}
			return testCase{
				name: "WithLocalSource",
				args: args{appEnv: appEnv},
				run: func(t *testing.T, b *Builder) {
					source, err := b.WithLocalSource()()
					if !assert.NoError(t, err) {
						return
					}
					if !assert.IsType(t, &localSource{}, source) {
						return
					}
					localSrc := source.(*localSource)
					assert.Equal(t, true, localSrc.ignoreDefaultService)
					assert.Equal(t, appEnv.ServiceName, localSrc.defaultService)
				},
			}
		},
		func() testCase {
			appEnv := AppEnv{
				ServiceName: "svc-" + faker.Word(),
				Name:        "dev",
			}
			return testCase{
				name: "WithRemoteSource(appEnv=dev)",
				args: args{appEnv: appEnv},
				run: func(t *testing.T, b *Builder) {
					source, err := b.WithRemoteSource()()
					if !assert.NoError(t, err) {
						return
					}
					assert.IsType(t, &localSource{}, source)
					localSrc := source.(*localSource)
					assert.Equal(t, false, localSrc.ignoreDefaultService)
					assert.Equal(t, appEnv.ServiceName, localSrc.defaultService)
				},
			}
		},
		func() testCase {
			appEnv := AppEnv{
				ServiceName: "svc-" + faker.Word(),
				Name:        "not-dev",
			}
			return testCase{
				name: "WithRemoteSource(appEnv!=dev)",
				args: args{appEnv: appEnv},
				run: func(t *testing.T, b *Builder) {
					source, err := b.WithRemoteSource()()
					if !assert.NoError(t, err) {
						return
					}
					assert.IsType(t, &awsSSMSource{}, source)
					awsSSMSource := source.(*awsSSMSource)
					assert.NotNil(t, awsSSMSource.ssmClient)
				},
			}
		},
		func() testCase {
			appEnv := AppEnv{
				ServiceName: "svc-" + faker.Word(),
				Name:        "app-env-" + faker.Word(),
			}
			return testCase{
				name: "LoadConfig",
				args: args{appEnv: appEnv},
				run: func(t *testing.T, b *Builder) {
					var source1 Source
					params1 := b.NewParamsBuilder(func() (Source, error) {
						return source1, nil
					})
					param11 := params1.NewParam("param11-" + faker.Word()).String()
					param12 := params1.NewParam("param12-" + faker.Word()).String()
					values1 := map[param]interface{}{
						param11: "val11-" + faker.Word(),
						param12: "val21-" + faker.Word(),
					}
					source1 = &mockSource{parameters: values1}

					var source2 Source
					params2 := b.NewParamsBuilder(func() (Source, error) {
						return source2, nil
					})
					param21 := params2.NewParam("param21-" + faker.Word()).String()
					param22 := params2.NewParam("param22-" + faker.Word()).String()
					values2 := map[param]interface{}{
						param21: "val21-" + faker.Word(),
						param22: "val22-" + faker.Word(),
					}
					source2 = &mockSource{parameters: values2}
					stop := make(chan bool)
					cfg, err := b.LoadConfig(withStop(stop))
					if !assert.NoError(t, err) {
						return
					}
					defer func() { stop <- true }()

					assert.Equal(t, values1[param11], cfg.StringParam(param11).Value())
					assert.Equal(t, values1[param12], cfg.StringParam(param12).Value())

					assert.Equal(t, values2[param21], cfg.StringParam(param21).Value())
					assert.Equal(t, values2[param22], cfg.StringParam(param22).Value())
				},
			}
		},
		func() testCase {
			appEnv := AppEnv{
				ServiceName: "svc-" + faker.Word(),
				Name:        "app-env-" + faker.Word(),
			}
			return testCase{
				name: "LoadConfig source fails",
				args: args{appEnv: appEnv},
				run: func(t *testing.T, b *Builder) {
					err1 := errors.New(faker.Sentence())
					b.NewParamsBuilder(func() (Source, error) {
						return nil, err1
					})
					_, err := b.LoadConfig()
					if !assert.Error(t, err) {
						return
					}
					assert.EqualError(t, err, err1.Error())
				},
			}
		},

		// TODO: Load fails
	}
	for _, tt := range tests {
		tt := tt()
		t.Run(tt.name, func(t *testing.T) {
			b := NewBuilder(tt.args.appEnv)
			tt.run(t, b)
		})
	}
}
