package services

import (
	"betterov2/models"
	"betterov2/repositories"
	"context"
	"time"

	"github.com/google/uuid"
)

type AccountRepository interface {
	// ListAccounts returns all accounts owned by the user, ordered by update recency.
	ListAccounts(context.Context, repositories.DBTX, uuid.UUID) ([]models.Account, error)
	// GetAccount returns an account only when it belongs to the given user.
	GetAccount(context.Context, repositories.DBTX, uuid.UUID, uuid.UUID) (models.Account, error)
	// LockAccount locks and returns an account for the duration of the current database transaction.
	LockAccount(context.Context, repositories.DBTX, uuid.UUID) (models.Account, error)
	// CountAccTransactions returns the number of an account's transactions matching the filter.
	CountAccTransactions(context.Context, repositories.DBTX, uuid.UUID, uuid.UUID, models.TransactionFilter) (int, error)
	// ListAccTransactions returns one paginated page of an account's transactions matching the filter.
	ListAccTransactions(context.Context, repositories.DBTX, uuid.UUID, uuid.UUID, models.TransactionFilter, int64) ([]models.Transaction, error)
	// InsertAccount creates and returns an account owned by the given user.
	InsertAccount(context.Context, repositories.DBTX, uuid.UUID, models.PostAccountBody) (models.Account, error)
	// UpdateAccount updates and returns an account without allowing its type to change.
	UpdateAccount(context.Context, repositories.DBTX, uuid.UUID, uuid.UUID, models.PutAccountBody) (models.Account, error)
	// DeleteAccount deletes and returns an account owned by the given user.
	DeleteAccount(context.Context, repositories.DBTX, uuid.UUID, uuid.UUID) (models.Account, error)
	// UpdateAccountBalance applies a net change and returns the resulting balance.
	// A positive net change decreases debit balances and increases credit balances.
	UpdateAccountBalance(context.Context, repositories.DBTX, uuid.UUID, models.Money) (models.Money, error)
}

type AccountHistoryRepository interface {
	// ListHistories returns the balance history of an account owned by the given user.
	ListHistories(context.Context, repositories.DBTX, uuid.UUID, uuid.UUID) ([]models.AccountHistory, error)
	// InsertHistory creates and returns an account balance checkpoint.
	InsertHistory(context.Context, repositories.DBTX, models.PostAccHistBody) (models.AccountHistory, error)
}

type TransactionRepository interface {
	// CountTransactions returns the number of a user's transactions matching the filter.
	CountTransactions(context.Context, repositories.DBTX, uuid.UUID, models.TransactionFilter) (int, error)
	// FilterTransactions returns one paginated page of a user's transactions matching the filter.
	FilterTransactions(context.Context, repositories.DBTX, uuid.UUID, models.TransactionFilter, int) ([]models.Transaction, error)
	// ListSuggestions returns autocomplete suggestions from a user's transaction data.
	ListSuggestions(context.Context, repositories.DBTX, uuid.UUID, string) ([]models.Suggestion, error)
	// GetTransaction locks and returns a transaction with its nested account when owned by the user.
	GetTransaction(context.Context, repositories.DBTX, uuid.UUID, uuid.UUID) (models.Transaction, error)
	// InsertTransaction creates and returns a transaction for an account owned by the user.
	InsertTransaction(context.Context, repositories.DBTX, uuid.UUID, models.PostTransactionBody) (models.Transaction, error)
	// UpdateTransaction updates and returns a transaction without allowing its account to change.
	UpdateTransaction(context.Context, repositories.DBTX, uuid.UUID, uuid.UUID, models.PutTransactionBody) (models.Transaction, error)
	// DeleteTransaction deletes and returns a transaction owned by the user.
	DeleteTransaction(context.Context, repositories.DBTX, uuid.UUID, uuid.UUID) (models.Transaction, error)
}

