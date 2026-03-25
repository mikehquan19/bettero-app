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

// ListTransactions returns list of transactions of the user from the database
func ListTransactions(ctx context.Context, userId int64, offset int) ([]models.Transaction, error) {
	var transactions []models.Transaction

	rows, err := database.Query(ctx, `
	SELECT
		t.id, t.merchant, t.tran_description, t.category, t.amount, t.created_at, t.updated_at,
		json_build_object(
			'id', a.id,
			'acc_number', a.acc_number,
			'acc_name', a.acc_name,
			'institution', a.institution,
			'type', a.type
		) AS account
	FROM transactions t 
	JOIN accounts a ON t.account_id = a.id
	WHERE a.user_id = $1
	ORDER BY t.created_at DESC
	LIMIT 20 
	OFFSET $2;
	`, userId, offset)
	if err != nil {
		return transactions, err
	}

	transactions, err = pgx.CollectRows(rows, pgx.RowToStructByName[models.Transaction])
	if err != nil {
		return transactions, err
	}

	return transactions, nil
}

// CreateTransaction inserts transaction into the database, and update the account's balance.
// If transaction references a non-existent account, it throws a custom not found error
//
// For an income transaction, by transaction amount:
//   - Debit card's balance will increase (income check, tax refund, allowance, etc)
//   - Credit card's balance will decrease (monthly payment, refund, etc.)
//
// For an expense transaction, by transaction amount:
//   - Debit card's balance will decrease
//   - Credit card's balance will increase
func CreateTransaction(ctx context.Context, body models.TransactionBody, createdAt time.Time) (models.Transaction, error) {
	var newTransaction models.Transaction

	pgTran, err := database.Begin(ctx)
	if err != nil {
		return newTransaction, err
	}
	defer pgTran.Rollback(ctx)

	// Insert the new transaction, get its id, category, and amount
	var (
		transactionId int64
		category      string
		amount        float64
	)
	newTranRow := pgTran.QueryRow(ctx, `
	INSERT INTO transactions (
		account_id, merchant, tran_description, category, amount, created_at
	)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id, category, amount;
	`,
		body.AccountID,
		body.Merchant,
		body.TranDescription,
		body.Category,
		body.Amount,
		createdAt,
	)
	if err = newTranRow.Scan(&transactionId, &category, &amount); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Insert a transaction tht references non-existent account
			if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23503" {
				return newTransaction, fmt.Errorf("Account %d not found", body.AccountID)
			}
		}
		return newTransaction, err
	}

	// Update the account balance
	netChange := ternary(category == "Income", -amount, amount)
	if err = updateAccountBalanceWithTx(ctx, pgTran, int64(body.AccountID), netChange); err != nil {
		return newTransaction, err
	}

	// Get the created transaction with nested account
	newTransaction, err = getTransactionWithTx(ctx, pgTran, transactionId)
	if err != nil {
		return newTransaction, err
	}

	if err = pgTran.Commit(ctx); err != nil {
		return newTransaction, err
	}

	return newTransaction, nil
}

