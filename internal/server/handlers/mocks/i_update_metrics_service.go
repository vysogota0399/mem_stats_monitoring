// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/vysogota0399/mem_stats_monitoring/internal/server/handlers (interfaces: IUpdateMetricsService)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	service "github.com/vysogota0399/mem_stats_monitoring/internal/server/service"
)

// MockIUpdateMetricsService is a mock of IUpdateMetricsService interface.
type MockIUpdateMetricsService struct {
	ctrl     *gomock.Controller
	recorder *MockIUpdateMetricsServiceMockRecorder
}

// MockIUpdateMetricsServiceMockRecorder is the mock recorder for MockIUpdateMetricsService.
type MockIUpdateMetricsServiceMockRecorder struct {
	mock *MockIUpdateMetricsService
}

// NewMockIUpdateMetricsService creates a new mock instance.
func NewMockIUpdateMetricsService(ctrl *gomock.Controller) *MockIUpdateMetricsService {
	mock := &MockIUpdateMetricsService{ctrl: ctrl}
	mock.recorder = &MockIUpdateMetricsServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIUpdateMetricsService) EXPECT() *MockIUpdateMetricsServiceMockRecorder {
	return m.recorder
}

// Call mocks base method.
func (m *MockIUpdateMetricsService) Call(arg0 context.Context, arg1 service.UpdateMetricsServiceParams) (*service.UpdateMetricsServiceResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Call", arg0, arg1)
	ret0, _ := ret[0].(*service.UpdateMetricsServiceResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Call indicates an expected call of Call.
func (mr *MockIUpdateMetricsServiceMockRecorder) Call(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Call", reflect.TypeOf((*MockIUpdateMetricsService)(nil).Call), arg0, arg1)
}
