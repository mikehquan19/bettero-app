package services

import (
	"betterov2/models"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Bill Service
type BillService struct {
	database *pgxpool.Pool
}

// Generate a new bill service
func NewBillService(database *pgxpool.Pool) *BillService {
	return &BillService{
		database: database,
	}
}

// ListBills gets the list of bills orded by its due date
func (s *BillService) ListBills(ctx context.Context, userId int64) ([]models.Bill, error) {
	var bills []models.Bill

	listBillQuery := `
	SELECT
		b.id, b.merchant, b.description, b.category, b.amount, b.due_date,
		json_build_object(
			'id', a.id,
			'acc_number', a.acc_number,
			'acc_name', a.acc_name,
			'institution', a.institution,
			'type', a.type
		) AS account
	FROM bills b 
	JOIN accounts a ON b.account_id = a.id
	WHERE a.user_id = $1
	ORDER BY b.due_date ASC;
	`
	billRows, err := s.database.Query(ctx, listBillQuery, userId)
	if err != nil {
		return bills, err
	}

	bills, err = pgx.CollectRows(billRows, pgx.RowToStructByName[models.Bill])
	if err != nil {
		return bills, err
	}

	return bills, nil
}

// CreateBill inserts the bill into the database
func (s *BillService) CreateBill(ctx context.Context, body models.BillBody) (models.Bill, error) {
	var newBill models.Bill

	pgTran, err := s.database.Begin(ctx)
	if err != nil {
		return newBill, err
	}
	defer func() {
		if err := pgTran.Rollback(ctx); err != nil {
			panic(err)
		}
	}()

	// Insert the bill into the database and obtains the bill's Id
	var insertedBillId int64
	createBillQuery := `
	INSERT INTO bills (
		account_id, 
		merchant, 
		description, 
		category, 
		amount, 
		due_date
	)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id;
	`
	insertedBillRow := pgTran.QueryRow(ctx, createBillQuery,
		body.AccountID,
		body.Merchant,
		body.Description,
		body.Category,
		body.Amount,
		body.DueDate,
	)
	if err := insertedBillRow.Scan(&insertedBillId); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23503" {
			// Inserts a bill that references non-existent account
			accId := int64(body.AccountID)
			return newBill, models.GetForeignKey[models.Account](accId)
		}
		return newBill, err
	}

	// Fetch the created bill with nested account for returning
	newBill, err = getBill(ctx, pgTran, insertedBillId)
	if err != nil {
		return newBill, err
	}

	if err = pgTran.Commit(ctx); err != nil {
		return newBill, err
	}

	return newBill, nil
}

// UpdateBill modifies the bill info.
// Change of account's Id mean that a different account will be responsible for paying for bill
// when it's due and is deleted from the database.
func (s *BillService) UpdateBill(
	ctx context.Context, id int64, body models.BillBody,
) (models.Bill, error) {
	var updatedBill models.Bill

	pgTran, err := s.database.Begin(ctx)
	if err != nil {
		return updatedBill, err
	}
	defer func() {
		if err := pgTran.Rollback(ctx); err != nil {
			panic(err)
		}
	}()

	// Update the bill
	updateBillQuery := `
	UPDATE bills
	SET account_id = $1, 
		merchant = $2, 
		description = $3, 
		category = $4, 
		amount = $5, 
		due_date = $6
	WHERE id = $7
	`
	cmdTag, err := pgTran.Exec(ctx, updateBillQuery,
		body.AccountID,
		body.Merchant,
		body.Description,
		body.Category,
		body.Amount,
		body.DueDate,
		id,
	)
	if err != nil || cmdTag.RowsAffected() == 0 {
		updateErr := ternary(err != nil, err, models.GetNotFound[models.Bill](id))
		return updatedBill, updateErr
	}

	// Fetch the updated bill with nested account (which might change too)
	updatedBill, err = getBill(ctx, pgTran, id)
	if err != nil {
		return updatedBill, err
	}

	if err = pgTran.Commit(ctx); err != nil {
		return updatedBill, err
	}

	return updatedBill, nil
}