type BillRepository interface {
	// ListBills returns a user's bills ordered by due date.
	ListBills(context.Context, repositories.DBTX, uuid.UUID) ([]models.Bill, error)
	// InsertBill creates and returns a bill owned by the user.
	InsertBill(context.Context, repositories.DBTX, uuid.UUID, models.BillBody) (models.Bill, error)
	// UpdateBill updates and returns a bill owned by the user.
	UpdateBill(context.Context, repositories.DBTX, uuid.UUID, uuid.UUID, models.BillBody) (models.Bill, error)
	// DeleteBill deletes and returns a bill owned by the user.
	DeleteBill(context.Context, repositories.DBTX, uuid.UUID, uuid.UUID) (models.Bill, error)
}

type BudgetRepository interface {
	// GetBudgetPlan returns a user's budget plan for the given interval.
	GetBudgetPlan(context.Context, repositories.DBTX, uuid.UUID, models.IntervalType) (models.BudgetPlan, error)
	// InsertBudgetPlan creates and returns a budget plan for the user.
	InsertBudgetPlan(context.Context, repositories.DBTX, uuid.UUID, models.PostBudgetPlanBody) (models.BudgetPlan, error)
	// UpdateBudgetPlan updates and returns a user's budget plan for the given interval.
	UpdateBudgetPlan(context.Context, repositories.DBTX, uuid.UUID, models.IntervalType, models.PutBudgetPlanBody) (models.BudgetPlan, error)
	// DeleteBudgetPlan deletes and returns a user's budget plan for the given interval.
	DeleteBudgetPlan(context.Context, repositories.DBTX, uuid.UUID, models.IntervalType) (models.BudgetPlan, error)
}

type SummaryRepository interface {
	// GetBasicAnalysis returns aggregate financial totals for a user between two dates.
	GetBasicAnalysis(context.Context, repositories.DBTX, uuid.UUID, time.Time, time.Time) (models.BasicAnalysis, error)
	// GetDateToAmount returns total transaction amounts grouped by date for a user or account.
	GetDateToAmount(context.Context, repositories.DBTX, models.ObjectType, uuid.UUID, time.Time, time.Time) (map[string]models.Money, error)
	// GetCategoryToAmount returns total transaction amounts grouped by category for a user or account.
	GetCategoryToAmount(context.Context, repositories.DBTX, models.ObjectType, uuid.UUID, time.Time, time.Time) (map[models.TransactionCategory]models.Money, error)
}

type CronAccountRepository interface {
	// ListAccountsByStatus returns all accounts matching the requested cron status filter.
	ListAccountsByStatus(context.Context, repositories.DBTX, repositories.Status) ([]models.Account, error)
	// LockAccount locks and returns an account for the duration of the current database transaction.
	LockAccount(context.Context, repositories.DBTX, uuid.UUID) (models.Account, error)
	// MoveAccountsDueDate advances the due date of each supplied account by one month.
	MoveAccountsDueDate(context.Context, repositories.DBTX, []uuid.UUID) (int, error)
	// SetAccountDiscrepancy stores a signed discrepancy and flags the account when it is nonzero.
	// Passing zero clears both the discrepancy amount and flag.
	SetAccountDiscrepancy(context.Context, repositories.DBTX, uuid.UUID, models.Money) (models.Account, error)
}

type CronAccountHistoryRepository interface {
	// GetLatestHistory returns an account's most recent balance checkpoint.
	GetLatestHistory(context.Context, repositories.DBTX, uuid.UUID) (models.AccountHistory, error)
	// InsertHistory creates and returns an account balance checkpoint.
	InsertHistory(context.Context, repositories.DBTX, models.PostAccHistBody) (models.AccountHistory, error)
	// DeleteOutdatedHistories deletes histories older than the retention period and returns the count.
	DeleteOutdatedHistories(context.Context, repositories.DBTX, uuid.UUID) (int, error)
}

type CronTransactionRepository interface {
	// GetTransactionSum returns the signed transaction sum for an account after the given checkpoint.
	GetTransactionSum(context.Context, repositories.DBTX, uuid.UUID, time.Time) (models.Money, error)
	// DeleteOutdatedTransactions deletes transactions older than the retention period and returns the count.
	DeleteOutdatedTransactions(context.Context, repositories.DBTX, uuid.UUID) (int, error)
}
