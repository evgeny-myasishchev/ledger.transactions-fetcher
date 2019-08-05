package config

import (
	"context"
	"net/http"
	"os"

	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/version"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/lib-core-golang/diag"
)

type ssmClient interface {
	GetParameters(input *ssm.GetParametersInput) (*ssm.GetParametersOutput, error)
}

type ssmClientAuthTokenMiddleware func(req *http.Request) (*http.Response, error)

func (rt ssmClientAuthTokenMiddleware) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt(req)
}

func newSSMClientAuthTokenMiddleware(authToken string, next http.RoundTripper) http.RoundTripper {
	return ssmClientAuthTokenMiddleware(func(req *http.Request) (*http.Response, error) {
		req.Header.Add(awsSSMEndpointTokenHeaderName, authToken)
		// for debugging purposes mostly
		req.Header.Add("x-requested-by", version.AppName+"("+version.Version+")")
		return next.RoundTrip(req)
	})
}

func newSSMClient() (ssmClient, error) {
	s, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	clientCfg := aws.NewConfig()

	configURL := os.Getenv(awsSSMEndpointURLVar)

	if configURL != "" {
		logger.Info(nil, "Using service config: %v", configURL)
		httpClient := &http.Client{
			Transport: newSSMClientAuthTokenMiddleware(
				os.Getenv(awsSSMEndpointTokenVar),
				http.DefaultTransport,
			),
		}
		clientCfg = clientCfg.
			WithEndpoint(configURL).
			WithHTTPClient(httpClient)
	}

	client := ssm.New(s, clientCfg)
	return client, nil
}

type awsSSMSource struct {
	appEnv    AppEnv
	ssmClient ssmClient
}

func (s *awsSSMSource) GetParameters(ctx context.Context, params []param) (map[paramID]interface{}, error) {
	envName := s.appEnv.Name
	clusterName := s.appEnv.ClusterName
	names := make([]*string, 0, len(params)*2)
	type paramValuePrio struct {
		param param
		prio  int
	}

	nameToScore := make(map[string]*paramValuePrio, len(params))
	clusterNameToScore := make(map[string]*paramValuePrio, len(params))

	for _, param := range params {
		serviceScopedName := "/" + envName + "/" + param.service + "/" + param.key
		names = append(names, aws.String(serviceScopedName))
		score := &paramValuePrio{param: param}
		nameToScore[serviceScopedName] = score
		if clusterName != "" {
			clusterScopedName := "/" + envName + "/" + clusterName + "/" + param.service + "/" + param.key
			names = append(names, aws.String(clusterScopedName))
			clusterNameToScore[clusterScopedName] = score
		}
	}
	logger.WithData(diag.MsgData{"paths": names}).Debug(ctx, "Attempting to get SSM parameters")
	output, err := s.ssmClient.GetParameters(&ssm.GetParametersInput{
		Names: names,
	})
	if err != nil {
		return nil, err
	}

	result := make(map[paramID]interface{}, len(output.Parameters))
	for _, awsParam := range output.Parameters {
		score, ok := nameToScore[*awsParam.Name]
		if ok && score.prio < 1 {
			result[score.param.paramID] = *awsParam.Value
			score.prio = 1
		}
		score, ok = clusterNameToScore[*awsParam.Name]
		if ok {
			result[score.param.paramID] = *awsParam.Value
			score.prio = 2
		}
	}
	return result, nil
}

// AwsSSMOpt is an option of a local config source
type AwsSSMOpt func(s *awsSSMSource)

// AwsSSMOpts are options of a local source
var AwsSSMOpts = struct {
	// WithAppEnv option will sent the app env
	WithAppEnv func(appEnv AppEnv) AwsSSMOpt

	withSSMClient func(client ssmClient) AwsSSMOpt
}{
	WithAppEnv: func(appEnv AppEnv) AwsSSMOpt {
		return func(s *awsSSMSource) {
			s.appEnv = appEnv
		}
	},
	withSSMClient: func(client ssmClient) AwsSSMOpt {
		return func(s *awsSSMSource) {
			s.ssmClient = client
		}
	},
}

// NewAWSSSMSource creates a source that reads params from aws SSM.
func NewAWSSSMSource(opts ...AwsSSMOpt) SourceFactory {
	return func() (Source, error) {
		source := &awsSSMSource{}

		for _, opt := range opts {
			opt(source)
		}

		if source.ssmClient == nil {
			client, err := newSSMClient()
			if err != nil {
				return nil, err
			}
			source.ssmClient = client
		}

		return source, nil
	}
}
