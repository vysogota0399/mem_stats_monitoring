// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/vysogota0399/mem_stats_monitoring/internal/server/storage (interfaces: DBAble)

// Package mocks is a generated GoMock package.
package mocks

import (
	sql "database/sql"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockDBAble is a mock of DBAble interface.
type MockDBAble struct {
	ctrl     *gomock.Controller
	recorder *MockDBAbleMockRecorder
}

// MockDBAbleMockRecorder is the mock recorder for MockDBAble.
type MockDBAbleMockRecorder struct {
	mock *MockDBAble
}

// NewMockDBAble creates a new mock instance.
func NewMockDBAble(ctrl *gomock.Controller) *MockDBAble {
	mock := &MockDBAble{ctrl: ctrl}
	mock.recorder = &MockDBAbleMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDBAble) EXPECT() *MockDBAbleMockRecorder {
	return m.recorder
}

// All mocks base method.
func (m *MockDBAble) All() map[string]map[string][]string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "All")
	ret0, _ := ret[0].(map[string]map[string][]string)
	return ret0
}

// All indicates an expected call of All.
func (mr *MockDBAbleMockRecorder) All() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "All", reflect.TypeOf((*MockDBAble)(nil).All))
}

// DB mocks base method.
func (m *MockDBAble) DB() *sql.DB {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DB")
	ret0, _ := ret[0].(*sql.DB)
	return ret0
}

// DB indicates an expected call of DB.
func (mr *MockDBAbleMockRecorder) DB() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DB", reflect.TypeOf((*MockDBAble)(nil).DB))
}

// Last mocks base method.
func (m *MockDBAble) Last(arg0, arg1 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Last", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Last indicates an expected call of Last.
func (mr *MockDBAbleMockRecorder) Last(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Last", reflect.TypeOf((*MockDBAble)(nil).Last), arg0, arg1)
}

// Ping mocks base method.
func (m *MockDBAble) Ping() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Ping")
	ret0, _ := ret[0].(error)
	return ret0
}

// Ping indicates an expected call of Ping.
func (mr *MockDBAbleMockRecorder) Ping() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ping", reflect.TypeOf((*MockDBAble)(nil).Ping))
}

// Push mocks base method.
func (m *MockDBAble) Push(arg0, arg1 string, arg2 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Push", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Push indicates an expected call of Push.
func (mr *MockDBAbleMockRecorder) Push(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Push", reflect.TypeOf((*MockDBAble)(nil).Push), arg0, arg1, arg2)
}