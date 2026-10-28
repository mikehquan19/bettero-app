package services

import (
	"betterov2/models"
	"betterov2/repositories"
	"context"
	"time"

	"github.com/google/uuid"
)

type fakeUnitOfWork struct {
	err    error
	called bool
}

func (f *fakeUnitOfWork) Do(_ context.Context, work func(repositories.DBTX) error) error {
	f.called = true
	if f.err != nil {
		return f.err
	}
	return work(nil)
}

type fakeAccountRepository struct {
	listAccounts         func(context.Context, repositories.DBTX, uuid.UUID) ([]models.Account, error)
	getAccount           func(context.Context, repositories.DBTX, uuid.UUID, uuid.UUID) (models.Account, error)
	lockAccount          func(context.Context, repositories.DBTX, uuid.UUID) (models.Account, error)
	countTransactions    func(context.Context, repositories.DBTX, uuid.UUID, uuid.UUID, models.TransactionFilter) (int, error)
	listTransactions     func(context.Context, repositories.DBTX, uuid.UUID, uuid.UUID, models.TransactionFilter, int64) ([]models.Transaction, error)
	insertAccount        func(context.Context, repositories.DBTX, uuid.UUID, models.PostAccountBody) (models.Account, error)
	updateAccount        func(context.Context, repositories.DBTX, uuid.UUID, uuid.UUID, models.PutAccountBody) (models.Account, error)
	deleteAccount        func(context.Context, repositories.DBTX, uuid.UUID, uuid.UUID) (models.Account, error)
	updateAccountBalance func(context.Context, repositories.DBTX, uuid.UUID, models.Money) (models.Money, error)
}

func (f *fakeAccountRepository) ListAccounts(ctx context.Context, db repositories.DBTX, userID uuid.UUID) ([]models.Account, error) {
	return f.listAccounts(ctx, db, userID)
}
func (f *fakeAccountRepository) GetAccount(ctx context.Context, db repositories.DBTX, userID, accountID uuid.UUID) (models.Account, error) {
	return f.getAccount(ctx, db, userID, accountID)
}
func (f *fakeAccountRepository) LockAccount(ctx context.Context, db repositories.DBTX, accountID uuid.UUID) (models.Account, error) {
	if f.lockAccount == nil {
		panic("unexpected LockAccount call")
	}
	return f.lockAccount(ctx, db, accountID)
}
func (f *fakeAccountRepository) CountAccTransactions(ctx context.Context, db repositories.DBTX, userID, accountID uuid.UUID, filter models.TransactionFilter) (int, error) {
	return f.countTransactions(ctx, db, userID, accountID, filter)
}
func (f *fakeAccountRepository) ListAccTransactions(ctx context.Context, db repositories.DBTX, userID, accountID uuid.UUID, filter models.TransactionFilter, offset int64) ([]models.Transaction, error) {
	return f.listTransactions(ctx, db, userID, accountID, filter, offset)
}
func (f *fakeAccountRepository) InsertAccount(ctx context.Context, db repositories.DBTX, userID uuid.UUID, body models.PostAccountBody) (models.Account, error) {
	if f.insertAccount == nil {
		panic("unexpected InsertAccount call")
	}
	return f.insertAccount(ctx, db, userID, body)
}
func (f *fakeAccountRepository) UpdateAccount(ctx context.Context, db repositories.DBTX, userID, accountID uuid.UUID, body models.PutAccountBody) (models.Account, error) {
	if f.updateAccount == nil {
		panic("unexpected UpdateAccount call")
	}
	return f.updateAccount(ctx, db, userID, accountID, body)
}
func (f *fakeAccountRepository) DeleteAccount(ctx context.Context, db repositories.DBTX, userID, accountID uuid.UUID) (models.Account, error) {
	return f.deleteAccount(ctx, db, userID, accountID)
}
func (f *fakeAccountRepository) UpdateAccountBalance(ctx context.Context, db repositories.DBTX, accountID uuid.UUID, amount models.Money) (models.Money, error) {
	if f.updateAccountBalance == nil {
		panic("unexpected UpdateAccountBalance call")
	}
	return f.updateAccountBalance(ctx, db, accountID, amount)
}

type fakeAccountHistoryRepository struct {
	listHistories func(context.Context, repositories.DBTX, uuid.UUID, uuid.UUID) ([]models.AccountHistory, error)
	insertHistory func(context.Context, repositories.DBTX, models.PostAccHistBody) (models.AccountHistory, error)
}

