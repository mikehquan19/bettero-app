package repositories

import (
	"betterov2/models"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type AccountHistoryRepo struct{}

func NewAccountHistoryRepo() *AccountHistoryRepo {
	return &AccountHistoryRepo{}
}

// ListHistories returns list of balance history of the account over time
func (r *AccountHistoryRepo) ListHistories(ctx context.Context, dbtx DBTX, accountId int64) ([]models.AccountHistory, error) {
	var histories []models.AccountHistory

	const listHistorySQL = `
	SELECT * FROM account_histories
	WHERE account_id = $1 
	ORDER BY logged_time ASC;`

	rows, err := dbtx.Query(ctx, listHistorySQL, accountId)
	if err != nil {
		return nil, err
	}
	histories, err = pgx.CollectRows(rows, pgx.RowToStructByName[models.AccountHistory])
	if err != nil {
		return nil, err
	}

	return histories, nil
}

// GetLatest gets the most recent account history, used for validating primarily
func (r *AccountHistoryRepo) GetLatestHistory(ctx context.Context, dbtx DBTX, accountId int64) (models.AccountHistory, error) {
	var latestHistory models.AccountHistory

	const getLatestHistorySQL = `
	SELECT * FROM account_histories
	WHERE account_id = $1
	ORDER BY logged_time DESC 
	LIMIT 1;`

	row := dbtx.QueryRow(ctx, getLatestHistorySQL, accountId)
	if err := scanAccHistory(row, &latestHistory); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.AccountHistory{}, models.ErrNotFound
		}
		return models.AccountHistory{}, err
	}

	return latestHistory, nil
}

// InsertHistory inserts the account history to database
func (r *AccountHistoryRepo) InsertHistory(ctx context.Context, dbtx DBTX, body models.PostAccHistBody) (models.AccountHistory, error) {
	var newAccountHistory models.AccountHistory

	// Don't need to include time because it's auto now
	const insertHistorySQL = `
	INSERT INTO account_histories (
		account_id,
		logged_time,
		balance
	)
	VALUES ($1, $2, $3)
	RETURNING *;`

	row := dbtx.QueryRow(ctx, insertHistorySQL,
		body.AccountId,
		body.LoggedTime,
		body.Balance,
	)
	if err := scanAccHistory(row, &newAccountHistory); err != nil {
		if isForeignKeyViolation(err) {
			return models.AccountHistory{}, models.ErrForeignKey
		}
		return models.AccountHistory{}, err
	}

	return newAccountHistory, nil
}

// DeleteOutdatedHistories deletes the outdated balance history of account.
// Returns the number of successfully deleted history.
func (r *AccountHistoryRepo) DeleteOutdatedHistories(ctx context.Context, dbtx DBTX, accountId int64) (int, error) {
	const deleteHistSQL = `
	DELETE FROM account_histories 
	WHERE
		logged_time < CURRENT_DATE - INTERVAL '6 months'
		AND account_id = $1
	RETURNING id;`

	rows, err := dbtx.Query(ctx, deleteHistSQL, accountId)
	if err != nil {
		return -1, err
	}
	deleted, err := pgx.CollectRows(rows, pgx.RowTo[int64])
	if err != nil {
		return -1, err
	}

	return len(deleted), nil
}

// ScanAccHistory parses the returned db row into account history struct and destinations
func scanAccHistory(accHistRow pgx.Row, accHist *models.AccountHistory) error {
	return accHistRow.Scan(
		&accHist.ID,
		&accHist.AccountId,
		&accHist.LoggedTime,
		&accHist.Balance,
	)
}
