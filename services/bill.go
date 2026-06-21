package services

import (
	"betterov2/models"
	"betterov2/repositories"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Bill Service
type BillService struct {
	db       *pgxpool.Pool
	billRepo *repositories.BillRepo
	accRepo  *repositories.AccountRepo
	tranRepo *repositories.TransactionRepo
}

// Generate a new bill service
func NewBillService(
	database *pgxpool.Pool, billRepo *repositories.BillRepo, accRepo *repositories.AccountRepo, tranRepo *repositories.TransactionRepo,
) *BillService {
	return &BillService{
		db:       database,
		billRepo: billRepo,
		accRepo:  accRepo,
		tranRepo: tranRepo,
	}
}

// ListBills gets the list of bills ordered by its due date
func (s *BillService) ListBills(ctx context.Context, userId int64) ([]models.Bill, error) {
	bills, err := s.billRepo.ListBills(ctx, s.db, userId)
	if err != nil {
		return bills, err
	}
	return bills, nil
}

// CreateBill inserts the bill into the database and returns the nested bill
func (s *BillService) CreateBill(ctx context.Context, body models.BillBody) (models.Bill, error) {
	var createdBill models.Bill

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return createdBill, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	createdBill, err = s.billRepo.InsertBill(ctx, tx, body)
	if err != nil {
		return createdBill, err
	}

	if err = tx.Commit(ctx); err != nil {
		return createdBill, err
	}

	return createdBill, nil
}

// UpdateBill modifies the bill info.
// Change of account's Id means that a different account will be responsible for
// paying the bill when it's due.
func (s *BillService) UpdateBill(ctx context.Context, id int64, body models.BillBody) (models.Bill, error) {
	var updatedBill models.Bill

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return updatedBill, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	updatedBill, err = s.billRepo.UpdateBill(ctx, tx, id, body)
	if err != nil {
		return updatedBill, err
	}

	if err = tx.Commit(ctx); err != nil {
		return updatedBill, err
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
func (s *BillService) DeleteBill(ctx context.Context, id int64, pay bool, recurring bool) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Delete the bill, get the non-nested deleted bill
	deleted, err := s.billRepo.DeleteBill(ctx, tx, id)
	if err != nil {
		return err
	}

	if pay {
		// Create the transaction representing bill payment
		paymentBody := models.PostTransactionBody{
			AccountID:       deleted.Account.Id,
			Merchant:        deleted.Merchant,
			TranDescription: fmt.Sprintf("Payment to %s", deleted.Description),
			Category:        deleted.Category,
			Amount:          deleted.Amount,
			CreatedAt:       time.Now(),
		}
		inserted, err := s.tranRepo.InsertTransaction(ctx, tx, paymentBody)
		if err != nil {
			return err
		}
		log.Printf("Transaction %s has been inserted\n", inserted.TranDescription)

		// Bill payment is considered expense, so auto-update the balance
		balance, err := s.accRepo.UpdateAccountBalance(ctx, tx, deleted.Account.Id, deleted.Amount)
		if err != nil {
			return err
		}
		log.Printf("Balance changes to: %f", balance)
	}
	if recurring {
		// Insert the recurring bill that is due next month
		recurringBody := models.BillBody{
			AccountID:   deleted.Account.Id,
			Merchant:    deleted.Merchant,
			Description: deleted.Description,
			Category:    deleted.Category,
			Amount:      deleted.Amount,
			DueDate:     deleted.DueDate.AddDate(0, 1, 0),
		}
		recurred, err := s.billRepo.InsertBill(ctx, tx, recurringBody)
		if err != nil {
			return err
		}
		log.Printf("Bill %s has been recurred\n", recurred.Description)
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}
