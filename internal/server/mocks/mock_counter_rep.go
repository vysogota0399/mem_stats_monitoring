// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/vysogota0399/mem_stats_monitoring/internal/server/service (interfaces: CntrRep)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	models "github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
)

// MockCntrRep is a mock of CntrRep interface.
type MockCntrRep struct {
	ctrl     *gomock.Controller
	recorder *MockCntrRepMockRecorder
}

// MockCntrRepMockRecorder is the mock recorder for MockCntrRep.
type MockCntrRepMockRecorder struct {
	mock *MockCntrRep
}

// NewMockCntrRep creates a new mock instance.
func NewMockCntrRep(ctrl *gomock.Controller) *MockCntrRep {
	mock := &MockCntrRep{ctrl: ctrl}
	mock.recorder = &MockCntrRepMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCntrRep) EXPECT() *MockCntrRepMockRecorder {
	return m.recorder
}

// SaveCollection mocks base method.
func (m *MockCntrRep) SaveCollection(arg0 context.Context, arg1 []models.Counter) ([]models.Counter, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveCollection", arg0, arg1)
	ret0, _ := ret[0].([]models.Counter)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SaveCollection indicates an expected call of SaveCollection.
func (mr *MockCntrRepMockRecorder) SaveCollection(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveCollection", reflect.TypeOf((*MockCntrRep)(nil).SaveCollection), arg0, arg1)
}