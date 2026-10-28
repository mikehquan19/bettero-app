package services

import (
	"betterov2/models"
	"betterov2/repositories"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

type fakeCronAccountRepository struct {
	listAccounts          func(context.Context, repositories.DBTX, repositories.Status) ([]models.Account, error)
	lockAccount           func(context.Context, repositories.DBTX, uuid.UUID) (models.Account, error)
	moveDueDates          func(context.Context, repositories.DBTX, []uuid.UUID) (int, error)
	setAccountDiscrepancy func(context.Context, repositories.DBTX, uuid.UUID, models.Money) (models.Account, error)
}

func (f *fakeCronAccountRepository) ListAccountsByStatus(ctx context.Context, db repositories.DBTX, status repositories.Status) ([]models.Account, error) {
	return f.listAccounts(ctx, db, status)
}
func (f *fakeCronAccountRepository) LockAccount(ctx context.Context, db repositories.DBTX, accountID uuid.UUID) (models.Account, error) {
	return f.lockAccount(ctx, db, accountID)
}
func (f *fakeCronAccountRepository) MoveAccountsDueDate(ctx context.Context, db repositories.DBTX, accountIDs []uuid.UUID) (int, error) {
	return f.moveDueDates(ctx, db, accountIDs)
}
func (f *fakeCronAccountRepository) SetAccountDiscrepancy(ctx context.Context, db repositories.DBTX, accountID uuid.UUID, discrepancy models.Money) (models.Account, error) {
	return f.setAccountDiscrepancy(ctx, db, accountID, discrepancy)
}

type fakeCronAccountHistoryRepository struct {
	getLatestHistory        func(context.Context, repositories.DBTX, uuid.UUID) (models.AccountHistory, error)
	insertHistory           func(context.Context, repositories.DBTX, models.PostAccHistBody) (models.AccountHistory, error)
	deleteOutdatedHistories func(context.Context, repositories.DBTX, uuid.UUID) (int, error)
}

func (f *fakeCronAccountHistoryRepository) GetLatestHistory(ctx context.Context, db repositories.DBTX, accountID uuid.UUID) (models.AccountHistory, error) {
	return f.getLatestHistory(ctx, db, accountID)
}
func (f *fakeCronAccountHistoryRepository) InsertHistory(ctx context.Context, db repositories.DBTX, body models.PostAccHistBody) (models.AccountHistory, error) {
	return f.insertHistory(ctx, db, body)
}
func (f *fakeCronAccountHistoryRepository) DeleteOutdatedHistories(ctx context.Context, db repositories.DBTX, accountID uuid.UUID) (int, error) {
	return f.deleteOutdatedHistories(ctx, db, accountID)
}

type fakeCronTransactionRepository struct {
	getTransactionSum          func(context.Context, repositories.DBTX, uuid.UUID, time.Time) (models.Money, error)
	deleteOutdatedTransactions func(context.Context, repositories.DBTX, uuid.UUID) (int, error)
}

func (f *fakeCronTransactionRepository) GetTransactionSum(ctx context.Context, db repositories.DBTX, accountID uuid.UUID, start time.Time) (models.Money, error) {
	return f.getTransactionSum(ctx, db, accountID, start)
}
func (f *fakeCronTransactionRepository) DeleteOutdatedTransactions(ctx context.Context, db repositories.DBTX, accountID uuid.UUID) (int, error) {
	return f.deleteOutdatedTransactions(ctx, db, accountID)
}

func TestCalculateDiscrepancy(t *testing.T) {
	tests := []struct {
		name           string
		accountType    models.AccountType
		balance        models.Money
		historyBalance models.Money
		transactionSum models.Money
		expected       models.Money
	}{
		{
			name:           "consistent debit",
			accountType:    models.Debit,
			balance:        8000,
			historyBalance: 10000,
			transactionSum: 2000,
			expected:       0,
		},
		{
			name:           "inconsistent debit",
			accountType:    models.Debit,
			balance:        8100,
			historyBalance: 10000,
			transactionSum: 2000,
			expected:       100,
		},
		{
			name:           "consistent credit",
			accountType:    models.Credit,
			balance:        12000,
			historyBalance: 10000,
			transactionSum: 2000,
			expected:       0,
		},
		{
			name:           "inconsistent credit",
			accountType:    models.Credit,
			balance:        11900,
			historyBalance: 10000,
			transactionSum: 2000,
			expected:       -100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := calculateDiscrepancy(
				models.Account{
					Type:         tt.accountType,
					BalanceCents: tt.balance,
				},
				models.AccountHistory{
					BalanceCents: tt.historyBalance,
				},
				tt.transactionSum,
			)
			if actual != tt.expected {
				t.Errorf("calculateDiscrepancy() = %d, expected %d", actual, tt.expected)
			}
		})
	}
}

