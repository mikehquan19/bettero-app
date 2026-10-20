package services

import (
	"betterov2/models"
	"betterov2/repositories"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

// Bill Service
type BillService struct {
	db              repositories.DBTX
	unitOfWork      repositories.UnitOfWork
	billRepo        BillRepository
	accountRepo     AccountRepository
	transactionRepo TransactionRepository
}

// Generate a new bill service
func NewBillService(
	database repositories.DBTX,
	unitOfWork repositories.UnitOfWork,
	billRepo BillRepository,
	accountRepo AccountRepository,
	transactionRepo TransactionRepository,
) *BillService {
	return &BillService{
		db:              database,
		unitOfWork:      unitOfWork,
		billRepo:        billRepo,
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
	}
}

// ListBills gets the list of bills ordered by its due date
func (s *BillService) ListBills(ctx context.Context, userId uuid.UUID) ([]models.Bill, error) {
	bills, err := s.billRepo.ListBills(ctx, s.db, userId)
	if err != nil {
		return bills, err
	}
	return bills, nil
}

// CreateBill inserts the bill into the database and returns the nested bill
func (s *BillService) CreateBill(ctx context.Context, userId uuid.UUID, body models.BillBody) (models.Bill, error) {
	newBill, err := s.billRepo.InsertBill(ctx, s.db, userId, body)
	if err != nil {
		return models.Bill{}, err
	}

	return newBill, nil
}

// UpdateBill modifies the bill info.
// Change of account's Id means that a different account will be responsible for
// paying the bill when it's due.
func (s *BillService) UpdateBill(ctx context.Context, userId, id uuid.UUID, body models.BillBody) (models.Bill, error) {
	var updatedBill models.Bill

	updatedBill, err := s.billRepo.UpdateBill(ctx, s.db, userId, id, body)
	if err != nil {
		return models.Bill{}, err
	}

	return updatedBill, nil
}

// DeleteBill deletes the bill.
//
// - If the user pays bill, creates a transaction as payment to the bill.
// - If the user only deletes the bill, no new transaction is created.
//
// The created transaction has the same account's Id, category, and amount as the bill.
// The paying account's balance also updates to reflect the transaction.
func (s *BillService) DeleteBill(ctx context.Context, userId, id uuid.UUID, pay bool, recurring bool) error {
	return s.unitOfWork.Do(ctx, func(tx repositories.DBTX) error {
		// Delete the bill, get the non-nested deleted bill
		deletedBill, uowErr := s.billRepo.DeleteBill(ctx, tx, userId, id)
		if uowErr != nil {
			return uowErr
		}
		if (pay || recurring) && deletedBill.Account == nil {
			return models.ErrForeignKey
		}

		if pay {
			if _, uowErr = s.accountRepo.LockAccount(ctx, tx, deletedBill.Account.ID); uowErr != nil {
				return uowErr
			}

			// Create the transaction representing bill payment
			postTranBody := models.PostTransactionBody{
				AccountID:   deletedBill.Account.ID,
				Merchant:    deletedBill.Merchant,
				Description: fmt.Sprintf("Payment to %s", deletedBill.Description),
				Category:    deletedBill.Category,
				AmountCents: deletedBill.AmountCents,
				CreatedAt:   time.Now(),
			}
			insertedTransaction, uowErr := s.transactionRepo.InsertTransaction(ctx, tx, userId, postTranBody)
			if uowErr != nil {
				return uowErr
			}
			log.Printf("Transaction %s has been inserted\n", insertedTransaction.Description)

			// Bill payment is considered expense, so auto-update the balance
			balance, uowErr := s.accountRepo.UpdateAccountBalance(ctx, tx, deletedBill.Account.ID, deletedBill.AmountCents)
			if uowErr != nil {
				return uowErr
			}
			log.Printf("Balance changes to: %d cents\n", balance)
		}

		if recurring {
			// Insert the recurring bill that is due next month
			postBillBody := models.BillBody{
				AccountID:   deletedBill.Account.ID,
				Merchant:    deletedBill.Merchant,
				Description: deletedBill.Description,
				Category:    deletedBill.Category,
				AmountCents: deletedBill.AmountCents,
				DueDate:     deletedBill.DueDate.AddDate(0, 1, 0),
			}
			recurredBill, uowErr := s.billRepo.InsertBill(ctx, tx, userId, postBillBody)
			if uowErr != nil {
				return uowErr
			}
			log.Printf("Bill %s has been recurred\n", recurredBill.Description)
		}

		return nil
	})
}
