package config

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/evgeny-myasishchev/pkg/lib-core-golang/diag"
	uuid "github.com/satori/go.uuid"
)

const (
	appEnvVar = "APP_ENV"

	facetVar = "APP_ENV_FACET"

	clusterNameVar = "CLUSTER_NAME"
)

var logger = diag.CreateLogger()

// TODO: Generally do logging (maybe even with fmt)
// to be able to understand some failure cases (like ebad type)

// AppEnv represents app env
type AppEnv struct {
	// ServiceName is a name of a current service
	ServiceName string

	// Name is a env name. By default taken from APP_ENV. Corresponds to NODE_ENV
	Name string

	// Facet is a env facet like preprod (for production). By default taken from APP_ENV_FACED. Corresponds to NODE_APP_INSTANCE
	Facet string

	// Name of a cluster where service is running
	ClusterName string
}

type appEnvCfg struct {
	lookupFlag func(name string) *flag.Flag
}

type appEnvOpt func(*appEnvCfg)

func withLookupFlag(lookupFlag func(name string) *flag.Flag) appEnvOpt {
	return func(cfg *appEnvCfg) {
		cfg.lookupFlag = lookupFlag
	}
}

// NewAppEnv creates a new instance of the app env from os env
// Will use "dev" by default
func NewAppEnv(serviceName string, opts ...appEnvOpt) AppEnv {
	cfg := appEnvCfg{
		lookupFlag: flag.Lookup,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	appEnv := os.Getenv(appEnvVar)
	if appEnv == "" {
		if v := cfg.lookupFlag("test.v"); v == nil {
			appEnv = "dev"
		} else {
			appEnv = "test"
		}
	}
	facet := os.Getenv(facetVar)
	clusterName := os.Getenv(clusterNameVar)
	return AppEnv{
		Name:        appEnv,
		Facet:       facet,
		ClusterName: clusterName,
		ServiceName: serviceName,
	}
}

// ServiceConfig is a service config abstraction
type ServiceConfig interface {
	StringParam(key StringParam) StringVal
	IntParam(key IntParam) IntVal
	BoolParam(key BoolParam) BoolVal
}

// Source is an abstraction to read params
// TODO: Make sure to define String() for sources since they're being logged
// TODO: Refactor to accept context as a first arg
type Source interface {
	GetParameters(params []param) (map[param]interface{}, error)
}

type sourceBinding struct {
	params []param
	source Source
}

type serviceConfig struct {
	sources       []sourceBinding
	values        map[param]paramValue
	refreshTicker *time.Ticker
	refreshed     chan bool // TODO: Send only
	stop          <-chan bool
}

func (cfg *serviceConfig) getParamValue(key param) paramValue {
	if val, ok := cfg.values[key]; ok {
		return val
	}
	logger.Error(nil, "Unknown parameter: %v", key)
	panic(fmt.Sprintf("Unknown parameter: %v", key))
}

func (cfg *serviceConfig) StringParam(key StringParam) StringVal {
	return cfg.getParamValue(key).(StringVal)
}
func (cfg *serviceConfig) IntParam(key IntParam) IntVal {
	return cfg.getParamValue(key).(IntVal)
}
func (cfg *serviceConfig) BoolParam(key BoolParam) BoolVal {
	return cfg.getParamValue(key).(BoolVal)
}

// ServiceConfigOpt represents option of a service config
type ServiceConfigOpt func(cfg *serviceConfig)

// WithSource will add a source to fetch params from
func WithSource(source sourceBinding) ServiceConfigOpt {
	return func(cfg *serviceConfig) {
		cfg.sources = append(cfg.sources, source)
	}
}

// Used mostly for testing
func withTicker(ticker *time.Ticker) ServiceConfigOpt {
	return func(cfg *serviceConfig) {
		cfg.refreshTicker = ticker
	}
}

// Used mostly for testing
func withRefreshed(refreshed chan bool) ServiceConfigOpt {
	return func(cfg *serviceConfig) {
		cfg.refreshed = refreshed
	}
}

// Used mostly for testing
func withStop(stop chan bool) ServiceConfigOpt {
	return func(cfg *serviceConfig) {
		cfg.stop = stop
	}
}

func loadInitialValues(cfg *serviceConfig) error {
	for _, binding := range cfg.sources {
		rawValues, err := binding.source.GetParameters(binding.params)
		if err != nil {
			return err
		}
		for _, param := range binding.params {
			rawValue, ok := rawValues[param]
			if !ok {
				return fmt.Errorf("Parameter %v not found", param)
			}
			value := param.emptyValue()
			if err := value.setValue(rawValue); err != nil {
				return fmt.Errorf("Failed to set value for parameter %v: %v", param, err)
			}
			cfg.values[param] = value
		}
	}
	return nil
}

func startRefreshingValues(cfg *serviceConfig) {
	// TODO: do not refresh local source somehow
	// Perhaps add option for sourceBinding
	ctx := diag.ContextWithRequestID(context.Background(), uuid.NewV4().String())
	logger.Info(ctx, "Starting refreshing params")
	go func() {
		notifyRefreshed := func() {
			if cfg.refreshed != nil {
				cfg.refreshed <- true
			}
		}
		for {
			select {
			case <-cfg.refreshTicker.C:
				logger.Debug(ctx, "Refreshing config parameters")
				for _, binding := range cfg.sources {
					rawValues, err := binding.source.GetParameters(binding.params)
					if err != nil {
						logger.WithError(err).Error(ctx, "Failed to fetch params from source %v", binding.source)
						continue
					}

					for _, param := range binding.params {
						newVal, ok := rawValues[param]
						if !ok {
							logger.Error(ctx, "Missing value for param %v", param)
							continue
						}
						if err := cfg.values[param].setValue(newVal); err != nil {
							logger.WithError(err).Error(ctx, "Failed to refresh param: %v", param)
						}
					}
				}
				notifyRefreshed()
			case <-cfg.stop:
				logger.Debug(ctx, "Refresh stopped")
				break
			}
		}
	}()
}

func newServiceConfig(opts ...ServiceConfigOpt) *serviceConfig {
	cfg := &serviceConfig{
		sources: []sourceBinding{},
		values:  map[param]paramValue{},
	}
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.refreshTicker == nil {
		rnd := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

		// Using rand refresh intervals
		refreshIntervalSeconds := 60 + rnd.Intn(120)

		logger.Info(nil, "Using %v sec config refresh interval", refreshIntervalSeconds)

		cfg.refreshTicker = time.NewTicker(time.Duration(refreshIntervalSeconds) * time.Second)
	}

	return cfg
}

// Load will return the config with given keys
func Load(opts ...ServiceConfigOpt) (ServiceConfig, error) {
	cfg := newServiceConfig(opts...)

	if err := loadInitialValues(cfg); err != nil {
		return nil, err
	}

	startRefreshingValues(cfg)

	return cfg, nil
}
