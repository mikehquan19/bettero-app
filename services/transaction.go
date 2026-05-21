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

// Transaction Service
type TransactionService struct {
	database *pgxpool.Pool
}

// Generate a new transaction service
func NewTransactionService(database *pgxpool.Pool) *TransactionService {
	return &TransactionService{
		database: database,
	}
}

// FilterTransactions returns list of transactions of category between 2 dates
func (s *TransactionService) FilterTransactions(
	ctx context.Context, userId int64, tranFilter models.TransactionFilter, offset int,
) (int, []models.Transaction, error) {
	var totalCount int
	var transactions []models.Transaction

	pgTran, err := s.database.Begin(ctx)
	if err != nil {
		return totalCount, nil, err
	}
	defer func() {
		if err := pgTran.Rollback(ctx); err != nil {
			panic(err)
		}
	}()

	// Dynamically build the filter based on values from query params
	filter, args := buildFilterSQL("user_id = $1", userId, tranFilter)

	// Fetch the total number of transactions from this filter
	countQuery := `
	SELECT COUNT(*) 
	FROM transactions t 
	JOIN accounts a ON t.account_id = a.id
	`
	countRow := pgTran.QueryRow(ctx, countQuery+filter, args...)
	if err := countRow.Scan(&totalCount); err != nil {
		return totalCount, nil, err
	}

	// List the page of transactions from this filter
	listTranQuery := `
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
	`
	page := fmt.Sprintf(" ORDER BY t.created_at DESC LIMIT 15 OFFSET $%d", len(args)+1)
	args = append(args, offset)

	tranRows, err := pgTran.Query(ctx, listTranQuery+filter+page, args...)
	if err != nil {
		return totalCount, nil, err
	}
	transactions, err = pgx.CollectRows(tranRows, pgx.RowToStructByName[models.Transaction])
	if err != nil {
		return totalCount, nil, err
	}

	if err = pgTran.Commit(ctx); err != nil {
		return totalCount, nil, err
	}

	return totalCount, transactions, nil
}

// ListSuggestions returns the list of transaction description
func (s *TransactionService) ListSuggestions(ctx context.Context, userId int64, q string) ([]models.Suggestion, error) {
	var suggestions []models.Suggestion

	pgTran, err := s.database.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := pgTran.Rollback(ctx); err != nil {
			panic(err)
		}
	}()

	// Set the threshold per query since this is session-scoped
	_, err = pgTran.Exec(ctx, "SET pg_trgm.similarity_threshold = 0.2;")
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
	rows, err := pgTran.Query(ctx, autocompleteQuery, userId, q)
	if err != nil {
		return nil, err
	}
	suggestions, err = pgx.CollectRows(rows, pgx.RowToStructByName[models.Suggestion])
	if err != nil {
		return nil, err
	}

	if err = pgTran.Commit(ctx); err != nil {
		return nil, err
	}

	return suggestions, nil
}

// CreateTransaction inserts transaction into the database, and update the account's balance.
//
// For an income transaction, by transaction amount:
//   - Debit card's balance will increase (income check, tax refund, allowance, etc)
//   - Credit card's balance will decrease (monthly payment, refund, etc.)
//
// For an expense transaction, by transaction amount:
//   - Debit card's balance will decrease
//   - Credit card's balance will increase
func (s *TransactionService) CreateTransaction(ctx context.Context, body models.PostTransactionBody) (models.Transaction, error) {
	var newTransaction models.Transaction

	pgTran, err := s.database.Begin(ctx)
	if err != nil {
		return newTransaction, err
	}
	defer func() {
		if err := pgTran.Rollback(ctx); err != nil {
			panic(err)
		}
	}()

	// Insert the new transaction, get its id, category, and amount
	var transactionId int64
	insertTranQuery := `
	INSERT INTO transactions (
		account_id, 
		merchant, 
		tran_description, 
		category, 
		amount, 
		created_at
	)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id;
	`
	newTranRow := pgTran.QueryRow(ctx, insertTranQuery,
		body.AccountID,
		body.Merchant,
		body.TranDescription,
		body.Category,
		body.Amount,
		body.CreatedAt,
	)
	if err = newTranRow.Scan(&transactionId); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23503" {
			// Insert a transaction for non-existent account
			return newTransaction, models.GetForeignKey[models.Account](int64(body.AccountID))
		}
		return newTransaction, err
	}

	// Update the account balance
	netChange := ternary(body.Category == "Income", -body.Amount, body.Amount)
	accId := int64(body.AccountID)
	if err = updateAccountBalance(ctx, pgTran, accId, netChange); err != nil {
		return newTransaction, err
	}

	// Get the created transaction with nested account
	newTransaction, err = getTransaction(ctx, pgTran, transactionId)
	if err != nil {
		return newTransaction, err
	}

	if err = pgTran.Commit(ctx); err != nil {
		return newTransaction, err
	}

	return newTransaction, nil
}

