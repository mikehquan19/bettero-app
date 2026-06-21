package repositories

import (
	"betterov2/models"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionRepo struct {
}

func NewTransactionRepo() *TransactionRepo {
	return &TransactionRepo{}
}

// FilterTransactions returns the list of filtered transactions with nested account
func (r *TransactionRepo) FilterTransactions(
	ctx context.Context, db *pgxpool.Pool, userId int64, filter models.TransactionFilter, offset int,
) (int, []models.Transaction, error) {
	var count int
	var transactions []models.Transaction

	tx, err := db.Begin(ctx)
	if err != nil {
		return -1, nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Dynamically build the filter based on values from query params
	condition, args := buildDynamicFilter("a.user_id = $1", userId, filter)

	// Fetch the total number of transactions
	countQuery := fmt.Sprintf(`
	SELECT COUNT(*)
	FROM transactions t 
	JOIN accounts a ON t.account_id = a.id
	%s
	`, condition)
	row := tx.QueryRow(ctx, countQuery, args...)
	if err := row.Scan(&count); err != nil {
		return -1, nil, err
	}

	// List the page of transactions from this filter
	paginate := fmt.Sprintf("ORDER BY t.created_at DESC LIMIT 15 OFFSET $%d", len(args)+1)
	listQuery := fmt.Sprintf(`
	SELECT 
		t.id, 
		t.merchant, t.tran_description, t.category, t.amount, 
		t.created_at, t.updated_at,
		json_build_object(
			'id', a.id,
			'acc_number', a.acc_number,
			'acc_name', a.acc_name,
			'institution', a.institution,
			'type', a.type
		) AS account
	FROM transactions t
	JOIN accounts a ON t.account_id = a.id
	%s
	%s
	`, condition, paginate)
	args = append(args, offset)

	rows, err := tx.Query(ctx, listQuery, args...)
	if err != nil {
		return -1, nil, err
	}
	transactions, err = pgx.CollectRows(rows, pgx.RowToStructByName[models.Transaction])
	if err != nil {
		return -1, nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return -1, nil, err
	}

	return count, transactions, err
}

// Get the transaction with nested account data of a given id
func (r *TransactionRepo) GetTransaction(ctx context.Context, tx pgx.Tx, id int64) (models.Transaction, error) {
	var transaction models.Transaction

	const getTransactionQuery = `
	SELECT
		t.id,
		json_build_object(
			'id', a.id,
			'acc_number', a.acc_number,
			'acc_name', a.acc_name,
			'institution', a.institution,
			'type', a.type
		) AS account,
		t.merchant, t.tran_description, t.category, t.amount, 
		t.created_at, t.updated_at
	FROM transactions t
	JOIN accounts a ON t.account_id = a.id
	WHERE t.id = $1;
	`
	row := tx.QueryRow(ctx, getTransactionQuery, id)
	if err := models.ScanTransaction(row, &transaction); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return transaction, models.GetNotFound[models.Transaction](id)
		}
		return transaction, err
	}

	return transaction, nil
}

// Returns the list of results for autocompletes to search for transactions
func (r *TransactionRepo) ListSuggestions(
	ctx context.Context, db *pgxpool.Pool, userId int64, keyword string,
) ([]models.Suggestion, error) {
	var suggestions []models.Suggestion

	tx, err := db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Set the threshold per query since this is session-scoped
	// NOTE: Keep low enough so it can be diverse, but also high enough so it doesn't return
	// completely unrelated result
	_, err = tx.Exec(ctx, "SET pg_trgm.similarity_threshold = 0.2;")
	if err != nil {
		return nil, err
	}

	autocompleteQuery := `
	SELECT type, name
	FROM (
		SELECT DISTINCT
			'description' AS type,
			t.tran_description AS name,
			similarity(t.tran_description, $2) AS score
		FROM transactions t
		JOIN accounts a ON t.account_id = a.id
		WHERE t.tran_description % $2 AND a.user_id = $1

		UNION ALL

		SELECT DISTINCT
			'merchant' AS type,
			t.merchant AS name,
			similarity(t.merchant, $2) AS score
		FROM transactions t
		JOIN accounts a ON t.account_id = a.id
		WHERE t.merchant % $2 AND a.user_id = $1
	)
	ORDER BY score DESC
	LIMIT 10;
	`
	rows, err := tx.Query(ctx, autocompleteQuery, userId, keyword)
	if err != nil {
		return nil, err
	}
	suggestions, err = pgx.CollectRows(rows, pgx.RowToStructByName[models.Suggestion])
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return suggestions, nil
}