// UpateTransaction updates the transaction's info and update the account's balance.
// If transaction doesn't exist, it returns a custom not found error
//
// Net change to be updated is computed as follows:
//
// # net change = current effect - previous effect
//
//   - previous effect: amount if the transaction with previous info was deleted
//   - current effect: amount if the transaction with updated info was inserted
func UpdateTransaction(
	ctx context.Context, id int64, body models.TransactionBody, createdAt time.Time,
) (models.Transaction, error) {
	var updatedTransaction models.Transaction

	pgTran, err := database.Begin(ctx)
	if err != nil {
		return updatedTransaction, err
	}
	defer pgTran.Rollback(ctx)

	// Store the previous attributes before updating
	var (
		prevCategory string
		prevAmount   float64
	)
	prevRow := pgTran.QueryRow(ctx, `SELECT category, amount from transactions WHERE id = $1;`, id)
	if err = prevRow.Scan(&prevCategory, &prevAmount); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Update a non-existent transaction
			return updatedTransaction, fmt.Errorf("Transaction %d not found", id)
		}
		return updatedTransaction, err
	}

	// Update the transaction, and get the current category and amount
	var (
		accountId    int64
		currCategory string
		currAmount   float64
	)
	updatedTranRow := pgTran.QueryRow(ctx, `
	UPDATE transactions
	SET merchant = $2, 
		tran_description = $3, 
		category = $4, 
		amount = $5, 
		created_at = $6, 
		updated_at = NOW()
	WHERE id = $1
	RETURNING account_id, category, amount;
	`,
		id,
		body.Merchant,
		body.TranDescription,
		body.Category,
		body.Amount,
		createdAt,
	)
	if err = updatedTranRow.Scan(&accountId, &currCategory, &currAmount); err != nil {
		return updatedTransaction, err
	}

	// Compute the amount to update the account balance
	prevChange := ternary(prevCategory == "Income", -prevAmount, +prevAmount)
	currChange := ternary(currCategory == "Income", -currAmount, +currAmount)
	err = updateAccountBalanceWithTx(ctx, pgTran, accountId, currChange-prevChange)
	if err != nil {
		return updatedTransaction, err
	}

	// Get the updated transaction with nested account
	updatedTransaction, err = getTransactionWithTx(ctx, pgTran, id)
	if err != nil {
		return updatedTransaction, err
	}

	if err = pgTran.Commit(ctx); err != nil {
		return updatedTransaction, err
	}

	return updatedTransaction, nil
}

// DeleteTransaction deletes the transaction from database and update balance.
// If transaction doesn't exist, it returns a custom not found error.
//
// For an income transaction,
//   - Debit card's balance will decrease (income check, tax refund, allowance, etc)
//   - Credit card's balance will increase (monthly payment, refund, etc.)
//
// For an expense transaction,
//   - Debit card's balance will increase
//   - Credit card's balance will decrease
func DeleteTransaction(ctx context.Context, id int64) error {
	pgTran, err := database.Begin(ctx)
	if err != nil {
		return err
	}
	defer pgTran.Rollback(ctx)

	// Delete the transaction with a given Id
	var (
		accountId int64
		category  string
		amount    float64
	)
	deletedTranRow := pgTran.QueryRow(ctx, `
	DELETE FROM transactions WHERE id = $1 RETURNING account_id, category, amount;
	`, id)
	if err = deletedTranRow.Scan(&accountId, &category, &amount); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("Transaction %d not found", id)
		}
		return err
	}

	// Reverse the effect of CreateTransaction
	netChange := ternary(category == "Income", amount, -amount)
	if err = updateAccountBalanceWithTx(ctx, pgTran, accountId, netChange); err != nil {
		return err
	}

	if err = pgTran.Commit(ctx); err != nil {
		return err
	}

	return nil
}

// GetTransactionWithTx returns a transaction with nested account data.
//
// It uses transaction (not database) to execute SQL query because this is used with
// other queries inside an atromic transaction.
//
// If transaction doesn't exist, it sends a custom not found error.
func getTransactionWithTx(ctx context.Context, tx pgx.Tx, id int64) (models.Transaction, error) {
	var transaction models.Transaction

	tranRow := tx.QueryRow(ctx, `
	SELECT
		t.id,
		json_build_object(
			'id', a.id,
			'acc_number', a.acc_number,
			'acc_name', a.acc_name,
			'institution', a.institution,
			'type', a.type
		) AS account,
		t.merchant, t.tran_description, t.category, t.amount, t.created_at, t.updated_at
	FROM transactions t
	JOIN accounts a ON t.account_id = a.id
	WHERE t.id = $1;
	`, id)
	if err := models.ScanTransaction(tranRow, &transaction); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return transaction, fmt.Errorf("Transaction %d not found", id)
		}
		return transaction, err
	}

	return transaction, nil
}
