package config

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockSSMClient struct {
	mock.Mock
}

func (m *mockSSMClient) GetParameters(input *ssm.GetParametersInput) (*ssm.GetParametersOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*ssm.GetParametersOutput), args.Error(1)
}

func Test_awsSSMSource_GetParameters(t *testing.T) {
	mockAppEnv := func() AppEnv {
		return AppEnv{Name: "env" + faker.Word(), ClusterName: "cluster-" + faker.Word()}
	}

	mockParam := func(key string) (param, string) {
		fullKey := key + "-" + faker.Word()
		p := param{paramID: paramID{key: fullKey, service: "svc-" + faker.Word()}}
		return p, fullKey + "-value-" + faker.Word()
	}

	type fields struct {
		appEnv    AppEnv
		ssmClient ssmClient
	}
	type args struct {
		params []param
	}
	type testCase struct {
		name   string
		fields fields
		args   args
		assert func(t *testing.T, values map[paramID]interface{}, err error)
	}
	tests := []func() testCase{
		func() testCase {
			appEnv := mockAppEnv()
			appEnv.ClusterName = ""
			mockClient := &mockSSMClient{}
			p1, p1val := mockParam("p1")
			p2, p2val := mockParam("p2")
			p3, p3val := mockParam("p3")

			p1Name := aws.String("/" + appEnv.Name + "/" + p1.service + "/" + p1.key)
			p2Name := aws.String("/" + appEnv.Name + "/" + p2.service + "/" + p2.key)
			p3Name := aws.String("/" + appEnv.Name + "/" + p3.service + "/" + p3.key)

			input := &ssm.GetParametersInput{
				Names: []*string{p1Name, p2Name, p3Name},
			}

			output := &ssm.GetParametersOutput{
				Parameters: []*ssm.Parameter{
					&ssm.Parameter{Name: p1Name, Value: aws.String(p1val)},
					&ssm.Parameter{Name: p2Name, Value: aws.String(p2val)},
					&ssm.Parameter{Name: p3Name, Value: aws.String(p3val)},
				},
			}

			mockClient.On("GetParameters", input).Return(output, nil)

			return testCase{
				name:   "fetch service params",
				fields: fields{appEnv: appEnv, ssmClient: mockClient},
				args:   args{params: []param{p1, p2, p3}},
				assert: func(t *testing.T, values map[paramID]interface{}, err error) {
					if !assert.NoError(t, err) {
						return
					}
					assert.Len(t, values, 3)
					assert.Equal(t, p1val, values[p1.paramID])
					assert.Equal(t, p2val, values[p2.paramID])
					assert.Equal(t, p3val, values[p3.paramID])
					mockClient.AssertExpectations(t)
				},
			}
		},
		func() testCase {
			appEnv := mockAppEnv()
			mockClient := &mockSSMClient{}
			p1, p1val := mockParam("p1")
			p2, p2val := mockParam("p2")
			p3, p3val := mockParam("p3")

			p1cVal, p2cVal := p1val+"-cluster", p2val+"-cluster"

			p1Name := aws.String("/" + appEnv.Name + "/" + p1.service + "/" + p1.key)
			p2Name := aws.String("/" + appEnv.Name + "/" + p2.service + "/" + p2.key)
			p3Name := aws.String("/" + appEnv.Name + "/" + p3.service + "/" + p3.key)

			p1cName := aws.String("/" + appEnv.Name + "/" + appEnv.ClusterName + "/" + p1.service + "/" + p1.key)
			p2cName := aws.String("/" + appEnv.Name + "/" + appEnv.ClusterName + "/" + p2.service + "/" + p2.key)
			p3cName := aws.String("/" + appEnv.Name + "/" + appEnv.ClusterName + "/" + p3.service + "/" + p3.key)

			input := &ssm.GetParametersInput{
				Names: []*string{
					p1Name, p1cName,
					p2Name, p2cName,
					p3Name, p3cName,
				},
			}

			output := &ssm.GetParametersOutput{
				Parameters: []*ssm.Parameter{
					&ssm.Parameter{Name: p1Name, Value: aws.String(p1val)},
					&ssm.Parameter{Name: p2Name, Value: aws.String(p2val)},
					&ssm.Parameter{Name: p3Name, Value: aws.String(p3val)},

					&ssm.Parameter{Name: p1cName, Value: aws.String(p1cVal)},
					&ssm.Parameter{Name: p2cName, Value: aws.String(p2cVal)},
				},
			}

			mockClient.On("GetParameters", input).Return(output, nil)
			return testCase{
				name:   "prefer cluster specific params",
				fields: fields{appEnv: appEnv, ssmClient: mockClient},
				args:   args{params: []param{p1, p2, p3}},
				assert: func(t *testing.T, values map[paramID]interface{}, err error) {
					if !assert.NoError(t, err) {
						return
					}
					assert.Len(t, values, 3)
					assert.Equal(t, p1cVal, values[p1.paramID])
					assert.Equal(t, p2cVal, values[p2.paramID])
					assert.Equal(t, p3val, values[p3.paramID])
					mockClient.AssertExpectations(t)
				},
			}
		},
	}
	for _, tt := range tests {
		tt := tt()
		t.Run(tt.name, func(t *testing.T) {
			s, err := NewAWSSSMSource(AwsSSMOpts.WithAppEnv(tt.fields.appEnv), AwsSSMOpts.withSSMClient(tt.fields.ssmClient))()
			if !assert.NoError(t, err) {
				return
			}
			got, err := s.GetParameters(context.Background(), tt.args.params)
			tt.assert(t, got, err)
		})
	}
}
