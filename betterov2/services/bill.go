package services

import (
	"betterov2/models"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// ListBills gets the list of bills orded by its due date
func ListBills(ctx context.Context, userId int64) ([]models.Bill, error) {
	var bills []models.Bill

	billRows, err := database.Query(ctx, `
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
	`, userId)
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
// It returns not found error if bill references non-existent account
func CreateBill(ctx context.Context, body models.BillBody, dueDate time.Time) (models.Bill, error) {
	var newBill models.Bill

	pgTran, err := database.Begin(ctx)
	if err != nil {
		return newBill, err
	}

	// Insert the bill into the database and obtains the bill's Id
	var insertedBillId int64
	insertedBillRow := pgTran.QueryRow(ctx, `
	INSERT INTO bills (
		account_id, merchant, description, category, amount, due_date
	)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id;
	`,
		body.AccountID, body.Merchant, body.Description,
		body.Category, body.Amount, dueDate,
	)
	if err := insertedBillRow.Scan(&insertedBillId); err != nil {
		// Inserts a bill which references non-existent account
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23503" {
			return newBill, fmt.Errorf("Account %d not found", body.AccountID)
		}
		return newBill, err
	}

	// Fetch the created bill with nested account for returning
	newBill, err = GetBillWithTx(ctx, pgTran, insertedBillId)
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
func UpdateBill(
	ctx context.Context, id int64, body models.BillBody, dueDate time.Time,
) (models.Bill, error) {
	var updatedBill models.Bill

	pgTran, err := database.Begin(ctx)
	if err != nil {
		return updatedBill, err
	}

	// Update the bill
	cmdTag, err := pgTran.Exec(ctx, `
	UPDATE bills
	SET account_id = $1, merchant = $2, description = $3, 
		category = $4, amount = $5, due_date = $6
	WHERE id = $7
	`,
		body.AccountID, body.Merchant, body.Description,
		body.Category, body.Amount, dueDate,
		id,
	)
	if err != nil {
		return updatedBill, err
	}
	if cmdTag.RowsAffected() == 0 {
		return updatedBill, fmt.Errorf("Bill %d not found", id)
	}

	// Fetch the updated bill with nested account (which might change too)
	updatedBill, err = GetBillWithTx(ctx, pgTran, id)
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
func DeleteBill(ctx context.Context, id int64, pay bool, recurring bool) error {
	pgTran, err := database.Begin(ctx)
	if err != nil {
		return err
	}

	// Store the bill to be deleted, since its details will be used later
	deletedBill, err := GetBillWithTx(ctx, pgTran, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("Bill %d not found", id)
		}
		return err
	}
	cmdTag, err := pgTran.Exec(ctx, `DELETE FROM bills WHERE id = $1;`, id)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("Error updating bill %d", id)
	}

	if pay {
		// Create the transaction in the database representing bill payment
		cmdTag, err := pgTran.Exec(ctx, `
		INSERT INTO transactions (
			account_id, merchant, tran_description, category, amount
		)
		VALUES ($1, $2, $3, $4, $5);
		`,
			deletedBill.Account.Id,
			deletedBill.Merchant,
			fmt.Sprintf("Payment to %s", deletedBill.Description),
			deletedBill.Category,
			deletedBill.Amount,
		)
		if err != nil {
			return err
		}
		if cmdTag.RowsAffected() == 0 {
			return fmt.Errorf("Error creating payment for bill %d", deletedBill.ID)
		}
		// Helper updateAccBalance from transaction repo
		// Bill payment is considered expense
		if err = updateAccountBalanceWithTx(
			ctx, pgTran, deletedBill.Account.Id, deletedBill.Amount); err != nil {
			return err
		}
	}
	if recurring {
		cmdTag, err := pgTran.Exec(ctx, `
		INSERT INTO bills (
			account_id, merchant, description, category, amount, due_date
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		`,
			deletedBill.Account.Id,
			deletedBill.Merchant,
			deletedBill.Description,
			deletedBill.Category,
			deletedBill.Amount,
			deletedBill.DueDate.AddDate(0, 1, 0),
		)
		if err != nil {
			return err
		}
		if cmdTag.RowsAffected() == 0 {
			return fmt.Errorf("Error creating reccurring bill for %d", deletedBill.ID)
		}
	}

	if err = pgTran.Commit(ctx); err != nil {
		return err
	}

	return nil
}

// GetBillWithTx returns a bill with nested account data.
//
// It uses transaction (not database instance) to execute SQL query because this is used with
// other queries inside an atromic transaction.
//
// If bill doesn't exist, it returns a custom not found error.
func GetBillWithTx(ctx context.Context, tx pgx.Tx, id int64) (models.Bill, error) {
	var bill models.Bill

	billRow := tx.QueryRow(ctx, `
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
	`, id)
	if err := models.ScanBill(billRow, &bill); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return bill, fmt.Errorf("Bill %d not found", id)
		}
		return bill, err
	}

	return bill, nil
}
