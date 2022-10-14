// Code generated by MockGen. DO NOT EDIT.
// Source: internal/clients/tg/tgclient.go

// Package mock_tg is a generated GoMock package.
package mock_tg

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	telebot "gopkg.in/telebot.v3"
)

// MocktelebotReducedContext is a mock of telebotReducedContext interface.
type MocktelebotReducedContext struct {
	ctrl     *gomock.Controller
	recorder *MocktelebotReducedContextMockRecorder
}

// MocktelebotReducedContextMockRecorder is the mock recorder for MocktelebotReducedContext.
type MocktelebotReducedContextMockRecorder struct {
	mock *MocktelebotReducedContext
}

// NewMocktelebotReducedContext creates a new mock instance.
func NewMocktelebotReducedContext(ctrl *gomock.Controller) *MocktelebotReducedContext {
	mock := &MocktelebotReducedContext{ctrl: ctrl}
	mock.recorder = &MocktelebotReducedContextMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MocktelebotReducedContext) EXPECT() *MocktelebotReducedContextMockRecorder {
	return m.recorder
}

// Args mocks base method.
func (m *MocktelebotReducedContext) Args() []string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Args")
	ret0, _ := ret[0].([]string)
	return ret0
}

// Args indicates an expected call of Args.
func (mr *MocktelebotReducedContextMockRecorder) Args() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Args", reflect.TypeOf((*MocktelebotReducedContext)(nil).Args))
}

// Message mocks base method.
func (m *MocktelebotReducedContext) Message() *telebot.Message {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Message")
	ret0, _ := ret[0].(*telebot.Message)
	return ret0
}

// Message indicates an expected call of Message.
func (mr *MocktelebotReducedContextMockRecorder) Message() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Message", reflect.TypeOf((*MocktelebotReducedContext)(nil).Message))
}

// Send mocks base method.
func (m *MocktelebotReducedContext) Send(what interface{}, opts ...interface{}) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{what}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Send", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send.
func (mr *MocktelebotReducedContextMockRecorder) Send(what interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{what}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MocktelebotReducedContext)(nil).Send), varargs...)
}
