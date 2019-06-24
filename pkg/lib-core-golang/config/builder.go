package config

// Builder is a tool to setup config
type Builder struct {
	appEnv         AppEnv
	paramsBuilders []*ParamsBuilder
}

// NewBuilder returns an instance of a config builder
func NewBuilder(appEnv AppEnv) *Builder {
	return &Builder{appEnv: appEnv}
}

// WithLocalSource creates a source factory for a local source
// that will point on configs dir
func (b *Builder) WithLocalSource() SourceFactory {
	return func() (Source, error) {
		return NewLocalSource(
			LocalOpts.WithAppEnv(b.appEnv),
			LocalOpts.WithIgnoreDefaultService(),
		)
	}
}

// WithRemoteSource creates a source factory for a remote source
// for dev env it will actually point on local source
func (b *Builder) WithRemoteSource() SourceFactory {
	return func() (Source, error) {
		// When running locally using local source
		if b.appEnv.Name == "dev" {
			return NewLocalSource(
				LocalOpts.WithAppEnv(b.appEnv),
				LocalOpts.WithDir("config"),
			)
		}
		logger.Info(nil, "Using AWS SSM as a remote params source")
		return NewAWSSSMSource(AwsSSMOpts.WithAppEnv(b.appEnv))
	}
}

// SourceFactory is a func that creates an instance of a source
type SourceFactory func() (Source, error)

// NewParamsBuilder is a builder to to build params bound to a given source
func (b *Builder) NewParamsBuilder(sourceFactory SourceFactory) *ParamsBuilder {
	pb := &ParamsBuilder{
		params:        []param{},
		serviceName:   b.appEnv.ServiceName,
		sourceFactory: sourceFactory,
	}
	b.paramsBuilders = append(b.paramsBuilders, pb)
	return pb
}

// LoadConfig loads the config with sources and params built
func (b *Builder) LoadConfig(loadOpts ...ServiceConfigOpt) (ServiceConfig, error) {
	// copy
	sourceOpts := make([]ServiceConfigOpt, 0, len(b.paramsBuilders))
	for _, paramsBuilder := range b.paramsBuilders {
		source, err := paramsBuilder.sourceFactory()
		if err != nil {
			return nil, err
		}
		sourceOpts = append(sourceOpts, WithSource(sourceBinding{
			params: paramsBuilder.params,
			source: source,
		}))
	}

	cfg, err := Load(append(loadOpts, sourceOpts...)...)
	if err != nil {
		logger.WithError(err).Error(nil, "Failed to load config")
		return nil, err
	}
	return cfg, nil
}

// ParamsBuilder is a tool to build params bound to particular source
type ParamsBuilder struct {
	// List of all built params
	params []param

	serviceName string

	sourceFactory SourceFactory
}

func (b *ParamsBuilder) appendParam(p param) param {
	b.params = append(b.params, p)
	return p
}

// NewParam returns an instance of a param builder
func (b *ParamsBuilder) NewParam(key string) *ParamBuilder {
	return &ParamBuilder{
		paramKey: key,
		paramSvc: b.serviceName,
		pb:       b,
	}
}

// ParamBuilder is a tool to build params
type ParamBuilder struct {
	paramKey string
	paramSvc string
	pb       *ParamsBuilder
}

// WithService binds param to a given service
// This is optional, by default will service that was used to
// initialize toplevel config
func (b *ParamBuilder) WithService(service string) *ParamBuilder {
	b.paramSvc = service
	return b
}

// Int creates an instance of an int param
func (b *ParamBuilder) Int() IntParam {
	p := newIntParam(b.paramKey, b.paramSvc)
	b.pb.appendParam(p)
	return p
}

// String creates an instance of a string param
func (b *ParamBuilder) String() StringParam {
	p := newStringParam(b.paramKey, b.paramSvc)
	b.pb.appendParam(p)
	return p
}

// Bool creates an instance of a bool param
func (b *ParamBuilder) Bool() BoolParam {
	p := newBoolParam(b.paramKey, b.paramSvc)
	b.pb.appendParam(p)
	return p
}
