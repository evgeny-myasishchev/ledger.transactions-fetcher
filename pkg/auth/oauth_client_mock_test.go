// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/auth (interfaces: OAuthClient)

// Package auth is a generated GoMock package.
package auth

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
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
func (m *MockOAuthClient) PerformAuthCodeExchangeFlow(arg0 context.Context, arg1 string) (*AccessToken, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PerformAuthCodeExchangeFlow", arg0, arg1)
	ret0, _ := ret[0].(*AccessToken)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PerformAuthCodeExchangeFlow indicates an expected call of PerformAuthCodeExchangeFlow
func (mr *MockOAuthClientMockRecorder) PerformAuthCodeExchangeFlow(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PerformAuthCodeExchangeFlow", reflect.TypeOf((*MockOAuthClient)(nil).PerformAuthCodeExchangeFlow), arg0, arg1)
}

// PerformRefreshFlow mocks base method
func (m *MockOAuthClient) PerformRefreshFlow(arg0 context.Context, arg1 string) (*RefreshedToken, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PerformRefreshFlow", arg0, arg1)
	ret0, _ := ret[0].(*RefreshedToken)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PerformRefreshFlow indicates an expected call of PerformRefreshFlow
func (mr *MockOAuthClientMockRecorder) PerformRefreshFlow(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PerformRefreshFlow", reflect.TypeOf((*MockOAuthClient)(nil).PerformRefreshFlow), arg0, arg1)
}
