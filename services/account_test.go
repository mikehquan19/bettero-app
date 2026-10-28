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

func TestValidateAccount(t *testing.T) {
	creditLimit := models.Money(100000)
	dueDate := time.Now()

	tests := []struct {
		name        string
		accountType models.AccountType
		body        models.AccountBody
		expectedErr error
	}{
		{
			name:        "valid debit",
			accountType: models.Debit,
			body:        models.AccountBody{},
		},
		{
			name:        "debit contains credit fields",
			accountType: models.Debit,
			body: models.AccountBody{
				CreditLimitCents: &creditLimit,
			},
			expectedErr: models.ErrDebitCardWithCreditInfo,
		},
		{
			name:        "valid credit",
			accountType: models.Credit,
			body: models.AccountBody{
				CreditLimitCents: &creditLimit,
				NextDue:          &dueDate,
			},
		},
		{
			name:        "credit is missing credit fields",
			accountType: models.Credit,
			body:        models.AccountBody{},
			expectedErr: models.ErrCreditCardWithoutCreditInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAccount(tt.accountType, tt.body)
			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("validateAccount() error = %v, expected %v", err, tt.expectedErr)
			}
		})
	}
}

func TestAccountServiceCreateAccount(t *testing.T) {
	userID := uuid.New()
	createdAt := time.Now()
	expectedAccount := models.Account{
		ID:           uuid.New(),
		UserID:       userID,
		Type:         models.Debit,
		BalanceCents: 25000,
		CreatedAt:    createdAt,
	}

	tests := []struct {
		name        string
		insertErr   error
		historyErr  error
		expectedErr error
	}{
		{
			name: "creates account and initial history",
		},
		{
			name:        "returns account insertion error",
			insertErr:   models.ErrForeignKey,
			expectedErr: models.ErrForeignKey,
		},
		{
			name:        "returns history insertion error",
			historyErr:  errors.New("insert history"),
			expectedErr: errors.New("insert history"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var historyBody models.PostAccHistBody
			accountRepo := &fakeAccountRepository{
				insertAccount: func(context.Context, repositories.DBTX, uuid.UUID, models.PostAccountBody) (models.Account, error) {
					if tt.insertErr != nil {
						return models.Account{}, tt.insertErr
					}
					return expectedAccount, nil
				},
			}
			historyRepo := &fakeAccountHistoryRepository{
				insertHistory: func(_ context.Context, _ repositories.DBTX, body models.PostAccHistBody) (models.AccountHistory, error) {
					historyBody = body
					if tt.historyErr != nil {
						return models.AccountHistory{}, tt.historyErr
					}
					return models.AccountHistory{
						AccountID:    body.AccountID,
						LoggedTime:   body.LoggedTime,
						BalanceCents: body.BalanceCents,
					}, nil
				},
			}
			service := NewAccountService(nil, &fakeUnitOfWork{}, accountRepo, historyRepo, nil)

			account, err := service.CreateAccount(context.Background(), userID, models.PostAccountBody{
				Type: models.Debit,
			})

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Fatalf("CreateAccount() error = %v, expected %v", err, tt.expectedErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("CreateAccount() error = %v", err)
			}
			if !reflect.DeepEqual(account, expectedAccount) {
				t.Errorf("account = %#v, expected %#v", account, expectedAccount)
			}
			expectedHistoryBody := models.PostAccHistBody{
				AccountID:    expectedAccount.ID,
				LoggedTime:   expectedAccount.CreatedAt,
				BalanceCents: expectedAccount.BalanceCents,
			}
			if !reflect.DeepEqual(historyBody, expectedHistoryBody) {
				t.Errorf("history body = %#v, expected %#v", historyBody, expectedHistoryBody)
			}
		})
	}
}

func TestAccountServiceUpdateAccountReconciliation(t *testing.T) {
	userID := uuid.New()
	accountID := uuid.New()

	tests := []struct {
		name                string
		accountType         models.AccountType
		previousBalance     models.Money
		updatedBalance      models.Money
		expectedTransaction bool
		expectedAmount      models.Money
		expectedCategory    models.TransactionCategory
		expectedDescription string
	}{
		{
			name:                "debit increase is income",
			accountType:         models.Debit,
			previousBalance:     10000,
			updatedBalance:      12500,
			expectedTransaction: true,
			expectedAmount:      2500,
			expectedCategory:    models.Income,
			expectedDescription: "Balance reconciled (+25.00)",
		},
		{
			name:                "debit decrease is expense",
			accountType:         models.Debit,
			previousBalance:     12500,
			updatedBalance:      10000,
			expectedTransaction: true,
			expectedAmount:      2500,
			expectedCategory:    models.Others,
			expectedDescription: "Balance reconciled (-25.00)",
		},
		{
			name:                "credit increase is expense",
			accountType:         models.Credit,
			previousBalance:     10000,
			updatedBalance:      12500,
			expectedTransaction: true,
			expectedAmount:      2500,
			expectedCategory:    models.Others,
			expectedDescription: "Balance reconciled (+25.00)",
		},
		{
			name:                "credit decrease is income",
			accountType:         models.Credit,
			previousBalance:     12500,
			updatedBalance:      10000,
			expectedTransaction: true,
			expectedAmount:      2500,
			expectedCategory:    models.Income,
			expectedDescription: "Balance reconciled (-25.00)",
		},
		{
			name:            "unchanged balance creates no transaction",
			accountType:     models.Debit,
			previousBalance: 10000,
			updatedBalance:  10000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creditLimit := models.Money(100000)
			dueDate := time.Now()
			body := models.PutAccountBody{
				AccountBody: models.AccountBody{
					BalanceCents: tt.updatedBalance,
				},
			}
			if tt.accountType == models.Credit {
				body.CreditLimitCents = &creditLimit
				body.NextDue = &dueDate
			}

			previous := models.Account{
				ID:           accountID,
				UserID:       userID,
				Type:         tt.accountType,
				BalanceCents: tt.previousBalance,
			}
			expectedAccount := models.Account{
				ID:           accountID,
				UserID:       userID,
				Type:         tt.accountType,
				Institution:  "Test Bank",
				BalanceCents: tt.updatedBalance,
			}
			accountRepo := &fakeAccountRepository{
				lockAccount: func(context.Context, repositories.DBTX, uuid.UUID) (models.Account, error) {
					return previous, nil
				},
				updateAccount: func(context.Context, repositories.DBTX, uuid.UUID, uuid.UUID, models.PutAccountBody) (models.Account, error) {
					return expectedAccount, nil
				},
			}
			transactionCalled := false
			transactionRepo := &fakeTransactionRepository{
				insertTransaction: func(_ context.Context, _ repositories.DBTX, _ uuid.UUID, body models.PostTransactionBody) (models.Transaction, error) {
					transactionCalled = true
					if body.AmountCents != tt.expectedAmount || body.Category != tt.expectedCategory || body.Description != tt.expectedDescription {
						t.Errorf("transaction body = %#v", body)
					}
					return models.Transaction{
						ID: uuid.New(),
					}, nil
				},
			}
			service := NewAccountService(nil, &fakeUnitOfWork{}, accountRepo, nil, transactionRepo)

			account, err := service.UpdateAccount(context.Background(), userID, accountID, body)

			if err != nil {
				t.Fatalf("UpdateAccount() error = %v", err)
			}
			if !reflect.DeepEqual(account, expectedAccount) {
				t.Errorf("account = %#v, expected %#v", account, expectedAccount)
			}
			if transactionCalled != tt.expectedTransaction {
				t.Errorf("transaction called = %v, expected %v", transactionCalled, tt.expectedTransaction)
			}
		})
	}
}
