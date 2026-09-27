package repositories

import (
	"betterov2/models"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionRepo struct{}

func NewTransactionRepo() *TransactionRepo {
	return &TransactionRepo{}
}

// FilterTransactions returns the list of paginated transactions of the user
func (r *TransactionRepo) FilterTransactions(
	ctx context.Context,
	db *pgxpool.Pool,
	userId uuid.UUID,
	filter models.TransactionFilter,
	offset int,
) (int, []models.Transaction, error) {
	var transactionCount int
	var transactions []models.Transaction

	tx, err := db.Begin(ctx)
	if err != nil {
		return -1, nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	condition, args := buildTransactionFilter("a.user_id = $1", userId, filter)

	// Fetch the total number of transactions
	countTransactionSQL := fmt.Sprintf(`
	SELECT COUNT(*)
	FROM transactions t 
	JOIN accounts a ON t.account_id = a.id
	WHERE %s;`, condition)

	row := tx.QueryRow(ctx, countTransactionSQL, args...)
	if err := row.Scan(&transactionCount); err != nil {
		return -1, nil, err
	}

	// List the page of transactions from this filter
	// TODO: Don't use OFFSET, change to a more scalable approach
	listTransactionSQL := fmt.Sprintf(`
	SELECT 
		t.id, 
		json_build_object(
			'id', a.id,
			'account_number', a.account_number,
			'name', a.name,
			'institution', a.institution,
			'type', a.type
		) AS account,
		t.merchant, 
		t.description,
		t.category, 
		t.amount_cents,
		t.created_at, 
		t.updated_at
	FROM transactions t
	JOIN accounts a ON t.account_id = a.id
	WHERE %s
	ORDER BY t.created_at DESC 
	LIMIT 20 OFFSET $%d;`, condition, len(args)+1)
	args = append(args, offset)

	rows, err := tx.Query(ctx, listTransactionSQL, args...)
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

	return transactionCount, transactions, err
}

// Get the transaction with nested account data of a given id
func (r *TransactionRepo) GetTransaction(ctx context.Context, dbtx DBTX, id uuid.UUID) (models.Transaction, error) {
	var transaction models.Transaction

	const getTransactionSQL = `
	SELECT
		t.id,
		json_build_object(
			'id', a.id,
			'account_number', a.account_number,
			'name', a.name,
			'institution', a.institution,
			'type', a.type
		) AS account,
		t.merchant, 
		t.description,
		t.category, 
		t.amount_cents,
		t.created_at, 
		t.updated_at
	FROM transactions t
	JOIN accounts a ON t.account_id = a.id
	WHERE t.id = $1;`

	row := dbtx.QueryRow(ctx, getTransactionSQL, id)
	if err := scanTransaction(row, &transaction); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Transaction{}, models.ErrNotFound
		}
		return models.Transaction{}, err
	}

	return transaction, nil
}

// ListSuggestions returns the list of results for autocompletes to search for transactions
func (r *TransactionRepo) ListSuggestions(
	ctx context.Context,
	dbtx DBTX,
	userId uuid.UUID,
	keyword string,
) ([]models.Suggestion, error) {
	var suggestions []models.Suggestion

	tx, err := dbtx.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Set the threshold per query since this is session-scoped
	// NOTE: Keep low enough so it can be diverse, but also high enough
	// so it doesn't return completely unrelated result
	_, err = tx.Exec(ctx, "SET pg_trgm.similarity_threshold = 0.2;")
	if err != nil {
		return nil, err
	}

	autocompleteSQL := `
	SELECT type, name
	FROM (
		SELECT DISTINCT
			'description' AS type,
			t.description AS name,
			similarity(t.description, $2) AS score
		FROM transactions t
		JOIN accounts a ON t.account_id = a.id
		WHERE t.description % $2 AND a.user_id = $1

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
	LIMIT 10;`

	rows, err := tx.Query(ctx, autocompleteSQL, userId, keyword)
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
func (r *TransactionRepo) InsertTransaction(ctx context.Context, dbtx DBTX, body models.PostTransactionBody) (models.Transaction, error) {
	var newTransaction models.Transaction

	// Insert the new transaction, get its id, category, and amount
	const insertTransactionSQL = `
	WITH new_transaction AS (
		INSERT INTO transactions (
			account_id,
			merchant,
			description,
			category,
			amount_cents,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING *
	)
	SELECT
		t.id,
		json_build_object(
			'id', a.id,
			'account_number', a.account_number,
			'name', a.name,
			'institution', a.institution,
			'type', a.type
		) AS account,
		t.merchant, 
		t.description,
		t.category, 
		t.amount_cents,
		t.created_at, 
		t.updated_at
	FROM new_transaction t
	JOIN accounts a ON a.id = t.account_id;`

	row := dbtx.QueryRow(ctx, insertTransactionSQL,
		body.AccountID,
		body.Merchant,
		body.Description,
		body.Category,
		body.AmountCents,
		body.CreatedAt,
	)
	if err := scanTransaction(row, &newTransaction); err != nil {
		if isForeignKeyViolation(err) {
			return models.Transaction{}, models.ErrForeignKey
		}
		return models.Transaction{}, err
	}

	return newTransaction, nil
}

// UpdateTransaction updates and returns the transaction. Doesn't allow for updating account ID
func (r *TransactionRepo) UpdateTransaction(
	ctx context.Context, dbtx DBTX, id uuid.UUID, body models.PutTransactionBody,
) (models.Transaction, error) {
	var updatedTransaction models.Transaction

	// Store the previous category and amount before updating
	const updateTransactionSQL = `
	WITH updated_transaction AS (
		UPDATE transactions
		SET merchant = $2, 
			description = $3,
			category = $4, 
			amount_cents = $5,
			created_at = $6,
			updated_at = NOW()
		WHERE id = $1
		RETURNING *
	)
	SELECT
		t.id,
		json_build_object(
			'id', a.id,
			'account_number', a.account_number,
			'name', a.name,
			'institution', a.institution,
			'type', a.type
		) AS account,
		t.merchant, 
		t.description,
		t.category, 
		t.amount_cents,
		t.created_at, 
		t.updated_at
	FROM updated_transaction t
	JOIN accounts a ON a.id = t.account_id;`

	row := dbtx.QueryRow(ctx, updateTransactionSQL,
		id,
		body.Merchant,
		body.Description,
		body.Category,
		body.AmountCents,
		body.CreatedAt,
	)
	if err := scanTransaction(row, &updatedTransaction); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Transaction{}, models.ErrNotFound
		}
		return models.Transaction{}, err
	}

	return updatedTransaction, nil
}

