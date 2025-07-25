// Code generated by MockGen. DO NOT EDIT.
// Source: api/event.go
//
// Generated by this command:
//
//	mockgen -source=api/event.go -destination=tests/mocks/mock_event.go -package=mocks
//

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockEventDispatcher is a mock of EventDispatcher interface.
type MockEventDispatcher struct {
	ctrl     *gomock.Controller
	recorder *MockEventDispatcherMockRecorder
	isgomock struct{}
}

// MockEventDispatcherMockRecorder is the mock recorder for MockEventDispatcher.
type MockEventDispatcherMockRecorder struct {
	mock *MockEventDispatcher
}

// NewMockEventDispatcher creates a new mock instance.
func NewMockEventDispatcher(ctrl *gomock.Controller) *MockEventDispatcher {
	mock := &MockEventDispatcher{ctrl: ctrl}
	mock.recorder = &MockEventDispatcherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockEventDispatcher) EXPECT() *MockEventDispatcherMockRecorder {
	return m.recorder
}

// AddListener mocks base method.
func (m *MockEventDispatcher) AddListener(arg0, arg1 any) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddListener", arg0, arg1)
}

// AddListener indicates an expected call of AddListener.
func (mr *MockEventDispatcherMockRecorder) AddListener(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddListener", reflect.TypeOf((*MockEventDispatcher)(nil).AddListener), arg0, arg1)
}

// Dispatch mocks base method.
func (m *MockEventDispatcher) Dispatch(arg0 context.Context, arg1 any) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Dispatch", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Dispatch indicates an expected call of Dispatch.
func (mr *MockEventDispatcherMockRecorder) Dispatch(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Dispatch", reflect.TypeOf((*MockEventDispatcher)(nil).Dispatch), arg0, arg1)
}