func (f *fakeAccountHistoryRepository) ListHistories(ctx context.Context, db repositories.DBTX, userID, accountID uuid.UUID) ([]models.AccountHistory, error) {
	return f.listHistories(ctx, db, userID, accountID)
}
func (f *fakeAccountHistoryRepository) InsertHistory(ctx context.Context, db repositories.DBTX, body models.PostAccHistBody) (models.AccountHistory, error) {
	if f.insertHistory == nil {
		panic("unexpected InsertHistory call")
	}
	return f.insertHistory(ctx, db, body)
}

type fakeTransactionRepository struct {
	countTransactions  func(context.Context, repositories.DBTX, uuid.UUID, models.TransactionFilter) (int, error)
	filterTransactions func(context.Context, repositories.DBTX, uuid.UUID, models.TransactionFilter, int) ([]models.Transaction, error)
	listSuggestions    func(context.Context, repositories.DBTX, uuid.UUID, string) ([]models.Suggestion, error)
	getTransaction     func(context.Context, repositories.DBTX, uuid.UUID, uuid.UUID) (models.Transaction, error)
	insertTransaction  func(context.Context, repositories.DBTX, uuid.UUID, models.PostTransactionBody) (models.Transaction, error)
	updateTransaction  func(context.Context, repositories.DBTX, uuid.UUID, uuid.UUID, models.PutTransactionBody) (models.Transaction, error)
	deleteTransaction  func(context.Context, repositories.DBTX, uuid.UUID, uuid.UUID) (models.Transaction, error)
}

func (f *fakeTransactionRepository) CountTransactions(ctx context.Context, db repositories.DBTX, userID uuid.UUID, filter models.TransactionFilter) (int, error) {
	return f.countTransactions(ctx, db, userID, filter)
}
func (f *fakeTransactionRepository) FilterTransactions(ctx context.Context, db repositories.DBTX, userID uuid.UUID, filter models.TransactionFilter, offset int) ([]models.Transaction, error) {
	return f.filterTransactions(ctx, db, userID, filter, offset)
}
func (f *fakeTransactionRepository) ListSuggestions(ctx context.Context, db repositories.DBTX, userID uuid.UUID, keyword string) ([]models.Suggestion, error) {
	return f.listSuggestions(ctx, db, userID, keyword)
}
func (f *fakeTransactionRepository) GetTransaction(ctx context.Context, db repositories.DBTX, userID, transactionID uuid.UUID) (models.Transaction, error) {
	if f.getTransaction == nil {
		panic("unexpected GetTransaction call")
	}
	return f.getTransaction(ctx, db, userID, transactionID)
}
func (f *fakeTransactionRepository) InsertTransaction(ctx context.Context, db repositories.DBTX, userID uuid.UUID, body models.PostTransactionBody) (models.Transaction, error) {
	if f.insertTransaction == nil {
		panic("unexpected InsertTransaction call")
	}
	return f.insertTransaction(ctx, db, userID, body)
}
func (f *fakeTransactionRepository) UpdateTransaction(ctx context.Context, db repositories.DBTX, userID, transactionID uuid.UUID, body models.PutTransactionBody) (models.Transaction, error) {
	if f.updateTransaction == nil {
		panic("unexpected UpdateTransaction call")
	}
	return f.updateTransaction(ctx, db, userID, transactionID, body)
}
func (f *fakeTransactionRepository) DeleteTransaction(ctx context.Context, db repositories.DBTX, userID, transactionID uuid.UUID) (models.Transaction, error) {
	if f.deleteTransaction == nil {
		panic("unexpected DeleteTransaction call")
	}
	return f.deleteTransaction(ctx, db, userID, transactionID)
}

type fakeBillRepository struct {
	listBills  func(context.Context, repositories.DBTX, uuid.UUID) ([]models.Bill, error)
	insertBill func(context.Context, repositories.DBTX, uuid.UUID, models.BillBody) (models.Bill, error)
	updateBill func(context.Context, repositories.DBTX, uuid.UUID, uuid.UUID, models.BillBody) (models.Bill, error)
	deleteBill func(context.Context, repositories.DBTX, uuid.UUID, uuid.UUID) (models.Bill, error)
}

