package services

import (
	"betterov2/models"
	"betterov2/repositories"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Bill Service
type BillService struct {
	db              *pgxpool.Pool
	billRepo        *repositories.BillRepo
	accountRepo     *repositories.AccountRepo
	transactionRepo *repositories.TransactionRepo
}

// Generate a new bill service
func NewBillService(
	database *pgxpool.Pool,
	billRepo *repositories.BillRepo,
	accountRepo *repositories.AccountRepo,
	transactionRepo *repositories.TransactionRepo,
) *BillService {
	return &BillService{
		db:              database,
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
func (s *BillService) CreateBill(ctx context.Context, body models.BillBody) (models.Bill, error) {
	newBill, err := s.billRepo.InsertBill(ctx, s.db, body)
	if err != nil {
		return models.Bill{}, err
	}

	return newBill, nil
}

// UpdateBill modifies the bill info.
// Change of account's Id means that a different account will be responsible for
// paying the bill when it's due.
func (s *BillService) UpdateBill(ctx context.Context, id uuid.UUID, body models.BillBody) (models.Bill, error) {
	var updatedBill models.Bill

	updatedBill, err := s.billRepo.UpdateBill(ctx, s.db, id, body)
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
func (s *BillService) DeleteBill(ctx context.Context, id uuid.UUID, pay bool, recurring bool) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Delete the bill, get the non-nested deleted bill
	deletedBill, err := s.billRepo.DeleteBill(ctx, tx, id)
	if err != nil {
		return err
	}
	if (pay || recurring) && deletedBill.Account == nil {
		return models.ErrForeignKey
	}

	if pay {
		// Create the transaction representing bill payment
		postTranBody := models.PostTransactionBody{
			AccountID:   deletedBill.Account.ID,
			Merchant:    deletedBill.Merchant,
			Description: fmt.Sprintf("Payment to %s", deletedBill.Description),
			Category:    deletedBill.Category,
			AmountCents: deletedBill.AmountCents,
			CreatedAt:   time.Now(),
		}
		insertedTransaction, err := s.transactionRepo.InsertTransaction(ctx, tx, postTranBody)
		if err != nil {
			return err
		}

		log.Printf("Transaction %s has been inserted\n", insertedTransaction.Description)

		// Bill payment is considered expense, so auto-update the balance
		balance, err := s.accountRepo.UpdateAccountBalance(ctx, tx, deletedBill.Account.ID, deletedBill.AmountCents)
		if err != nil {
			return err
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
		recurredBill, err := s.billRepo.InsertBill(ctx, tx, postBillBody)
		if err != nil {
			return err
		}
		log.Printf("Bill %s has been recurred\n", recurredBill.Description)
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}
