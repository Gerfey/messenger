// Code generated by MockGen. DO NOT EDIT.
// Source: api/bus.go
//
// Generated by this command:
//
//	mockgen -source=api/bus.go -destination=tests/mocks/mock_bus.go -package=mocks
//

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	api "github.com/gerfey/messenger/api"
	gomock "go.uber.org/mock/gomock"
)

// MockMessageBus is a mock of MessageBus interface.
type MockMessageBus struct {
	ctrl     *gomock.Controller
	recorder *MockMessageBusMockRecorder
	isgomock struct{}
}

// MockMessageBusMockRecorder is the mock recorder for MockMessageBus.
type MockMessageBusMockRecorder struct {
	mock *MockMessageBus
}

// NewMockMessageBus creates a new mock instance.
func NewMockMessageBus(ctrl *gomock.Controller) *MockMessageBus {
	mock := &MockMessageBus{ctrl: ctrl}
	mock.recorder = &MockMessageBusMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMessageBus) EXPECT() *MockMessageBusMockRecorder {
	return m.recorder
}

// Dispatch mocks base method.
func (m *MockMessageBus) Dispatch(arg0 context.Context, arg1 any, arg2 ...api.Stamp) (api.Envelope, error) {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Dispatch", varargs...)
	ret0, _ := ret[0].(api.Envelope)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Dispatch indicates an expected call of Dispatch.
func (mr *MockMessageBusMockRecorder) Dispatch(arg0, arg1 any, arg2 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Dispatch", reflect.TypeOf((*MockMessageBus)(nil).Dispatch), varargs...)
}

// MockBusLocator is a mock of BusLocator interface.
type MockBusLocator struct {
	ctrl     *gomock.Controller
	recorder *MockBusLocatorMockRecorder
	isgomock struct{}
}

// MockBusLocatorMockRecorder is the mock recorder for MockBusLocator.
type MockBusLocatorMockRecorder struct {
	mock *MockBusLocator
}

// NewMockBusLocator creates a new mock instance.
func NewMockBusLocator(ctrl *gomock.Controller) *MockBusLocator {
	mock := &MockBusLocator{ctrl: ctrl}
	mock.recorder = &MockBusLocatorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBusLocator) EXPECT() *MockBusLocatorMockRecorder {
	return m.recorder
}

// Get mocks base method.
func (m *MockBusLocator) Get(arg0 string) (api.MessageBus, bool) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", arg0)
	ret0, _ := ret[0].(api.MessageBus)
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockBusLocatorMockRecorder) Get(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockBusLocator)(nil).Get), arg0)
}

// GetAll mocks base method.
func (m *MockBusLocator) GetAll() []api.MessageBus {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAll")
	ret0, _ := ret[0].([]api.MessageBus)
	return ret0
}

// GetAll indicates an expected call of GetAll.
func (mr *MockBusLocatorMockRecorder) GetAll() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAll", reflect.TypeOf((*MockBusLocator)(nil).GetAll))
}

// Register mocks base method.
func (m *MockBusLocator) Register(arg0 string, arg1 api.MessageBus) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Register", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Register indicates an expected call of Register.
func (mr *MockBusLocatorMockRecorder) Register(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Register", reflect.TypeOf((*MockBusLocator)(nil).Register), arg0, arg1)
}
