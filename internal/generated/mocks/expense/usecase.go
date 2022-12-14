// Code generated by MockGen. DO NOT EDIT.
// Source: internal/expense/usecase.go

// Package mock_expense is a generated GoMock package.
package mock_expense

import (
	context "context"
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
	expense "gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/expense"
	models "gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
)

// MockUseCase is a mock of UseCase interface.
type MockUseCase struct {
	ctrl     *gomock.Controller
	recorder *MockUseCaseMockRecorder
}

// MockUseCaseMockRecorder is the mock recorder for MockUseCase.
type MockUseCaseMockRecorder struct {
	mock *MockUseCase
}

// NewMockUseCase creates a new mock instance.
func NewMockUseCase(ctrl *gomock.Controller) *MockUseCase {
	mock := &MockUseCase{ctrl: ctrl}
	mock.recorder = &MockUseCaseMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUseCase) EXPECT() *MockUseCaseMockRecorder {
	return m.recorder
}

// AddExpense mocks base method.
func (m *MockUseCase) AddExpense(ctx context.Context, userID models.UserID, expense models.Expense) (models.Expense, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddExpense", ctx, userID, expense)
	ret0, _ := ret[0].(models.Expense)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AddExpense indicates an expected call of AddExpense.
func (mr *MockUseCaseMockRecorder) AddExpense(ctx, userID, expense interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddExpense", reflect.TypeOf((*MockUseCase)(nil).AddExpense), ctx, userID, expense)
}

// GetExpensesAscendSinceTill mocks base method.
func (m *MockUseCase) GetExpensesAscendSinceTill(ctx context.Context, userID models.UserID, since, till time.Time, max int) ([]models.Expense, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetExpensesAscendSinceTill", ctx, userID, since, till, max)
	ret0, _ := ret[0].([]models.Expense)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetExpensesAscendSinceTill indicates an expected call of GetExpensesAscendSinceTill.
func (mr *MockUseCaseMockRecorder) GetExpensesAscendSinceTill(ctx, userID, since, till, max interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetExpensesAscendSinceTill", reflect.TypeOf((*MockUseCase)(nil).GetExpensesAscendSinceTill), ctx, userID, since, till, max)
}

// GetExpensesSummaryByCategorySince mocks base method.
func (m *MockUseCase) GetExpensesSummaryByCategorySince(ctx context.Context, userID models.UserID, since, till time.Time) (expense.SummaryReport, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetExpensesSummaryByCategorySince", ctx, userID, since, till)
	ret0, _ := ret[0].(expense.SummaryReport)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetExpensesSummaryByCategorySince indicates an expected call of GetExpensesSummaryByCategorySince.
func (mr *MockUseCaseMockRecorder) GetExpensesSummaryByCategorySince(ctx, userID, since, till interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetExpensesSummaryByCategorySince", reflect.TypeOf((*MockUseCase)(nil).GetExpensesSummaryByCategorySince), ctx, userID, since, till)
}

// MockExtendedUseCase is a mock of ExtendedUseCase interface.
type MockExtendedUseCase struct {
	ctrl     *gomock.Controller
	recorder *MockExtendedUseCaseMockRecorder
}

// MockExtendedUseCaseMockRecorder is the mock recorder for MockExtendedUseCase.
type MockExtendedUseCaseMockRecorder struct {
	mock *MockExtendedUseCase
}

// NewMockExtendedUseCase creates a new mock instance.
func NewMockExtendedUseCase(ctrl *gomock.Controller) *MockExtendedUseCase {
	mock := &MockExtendedUseCase{ctrl: ctrl}
	mock.recorder = &MockExtendedUseCaseMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockExtendedUseCase) EXPECT() *MockExtendedUseCaseMockRecorder {
	return m.recorder
}

// AddExpense mocks base method.
func (m *MockExtendedUseCase) AddExpense(ctx context.Context, userID models.UserID, expense models.Expense) (models.Expense, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddExpense", ctx, userID, expense)
	ret0, _ := ret[0].(models.Expense)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AddExpense indicates an expected call of AddExpense.
func (mr *MockExtendedUseCaseMockRecorder) AddExpense(ctx, userID, expense interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddExpense", reflect.TypeOf((*MockExtendedUseCase)(nil).AddExpense), ctx, userID, expense)
}

// GetExpensesAscendSinceTill mocks base method.
func (m *MockExtendedUseCase) GetExpensesAscendSinceTill(ctx context.Context, userID models.UserID, since, till time.Time, max int) ([]models.Expense, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetExpensesAscendSinceTill", ctx, userID, since, till, max)
	ret0, _ := ret[0].([]models.Expense)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetExpensesAscendSinceTill indicates an expected call of GetExpensesAscendSinceTill.
func (mr *MockExtendedUseCaseMockRecorder) GetExpensesAscendSinceTill(ctx, userID, since, till, max interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetExpensesAscendSinceTill", reflect.TypeOf((*MockExtendedUseCase)(nil).GetExpensesAscendSinceTill), ctx, userID, since, till, max)
}

// GetExpensesSummaryByCategorySince mocks base method.
func (m *MockExtendedUseCase) GetExpensesSummaryByCategorySince(ctx context.Context, userID models.UserID, since, till time.Time) (expense.SummaryReport, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetExpensesSummaryByCategorySince", ctx, userID, since, till)
	ret0, _ := ret[0].(expense.SummaryReport)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetExpensesSummaryByCategorySince indicates an expected call of GetExpensesSummaryByCategorySince.
func (mr *MockExtendedUseCaseMockRecorder) GetExpensesSummaryByCategorySince(ctx, userID, since, till interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetExpensesSummaryByCategorySince", reflect.TypeOf((*MockExtendedUseCase)(nil).GetExpensesSummaryByCategorySince), ctx, userID, since, till)
}

// SendGetExpensesSummaryByCategorySinceRequest mocks base method.
func (m *MockExtendedUseCase) SendGetExpensesSummaryByCategorySinceRequest(ctx context.Context, chatID int64, userID models.UserID, since, till time.Time) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendGetExpensesSummaryByCategorySinceRequest", ctx, chatID, userID, since, till)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendGetExpensesSummaryByCategorySinceRequest indicates an expected call of SendGetExpensesSummaryByCategorySinceRequest.
func (mr *MockExtendedUseCaseMockRecorder) SendGetExpensesSummaryByCategorySinceRequest(ctx, chatID, userID, since, till interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendGetExpensesSummaryByCategorySinceRequest", reflect.TypeOf((*MockExtendedUseCase)(nil).SendGetExpensesSummaryByCategorySinceRequest), ctx, chatID, userID, since, till)
}

// MockReportsCache is a mock of ReportsCache interface.
type MockReportsCache struct {
	ctrl     *gomock.Controller
	recorder *MockReportsCacheMockRecorder
}

// MockReportsCacheMockRecorder is the mock recorder for MockReportsCache.
type MockReportsCacheMockRecorder struct {
	mock *MockReportsCache
}

// NewMockReportsCache creates a new mock instance.
func NewMockReportsCache(ctrl *gomock.Controller) *MockReportsCache {
	mock := &MockReportsCache{ctrl: ctrl}
	mock.recorder = &MockReportsCacheMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockReportsCache) EXPECT() *MockReportsCacheMockRecorder {
	return m.recorder
}

// AddToCache mocks base method.
func (m *MockReportsCache) AddToCache(ctx context.Context, userID models.UserID, since, till time.Time, report expense.SummaryReport) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddToCache", ctx, userID, since, till, report)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddToCache indicates an expected call of AddToCache.
func (mr *MockReportsCacheMockRecorder) AddToCache(ctx, userID, since, till, report interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddToCache", reflect.TypeOf((*MockReportsCache)(nil).AddToCache), ctx, userID, since, till, report)
}

// DropCacheForUserID mocks base method.
func (m *MockReportsCache) DropCacheForUserID(ctx context.Context, userID models.UserID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DropCacheForUserID", ctx, userID)
	ret0, _ := ret[0].(error)
	return ret0
}

// DropCacheForUserID indicates an expected call of DropCacheForUserID.
func (mr *MockReportsCacheMockRecorder) DropCacheForUserID(ctx, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DropCacheForUserID", reflect.TypeOf((*MockReportsCache)(nil).DropCacheForUserID), ctx, userID)
}

// GetFromCache mocks base method.
func (m *MockReportsCache) GetFromCache(ctx context.Context, userID models.UserID, since, till time.Time) (expense.SummaryReport, bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFromCache", ctx, userID, since, till)
	ret0, _ := ret[0].(expense.SummaryReport)
	ret1, _ := ret[1].(bool)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetFromCache indicates an expected call of GetFromCache.
func (mr *MockReportsCacheMockRecorder) GetFromCache(ctx, userID, since, till interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFromCache", reflect.TypeOf((*MockReportsCache)(nil).GetFromCache), ctx, userID, since, till)
}
