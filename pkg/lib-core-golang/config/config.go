package config

import (
	"context"
	"flag"
	"math/rand"
	"os"
	"time"

	"github.com/pkg/errors"

	uuid "github.com/satori/go.uuid"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/lib-core-golang/diag"
)

const (
	appEnvVar = "APP_ENV"

	facetVar = "APP_ENV_FACET"

	clusterNameVar = "CLUSTER_NAME"

	awsSSMEndpointURLVar          = "AWS_SSM_ENDPOINT_URL"
	awsSSMEndpointTokenVar        = "AWS_SSM_ENDPOINT_TOKEN"
	awsSSMEndpointTokenHeaderName = "x-access-token"
)

var logger = diag.CreateLogger()

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

// SourceFactory is a func that creates an instance of a source
type SourceFactory func() (Source, error)

// Source is an abstraction to read params
type Source interface {
	GetParameters(ctx context.Context, params []param) (map[paramID]interface{}, error)
}

type binding struct {
	sources        map[string]Source
	paramsBySource map[string][]param
	refreshSignal  <-chan time.Time
	onRefreshed    chan struct{}
	stopSignal     <-chan struct{}
}

func (b *binding) loadInitialValues() error {
	ctx := diag.ContextWithRequestID(context.Background(), uuid.NewV4().String())
	logger.Info(ctx, "Loading initial config values")
	for sourceName, source := range b.sources {
		sourceParams := b.paramsBySource[sourceName]
		values, err := source.GetParameters(ctx, sourceParams)
		if err != nil {
			return errors.Wrapf(err, "Failed to fetch from source %v", sourceName)
		}
		logger.WithData(diag.MsgData{
			"params": sourceParams,
		}).Debug(ctx, "Fetched %v (of %v requested) values from %v source", len(values), len(sourceParams), sourceName)
		for _, sourceParam := range sourceParams {
			value, ok := values[sourceParam.paramID]
			if !ok {
				return errors.Errorf("Parameter %v not found (source=%v)", sourceParam.paramID, sourceName)
			}

			if err := sourceParam.setValue(value); err != nil {
				return errors.Wrapf(err, "Failed to set parameter %v value (source=%v)", sourceParam.paramID, sourceName)
			}
		}
	}
	return nil
}

func (b *binding) startRefreshingValues() {
	go func() {
		if b.refreshSignal == nil {
			rnd := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

			// Using rand refresh intervals
			refreshIntervalSeconds := 60 + rnd.Intn(120)

			logger.Info(nil, "Using %v sec config refresh interval", refreshIntervalSeconds)

			b.refreshSignal = time.NewTicker(time.Duration(refreshIntervalSeconds) * time.Second).C
		}

		shouldLoop := true
		for shouldLoop {
			ctx := diag.ContextWithRequestID(context.Background(), uuid.NewV4().String())
			select {
			case <-b.refreshSignal:
				logger.Info(ctx, "Refreshing config parameters")
				for sourceName, source := range b.sources {
					sourceParams := b.paramsBySource[sourceName]
					values, err := source.GetParameters(ctx, sourceParams)
					if err != nil {
						logger.WithError(err).Error(ctx, "Failed to fetch from source: %v", sourceName)
						continue
					}
					logger.Debug(ctx, "Fetched %v (of %v requested) values from %v source", len(values), len(sourceParams), sourceName)
					for _, sourceParam := range sourceParams {
						value, ok := values[sourceParam.paramID]
						if !ok {
							logger.Error(ctx, "Parameter %v not found (source=%v)", sourceParam.paramID, sourceName)
							continue
						}

						if err := sourceParam.setValue(value); err != nil {
							logger.WithError(err).Error(ctx, "Failed to update parameter %v (source=%v)", sourceParam.paramID, sourceName)
						}
					}
				}
				b.onRefreshed <- struct{}{}
			case <-b.stopSignal:
				logger.Warn(ctx, "Refresh stopped") //We should not see this on prod
				shouldLoop = false
			}
		}
	}()
}

// BindOpt represents binding option
type BindOpt func(b *binding) error

// WithSource is a binding option
func WithSource(name string, factory SourceFactory) BindOpt {
	return func(b *binding) error {
		source, err := factory()
		if err != nil {
			return err
		}
		b.sources[name] = source
		return nil
	}
}

func withSignals(refreshSignal <-chan time.Time, onRefreshed chan struct{}, stopSignal <-chan struct{}) BindOpt {
	return func(b *binding) error {
		b.refreshSignal = refreshSignal
		b.onRefreshed = onRefreshed
		b.stopSignal = stopSignal
		return nil
	}
}

// Bind will bind the receiver config to sources and refresh values
func Bind(receiver interface{}, appEnv AppEnv, opts ...BindOpt) error {
	b := &binding{
		sources:        map[string]Source{},
		paramsBySource: map[string][]param{},
	}

	for _, opt := range opts {
		if err := opt(b); err != nil {
			return errors.Wrap(err, "Failed to process bind option")
		}
	}

	params, err := bindParamsToReceiver(receiver, appEnv.ServiceName)
	if err != nil {
		return err
	}

	for _, param := range params {
		sourceParams := b.paramsBySource[param.source]
		b.paramsBySource[param.source] = append(sourceParams, param)
	}

	if err := b.loadInitialValues(); err != nil {
		return err
	}

	b.startRefreshingValues()

	return nil
}
