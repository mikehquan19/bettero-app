package repositories

import (
	"betterov2/models"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type AccountHistoryRepo struct{}

func NewAccountHistoryRepo() *AccountHistoryRepo {
	return &AccountHistoryRepo{}
}

// ListHistories returns list of balance history of the account over time
func (r *AccountHistoryRepo) ListHistories(ctx context.Context, db models.DBTX, accountId int64) ([]models.AccountHistory, error) {
	var histories []models.AccountHistory

	const listHistoryQuery = `
	SELECT * FROM account_histories
	WHERE account_id = $1 
	ORDER BY logged_time ASC;
	`
	rows, err := db.Query(ctx, listHistoryQuery, accountId)
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
func (r *AccountHistoryRepo) GetLatestHistory(ctx context.Context, db models.DBTX, accountId int64) (models.AccountHistory, error) {
	var latestHistory models.AccountHistory

	const getLatestHistoryQuery = `
	SELECT * FROM account_histories
	WHERE account_id = $1
	ORDER BY logged_time DESC 
	LIMIT 1;
	`
	row := db.QueryRow(ctx, getLatestHistoryQuery, accountId)
	if err := models.ScanAccHistory(row, &latestHistory); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// No history found for the account
			return latestHistory, models.ErrNotFound
		}
		return latestHistory, err
	}

	return latestHistory, nil
}

// InsertHistory inserts the account history to database
func (r *AccountHistoryRepo) InsertHistory(ctx context.Context, db models.DBTX, body models.PostAccHistBody) (models.AccountHistory, error) {
	var newAccHistory models.AccountHistory

	// Don't need to include time because it's auto now
	const insertHistoryQuery = `
	INSERT INTO account_histories (
		account_id,
		logged_time,
		balance
	)
	VALUES ($1, $2, $3)
	RETURNING *;
	`
	row := db.QueryRow(ctx, insertHistoryQuery,
		body.AccountId,
		body.LoggedTime,
		body.Balance,
	)
	if err := models.ScanAccHistory(row, &newAccHistory); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23503" {
			// Insert history that references non-existent account
			return newAccHistory, models.ErrForeignKey
		}
		return newAccHistory, err
	}

	return newAccHistory, nil
}

// DeleteOutdatedHistories deletes the outdated balance history of account.
// Returns the number of successfully deleted history.
func (r *AccountHistoryRepo) DeleteOutdatedHistories(ctx context.Context, db models.DBTX, accountId int64) (int, error) {
	const deleteHistQuery = `
	DELETE FROM account_histories 
	WHERE
		logged_time < CURRENT_DATE - INTERVAL '6 months'
		AND account_id = $1
	RETURNING ids;
	`
	rows, err := db.Query(ctx, deleteHistQuery, accountId)
	if err != nil {
		return -1, err
	}
	deleted, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[int64])
	if err != nil {
		return -1, err
	}

	return len(deleted), nil
}