// DeleteBill deletes the bill.
//
//   - If the user pays bill, creates a transaction as payment to the bill.
//   - If the user only, deletes the bill, no new transaction is created.
//
// The created transaction has the same account's Id, category, and amount as the bill.
// The paying account's balance also updates to reflect the transaction.
func (s *BillService) DeleteBill(ctx context.Context, id int64, pay bool, recurring bool) error {
	pgTran, err := s.database.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := pgTran.Rollback(ctx); err != nil {
			panic(err)
		}
	}()

	// Store the bill to be deleted, info will be used later
	deletedBill, err := getBill(ctx, pgTran, id)
	if err != nil {
		return err
	}
	cmdTag, err := pgTran.Exec(ctx, "DELETE FROM bills WHERE id = $1;", id)
	if err != nil || cmdTag.RowsAffected() == 0 {
		return ternary(err != nil, err, fmt.Errorf("error updating bill %d", id))
	}

	if pay {
		// Create the transaction in the database representing bill payment
		insertPaymentTranQuery := `
		INSERT INTO transactions (
			account_id, 
			merchant, 
			tran_description, 
			category, 
			amount
		)
		VALUES ($1, $2, $3, $4, $5);
		`
		cmdTag, err := pgTran.Exec(ctx, insertPaymentTranQuery,
			deletedBill.Account.Id,
			deletedBill.Merchant,
			fmt.Sprintf("Payment to %s", deletedBill.Description),
			deletedBill.Category,
			deletedBill.Amount,
		)
		if err != nil || cmdTag.RowsAffected() == 0 {
			// Err not nil is always prioritized over the no row err
			return ternary(err != nil,
				err,
				// Can't create the payment, for some reason
				fmt.Errorf("error creating payment for bill %d", deletedBill.ID),
			)
		}
		// Bill payment is considered expense, so auto-update the balance
		err = updateAccountBalance(
			ctx, pgTran, deletedBill.Account.Id, deletedBill.Amount)
		if err != nil {
			return err
		}
	}
	if recurring {
		// Insert the recurring bill that is due next month
		insertRecurringBillQuery := `
		INSERT INTO bills (
			account_id, 
			merchant, 
			description, 
			category, 
			amount, 
			due_date
		)
		VALUES ($1, $2, $3, $4, $5, $6);
		`
		cmdTag, err := pgTran.Exec(ctx, insertRecurringBillQuery,
			deletedBill.Account.Id,
			deletedBill.Merchant,
			deletedBill.Description,
			deletedBill.Category,
			deletedBill.Amount,
			deletedBill.DueDate.AddDate(0, 1, 0),
		)
		if err != nil || cmdTag.RowsAffected() == 0 {
			return ternary(err != nil,
				err,
				// Can't insert the bil with same data for next deadline
				fmt.Errorf("error reccurring bill for %d", deletedBill.ID),
			)
		}
	}

	if err = pgTran.Commit(ctx); err != nil {
		return err
	}

	return nil
}

// # Helper function
//
// getBill returns a bill with nested account data.
// It uses transaction to execute SQL query, used in an atromic transaction.
func getBill(ctx context.Context, tx pgx.Tx, id int64) (models.Bill, error) {
	var bill models.Bill

	getNestedBillQuery := `
	SELECT
		b.id,
		json_build_object(
			'id', a.id,
			'acc_number', a.acc_number,
			'acc_name', a.acc_name,
			'institution', a.institution,
			'type', a.type,
		) AS account,
		b.merchant, b.description, b.category, b.amount, b.due_date
	FROM bills b
	JOIN accounts a ON b.account_id = a.id
	WHERE b.id = $1;
	`
	billRow := tx.QueryRow(ctx, getNestedBillQuery, id)
	if err := models.ScanBill(billRow, &bill); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return bill, models.GetNotFound[models.Bill](id)
		}
		return bill, err
	}

	return bill, nil
}