// InsertTransaction inserts the transaction, and returns the inserted transaction
// that matches the schema in the database (no nested account)
func (r *TransactionRepo) InsertTransaction(ctx context.Context, tx pgx.Tx, body models.PostTransactionBody) (models.Transaction, error) {
	var newTran models.Transaction

	// Insert the new transaction, get its id, category, and amount
	const insertTransactionQuery = `
	WITH new_transaction AS (
		INSERT INTO transactions (
			account_id,
			merchant,
			tran_description,
			category,
			amount,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING *
	)
	SELECT
		t.id,
		json_build_object(
			'id', a.id,
			'acc_number', a.acc_number,
			'acc_name', a.acc_name,
			'institution', a.institution,
			'type', a.type
		) AS account,
		t.merchant, t.tran_description, t.category, t.amount,
		t.created_at, t.updated_at
	FROM new_transaction t
	JOIN accounts a ON a.id = t.account_id;
	`
	row := tx.QueryRow(ctx, insertTransactionQuery,
		body.AccountID,
		body.Merchant,
		body.TranDescription,
		body.Category,
		body.Amount,
		body.CreatedAt,
	)
	if err := models.ScanTransaction(row, &newTran); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23503" {
			// Insert a transaction for non-existent account
			return newTran, models.GetForeignKey[models.Account](body.AccountID)
		}
		return newTran, err
	}

	return newTran, nil
}

// UpdateTransaction updates the transaction's info, and returns the updated transaction
// that matches the schema in the database (no nested account)
func (r *TransactionRepo) UpdateTransaction(
	ctx context.Context, tx pgx.Tx, id int64, body models.PutTransactionBody,
) (models.Transaction, error) {
	var updatedTran models.Transaction

	// Store the previous category and amount before updating
	const updateTranQuery = `
	WITH updated_transaction AS (
		UPDATE transactions
		SET merchant = $2, 
			tran_description = $3, 
			category = $4, 
			amount = $5, 
			created_at = $6,
			updated_at = NOW()
		WHERE id = $1
		RETURNING *;
	)
	SELECT
		t.id,
		json_build_object(
			'id', a.id,
			'acc_number', a.acc_number,
			'acc_name', a.acc_name,
			'institution', a.institution,
			'type', a.type
		) AS account,
		t.merchant, t.tran_description, t.category, t.amount,
		t.created_at, t.updated_at
	FROM updated_transaction t
	JOIN accounts a ON a.id = t.account_id;
	`
	row := tx.QueryRow(ctx, updateTranQuery,
		id,
		body.Merchant,
		body.TranDescription,
		body.Category,
		body.Amount,
		body.CreatedAt,
	)
	if err := models.ScanTransaction(row, &updatedTran); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Update a non-existent transaction
			return updatedTran, models.GetNotFound[models.Transaction](id)
		}
		return updatedTran, err
	}

	return updatedTran, nil
}

// DeleteTransaction deletes the transaction, and returns the deleted transaction
// that matches the schema in the database (no nested account)
func (r *TransactionRepo) DeleteTransaction(ctx context.Context, tx pgx.Tx, id int64) (models.Transaction, error) {
	var deletedTran models.Transaction

	const deleteTranQuery = `
	WITH deleted_transaction AS (
		DELETE FROM transactions WHERE id = $1 RETURNING *
	)
	SELECT
		t.id,
		json_build_object(
			'id', a.id,
			'acc_number', a.acc_number,
			'acc_name', a.acc_name,
			'institution', a.institution,
			'type', a.type
		) AS account,
		t.merchant, t.tran_description, t.category, t.amount,
		t.created_at, t.updated_at
	FROM deleted_transaction t
	JOIN accounts a ON a.id = t.account_id;
	`
	row := tx.QueryRow(ctx, deleteTranQuery, id)
	if err := models.ScanTransaction(row, &deletedTran); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return deletedTran, models.GetNotFound[models.Transaction](id)
		}
		return deletedTran, err
	}

	return deletedTran, nil
}

// Return the total sum of all transactions of an account from a given date.
// Used by cron job.
func (r *TransactionRepo) GetTransactionSum(ctx context.Context, tx pgx.Tx, accountId int64, from time.Time) (float64, error) {
	var transactionSum float64

	const totalSumQuery = `
	SELECT
		SUM(t.amount * CASE 
			WHEN t.category = 'Income' THEN -1 
			ELSE 1 
		END)
	FROM transactions t
	WHERE t.account_id = $1 AND t.created_at > $2
	`
	row := tx.QueryRow(ctx, totalSumQuery, accountId, from)
	if err := row.Scan(&transactionSum); err != nil {
		return -1, err
	}

	return transactionSum, nil
}

// DeleteOutdatedTransactions deleles the list of outdated transactions (older than 6 months ago)
// and returns the number of successfully deleted ones.
func (r *TransactionRepo) DeleteOutdatedTransactions(ctx context.Context, db *pgxpool.Pool, accountId int64) (int, error) {
	const deleteOutdatedTranQuery = `
	DELETE FROM transactions
	WHERE 
		created_at < CURRENT_DATE - INTERVAL '6 months'
		AND account_id = $1
	RETURNING id;
	`
	rows, err := db.Query(ctx, deleteOutdatedTranQuery, accountId)
	if err != nil {
		return -1, err
	}

	deleted, err := pgx.CollectRows(rows, pgx.RowTo[int64])
	if err != nil {
		return -1, err
	}

	return len(deleted), nil
}
