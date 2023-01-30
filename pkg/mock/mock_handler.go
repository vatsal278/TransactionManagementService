// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/vatsal278/TransactionManagementService/internal/handler (interfaces: TransactionManagementServiceHandler)

// Package mock is a generated GoMock package.
package mock

import (
	http "net/http"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockTransactionManagementServiceHandler is a mock of TransactionManagementServiceHandler interface.
type MockTransactionManagementServiceHandler struct {
	ctrl     *gomock.Controller
	recorder *MockTransactionManagementServiceHandlerMockRecorder
}

// MockTransactionManagementServiceHandlerMockRecorder is the mock recorder for MockTransactionManagementServiceHandler.
type MockTransactionManagementServiceHandlerMockRecorder struct {
	mock *MockTransactionManagementServiceHandler
}

// NewMockTransactionManagementServiceHandler creates a new mock instance.
func NewMockTransactionManagementServiceHandler(ctrl *gomock.Controller) *MockTransactionManagementServiceHandler {
	mock := &MockTransactionManagementServiceHandler{ctrl: ctrl}
	mock.recorder = &MockTransactionManagementServiceHandlerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTransactionManagementServiceHandler) EXPECT() *MockTransactionManagementServiceHandlerMockRecorder {
	return m.recorder
}

// GetTransactions mocks base method.
func (m *MockTransactionManagementServiceHandler) GetTransactions(arg0 http.ResponseWriter, arg1 *http.Request) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "GetTransactions", arg0, arg1)
}

// GetTransactions indicates an expected call of GetTransactions.
func (mr *MockTransactionManagementServiceHandlerMockRecorder) GetTransactions(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTransactions", reflect.TypeOf((*MockTransactionManagementServiceHandler)(nil).GetTransactions), arg0, arg1)
}

// HealthCheck mocks base method.
func (m *MockTransactionManagementServiceHandler) HealthCheck() (string, string, bool) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HealthCheck")
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(bool)
	return ret0, ret1, ret2
}

// HealthCheck indicates an expected call of HealthCheck.
func (mr *MockTransactionManagementServiceHandlerMockRecorder) HealthCheck() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HealthCheck", reflect.TypeOf((*MockTransactionManagementServiceHandler)(nil).HealthCheck))
}

// NewTransaction mocks base method.
func (m *MockTransactionManagementServiceHandler) NewTransaction(arg0 http.ResponseWriter, arg1 *http.Request) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "NewTransaction", arg0, arg1)
}

// NewTransaction indicates an expected call of NewTransaction.
func (mr *MockTransactionManagementServiceHandlerMockRecorder) NewTransaction(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewTransaction", reflect.TypeOf((*MockTransactionManagementServiceHandler)(nil).NewTransaction), arg0, arg1)
}
