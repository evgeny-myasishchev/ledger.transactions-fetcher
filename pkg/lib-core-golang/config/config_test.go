package config

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/assert"
)

type mockSource struct {
	err    error
	values map[string]interface{}
}

func (s *mockSource) GetParameters(ctx context.Context, params []param) (map[paramID]interface{}, error) {
	if s.err != nil {
		return nil, s.err
	}
	result := map[paramID]interface{}{}
	for key, val := range s.values {
		result[paramID{key: key}] = val
	}
	return result, nil
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
				want: AppEnv{
					Name:        "dev",
					Facet:       os.Getenv("APP_ENV_FACET"),
					ServiceName: serviceName,
				},
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

func TestBind(t *testing.T) {
	type args struct {
		receiver interface{}
		opts     []BindOpt
	}

	type params1 struct {
		Val1 string `config:"key=val11"`
		Val2 string `config:"key=val12"`
	}

	type params2 struct {
		Val1 string `config:"key=val21"`
		Val2 string `config:"key=val22"`
	}

	type config struct {
		P1 *params1 `config:"source=src1"`
		P2 *params2 `config:"source=src2"`
	}

	type mockSourceOpt func(src *mockSource)
	newMockSource := func(values map[string]interface{}, opts ...mockSourceOpt) SourceFactory {
		return func() (Source, error) {
			src := &mockSource{
				values: values,
			}
			for _, opt := range opts {
				opt(src)
			}
			return src, nil
		}
	}

	type testCase func(*testing.T) error
	tests := []func() (string, testCase){
		func() (string, testCase) {
			return "initial values", func(t *testing.T) error {
				values1 := map[string]interface{}{
					"val11": "val-11-" + faker.Word(),
					"val12": "val-12-" + faker.Word(),
				}
				values2 := map[string]interface{}{
					"val21": "val-21-" + faker.Word(),
					"val22": "val-22-" + faker.Word(),
				}
				var cfg config
				if err := Bind(&cfg, AppEnv{},
					WithSource("src1", newMockSource(values1)),
					WithSource("src2", newMockSource(values2)),
				); err != nil {
					return err
				}
				if !assert.NotNil(t, cfg.P1) || !assert.NotNil(t, cfg.P2) {
					return nil
				}
				assert.Equal(t, cfg.P1.Val1, values1["val11"])
				assert.Equal(t, cfg.P1.Val2, values1["val12"])
				assert.Equal(t, cfg.P2.Val1, values2["val21"])
				assert.Equal(t, cfg.P2.Val2, values2["val22"])
				return nil
			}
		},
		func() (string, testCase) {
			return "fail if no initial values", func(t *testing.T) error {
				values1 := map[string]interface{}{"val11": "val-11-" + faker.Word()}
				var cfg config
				err := Bind(&cfg, AppEnv{}, WithSource("src1", newMockSource(values1)))
				if !assert.Error(t, err) {
					return nil
				}
				assert.EqualError(t, err, fmt.Sprintf("Parameter %v not found (source=%v)", paramID{key: "val12"}, "src1"))
				return nil
			}
		},
		func() (string, testCase) {
			return "load initial fail if source factory fails", func(t *testing.T) error {
				wantErr := errors.New(faker.Sentence())
				var cfg config
				err := Bind(&cfg, AppEnv{}, WithSource("src", func() (Source, error) {
					return nil, wantErr
				}))
				assert.Equal(t, wantErr, errors.Cause(err))
				return nil
			}
		},
		func() (string, testCase) {
			return "load initial fail if source fails", func(t *testing.T) error {
				wantErr := errors.New(faker.Sentence())
				var cfg config
				err := Bind(&cfg, AppEnv{}, WithSource("src", func() (Source, error) {
					return &mockSource{err: wantErr}, nil
				}))
				assert.Equal(t, wantErr, errors.Cause(err))
				return nil
			}
		},
		func() (string, testCase) {
			return "load initial fails if bad value", func(t *testing.T) error {
				values1 := map[string]interface{}{"val11": 100, "val12": "val-12-" + faker.Word()}
				values2 := map[string]interface{}{"val21": "val-21-" + faker.Word(), "val22": "val-22-" + faker.Word()}
				var cfg config
				err := Bind(&cfg, AppEnv{},
					WithSource("src1", newMockSource(values1)),
					WithSource("src2", newMockSource(values2)),
				)
				if !assert.Error(t, err) {
					return nil
				}
				assert.EqualError(t, errors.Cause(err), "Expected string value but got: 100(int)")
				return nil
			}
		},
		func() (string, testCase) {
			return "refresh values", func(t *testing.T) error {
				values1 := map[string]interface{}{"val11": "val-11-" + faker.Word(), "val12": "val-12-" + faker.Word()}
				values2 := map[string]interface{}{"val21": "val-21-" + faker.Word(), "val22": "val-22-" + faker.Word()}
				refreshSignal := make(chan time.Time)
				onRefreshed := make(chan struct{})
				stopSignal := make(chan struct{})
				defer func() {
					stopSignal <- struct{}{}
				}()
				var cfg config
				if err := Bind(&cfg, AppEnv{},
					WithSource("src1", newMockSource(values1)),
					WithSource("src2", newMockSource(values2)),
					withSignals(refreshSignal, onRefreshed, stopSignal),
				); err != nil {
					return err
				}
				if !assert.NotNil(t, cfg.P1) || !assert.NotNil(t, cfg.P2) {
					return nil
				}
				values1["val11"] = "updated-val11-" + faker.Word()
				values2["val22"] = "updated-val22-" + faker.Word()
				refreshSignal <- time.Now()
				<-onRefreshed
				assert.Equal(t, cfg.P1.Val1, values1["val11"])
				assert.Equal(t, cfg.P2.Val2, values2["val22"])
				return nil
			}
		},
		func() (string, testCase) {
			return "refresh values from other sources if source fails", func(t *testing.T) error {
				values1 := map[string]interface{}{"val11": "val-11-" + faker.Word(), "val12": "val-12-" + faker.Word()}
				values2 := map[string]interface{}{"val21": "val-21-" + faker.Word(), "val22": "val-22-" + faker.Word()}
				refreshSignal := make(chan time.Time)
				onRefreshed := make(chan struct{})
				stopSignal := make(chan struct{})
				defer func() {
					stopSignal <- struct{}{}
				}()
				var cfg config
				var src1 *mockSource
				if err := Bind(&cfg, AppEnv{},
					WithSource("src1", newMockSource(values1, func(s *mockSource) {
						src1 = s
					})),
					WithSource("src2", newMockSource(values2)),
					withSignals(refreshSignal, onRefreshed, stopSignal),
				); err != nil {
					return err
				}
				if !assert.NotNil(t, cfg.P1) || !assert.NotNil(t, cfg.P2) {
					return nil
				}
				values2["val22"] = "updated-val22-" + faker.Word()
				src1.err = errors.New(faker.Sentence())
				refreshSignal <- time.Now()
				<-onRefreshed
				assert.Equal(t, cfg.P2.Val2, values2["val22"])
				return nil
			}
		},
		func() (string, testCase) {
			return "refresh values ignore missing values", func(t *testing.T) error {
				initialVal11 := "val-11-" + faker.Word()
				values1 := map[string]interface{}{"val11": initialVal11, "val12": "val-12-" + faker.Word()}
				values2 := map[string]interface{}{"val21": "val-21-" + faker.Word(), "val22": "val-22-" + faker.Word()}
				refreshSignal := make(chan time.Time)
				onRefreshed := make(chan struct{})
				stopSignal := make(chan struct{})
				defer func() {
					stopSignal <- struct{}{}
				}()
				var cfg config
				if err := Bind(&cfg, AppEnv{},
					WithSource("src1", newMockSource(values1)),
					WithSource("src2", newMockSource(values2)),
					withSignals(refreshSignal, onRefreshed, stopSignal),
				); err != nil {
					return err
				}
				if !assert.NotNil(t, cfg.P1) || !assert.NotNil(t, cfg.P2) {
					return nil
				}
				delete(values1, "val11")
				values1["val12"] = "updated-val12-" + faker.Word()
				refreshSignal <- time.Now()
				<-onRefreshed
				assert.Equal(t, initialVal11, cfg.P1.Val1)
				assert.Equal(t, cfg.P1.Val2, values1["val12"])
				return nil
			}
		},
		func() (string, testCase) {
			return "refresh values ignore bad values", func(t *testing.T) error {
				initialVal11 := "val-11-" + faker.Word()
				values1 := map[string]interface{}{"val11": initialVal11, "val12": "val-12-" + faker.Word()}
				values2 := map[string]interface{}{"val21": "val-21-" + faker.Word(), "val22": "val-22-" + faker.Word()}
				refreshSignal := make(chan time.Time)
				onRefreshed := make(chan struct{})
				stopSignal := make(chan struct{})
				defer func() {
					stopSignal <- struct{}{}
				}()
				var cfg config
				if err := Bind(&cfg, AppEnv{},
					WithSource("src1", newMockSource(values1)),
					WithSource("src2", newMockSource(values2)),
					withSignals(refreshSignal, onRefreshed, stopSignal),
				); err != nil {
					return err
				}
				if !assert.NotNil(t, cfg.P1) || !assert.NotNil(t, cfg.P2) {
					return nil
				}
				values1["val11"] = 100
				values1["val12"] = "updated-val12-" + faker.Word()
				refreshSignal <- time.Now()
				<-onRefreshed
				assert.Equal(t, initialVal11, cfg.P1.Val1)
				assert.Equal(t, cfg.P1.Val2, values1["val12"])
				return nil
			}
		},
	}
	for _, tt := range tests {
		name, tt := tt()
		t.Run(name, func(t *testing.T) {
			if err := tt(t); !assert.NoError(t, err) {
				return
			}
		})
	}
}