func (f *fakeBillRepository) ListBills(ctx context.Context, db repositories.DBTX, userID uuid.UUID) ([]models.Bill, error) {
	return f.listBills(ctx, db, userID)
}
func (f *fakeBillRepository) InsertBill(ctx context.Context, db repositories.DBTX, userID uuid.UUID, body models.BillBody) (models.Bill, error) {
	return f.insertBill(ctx, db, userID, body)
}
func (f *fakeBillRepository) UpdateBill(ctx context.Context, db repositories.DBTX, userID, billID uuid.UUID, body models.BillBody) (models.Bill, error) {
	return f.updateBill(ctx, db, userID, billID, body)
}
func (f *fakeBillRepository) DeleteBill(ctx context.Context, db repositories.DBTX, userID, billID uuid.UUID) (models.Bill, error) {
	return f.deleteBill(ctx, db, userID, billID)
}

type fakeBudgetRepository struct {
	getBudgetPlan    func(context.Context, repositories.DBTX, uuid.UUID, models.IntervalType) (models.BudgetPlan, error)
	insertBudgetPlan func(context.Context, repositories.DBTX, uuid.UUID, models.PostBudgetPlanBody) (models.BudgetPlan, error)
	updateBudgetPlan func(context.Context, repositories.DBTX, uuid.UUID, models.IntervalType, models.PutBudgetPlanBody) (models.BudgetPlan, error)
	deleteBudgetPlan func(context.Context, repositories.DBTX, uuid.UUID, models.IntervalType) (models.BudgetPlan, error)
}

func (f *fakeBudgetRepository) GetBudgetPlan(ctx context.Context, db repositories.DBTX, userID uuid.UUID, interval models.IntervalType) (models.BudgetPlan, error) {
	return f.getBudgetPlan(ctx, db, userID, interval)
}
func (f *fakeBudgetRepository) InsertBudgetPlan(ctx context.Context, db repositories.DBTX, userID uuid.UUID, body models.PostBudgetPlanBody) (models.BudgetPlan, error) {
	return f.insertBudgetPlan(ctx, db, userID, body)
}
func (f *fakeBudgetRepository) UpdateBudgetPlan(ctx context.Context, db repositories.DBTX, userID uuid.UUID, interval models.IntervalType, body models.PutBudgetPlanBody) (models.BudgetPlan, error) {
	return f.updateBudgetPlan(ctx, db, userID, interval, body)
}
func (f *fakeBudgetRepository) DeleteBudgetPlan(ctx context.Context, db repositories.DBTX, userID uuid.UUID, interval models.IntervalType) (models.BudgetPlan, error) {
	return f.deleteBudgetPlan(ctx, db, userID, interval)
}

type fakeSummaryRepository struct {
	getBasicAnalysis    func(context.Context, repositories.DBTX, uuid.UUID, time.Time, time.Time) (models.BasicAnalysis, error)
	getDateToAmount     func(context.Context, repositories.DBTX, models.ObjectType, uuid.UUID, time.Time, time.Time) (map[string]models.Money, error)
	getCategoryToAmount func(context.Context, repositories.DBTX, models.ObjectType, uuid.UUID, time.Time, time.Time) (map[models.TransactionCategory]models.Money, error)
}

func (f *fakeSummaryRepository) GetBasicAnalysis(ctx context.Context, db repositories.DBTX, userID uuid.UUID, start, end time.Time) (models.BasicAnalysis, error) {
	return f.getBasicAnalysis(ctx, db, userID, start, end)
}
func (f *fakeSummaryRepository) GetDateToAmount(ctx context.Context, db repositories.DBTX, objectType models.ObjectType, objectID uuid.UUID, start, end time.Time) (map[string]models.Money, error) {
	return f.getDateToAmount(ctx, db, objectType, objectID, start, end)
}
func (f *fakeSummaryRepository) GetCategoryToAmount(ctx context.Context, db repositories.DBTX, objectType models.ObjectType, objectID uuid.UUID, start, end time.Time) (map[models.TransactionCategory]models.Money, error) {
	return f.getCategoryToAmount(ctx, db, objectType, objectID, start, end)
}

var _ AccountRepository = (*fakeAccountRepository)(nil)
var _ AccountHistoryRepository = (*fakeAccountHistoryRepository)(nil)
var _ TransactionRepository = (*fakeTransactionRepository)(nil)
var _ BillRepository = (*fakeBillRepository)(nil)
var _ BudgetRepository = (*fakeBudgetRepository)(nil)
var _ SummaryRepository = (*fakeSummaryRepository)(nil)