func TestShouldCreateHistory(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name       string
		loggedTime time.Time
		expected   bool
	}{
		{
			name:       "thirteen days old",
			loggedTime: now.AddDate(0, 0, -13),
			expected:   false,
		},
		{
			name:       "exactly fourteen days old",
			loggedTime: now.AddDate(0, 0, -14),
			expected:   true,
		},
		{
			name:       "older than fourteen days",
			loggedTime: now.AddDate(0, 0, -30),
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := shouldCreateHistory(models.AccountHistory{
				LoggedTime: tt.loggedTime,
			}, now)
			if actual != tt.expected {
				t.Errorf("shouldCreateHistory() = %v, expected %v", actual, tt.expected)
			}
		})
	}
}

func TestCronServiceValidateAndUpdateHistory(t *testing.T) {
	accountID := uuid.New()

	tests := []struct {
		name                   string
		balance                models.Money
		transactionSum         models.Money
		historyAgeDays         int
		expectedDiscrepancy    models.Money
		expectedHistoryCreated bool
		expectedCleanup        bool
	}{
		{
			name:                   "valid account creates eligible history",
			balance:                8000,
			transactionSum:         2000,
			historyAgeDays:         15,
			expectedHistoryCreated: true,
			expectedCleanup:        true,
		},
		{
			name:                "discrepancy flags and skips history",
			balance:             8100,
			transactionSum:      2000,
			historyAgeDays:      15,
			expectedDiscrepancy: 100,
		},
		{
			name:            "recent valid history is not recreated",
			balance:         8000,
			transactionSum:  2000,
			historyAgeDays:  5,
			expectedCleanup: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			latestHistory := models.AccountHistory{
				ID:           uuid.New(),
				AccountID:    accountID,
				BalanceCents: 10000,
				LoggedTime:   time.Now().AddDate(0, 0, -tt.historyAgeDays),
			}
			accountRepo := &fakeCronAccountRepository{
				lockAccount: func(context.Context, repositories.DBTX, uuid.UUID) (models.Account, error) {
					return models.Account{
						ID:           accountID,
						Type:         models.Debit,
						BalanceCents: tt.balance,
					}, nil
				},
				setAccountDiscrepancy: func(_ context.Context, _ repositories.DBTX, _ uuid.UUID, discrepancy models.Money) (models.Account, error) {
					if discrepancy != tt.expectedDiscrepancy {
						t.Errorf("discrepancy = %d, expected %d", discrepancy, tt.expectedDiscrepancy)
					}
					return models.Account{
						ID: accountID,
					}, nil
				},
			}
			historyCreated := false
			cleanupCalled := false
			historyRepo := &fakeCronAccountHistoryRepository{
				getLatestHistory: func(context.Context, repositories.DBTX, uuid.UUID) (models.AccountHistory, error) {
					return latestHistory, nil
				},
				insertHistory: func(context.Context, repositories.DBTX, models.PostAccHistBody) (models.AccountHistory, error) {
					historyCreated = true
					return models.AccountHistory{
						ID: uuid.New(),
					}, nil
				},
				deleteOutdatedHistories: func(context.Context, repositories.DBTX, uuid.UUID) (int, error) {
					cleanupCalled = true
					return 0, nil
				},
			}
			transactionRepo := &fakeCronTransactionRepository{
				getTransactionSum: func(context.Context, repositories.DBTX, uuid.UUID, time.Time) (models.Money, error) {
					return tt.transactionSum, nil
				},
			}
			service := NewCronService(nil, &fakeUnitOfWork{}, accountRepo, historyRepo, transactionRepo)

			err := service.validateAndUpdateHistory(context.Background(), accountID)

			if err != nil {
				t.Fatalf("validateAndUpdateHistory() error = %v", err)
			}
			if historyCreated != tt.expectedHistoryCreated {
				t.Errorf("history created = %v, expected %v", historyCreated, tt.expectedHistoryCreated)
			}
			if cleanupCalled != tt.expectedCleanup {
				t.Errorf("cleanup called = %v, expected %v", cleanupCalled, tt.expectedCleanup)
			}
		})
	}
}

var _ CronAccountRepository = (*fakeCronAccountRepository)(nil)
var _ CronAccountHistoryRepository = (*fakeCronAccountHistoryRepository)(nil)
var _ CronTransactionRepository = (*fakeCronTransactionRepository)(nil)
