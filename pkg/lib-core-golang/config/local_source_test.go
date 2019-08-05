package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/assert"
)

func ensureTmpDir(name string) string {
	var tmpDir = path.Join("..", "..", "..", "tmp", name)
	os.RemoveAll(tmpDir)
	if err := os.MkdirAll(tmpDir, os.ModePerm); err != nil {
		panic(err)
	}
	return tmpDir
}

func TestLocalSource_NewLocalSource(t *testing.T) {
	type fields struct {
		opts []LocalOpt
	}
	type args struct {
		params []param
	}
	type testCase struct {
		name string
		run  func(t *testing.T, src *localSource, err error)
	}

	tests := []func() testCase{
		func() testCase {
			return testCase{
				name: "default dir",
				run: func(t *testing.T, src *localSource, err error) {
					var expectedSource string
					if _, file, _, ok := runtime.Caller(0); ok == true {
						expectedSource = filepath.Join(file, "..", "..", "..", "..", "config")
					} else {
						panic("Can not get project root")
					}
					assert.Equal(t, expectedSource, src.dir)
				},
			}
		},
	}

	for _, tt := range tests {
		tt := tt()
		t.Run(tt.name, func(t *testing.T) {
			src, err := NewLocalSource()()
			tt.run(t, src.(*localSource), err)
		})
	}
}

