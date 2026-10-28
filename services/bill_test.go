package services

import (
	"betterov2/models"
	"betterov2/repositories"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestBillServiceDeleteBill(t *testing.T) {
	userID := uuid.New()
	billID := uuid.New()
	accountID := uuid.New()
	dueDate := time.Now()

	tests := []struct {
		name                string
		pay                 bool
		recurring           bool
		missingAccount      bool
		expectedTransaction bool
		expectedRecurrence  bool
		expectedErr         error
	}{
		{
			name: "delete only",
		},
		{
			name:                "pay bill",
			pay:                 true,
			expectedTransaction: true,
		},
		{
			name:               "recur bill",
			recurring:          true,
			expectedRecurrence: true,
		},
		{
			name:                "pay and recur bill",
			pay:                 true,
			recurring:           true,
			expectedTransaction: true,
			expectedRecurrence:  true,
		},
		{
			name:           "pay requires account",
			pay:            true,
			missingAccount: true,
			expectedErr:    models.ErrForeignKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var account *models.NestedAccount
			if !tt.missingAccount {
				account = &models.NestedAccount{
					ID: accountID,
				}
			}
			deletedBill := models.Bill{
				ID:          billID,
				Account:     account,
				Merchant:    "Power Co",
				Description: "Power bill",
				Category:    models.Housing,
				AmountCents: 8500,
				DueDate:     dueDate,
			}
			billRepo := &fakeBillRepository{
				deleteBill: func(context.Context, repositories.DBTX, uuid.UUID, uuid.UUID) (models.Bill, error) {
					return deletedBill, nil
				},
			}
			recurrenceCalled := false
			billRepo.insertBill = func(_ context.Context, _ repositories.DBTX, _ uuid.UUID, body models.BillBody) (models.Bill, error) {
				recurrenceCalled = true
				expectedDueDate := dueDate.AddDate(0, 1, 0)
				if !body.DueDate.Equal(expectedDueDate) {
					t.Errorf("recurrence due date = %s, expected %s", body.DueDate, expectedDueDate)
				}
				return models.Bill{
					ID:          uuid.New(),
					Description: body.Description,
				}, nil
			}

			transactionCalled := false
			transactionRepo := &fakeTransactionRepository{
				insertTransaction: func(_ context.Context, _ repositories.DBTX, _ uuid.UUID, body models.PostTransactionBody) (models.Transaction, error) {
					transactionCalled = true
					if body.AccountID != accountID || body.AmountCents != deletedBill.AmountCents || body.Description != "Payment to Power bill" {
						t.Errorf("payment transaction = %#v", body)
					}
					return models.Transaction{
						ID:          uuid.New(),
						Description: body.Description,
					}, nil
				},
			}
			accountRepo := &fakeAccountRepository{
				lockAccount: func(context.Context, repositories.DBTX, uuid.UUID) (models.Account, error) {
					return models.Account{
						ID: accountID,
					}, nil
				},
				updateAccountBalance: func(_ context.Context, _ repositories.DBTX, _ uuid.UUID, amount models.Money) (models.Money, error) {
					if amount != deletedBill.AmountCents {
						t.Errorf("balance amount = %d, expected %d", amount, deletedBill.AmountCents)
					}
					return 10000, nil
				},
			}
			service := NewBillService(nil, &fakeUnitOfWork{}, billRepo, accountRepo, transactionRepo)

			err := service.DeleteBill(context.Background(), userID, billID, tt.pay, tt.recurring)

			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("DeleteBill() error = %v, expected %v", err, tt.expectedErr)
			}
			if transactionCalled != tt.expectedTransaction {
				t.Errorf("transaction called = %v, expected %v", transactionCalled, tt.expectedTransaction)
			}
			if recurrenceCalled != tt.expectedRecurrence {
				t.Errorf("recurrence called = %v, expected %v", recurrenceCalled, tt.expectedRecurrence)
			}
		})
	}
}
