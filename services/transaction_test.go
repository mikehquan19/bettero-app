package services

import (
	"betterov2/models"
	"betterov2/repositories"
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestTransactionServiceCreateTransactionBalanceEffect(t *testing.T) {
	userID := uuid.New()
	accountID := uuid.New()

	tests := []struct {
		name              string
		category          models.TransactionCategory
		expectedNetChange models.Money
	}{
		{
			name:              "expense applies positive net change",
			category:          models.Grocery,
			expectedNetChange: 2500,
		},
		{
			name:              "income applies negative net change",
			category:          models.Income,
			expectedNetChange: -2500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := models.PostTransactionBody{
				AccountID:   accountID,
				Category:    tt.category,
				AmountCents: 2500,
				CreatedAt:   time.Now(),
			}
			expectedTransaction := models.Transaction{
				ID: uuid.New(),
				Account: models.NestedAccount{
					ID: accountID,
				},
				Category:    tt.category,
				AmountCents: 2500,
				CreatedAt:   body.CreatedAt,
			}
			accountRepo := &fakeAccountRepository{
				lockAccount: func(context.Context, repositories.DBTX, uuid.UUID) (models.Account, error) {
					return models.Account{
						ID: accountID,
					}, nil
				},
				updateAccountBalance: func(_ context.Context, _ repositories.DBTX, gotAccountID uuid.UUID, amount models.Money) (models.Money, error) {
					if gotAccountID != accountID || amount != tt.expectedNetChange {
						t.Errorf("balance update = (%s, %d), expected (%s, %d)", gotAccountID, amount, accountID, tt.expectedNetChange)
					}
					return 10000, nil
				},
			}
			transactionRepo := &fakeTransactionRepository{
				insertTransaction: func(context.Context, repositories.DBTX, uuid.UUID, models.PostTransactionBody) (models.Transaction, error) {
					return expectedTransaction, nil
				},
			}
			service := NewTransactionService(nil, &fakeUnitOfWork{}, transactionRepo, accountRepo)

			transaction, err := service.CreateTransaction(context.Background(), userID, body)

			if err != nil {
				t.Fatalf("CreateTransaction() error = %v", err)
			}
			if !reflect.DeepEqual(transaction, expectedTransaction) {
				t.Errorf("transaction = %#v, expected %#v", transaction, expectedTransaction)
			}
		})
	}
}

func TestTransactionServiceUpdateTransactionBalanceEffect(t *testing.T) {
	userID := uuid.New()
	transactionID := uuid.New()
	accountID := uuid.New()

	tests := []struct {
		name                  string
		previousCategory      models.TransactionCategory
		previousAmount        models.Money
		updatedCategory       models.TransactionCategory
		updatedAmount         models.Money
		expectedBalanceUpdate bool
		expectedNetChange     models.Money
	}{
		{
			name:                  "expense amount increases",
			previousCategory:      models.Grocery,
			previousAmount:        1000,
			updatedCategory:       models.Grocery,
			updatedAmount:         1500,
			expectedBalanceUpdate: true,
			expectedNetChange:     500,
		},
		{
			name:                  "expense changes to income",
			previousCategory:      models.Grocery,
			previousAmount:        1000,
			updatedCategory:       models.Income,
			updatedAmount:         1500,
			expectedBalanceUpdate: true,
			expectedNetChange:     -2500,
		},
		{
			name:             "financial effect is unchanged",
			previousCategory: models.Grocery,
			previousAmount:   1000,
			updatedCategory:  models.Grocery,
			updatedAmount:    1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			previous := models.Transaction{
				ID: transactionID,
				Account: models.NestedAccount{
					ID: accountID,
				},
				Category:    tt.previousCategory,
				AmountCents: tt.previousAmount,
				CreatedAt:   time.Now(),
			}
			expectedTransaction := models.Transaction{
				ID: transactionID,
				Account: models.NestedAccount{
					ID: accountID,
				},
				Category:    tt.updatedCategory,
				AmountCents: tt.updatedAmount,
				CreatedAt:   previous.CreatedAt,
			}
			transactionRepo := &fakeTransactionRepository{
				getTransaction: func(context.Context, repositories.DBTX, uuid.UUID, uuid.UUID) (models.Transaction, error) {
					return previous, nil
				},
				updateTransaction: func(context.Context, repositories.DBTX, uuid.UUID, uuid.UUID, models.PutTransactionBody) (models.Transaction, error) {
					return expectedTransaction, nil
				},
			}
			balanceUpdated := false
			accountRepo := &fakeAccountRepository{
				lockAccount: func(context.Context, repositories.DBTX, uuid.UUID) (models.Account, error) {
					return models.Account{
						ID: accountID,
					}, nil
				},
				updateAccountBalance: func(_ context.Context, _ repositories.DBTX, _ uuid.UUID, amount models.Money) (models.Money, error) {
					balanceUpdated = true
					if amount != tt.expectedNetChange {
						t.Errorf("net change = %d, expected %d", amount, tt.expectedNetChange)
					}
					return 10000, nil
				},
			}
			service := NewTransactionService(nil, &fakeUnitOfWork{}, transactionRepo, accountRepo)

			transaction, err := service.UpdateTransaction(context.Background(), userID, transactionID, models.PutTransactionBody{})

			if err != nil {
				t.Fatalf("UpdateTransaction() error = %v", err)
			}
			if !reflect.DeepEqual(transaction, expectedTransaction) {
				t.Errorf("transaction = %#v, expected %#v", transaction, expectedTransaction)
			}
			if balanceUpdated != tt.expectedBalanceUpdate {
				t.Errorf("balance updated = %v, expected %v", balanceUpdated, tt.expectedBalanceUpdate)
			}
		})
	}
}

