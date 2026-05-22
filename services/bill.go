package services

import (
	"betterov2/models"
	"betterov2/repositories"
	"context"
	"fmt"
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
	var newBill models.Bill

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return newBill, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			panic(err)
		}
	}()

	// Insert the bill into the database
	inserted, err := s.billRepo.InsertBill(ctx, tx, body)
	if err != nil {
		return newBill, err
	}

	// Get the created bill with nested account
	newBill, err = s.billRepo.GetBill(ctx, tx, int64(inserted.ID))
	if err != nil {
		return newBill, err
	}

	if err = tx.Commit(ctx); err != nil {
		return newBill, err
	}

	return newBill, nil
}

// UpdateBill modifies the bill info.
// Change of account's Id means that a different account will be responsible for paying the bill
// when it's due.
func (s *BillService) UpdateBill(
	ctx context.Context, id int64, body models.BillBody,
) (models.Bill, error) {
	var updatedBill models.Bill

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return updatedBill, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			panic(err)
		}
	}()

	// Update the bill
	_, err = s.billRepo.UpdateBill(ctx, tx, id, body)
	if err != nil {
		return updatedBill, err
	}

	// Fetch the updated bill with nested account
	updatedBill, err = s.billRepo.GetBill(ctx, tx, id)
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
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			panic(err)
		}
	}()

	// Store the bill to be deleted, info will be used later
	deletedBill, err := s.billRepo.GetBill(ctx, tx, id)
	if err != nil {
		return err
	}

	// Delete the bill
	_, err = s.billRepo.DeleteBill(ctx, tx, id)
	if err != nil {
		return err
	}

	if pay {
		// Create the transaction in the database representing bill payment
		paymentBody := models.PostTransactionBody{
			AccountID:       int(deletedBill.Account.Id),
			Merchant:        deletedBill.Merchant,
			TranDescription: fmt.Sprintf("Payment to %s", deletedBill.Description),
			Category:        deletedBill.Category,
			Amount:          deletedBill.Amount,
			CreatedAt:       time.Now(),
		}
		_, err := s.tranRepo.InsertTransaction(ctx, tx, paymentBody)
		if err != nil {
			return err
		}
		// Bill payment is considered expense, so auto-update the balance
		err = s.accRepo.UpdateAccountBalance(
			ctx, tx, deletedBill.Account.Id, deletedBill.Amount)
		if err != nil {
			return err
		}
	}

	if recurring {
		// Insert the recurring bill that is due next month
		recurringBody := models.BillBody{
			AccountID:   int(deletedBill.Account.Id),
			Merchant:    deletedBill.Merchant,
			Description: deletedBill.Description,
			Category:    deletedBill.Category,
			Amount:      deletedBill.Amount,
			DueDate:     deletedBill.DueDate.AddDate(0, 1, 0),
		}
		_, err := s.billRepo.InsertBill(ctx, tx, recurringBody)
		if err != nil {
			return err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}