func TestLocalSource_GetParameters(t *testing.T) {
	type fields struct {
		opts []LocalOpt
	}
	type args struct {
		params []param
	}
	type testCase struct {
		name   string
		fields fields
		args   args
		want   func(t *testing.T, params map[paramID]interface{}, err error)
		after  func()
	}

	defaultConfigDir := ensureTmpDir("local-source-test")
	defaultCfg := map[string]interface{}{
		"key1": fmt.Sprint("default-key1-", faker.Word()),
		"key2": fmt.Sprint("default-key2-", faker.Word()),
		"deeply": map[string]interface{}{
			"nested": map[string]interface{}{
				"key3": fmt.Sprint("default-key3-", faker.Word()),
			},
		},
	}
	productionCfg := map[string]interface{}{
		"prod-key1": fmt.Sprint("new-production-key1-", faker.Word()),
		"key2":      fmt.Sprint("production-key2-", faker.Word()),
		"key3":      fmt.Sprint("production-key3-", faker.Word()),
	}
	productionPreprodCfg := map[string]interface{}{
		"prod-key1": fmt.Sprint("new-prod-preprod-key1-", faker.Word()),
		"key1":      fmt.Sprint("prod-preprod-key1-", faker.Word()),
		"deeply": map[string]interface{}{
			"nested": map[string]interface{}{
				"key3": fmt.Sprint("prod-preprod-key3-", faker.Word()),
			},
		},
	}

	writeConfig := func(configDir string, name string, value interface{}) bool {
		buffer, err := json.Marshal(value)
		if !assert.NoError(t, err) {
			return false
		}

		if err := ioutil.WriteFile(path.Join(configDir, name), buffer, os.ModePerm); !assert.NoError(t, err) {
			return false
		}

		return true
	}

	if !writeConfig(defaultConfigDir, "default.json", defaultCfg) {
		return
	}
	if !writeConfig(defaultConfigDir, "production.json", productionCfg) {
		return
	}
	if !writeConfig(defaultConfigDir, "production-preprod.json", productionPreprodCfg) {
		return
	}

	tests := []func() testCase{
		func() testCase {
			return testCase{
				name:   "default config",
				fields: fields{opts: []LocalOpt{LocalOpts.WithDir(defaultConfigDir)}},
				args: args{
					params: []param{
						param{paramID: paramID{key: "key1"}},
						param{paramID: paramID{key: "key2"}},
						param{paramID: paramID{service: "deeply", key: "nested/key3"}},
					},
				},
				want: func(t *testing.T, got map[paramID]interface{}, err error) {
					if !assert.NoError(t, err) {
						return
					}
					deeply := defaultCfg["deeply"].(map[string]interface{})
					nested := deeply["nested"].(map[string]interface{})
					assert.Equal(t, map[paramID]interface{}{
						paramID{key: "key1"}:                           defaultCfg["key1"],
						paramID{key: "key2"}:                           defaultCfg["key2"],
						paramID{service: "deeply", key: "nested/key3"}: nested["key3"],
					}, got)
				},
			}
		},
		func() testCase {
			serviceName := "default-svc-" + faker.Word()
			return testCase{
				name: "ignore default service",
				fields: fields{opts: []LocalOpt{
					LocalOpts.WithDir(defaultConfigDir),
					LocalOpts.WithAppEnv(AppEnv{ServiceName: serviceName}),
					LocalOpts.WithIgnoreDefaultService(),
				}},
				args: args{
					params: []param{
						param{paramID: paramID{key: "key1"}},
						param{paramID: paramID{key: "key2"}},
						param{paramID: paramID{service: serviceName, key: "deeply/nested/key3"}},
					},
				},
				want: func(t *testing.T, got map[paramID]interface{}, err error) {
					if !assert.NoError(t, err) {
						return
					}
					deeply := defaultCfg["deeply"].(map[string]interface{})
					nested := deeply["nested"].(map[string]interface{})
					assert.Equal(t, map[paramID]interface{}{
						paramID{key: "key1"}: defaultCfg["key1"],
						paramID{key: "key2"}: defaultCfg["key2"],
						paramID{service: serviceName, key: "deeply/nested/key3"}: nested["key3"],
					}, got)
				},
			}
		},
		func() testCase {
			return testCase{
				name: "env specific config",
				fields: fields{opts: []LocalOpt{
					LocalOpts.WithDir(defaultConfigDir),
					LocalOpts.WithAppEnv(AppEnv{Name: "production"}),
				}},
				args: args{
					params: []param{
						param{paramID: paramID{key: "prod-key1"}},
						param{paramID: paramID{key: "key1"}},
						param{paramID: paramID{key: "key2"}},
						param{paramID: paramID{service: "deeply", key: "nested/key3"}},
					},
				},
				want: func(t *testing.T, got map[paramID]interface{}, err error) {
					if !assert.NoError(t, err) {
						return
					}
					deeply := defaultCfg["deeply"].(map[string]interface{})
					nested := deeply["nested"].(map[string]interface{})
					assert.Equal(t, map[paramID]interface{}{
						paramID{key: "prod-key1"}:                      productionCfg["prod-key1"],
						paramID{key: "key1"}:                           defaultCfg["key1"],
						paramID{key: "key2"}:                           productionCfg["key2"],
						paramID{service: "deeply", key: "nested/key3"}: nested["key3"],
					}, got)
				},
			}
		},
		func() testCase {
			return testCase{
				name: "env flavor specific config",
				fields: fields{opts: []LocalOpt{
					LocalOpts.WithDir(defaultConfigDir),
					LocalOpts.WithAppEnv(AppEnv{Name: "production", Facet: "preprod"}),
				}},
				args: args{
					params: []param{
						param{paramID: paramID{key: "prod-key1"}},
						param{paramID: paramID{key: "key1"}},
						param{paramID: paramID{key: "key2"}},
						param{paramID: paramID{key: "key3"}},
						param{paramID: paramID{service: "deeply", key: "nested/key3"}},
					},
				},
				want: func(t *testing.T, got map[paramID]interface{}, err error) {
					if !assert.NoError(t, err) {
						return
					}
					deeply := productionPreprodCfg["deeply"].(map[string]interface{})
					nested := deeply["nested"].(map[string]interface{})
					assert.Equal(t, map[paramID]interface{}{
						paramID{key: "prod-key1"}:                      productionPreprodCfg["prod-key1"],
						paramID{key: "key1"}:                           productionPreprodCfg["key1"],
						paramID{key: "key2"}:                           productionCfg["key2"],
						paramID{key: "key3"}:                           productionCfg["key3"],
						paramID{service: "deeply", key: "nested/key3"}: nested["key3"],
					}, got)
				},
			}
		},
		func() testCase {
			return testCase{
				name:   "no error if no such key",
				fields: fields{opts: []LocalOpt{LocalOpts.WithDir(defaultConfigDir)}},
				args: args{
					params: []param{
						param{paramID: paramID{key: "no-key1"}},
						param{paramID: paramID{key: "key2"}},
					},
				},
				want: func(t *testing.T, got map[paramID]interface{}, err error) {
					if !assert.NoError(t, err) {
						return
					}
					assert.Equal(t, map[paramID]interface{}{
						paramID{key: "key2"}: defaultCfg["key2"],
					}, got)
				},
			}
		},
		func() testCase {
			return testCase{
				name:   "no error if no such nested key",
				fields: fields{opts: []LocalOpt{LocalOpts.WithDir(defaultConfigDir)}},
				args: args{
					params: []param{
						param{paramID: paramID{key: "key2"}},
						param{paramID: paramID{service: "deeply", key: "nested/no-key3"}},
					},
				},
				want: func(t *testing.T, got map[paramID]interface{}, err error) {
					if !assert.NoError(t, err) {
						return
					}
					assert.Equal(t, map[paramID]interface{}{
						paramID{key: "key2"}: defaultCfg["key2"],
					}, got)
				},
			}
		},
		func() testCase {
			return testCase{
				name:   "error if no default config",
				fields: fields{opts: []LocalOpt{LocalOpts.WithDir(defaultConfigDir + "-no-config")}},
				args: args{
					params: []param{param{paramID: paramID{key: "key1"}}},
				},
				want: func(t *testing.T, got map[paramID]interface{}, err error) {
					if !assert.Error(t, err) {
						return
					}
					pathErr, ok := err.(*os.PathError)
					if !assert.True(t, ok) {
						return
					}
					assert.Equal(t, "no such file or directory", pathErr.Err.Error())
				},
			}
		},
		func() testCase {
			return testCase{
				name: "no error if no env specific config",
				fields: fields{opts: []LocalOpt{
					LocalOpts.WithDir(defaultConfigDir),
					LocalOpts.WithAppEnv(AppEnv{Name: "no-staging", Facet: "no-facet"}),
				}},
				args: args{
					params: []param{
						param{paramID: paramID{key: "key1"}},
						param{paramID: paramID{key: "key2"}},
						param{paramID: paramID{service: "deeply", key: "nested/key3"}},
					},
				},
				want: func(t *testing.T, got map[paramID]interface{}, err error) {
					if !assert.NoError(t, err) {
						return
					}
					deeply := defaultCfg["deeply"].(map[string]interface{})
					nested := deeply["nested"].(map[string]interface{})
					assert.Equal(t, map[paramID]interface{}{
						paramID{key: "key1"}:                           defaultCfg["key1"],
						paramID{key: "key2"}:                           defaultCfg["key2"],
						paramID{service: "deeply", key: "nested/key3"}: nested["key3"],
					}, got)
				},
			}
		},
		func() testCase {
			key1EnvName := "TEST_PROD_ENVparamKey1_" + strings.ToUpper(faker.Word())
			key1EnvVal := "key1-env-override-" + faker.Word()
			key3EnvName := "TEST_PROD_ENV_NESTEDparamKey3_" + strings.ToUpper(faker.Word())
			key3EnvVal := "key3-env-override-" + faker.Word()
			customEnvVars := map[string]interface{}{
				"key1": key1EnvName,
				"deeply": map[string]interface{}{
					"nested": map[string]interface{}{
						"key3": key3EnvName,
					},
				},
			}
			if !writeConfig(defaultConfigDir, "custom-environment-variables.json", customEnvVars) {
				panic("Failed to write file")
			}
			if err := os.Setenv(key1EnvName, key1EnvVal); err != nil {
				panic(err)
			}
			if err := os.Setenv(key3EnvName, key3EnvVal); err != nil {
				panic(err)
			}

			return testCase{
				name: "env overrides",
				fields: fields{opts: []LocalOpt{
					LocalOpts.WithDir(defaultConfigDir),
					LocalOpts.WithAppEnv(AppEnv{Name: "production", Facet: "preprod"}),
				}},
				args: args{
					params: []param{
						param{paramID: paramID{key: "prod-key1"}},
						param{paramID: paramID{key: "key1"}},
						param{paramID: paramID{key: "key2"}},
						param{paramID: paramID{key: "key3"}},
						param{paramID: paramID{service: "deeply", key: "nested/key3"}},
					},
				},
				want: func(t *testing.T, got map[paramID]interface{}, err error) {
					if !assert.NoError(t, err) {
						return
					}
					assert.Equal(t, map[paramID]interface{}{
						paramID{key: "prod-key1"}:                      productionPreprodCfg["prod-key1"],
						paramID{key: "key1"}:                           key1EnvVal,
						paramID{key: "key2"}:                           productionCfg["key2"],
						paramID{key: "key3"}:                           productionCfg["key3"],
						paramID{service: "deeply", key: "nested/key3"}: key3EnvVal,
					}, got)
				},
				after: func() {
					os.Remove(path.Join(defaultConfigDir, "custom-environment-variables.json"))
					os.Unsetenv(key1EnvName)
					os.Unsetenv(key3EnvName)
				},
			}
		},
		func() testCase {
			defaultOverrides := map[string]interface{}{
				"key1": fmt.Sprint("overrides-key1-", faker.Word()),
				"key2": fmt.Sprint("overrides-key2-", faker.Word()),
				"deeply": map[string]interface{}{
					"nested": map[string]interface{}{
						"key3": fmt.Sprint("overrides-key3-", faker.Word()),
					},
				},
			}

			overridesConfigDir := ensureTmpDir("local-source-test-default-overrides")
			if !writeConfig(overridesConfigDir, "default.json", defaultCfg) {
				panic("Failed to write config")
			}
			if !writeConfig(overridesConfigDir, "local-default.json", defaultOverrides) {
				panic("Failed to write config")
			}

			return testCase{
				name:   "local overrides for default config",
				fields: fields{opts: []LocalOpt{LocalOpts.WithDir(overridesConfigDir)}},
				args: args{
					params: []param{
						param{paramID: paramID{key: "key1"}},
						param{paramID: paramID{key: "key2"}},
						param{paramID: paramID{service: "deeply", key: "nested/key3"}},
					},
				},
				want: func(t *testing.T, got map[paramID]interface{}, err error) {
					if !assert.NoError(t, err) {
						return
					}
					deeply := defaultOverrides["deeply"].(map[string]interface{})
					nested := deeply["nested"].(map[string]interface{})
					assert.Equal(t, map[paramID]interface{}{
						paramID{key: "key1"}:                           defaultOverrides["key1"],
						paramID{key: "key2"}:                           defaultOverrides["key2"],
						paramID{service: "deeply", key: "nested/key3"}: nested["key3"],
					}, got)
				},
			}
		},
		func() testCase {
			envOverrides := map[string]interface{}{
				"key1": fmt.Sprint("env-overrides-key1-", faker.Word()),
				"key2": fmt.Sprint("env-overrides-key2-", faker.Word()),
				"deeply": map[string]interface{}{
					"nested": map[string]interface{}{
						"key3": fmt.Sprint("env-overrides-key3-", faker.Word()),
					},
				},
			}

			overridesConfigDir := ensureTmpDir("local-source-test-env-overrides")
			if !writeConfig(overridesConfigDir, "default.json", defaultCfg) {
				panic("Failed to write config")
			}
			if !writeConfig(overridesConfigDir, "local-production.json", envOverrides) {
				panic("Failed to write config")
			}

			return testCase{
				name: "local overrides for env config",
				fields: fields{opts: []LocalOpt{
					LocalOpts.WithAppEnv(AppEnv{Name: "production"}),
					LocalOpts.WithDir(overridesConfigDir),
				}},
				args: args{
					params: []param{
						param{paramID: paramID{key: "key1"}},
						param{paramID: paramID{key: "key2"}},
						param{paramID: paramID{service: "deeply", key: "nested/key3"}},
					},
				},
				want: func(t *testing.T, got map[paramID]interface{}, err error) {
					if !assert.NoError(t, err) {
						return
					}
					deeply := envOverrides["deeply"].(map[string]interface{})
					nested := deeply["nested"].(map[string]interface{})
					assert.Equal(t, map[paramID]interface{}{
						paramID{key: "key1"}:                           envOverrides["key1"],
						paramID{key: "key2"}:                           envOverrides["key2"],
						paramID{service: "deeply", key: "nested/key3"}: nested["key3"],
					}, got)
				},
			}
		},
		func() testCase {
			envOverrides := map[string]interface{}{
				"key1": fmt.Sprint("env-facet-overrides-key1-", faker.Word()),
				"key2": fmt.Sprint("env-facet-overrides-key2-", faker.Word()),
				"deeply": map[string]interface{}{
					"nested": map[string]interface{}{
						"key3": fmt.Sprint("env-facet-overrides-key3-", faker.Word()),
					},
				},
			}

			overridesConfigDir := ensureTmpDir("local-source-test-env-facet-overrides")
			if !writeConfig(overridesConfigDir, "default.json", defaultCfg) {
				panic("Failed to write config")
			}
			if !writeConfig(overridesConfigDir, "local-production-preprod.json", envOverrides) {
				panic("Failed to write config")
			}

			return testCase{
				name: "local overrides for env facet config",
				fields: fields{opts: []LocalOpt{
					LocalOpts.WithAppEnv(AppEnv{Name: "production", Facet: "preprod"}),
					LocalOpts.WithDir(overridesConfigDir),
				}},
				args: args{
					params: []param{
						param{paramID: paramID{key: "key1"}},
						param{paramID: paramID{key: "key2"}},
						param{paramID: paramID{service: "deeply", key: "nested/key3"}},
					},
				},
				want: func(t *testing.T, got map[paramID]interface{}, err error) {
					if !assert.NoError(t, err) {
						return
					}
					deeply := envOverrides["deeply"].(map[string]interface{})
					nested := deeply["nested"].(map[string]interface{})
					assert.Equal(t, map[paramID]interface{}{
						paramID{key: "key1"}:                           envOverrides["key1"],
						paramID{key: "key2"}:                           envOverrides["key2"],
						paramID{service: "deeply", key: "nested/key3"}: nested["key3"],
					}, got)
				},
			}
		},
	}
	for _, ttFn := range tests {
		tt := ttFn()
		t.Run(tt.name, func(t *testing.T) {
			if tt.after != nil {
				defer tt.after()
			}
			source, err := NewLocalSource(tt.fields.opts...)()
			if !assert.NoError(t, err) {
				return
			}
			if !assert.NotNil(t, source, "Expected to get service instance") {
				return
			}
			params, err := source.GetParameters(context.Background(), tt.args.params)
			tt.want(t, params, err)
		})
	}
}