// UpateTransaction updates the transaction's info and update the account's balance.
//
// Net change to be updated is computed as follows:
//
// # net change = current effect - previous effect
//
//   - previous effect: amount if the transaction with previous info was deleted
//   - current effect: amount if the transaction with updated info was inserted
func (s *TransactionService) UpdateTransaction(ctx context.Context, id int64, body models.PutTransactionBody) (models.Transaction, error) {
	var updatedTransaction models.Transaction

	pgTran, err := s.database.Begin(ctx)
	if err != nil {
		return updatedTransaction, err
	}
	defer func() {
		if err := pgTran.Rollback(ctx); err != nil {
			panic(err)
		}
	}()

	// Store the previous category and amount before updating
	var prevCategory string
	var prevAmount float64
	getPrevTranQuery := "SELECT category, amount from transactions WHERE id = $1;"
	prevRow := pgTran.QueryRow(ctx, getPrevTranQuery, id)
	if err = prevRow.Scan(&prevCategory, &prevAmount); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Update a non-existent transaction
			return updatedTransaction, models.GetNotFound[models.Transaction](id)
		}
		return updatedTransaction, err
	}

	// Update the transaction, and get the current category and amount
	var accountId int64
	updateTranQuery := `
	UPDATE transactions
	SET merchant = $2, 
		tran_description = $3, 
		category = $4, 
		amount = $5, 
		created_at = $6, 
		updated_at = NOW()
	WHERE id = $1
	RETURNING account_id;
	`
	updatedTranRow := pgTran.QueryRow(ctx, updateTranQuery,
		id,
		body.Merchant,
		body.TranDescription,
		body.Category,
		body.Amount,
		body.CreatedAt,
	)
	if err = updatedTranRow.Scan(&accountId); err != nil {
		return updatedTransaction, err
	}

	// Compute the amount to update the account balance
	prevChange := ternary(prevCategory == "Income", -prevAmount, prevAmount)
	currChange := ternary(body.Category == "Income", -body.Amount, body.Amount)
	err = updateAccountBalance(ctx, pgTran, accountId, currChange-prevChange)
	if err != nil {
		return updatedTransaction, err
	}

	// Get the updated transaction with nested account
	updatedTransaction, err = getTransaction(ctx, pgTran, id)
	if err != nil {
		return updatedTransaction, err
	}

	if err = pgTran.Commit(ctx); err != nil {
		return updatedTransaction, err
	}

	return updatedTransaction, nil
}

// DeleteTransaction deletes the transaction from database and update balance.
//
// For an income transaction,
//   - Debit card's balance will decrease (income check, tax refund, allowance, etc)
//   - Credit card's balance will increase (monthly payment, refund, etc.)
//
// For an expense transaction,
//   - Debit card's balance will increase
//   - Credit card's balance will decrease
func (s *TransactionService) DeleteTransaction(ctx context.Context, id int64) error {
	pgTran, err := s.database.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := pgTran.Rollback(ctx); err != nil {
			panic(err)
		}
	}()

	// Delete the transaction with a given Id
	// Return account's info, category, and amount to update the acount
	var accountId int64
	var category string
	var amount float64
	deleteTranQuery := `
	DELETE FROM transactions WHERE id = $1 
	RETURNING account_id, category, amount;
	`
	deletedTranRow := pgTran.QueryRow(ctx, deleteTranQuery, id)
	if err = deletedTranRow.Scan(&accountId, &category, &amount); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.GetNotFound[models.Transaction](id)
		}
		return err
	}

	// Reverse the effect of creating the transaction
	netChange := ternary(category == "Income", amount, -amount)
	if err = updateAccountBalance(ctx, pgTran, accountId, netChange); err != nil {
		return err
	}

	if err = pgTran.Commit(ctx); err != nil {
		return err
	}

	return nil
}

// # HELPER FUNCTION
//
// getTransaction returns a transaction with nested account data.
// It uses transaction to execute SQL query with other queries.
func getTransaction(ctx context.Context, tx pgx.Tx, id int64) (models.Transaction, error) {
	var transaction models.Transaction
	getTranQuery := `
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
	tranRow := tx.QueryRow(ctx, getTranQuery, id)
	if err := models.ScanTransaction(tranRow, &transaction); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return transaction, models.GetNotFound[models.Transaction](id)
		}
		return transaction, err
	}

	return transaction, nil
}
