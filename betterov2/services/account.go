package services

import (
	"betterov2/models"
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// ListAccounts returns the list of accounts of the user
func ListAccounts(ctx context.Context, userId int64) ([]models.Account, error) {
	var accounts []models.Account

	getAccQuery := `
	SELECT * FROM accounts WHERE user_id = $1 ORDER BY updated_at DESC;
	`
	rows, err := database.Query(ctx, getAccQuery, userId)
	if err != nil {
		return accounts, err
	}
	accounts, err = pgx.CollectRows(rows, pgx.RowToStructByName[models.Account])
	if err != nil {
		return accounts, err
	}

	return accounts, err
}

// GetAccount returns details of account
// If the account doesn't exist, it returns a custom not found error
func GetAccount(ctx context.Context, id int64) (models.Account, error) {
	var account models.Account

	accountRow := database.QueryRow(ctx, `SELECT * FROM accounts WHERE id = $1;`, id)
	if err := models.ScanAccount(accountRow, &account); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return account, fmt.Errorf("account %d not found, %w", id, models.ErrNotFound)
		}
		return account, err
	}

	return account, nil
}

// CreateAccount inserts new account.
// If the account references a non-existent user, it returns a custom not found error.
func CreateAccount(
	ctx context.Context, userId int64, body models.PostAccountBody,
) (models.Account, error) {
	var newAccount models.Account

	if err := body.Validate(); err != nil {
		return newAccount, err
	}

	insertAccQuery := `
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
	newAccRow := database.QueryRow(ctx, insertAccQuery,
		userId,
		body.AccNumber,
		body.AccName,
		body.Institution,
		body.Type,
		body.Balance,
		body.CreditLimit,
		body.NextDue,
	)
	if err := models.ScanAccount(newAccRow, &newAccount); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23503" {
			// Insert account that references non-existent user
			fkErr := fmt.Errorf("user %d not found, %w", userId, models.ErrForeignKey)
			return newAccount, fkErr
		}
		return newAccount, err
	}

	return newAccount, nil
}

// UpdateAccount modifies the acount's details
// If balance changes, it insert a new transaction for this account that reflects the balance change.
//
// For credit card,
//   - If balance increases, transaction is considered an Others-type expense.
//   - If balance decreases, transaction is considered an income.
//
// For debit card,
//   - If balance increases, transaction is considered an income.
//   - If balance decreases, transaction is considered an Others-type expense.
//
// NOTE: The feature is to keep ledging consistency, but user won't likely update an account's balance.
// Most of the time, they will create a descriptive transaction.
func UpdateAccount(ctx context.Context, id int64, body models.PutAccountBody) (models.Account, error) {
	var updatedAccount models.Account

	// Wrap a transaction around multiple sql queries
	pgTran, err := database.Begin(ctx)
	if err != nil {
		return updatedAccount, err
	}
	defer pgTran.Rollback(ctx)

	// Store the type and old balance of the account
	var (
		accType     string
		prevBalance float64
	)
	getPrevAccQuery := `SELECT type, balance from accounts WHERE id = $1;`
	prevRow := pgTran.QueryRow(ctx, getPrevAccQuery, id)
	if err = prevRow.Scan(&accType, &prevBalance); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return updatedAccount, fmt.Errorf("account %d not found, %w", id, models.ErrNotFound)
		}
		return updatedAccount, err
	}
	if err = body.Validate(accType); err != nil {
		return updatedAccount, err
	}

	// Actually update the account
	updateAccQuery := `
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
	updatedAccRow := pgTran.QueryRow(ctx, updateAccQuery,
		id,
		body.AccNumber,
		body.AccName,
		body.Institution,
		body.Balance,
		body.CreditLimit,
		body.NextDue,
	)
	if err = models.ScanAccount(updatedAccRow, &updatedAccount); err != nil {
		return updatedAccount, err
	}

	// Create the transaction reflecting the balance change, if any
	change := updatedAccount.Balance - prevBalance
	if change != 0 {
		var description, category string
		if change > 0 {
			description = fmt.Sprintf("Balance increases %f", round(math.Abs(change)))
			category = ternary(updatedAccount.Type == "Credit", "Others", "Income")
		} else {
			description = fmt.Sprintf("Balance descreases %f", round(math.Abs(change)))
			category = ternary(updatedAccount.Type == "Credit", "Income", "Others")
		}

		insertReconcileTranQuery := `
		INSERT INTO transactions (
			account_id, 
			merchant, 
			tran_description, 
			category, 
			amount
		)
		VALUES ($1, $2, $3, $4, $5);
		`
		cmdTag, err := pgTran.Exec(ctx, insertReconcileTranQuery,
			id,
			updatedAccount.Institution,
			description,
			category,
			math.Abs(change),
		)
		if err != nil {
			return updatedAccount, err
		}
		if cmdTag.RowsAffected() == 0 {
			// Account is found but for some reason can't be updated
			return updatedAccount, fmt.Errorf("Error updating account %d", id)
		}
	}

	if err = pgTran.Commit(ctx); err != nil {
		return updatedAccount, err
	}

	return updatedAccount, nil
}

// DeleteAccount deletes the account and all of its transactions.
// It returns a custom not found error if account doesn't exist
func DeleteAccount(ctx context.Context, id int64) error {
	cmdTag, err := database.Exec(ctx, `DELETE FROM accounts WHERE id = $1;`, id)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("account %d not found, %w", id, models.ErrNotFound)
	}

	return err
}

// ListAccountTransactions returns the list of transactions of account
func ListAccountTransactions(ctx context.Context, id int64, offset int64) ([]models.Transaction, error) {
	var transactions []models.Transaction

	listTranQuery := `
	SELECT 
		t.id, t.merchant, t.tran_description, t.category, t.amount, 
		t.created_at, t.updated_at
		json_build_object(
			'id', a.id,
			'acc_number', a.acc_number,
			'acc_name', a.acc_name,
			'institution', a.institution,
			'type', a.type
		) AS account
	FROM transactions t
	JOIN accounts a ON t.account_id = a.id
	WHERE a.id = $1
	ORDER BY t.created_at DESC
	LIMIT 20 
	OFFSET $2;
	`
	rows, err := database.Query(ctx, listTranQuery, id, offset)
	if err != nil {
		return transactions, err
	}

	transactions, err = pgx.CollectRows(rows, pgx.RowToStructByName[models.Transaction])
	if err != nil {
		return transactions, err
	}

	return transactions, err
}

// updateAccountBalanceWithTx updates the balance by the net change
//
//   - Debit card's balance will decrease
//   - Credit card's balance will increase
//
// It uses transaction to execute SQL with other queries inside transaction.
func updateAccountBalanceWithTx(tx pgx.Tx, ctx context.Context, id int64, netChange float64) error {
	updateAccBalQuery := `
	UPDATE accounts a
	SET balance = CASE
			WHEN type = 'Debit' THEN balance - $2
			ELSE balance + $2
		END,
		updated_at = NOW()
	WHERE id = $1;
	`
	cmdTag, err := tx.Exec(ctx, updateAccBalQuery, id, netChange)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("Error updating account %d", id)
	}
	return err
}