func TestTransactionServiceDeleteTransactionBalanceEffect(t *testing.T) {
	tests := []struct {
		name              string
		category          models.TransactionCategory
		expectedNetChange models.Money
	}{
		{
			name:              "deleting expense reverses expense",
			category:          models.Grocery,
			expectedNetChange: -2500,
		},
		{
			name:              "deleting income reverses income",
			category:          models.Income,
			expectedNetChange: 2500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountID := uuid.New()
			transaction := models.Transaction{
				ID: uuid.New(),
				Account: models.NestedAccount{
					ID: accountID,
				},
				Category:    tt.category,
				AmountCents: 2500,
				CreatedAt:   time.Now(),
			}
			transactionRepo := &fakeTransactionRepository{
				getTransaction: func(context.Context, repositories.DBTX, uuid.UUID, uuid.UUID) (models.Transaction, error) {
					return transaction, nil
				},
				deleteTransaction: func(context.Context, repositories.DBTX, uuid.UUID, uuid.UUID) (models.Transaction, error) {
					return transaction, nil
				},
			}
			accountRepo := &fakeAccountRepository{
				lockAccount: func(context.Context, repositories.DBTX, uuid.UUID) (models.Account, error) {
					return models.Account{
						ID: accountID,
					}, nil
				},
				updateAccountBalance: func(_ context.Context, _ repositories.DBTX, _ uuid.UUID, amount models.Money) (models.Money, error) {
					if amount != tt.expectedNetChange {
						t.Errorf("net change = %d, expected %d", amount, tt.expectedNetChange)
					}
					return 10000, nil
				},
			}
			service := NewTransactionService(nil, &fakeUnitOfWork{}, transactionRepo, accountRepo)

			err := service.DeleteTransaction(context.Background(), uuid.New(), transaction.ID)
			if err != nil {
				t.Fatalf("DeleteTransaction() error = %v", err)
			}
		})
	}
}

func TestValidateTransactionTime(t *testing.T) {
	tests := []struct {
		name        string
		createdAt   time.Time
		expectedErr error
	}{
		{
			name:      "recent transaction",
			createdAt: time.Now(),
		},
		{
			name:        "transaction older than allowed window",
			createdAt:   time.Now().AddDate(0, 0, -14),
			expectedErr: models.ErrTransactionTooOld,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTransactionTime(tt.createdAt)
			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("validateTransactionTime() error = %v, expected %v", err, tt.expectedErr)
			}
		})
	}
}
