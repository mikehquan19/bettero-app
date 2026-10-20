package repositories

import (
	"betterov2/models"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type AccountRepo struct{}

func NewAccountRepo() *AccountRepo {
	return &AccountRepo{}
}

// ListAcccounts returns the list of accounts of given user,
// sorted by updated recency.
func (r *AccountRepo) ListAccounts(ctx context.Context, dbtx DBTX, userId uuid.UUID) ([]models.Account, error) {
	const listAccountSQL = `
	SELECT * FROM accounts WHERE user_id = $1 ORDER BY updated_at DESC;
	`
	rows, err := dbtx.Query(ctx, listAccountSQL, userId)
	if err != nil {
		return nil, err
	}
	accounts, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Account])
	if err != nil {
		return nil, err
	}

	return accounts, nil
}

// Available filter status, we can add more later
type Status string

const (
	PastDue   Status = "PAST_DUE"
	Unflagged Status = "UNFLAGGED"
	All       Status = "All"
)

// ListAccountsByStatus returns the list of accounts matching the given status.
func (r *AccountRepo) ListAccountsByStatus(ctx context.Context, dbtx DBTX, status Status) ([]models.Account, error) {
	var filter string
	switch status {
	case PastDue:
		filter = "WHERE next_due < CURRENT_DATE"
	case Unflagged:
		filter = "WHERE discrepancy_flagged = FALSE"
	case All:
		filter = ""
	}
	filterAccountSQL := fmt.Sprintf("SELECT * FROM accounts %s", filter)
	rows, err := dbtx.Query(ctx, filterAccountSQL)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, pgx.RowToStructByName[models.Account])
}

// GetAccount gets the account with given ID
func (r *AccountRepo) GetAccount(ctx context.Context, dbtx DBTX, userId, id uuid.UUID) (models.Account, error) {
	var account models.Account

	row := dbtx.QueryRow(ctx, "SELECT * FROM accounts WHERE user_id = $1 AND id = $2;", userId, id)
	if err := scanAccount(row, &account); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Account{}, models.ErrNotFound
		}
		return models.Account{}, err
	}

	return account, nil
}

// LockAccount locks and returns an account for a ledger-consistency validation
func (r *AccountRepo) LockAccount(ctx context.Context, dbtx DBTX, id uuid.UUID) (models.Account, error) {
	var account models.Account

	row := dbtx.QueryRow(ctx, "SELECT * FROM accounts WHERE id = $1 FOR UPDATE;", id)
	if err := scanAccount(row, &account); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Account{}, models.ErrNotFound
		}
		return models.Account{}, err
	}

	return account, nil
}

// CountAccountTransactions returns the number of an account's transactions that match the filter.
func (r *AccountRepo) CountAccTransactions(
	ctx context.Context,
	dbtx DBTX,
	userId, id uuid.UUID,
	filter models.TransactionFilter,
) (int, error) {
	var transactionCount int

	condition, args := buildTransactionFilter("a.user_id = $1", userId, filter)
	condition += fmt.Sprintf(" AND t.account_id = $%d", len(args)+1)
	args = append(args, id)

	countTransactionSQL := fmt.Sprintf(`
	SELECT COUNT(*)
	FROM transactions t JOIN accounts a ON t.account_id = a.id
	WHERE %s;`, condition)

	row := dbtx.QueryRow(ctx, countTransactionSQL, args...)
	if err := row.Scan(&transactionCount); err != nil {
		return -1, err
	}

	return transactionCount, nil
}

