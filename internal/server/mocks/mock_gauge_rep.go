// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/vysogota0399/mem_stats_monitoring/internal/server/service (interfaces: GGRep)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	models "github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
)

// MockGGRep is a mock of GGRep interface.
type MockGGRep struct {
	ctrl     *gomock.Controller
	recorder *MockGGRepMockRecorder
}

// MockGGRepMockRecorder is the mock recorder for MockGGRep.
type MockGGRepMockRecorder struct {
	mock *MockGGRep
}

// NewMockGGRep creates a new mock instance.
func NewMockGGRep(ctrl *gomock.Controller) *MockGGRep {
	mock := &MockGGRep{ctrl: ctrl}
	mock.recorder = &MockGGRepMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockGGRep) EXPECT() *MockGGRepMockRecorder {
	return m.recorder
}

// SaveCollection mocks base method.
func (m *MockGGRep) SaveCollection(arg0 context.Context, arg1 []models.Gauge) ([]models.Gauge, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveCollection", arg0, arg1)
	ret0, _ := ret[0].([]models.Gauge)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SaveCollection indicates an expected call of SaveCollection.
func (mr *MockGGRepMockRecorder) SaveCollection(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveCollection", reflect.TypeOf((*MockGGRep)(nil).SaveCollection), arg0, arg1)
}
