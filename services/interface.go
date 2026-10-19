package services

import (
	"betterov2/models"
	"betterov2/repositories"
	"context"
	"time"

	"github.com/google/uuid"
)

type AccountRepository interface {
	ListAccounts(context.Context, repositories.DBTX, uuid.UUID) ([]models.Account, error)
	GetAccount(context.Context, repositories.DBTX, uuid.UUID, uuid.UUID) (models.Account, error)
	CountAccTransactions(context.Context, repositories.DBTX, uuid.UUID, uuid.UUID, models.TransactionFilter) (int, error)
	ListAccTransactions(context.Context, repositories.DBTX, uuid.UUID, uuid.UUID, models.TransactionFilter, int64) ([]models.Transaction, error)
	InsertAccount(context.Context, repositories.DBTX, uuid.UUID, models.PostAccountBody) (models.Account, error)
	UpdateAccount(context.Context, repositories.DBTX, uuid.UUID, uuid.UUID, models.PutAccountBody) (models.Account, error)
	DeleteAccount(context.Context, repositories.DBTX, uuid.UUID, uuid.UUID) (models.Account, error)
	UpdateAccountBalance(context.Context, repositories.DBTX, uuid.UUID, models.Money) (models.Money, error)
}

type AccountHistoryRepository interface {
	ListHistories(context.Context, repositories.DBTX, uuid.UUID, uuid.UUID) ([]models.AccountHistory, error)
	InsertHistory(context.Context, repositories.DBTX, models.PostAccHistBody) (models.AccountHistory, error)
}

type TransactionRepository interface {
	CountTransactions(context.Context, repositories.DBTX, uuid.UUID, models.TransactionFilter) (int, error)
	FilterTransactions(context.Context, repositories.DBTX, uuid.UUID, models.TransactionFilter, int) ([]models.Transaction, error)
	ListSuggestions(context.Context, repositories.DBTX, uuid.UUID, string) ([]models.Suggestion, error)
	GetTransaction(context.Context, repositories.DBTX, uuid.UUID, uuid.UUID) (models.Transaction, error)
	InsertTransaction(context.Context, repositories.DBTX, uuid.UUID, models.PostTransactionBody) (models.Transaction, error)
	UpdateTransaction(context.Context, repositories.DBTX, uuid.UUID, uuid.UUID, models.PutTransactionBody) (models.Transaction, error)
	DeleteTransaction(context.Context, repositories.DBTX, uuid.UUID, uuid.UUID) (models.Transaction, error)
}

type BillRepository interface {
	ListBills(context.Context, repositories.DBTX, uuid.UUID) ([]models.Bill, error)
	InsertBill(context.Context, repositories.DBTX, uuid.UUID, models.BillBody) (models.Bill, error)
	UpdateBill(context.Context, repositories.DBTX, uuid.UUID, uuid.UUID, models.BillBody) (models.Bill, error)
	DeleteBill(context.Context, repositories.DBTX, uuid.UUID, uuid.UUID) (models.Bill, error)
}

type BudgetRepository interface {
	GetBudgetPlan(context.Context, repositories.DBTX, uuid.UUID, models.IntervalType) (models.BudgetPlan, error)
	InsertBudgetPlan(context.Context, repositories.DBTX, uuid.UUID, models.PostBudgetPlanBody) (models.BudgetPlan, error)
	UpdateBudgetPlan(context.Context, repositories.DBTX, uuid.UUID, models.IntervalType, models.PutBudgetPlanBody) (models.BudgetPlan, error)
	DeleteBudgetPlan(context.Context, repositories.DBTX, uuid.UUID, models.IntervalType) (models.BudgetPlan, error)
}

type SummaryRepository interface {
	GetBasicAnalysis(context.Context, repositories.DBTX, uuid.UUID, time.Time, time.Time) (models.BasicAnalysis, error)
	GetDateToAmount(context.Context, repositories.DBTX, models.ObjectType, uuid.UUID, time.Time, time.Time) (map[string]models.Money, error)
	GetCategoryToAmount(context.Context, repositories.DBTX, models.ObjectType, uuid.UUID, time.Time, time.Time) (map[models.TransactionCategory]models.Money, error)
}

type CronAccountRepository interface {
	FilterAccounts(context.Context, repositories.DBTX, repositories.Status) ([]models.Account, error)
	MoveAccountsDueDate(context.Context, repositories.DBTX, []uuid.UUID) (int, error)
	FlagAccount(context.Context, repositories.DBTX, uuid.UUID, models.Money) (models.Account, error)
}

type CronAccountHistoryRepository interface {
	GetLatestHistory(context.Context, repositories.DBTX, uuid.UUID) (models.AccountHistory, error)
	InsertHistory(context.Context, repositories.DBTX, models.PostAccHistBody) (models.AccountHistory, error)
	DeleteOutdatedHistories(context.Context, repositories.DBTX, uuid.UUID) (int, error)
}

type CronTransactionRepository interface {
	GetTransactionSum(context.Context, repositories.DBTX, uuid.UUID, time.Time) (models.Money, error)
	DeleteOutdatedTransactions(context.Context, repositories.DBTX, uuid.UUID) (int, error)
}