// ListAccountTransactions returns one page of an account's transactions that match the filter.
func (r *AccountRepo) ListAccTransactions(
	ctx context.Context,
	dbtx DBTX,
	userId, id uuid.UUID,
	filter models.TransactionFilter,
	offset int64,
) ([]models.Transaction, error) {
	condition, args := buildTransactionFilter("a.user_id = $1", userId, filter)
	condition += fmt.Sprintf(" AND t.account_id = $%d", len(args)+1)
	args = append(args, id)

	// List the page of transactions from this filter
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
	FROM transactions t JOIN accounts a ON t.account_id = a.id
	WHERE %s
	ORDER BY t.created_at DESC
	LIMIT 20 OFFSET $%d;`, condition, len(args)+1)
	args = append(args, offset)

	rows, err := dbtx.Query(ctx, listTransactionSQL, args...)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, pgx.RowToStructByName[models.Transaction])
}

// InsertAccount inserts and returns an account of the user
func (r *AccountRepo) InsertAccount(
	ctx context.Context,
	dbtx DBTX,
	userId uuid.UUID,
	body models.PostAccountBody,
) (models.Account, error) {
	var newAccount models.Account

	const insertAccountSQL = `
	INSERT INTO accounts (
		user_id,
		account_number,
		name,
		institution,
		type,
		balance_cents,
		credit_limit_cents,
		next_due
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	RETURNING *;`

	row := dbtx.QueryRow(ctx, insertAccountSQL,
		userId,
		body.AccountNumber,
		body.Name,
		body.Institution,
		body.Type,
		body.BalanceCents,
		body.CreditLimitCents,
		body.NextDue,
	)
	if err := scanAccount(row, &newAccount); err != nil {
		if isForeignKeyViolation(err) {
			return models.Account{}, models.ErrForeignKey
		}
		return models.Account{}, err
	}

	return newAccount, nil
}

// UpdateAccount updates and returns the account by ID. It doesn't allow for updating type.
func (r *AccountRepo) UpdateAccount(
	ctx context.Context,
	dbtx DBTX,
	userId, id uuid.UUID,
	body models.PutAccountBody,
) (models.Account, error) {
	var updatedAccount models.Account

	updateAccountSQL := `
	UPDATE accounts
	SET account_number = $3,
		name = $4,
		institution = $5,
		balance_cents = $6,
		credit_limit_cents = $7,
		next_due = $8,
		updated_at = NOW()
	WHERE user_id = $1 AND id = $2 RETURNING *;`

	row := dbtx.QueryRow(ctx, updateAccountSQL,
		userId,
		id,
		body.AccountNumber,
		body.Name,
		body.Institution,
		body.BalanceCents,
		body.CreditLimitCents,
		body.NextDue,
	)
	if err := scanAccount(row, &updatedAccount); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Account{}, models.ErrNotFound
		}
		return models.Account{}, err
	}

	return updatedAccount, nil
}

// UpdateAccountBalance updates the balance of the account by the net change.
//
//   - Debit card's balance will decrease
//   - Credit card's balance will increase
func (r *AccountRepo) UpdateAccountBalance(
	ctx context.Context,
	dbtx DBTX,
	id uuid.UUID,
	netChange models.Money,
) (models.Money, error) {
	var newBalance models.Money

	const updateBalanceSQL = `
	UPDATE accounts a
	SET balance_cents = CASE
			WHEN type = 'Debit' THEN balance_cents - $2
			ELSE balance_cents + $2
		END,
		updated_at = NOW()
	WHERE id = $1
	RETURNING balance_cents;`

	row := dbtx.QueryRow(ctx, updateBalanceSQL, id, netChange)
	if err := row.Scan(&newBalance); err != nil {
		return 0, err
	}

	return newBalance, nil
}

// MoveDueDateNextMonth bulk updates next due of the list of accounts to next month.
// Returns the number of successfully updated accounts.
func (r *AccountRepo) MoveAccountsDueDate(ctx context.Context, dbtx DBTX, ids []uuid.UUID) (int, error) {
	const updateDueDateSQL = `
	UPDATE accounts a
	SET next_due = next_due + INTERVAL '1 month'
	WHERE id = ANY($1)
	RETURNING id;`

	rows, err := dbtx.Query(ctx, updateDueDateSQL, ids)
	if err != nil {
		return 0, err
	}
	updatedIDs, err := pgx.CollectRows(rows, pgx.RowTo[uuid.UUID])
	if err != nil {
		return 0, err
	}

	return len(updatedIDs), nil
}

// SetAccountDiscrepancy stores the signed discrepancy and updates the flag.
func (r *AccountRepo) SetAccountDiscrepancy(
	ctx context.Context,
	dbtx DBTX,
	id uuid.UUID,
	discrepancyAmount models.Money,
) (models.Account, error) {
	var updatedAccount models.Account

	const setDiscrepancySQL = `
	UPDATE accounts
	SET discrepancy_flagged = ($2 <> 0),
		discrepancy_amount_cents = $2
	WHERE id = $1
	RETURNING *;`

	row := dbtx.QueryRow(ctx, setDiscrepancySQL, id, discrepancyAmount)
	if err := scanAccount(row, &updatedAccount); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Account{}, models.ErrNotFound
		}
		return models.Account{}, err
	}

	return updatedAccount, nil
}

// Update the account's data and returns the deleted account
func (r *AccountRepo) DeleteAccount(ctx context.Context, dbtx DBTX, userId, id uuid.UUID) (models.Account, error) {
	var deletedAccount models.Account

	const deleteAccountSQL = `
	DELETE FROM accounts WHERE user_id = $1 AND id = $2 RETURNING *;
	`
	row := dbtx.QueryRow(ctx, deleteAccountSQL, userId, id)
	if err := scanAccount(row, &deletedAccount); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Account{}, models.ErrNotFound
		}
		return models.Account{}, err
	}
	return deletedAccount, nil
}

// scanAccount parses the returned db row into account struct and destinations
func scanAccount(accRow pgx.Row, acc *models.Account) error {
	return accRow.Scan(
		&acc.ID,
		&acc.UserID,
		&acc.AccountNumber,
		&acc.Name,
		&acc.Institution,
		&acc.Type,
		&acc.BalanceCents,
		&acc.CreditLimitCents,
		&acc.NextDue,
		&acc.DiscrepancyFlagged,
		&acc.DiscrepancyAmountCents,
		&acc.CreatedAt,
		&acc.UpdatedAt,
	)
}
