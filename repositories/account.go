package repositories

import (
	"betterov2/models"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AccountRepo struct {
}

func NewAccountRepo() *AccountRepo {
	return &AccountRepo{}
}

// ListAcccounts returns the list of accounts of given user,
// sorted by updated recency.
func (r *AccountRepo) ListAccounts(ctx context.Context, db *pgxpool.Pool, userId int64) ([]models.Account, error) {
	var accounts []models.Account

	const listAccountQuery = `
	SELECT * FROM accounts 
	WHERE user_id = $1 ORDER BY updated_at DESC;
	`
	rows, err := db.Query(ctx, listAccountQuery, userId)
	if err != nil {
		return nil, err
	}
	accounts, err = pgx.CollectRows(rows, pgx.RowToStructByName[models.Account])
	if err != nil {
		return nil, err
	}

	return accounts, nil
}

// GetAccount gets the account with given ID
func (r *AccountRepo) GetAccount(ctx context.Context, tx pgx.Tx, id int64) (models.Account, error) {
	var account models.Account

	accountRow := tx.QueryRow(ctx, "SELECT * FROM accounts WHERE id = $1;", id)
	if err := models.ScanAccount(accountRow, &account); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return account, models.GetNotFound[models.Account](id)
		}
		return account, err
	}

	return account, nil
}

func (r *AccountRepo) ListTransactions(
	ctx context.Context, db *pgxpool.Pool, id int64, filter models.TransactionFilter, offset int64,
) (int, []models.Transaction, error) {
	var count int
	var transactions []models.Transaction

	tx, err := db.Begin(ctx)
	if err != nil {
		return -1, nil, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			panic(err)
		}
	}()

	// Dynamically build the filter based on values from query params
	sql, args := buildDynamicFilter("a.id = $1", id, filter)

	// Fetch the total number of transactions
	const baseCountQuery = `
	SELECT COUNT(*)
	FROM transactions t 
	JOIN accounts a ON t.account_id = a.id
	`
	row := tx.QueryRow(ctx, baseCountQuery+sql, args...)
	if err := row.Scan(&count); err != nil {
		return -1, nil, err
	}

	// List the page of transactions from this filter
	const baseListQuery = `
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

	rows, err := tx.Query(ctx, baseListQuery+sql+page, args...)
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

func (r *AccountRepo) InsertAccount(
	ctx context.Context, tx pgx.Tx, userId int64, body models.PostAccountBody,
) (models.Account, error) {
	var newAccount models.Account

	// If the account's body doesn't satify the requirements
	if err := body.Validate(); err != nil {
		return newAccount, err
	}

	const insertAccountQuery = `
	INSERT INTO accounts (
		user_id, 
		acc_number, 
		acc_name, 
		institution, 
		type, 
		balance, 
		credit_limit, 
		next_due
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	RETURNING *;
	`
	row := tx.QueryRow(ctx, insertAccountQuery,
		userId,
		body.AccNumber,
		body.AccName,
		body.Institution,
		body.Type,
		body.Balance,
		body.CreditLimit,
		body.NextDue,
	)
	if err := models.ScanAccount(row, &newAccount); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23503" {
			// Insert account that references non-existent user
			return newAccount, models.GetForeignKey[models.User](userId)
		}
		return newAccount, err
	}

	return newAccount, nil
}

// Update the account's data and returns the updated data, doesn't allow for updating type.
func (r *AccountRepo) UpdateAccount(
	ctx context.Context, tx pgx.Tx, id int64, body models.PutAccountBody,
) (models.Account, error) {
	var updatedAccount models.Account

	updateAccountQuery := `
	UPDATE accounts
	SET acc_number = $2, 
		acc_name = $3, 
		institution = $4, 
		balance = $5, 
		credit_limit = $6, 
		next_due = $7, 
		updated_at = NOW()
	WHERE id = $1
	RETURNING *;
	`
	row := tx.QueryRow(ctx, updateAccountQuery,
		id,
		body.AccNumber,
		body.AccName,
		body.Institution,
		body.Balance,
		body.CreditLimit,
		body.NextDue,
	)
	if err := models.ScanAccount(row, &updatedAccount); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return updatedAccount, models.GetNotFound[models.Account](id)
		}
		return updatedAccount, err
	}

	return updatedAccount, nil
}

// UpdateAccountBalance updates the balance of the account by the net change.
//
//   - Debit card's balance will decrease
//   - Credit card's balance will increase
func (r *AccountRepo) UpdateAccountBalance(
	ctx context.Context, tx pgx.Tx, id int64, netChange float64,
) error {
	const updateBalanceQuery = `
	UPDATE accounts a
	SET balance = CASE
			WHEN type = 'Debit' THEN balance - $2
			ELSE balance + $2
		END,
		updated_at = NOW()
	WHERE id = $1;
	`
	cmdTag, err := tx.Exec(ctx, updateBalanceQuery, id, netChange)
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("error updating balance of account %d", id)
	}
	return err
}

// Update the account's data and returns the deleted account
func (r *AccountRepo) DeleteAccount(ctx context.Context, tx pgx.Tx, id int64) (models.Account, error) {
	var deletedAccount models.Account

	const deleteAccountQuery = `
	DELETE FROM accounts WHERE id = $1
	RETURNING *;
	`
	row := tx.QueryRow(ctx, deleteAccountQuery, id)
	if err := models.ScanAccount(row, &deletedAccount); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return deletedAccount, models.GetNotFound[models.Account](id)
		}
		return deletedAccount, err
	}
	return deletedAccount, nil
}