// DeleteTransaction deletes and returns the transaction
func (r *TransactionRepo) DeleteTransaction(ctx context.Context, dbtx DBTX, id uuid.UUID) (models.Transaction, error) {
	var deletedTransaction models.Transaction

	const deleteTransactionSQL = `
	WITH deleted_transaction AS (
		DELETE FROM transactions WHERE id = $1 RETURNING *
	)
	SELECT
		t.id,
		json_build_object(
			'id', a.id,
			'account_number', a.account_number,
			'name', a.name,
			'institution', a.institution,
			'type', a.type
		) AS account,
		t.merchant, 
		t.description,
		t.category, 
		t.amount_cents,
		t.created_at, 
		t.updated_at
	FROM deleted_transaction t
	JOIN accounts a ON a.id = t.account_id;`

	row := dbtx.QueryRow(ctx, deleteTransactionSQL, id)
	if err := scanTransaction(row, &deletedTransaction); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Transaction{}, models.ErrNotFound
		}
		return models.Transaction{}, err
	}

	return deletedTransaction, nil
}

// GetTransactionSum returns the total sum of all transactions of the account from the date
func (r *TransactionRepo) GetTransactionSum(
	ctx context.Context,
	dbtx DBTX,
	accountId uuid.UUID,
	startDate time.Time,
) (models.Money, error) {
	var transactionSum models.Money

	const totalSumSQL = `
	SELECT
		COALESCE(SUM(t.amount_cents * CASE
			WHEN t.category = 'income' THEN -1
			ELSE 1 
		END), 0)
	FROM transactions t
	WHERE t.account_id = $1 AND t.created_at >= $2;`

	row := dbtx.QueryRow(ctx, totalSumSQL, accountId, startDate)
	if err := row.Scan(&transactionSum); err != nil {
		return 0, err
	}

	return transactionSum, nil
}

// DeleteOutdatedTransactions deleles the list of outdated transactions (older than 6 months ago)
// and returns the number of successfully deleted ones.
func (r *TransactionRepo) DeleteOutdatedTransactions(ctx context.Context, dbtx DBTX, accountId uuid.UUID) (int, error) {
	const deleteOutdatedTranSQL = `
	DELETE FROM transactions
	WHERE 
		created_at < CURRENT_DATE - INTERVAL '6 months'
		AND account_id = $1
	RETURNING id;`

	rows, err := dbtx.Query(ctx, deleteOutdatedTranSQL, accountId)
	if err != nil {
		return -1, err
	}

	deleted, err := pgx.CollectRows(rows, pgx.RowTo[uuid.UUID])
	if err != nil {
		return -1, err
	}

	return len(deleted), nil
}

// ScanTransaction parses the returned row into transaction and destinations
func scanTransaction(tranRow pgx.Row, transaction *models.Transaction) error {
	return tranRow.Scan(
		&transaction.ID,
		&transaction.Account,
		&transaction.Merchant,
		&transaction.Description,
		&transaction.Category,
		&transaction.AmountCents,
		&transaction.CreatedAt,
		&transaction.UpdatedAt,
	)
}
