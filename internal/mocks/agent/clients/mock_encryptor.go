// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/vysogota0399/mem_stats_monitoring/internal/agent/clients (interfaces: Encryptor)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockEncryptor is a mock of Encryptor interface.
type MockEncryptor struct {
	ctrl     *gomock.Controller
	recorder *MockEncryptorMockRecorder
}

// MockEncryptorMockRecorder is the mock recorder for MockEncryptor.
type MockEncryptorMockRecorder struct {
	mock *MockEncryptor
}

// NewMockEncryptor creates a new mock instance.
func NewMockEncryptor(ctrl *gomock.Controller) *MockEncryptor {
	mock := &MockEncryptor{ctrl: ctrl}
	mock.recorder = &MockEncryptorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockEncryptor) EXPECT() *MockEncryptorMockRecorder {
	return m.recorder
}

// Encrypt mocks base method.
func (m *MockEncryptor) Encrypt(arg0 []byte) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Encrypt", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Encrypt indicates an expected call of Encrypt.
func (mr *MockEncryptorMockRecorder) Encrypt(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Encrypt", reflect.TypeOf((*MockEncryptor)(nil).Encrypt), arg0)
}
