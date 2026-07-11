package repositories

import (
	"betterov2/models"
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AccountRepo struct{}

func NewAccountRepo() *AccountRepo {
	return &AccountRepo{}
}

// ListAcccounts returns the list of accounts of given user,
// sorted by updated recency.
func (r *AccountRepo) ListAccounts(ctx context.Context, db models.DBTX, userId int64) ([]models.Account, error) {
	var accounts []models.Account

	const listAccountQuery = `SELECT * FROM accounts WHERE user_id = $1 ORDER BY updated_at DESC;`
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

// Available filter status, we can add more later
type FilterStatus string

const (
	PastDue   FilterStatus = "PAST_DUE"
	Unflagged FilterStatus = "UNFLAGGED"
	All       FilterStatus = "All"
)

// ListAllAccounts returns the list of accounts based on filter status
// This method primarily is used by cron jobs.
func (r *AccountRepo) ListAllAccounts(ctx context.Context, db models.DBTX, status FilterStatus) ([]models.Account, error) {
	var accounts []models.Account

	var filter string
	switch status {
	case PastDue:
		filter = "WHERE next_due < CURRENT_DATE"
	case Unflagged:
		filter = "WHERE discrepancy_flagged = FALSE"
	case All:
		filter = ""
	}

	listAllAccountQuery := fmt.Sprintf("SELECT * FROM accounts %s", filter)
	rows, err := db.Query(ctx, listAllAccountQuery)
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
func (r *AccountRepo) GetAccount(ctx context.Context, db models.DBTX, id int64) (models.Account, error) {
	var account models.Account

	row := db.QueryRow(ctx, "SELECT * FROM accounts WHERE id = $1;", id)
	if err := models.ScanAccount(row, &account); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return account, models.GetNotFound[models.Account](id)
		}
		return account, err
	}

	return account, nil
}

// ListAccountTransactions returns the paginated list of transactions of the acount
func (r *AccountRepo) ListAccountTransactions(
	ctx context.Context,
	db *pgxpool.Pool,
	id int64,
	filter models.TransactionFilter,
	offset int64,
) (int, []models.Transaction, error) {
	var count int
	var transactions []models.Transaction

	tx, err := db.Begin(ctx)
	if err != nil {
		return -1, nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	condition, args := buildTransactionFilter("t.account_id = $1", id, filter)

	// Fetch the total number of transactions
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM transactions t WHERE %s;`, condition)
	log.Println(countQuery)
	row := tx.QueryRow(ctx, countQuery, args...)
	if err := row.Scan(&count); err != nil {
		return -1, nil, err
	}

	// List the page of transactions from this filter
	listQuery := fmt.Sprintf(`
	SELECT 
		t.id, 
		json_build_object(
			'id', a.id,
			'acc_number', a.acc_number,
			'acc_name', a.acc_name,
			'institution', a.institution,
			'type', a.type
		) AS account,
		t.merchant, 
		t.tran_description, 
		t.category, 
		t.amount, 
		t.created_at, 
		t.updated_at
	FROM transactions t
	JOIN accounts a ON t.account_id = a.id
	WHERE %s
	ORDER BY t.created_at DESC 
	LIMIT 20 
	OFFSET $%d;
	`, condition, len(args)+1)
	args = append(args, offset)
	log.Println(listQuery)

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

// InsertAccount inserts and returns an account of the user
func (r *AccountRepo) InsertAccount(
	ctx context.Context,
	db models.DBTX,
	userId int64,
	body models.PostAccountBody,
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
	row := db.QueryRow(ctx, insertAccountQuery,
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

// UpdateAccount updates and returns the account by ID. It doesn't allow for updating type.
func (r *AccountRepo) UpdateAccount(
	ctx context.Context,
	db models.DBTX,
	id int64,
	body models.PutAccountBody,
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
	row := db.QueryRow(ctx, updateAccountQuery,
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
	ctx context.Context,
	db models.DBTX,
	id int64,
	netChange float64,
) (float64, error) {
	var newBalance float64

	const updateBalanceQuery = `
	UPDATE accounts a
	SET balance = CASE
			WHEN type = 'Debit' THEN balance - $2
			ELSE balance + $2
		END,
		updated_at = NOW()
	WHERE id = $1
	RETURNING balance;
	`
	row := db.QueryRow(ctx, updateBalanceQuery, id, netChange)
	if err := row.Scan(&newBalance); err != nil {
		return newBalance, err
	}

	return newBalance, nil
}

// MoveDueDateNextMonth bulk updates next due of the list of accounts to next month.
// Returns the number of successfully updated accounts.
func (r *AccountRepo) MoveAccountsDueDate(ctx context.Context, db models.DBTX, ids []int64) (int, error) {
	const updateDueDateQuery = `
	UPDATE accounts a
	SET next_due = next_due + INTERVAL '1 month'
	WHERE id = ANY($1)
	RETURNING id;
	`
	rows, err := db.Query(ctx, updateDueDateQuery, ids)
	if err != nil {
		return 0, err
	}
	updatedIDs, err := pgx.CollectRows(rows, pgx.RowToStructByName[int64])
	if err != nil {
		return 0, err
	}

	return len(updatedIDs), nil
}

// FlagAccount will flag the account with the given discrepany amount.
func (r *AccountRepo) FlagAccount(
	ctx context.Context, db models.DBTX, id int64, discrepancyAmount float64,
) (models.Account, error) {
	var flaggedAccount models.Account

	const flagAccountQuery = `
	UPDATE accounts
	SET discrepancy_flagged = TRUE, 
		discrepancy_amount = $2
	WHERE id = $1
	RETURNING *;
	`
	row := db.QueryRow(ctx, flagAccountQuery, id, discrepancyAmount)
	if err := models.ScanAccount(row, &flaggedAccount); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return flaggedAccount, models.GetNotFound[models.Account](id)
		}
		return flaggedAccount, err
	}

	return flaggedAccount, nil
}

// Update the account's data and returns the deleted account
func (r *AccountRepo) DeleteAccount(ctx context.Context, db models.DBTX, id int64) (models.Account, error) {
	var deletedAccount models.Account

	const deleteAccountQuery = `DELETE FROM accounts WHERE id = $1 RETURNING *;`
	row := db.QueryRow(ctx, deleteAccountQuery, id)
	if err := models.ScanAccount(row, &deletedAccount); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return deletedAccount, models.GetNotFound[models.Account](id)
		}
		return deletedAccount, err
	}
	return deletedAccount, nil
}
