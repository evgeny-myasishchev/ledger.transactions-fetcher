// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/auth/oauth.go

// Package auth is a generated GoMock package.
package auth

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockOAuthClient is a mock of OAuthClient interface
type MockOAuthClient struct {
	ctrl     *gomock.Controller
	recorder *MockOAuthClientMockRecorder
}

// MockOAuthClientMockRecorder is the mock recorder for MockOAuthClient
type MockOAuthClientMockRecorder struct {
	mock *MockOAuthClient
}

// NewMockOAuthClient creates a new mock instance
func NewMockOAuthClient(ctrl *gomock.Controller) *MockOAuthClient {
	mock := &MockOAuthClient{ctrl: ctrl}
	mock.recorder = &MockOAuthClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockOAuthClient) EXPECT() *MockOAuthClientMockRecorder {
	return m.recorder
}

// BuildCodeGrantURL mocks base method
func (m *MockOAuthClient) BuildCodeGrantURL() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BuildCodeGrantURL")
	ret0, _ := ret[0].(string)
	return ret0
}

// BuildCodeGrantURL indicates an expected call of BuildCodeGrantURL
func (mr *MockOAuthClientMockRecorder) BuildCodeGrantURL() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BuildCodeGrantURL", reflect.TypeOf((*MockOAuthClient)(nil).BuildCodeGrantURL))
}

// PerformAuthCodeExchangeFlow mocks base method
func (m *MockOAuthClient) PerformAuthCodeExchangeFlow(ctx context.Context, code string) (*AccessToken, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PerformAuthCodeExchangeFlow", ctx, code)
	ret0, _ := ret[0].(*AccessToken)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PerformAuthCodeExchangeFlow indicates an expected call of PerformAuthCodeExchangeFlow
func (mr *MockOAuthClientMockRecorder) PerformAuthCodeExchangeFlow(ctx, code interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PerformAuthCodeExchangeFlow", reflect.TypeOf((*MockOAuthClient)(nil).PerformAuthCodeExchangeFlow), ctx, code)
}
