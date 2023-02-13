// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/vatsal278/TransactionManagementService/internal/logic (interfaces: TransactionManagementServiceLogicIer)

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	model "github.com/PereRohit/util/model"
	gomock "github.com/golang/mock/gomock"
	model0 "github.com/vatsal278/TransactionManagementService/internal/model"
)

// MockTransactionManagementServiceLogicIer is a mock of TransactionManagementServiceLogicIer interface.
type MockTransactionManagementServiceLogicIer struct {
	ctrl     *gomock.Controller
	recorder *MockTransactionManagementServiceLogicIerMockRecorder
}

// MockTransactionManagementServiceLogicIerMockRecorder is the mock recorder for MockTransactionManagementServiceLogicIer.
type MockTransactionManagementServiceLogicIerMockRecorder struct {
	mock *MockTransactionManagementServiceLogicIer
}

// NewMockTransactionManagementServiceLogicIer creates a new mock instance.
func NewMockTransactionManagementServiceLogicIer(ctrl *gomock.Controller) *MockTransactionManagementServiceLogicIer {
	mock := &MockTransactionManagementServiceLogicIer{ctrl: ctrl}
	mock.recorder = &MockTransactionManagementServiceLogicIerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTransactionManagementServiceLogicIer) EXPECT() *MockTransactionManagementServiceLogicIerMockRecorder {
	return m.recorder
}

// GetTransactions mocks base method.
func (m *MockTransactionManagementServiceLogicIer) GetTransactions(arg0 string, arg1, arg2 int) *model.Response {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTransactions", arg0, arg1, arg2)
	ret0, _ := ret[0].(*model.Response)
	return ret0
}

// GetTransactions indicates an expected call of GetTransactions.
func (mr *MockTransactionManagementServiceLogicIerMockRecorder) GetTransactions(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTransactions", reflect.TypeOf((*MockTransactionManagementServiceLogicIer)(nil).GetTransactions), arg0, arg1, arg2)
}

// HealthCheck mocks base method.
func (m *MockTransactionManagementServiceLogicIer) HealthCheck() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HealthCheck")
	ret0, _ := ret[0].(bool)
	return ret0
}

// HealthCheck indicates an expected call of HealthCheck.
func (mr *MockTransactionManagementServiceLogicIerMockRecorder) HealthCheck() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HealthCheck", reflect.TypeOf((*MockTransactionManagementServiceLogicIer)(nil).HealthCheck))
}

// NewTransaction mocks base method.
func (m *MockTransactionManagementServiceLogicIer) NewTransaction(arg0 model0.NewTransaction) *model.Response {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewTransaction", arg0)
	ret0, _ := ret[0].(*model.Response)
	return ret0
}

// NewTransaction indicates an expected call of NewTransaction.
func (mr *MockTransactionManagementServiceLogicIerMockRecorder) NewTransaction(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewTransaction", reflect.TypeOf((*MockTransactionManagementServiceLogicIer)(nil).NewTransaction), arg0)
}
