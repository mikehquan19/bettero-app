package repositories

import (
	"betterov2/models"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

/*
Account History Repository
*/

type AccountHistoryRepo struct{}

func NewAccountHistoryRepo() *AccountHistoryRepo {
	return &AccountHistoryRepo{}
}

// ListHistories returns list of balance history of the account over time
func (r *AccountHistoryRepo) ListHistories(ctx context.Context, db *pgxpool.Pool, accountId int64) ([]models.AccountHistory, error) {
	var histories []models.AccountHistory

	const listHistoryQuery = `
SELECT * FROM account_histories
WHERE account_id = $1 ORDER BY logged_time ASC;
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
func (r *AccountHistoryRepo) GetLatestHistory(ctx context.Context, tx pgx.Tx, accountId int64) (models.AccountHistory, error) {
	var latestHistory models.AccountHistory

	const getLatestQuery = `
SELECT * FROM account_histories
WHERE account_id = $1 
ORDER BY logged_time DESC LIMIT 1;
	`
	row := tx.QueryRow(ctx, getLatestQuery, accountId)
	if err := models.ScanAccHistory(row, &latestHistory); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Tell the system that there is no account
			return latestHistory, models.GetNotFound[models.Account](accountId)
		}
		return latestHistory, err
	}

	return latestHistory, nil
}

// InsertAccountHistory the account history to database
func (r *AccountHistoryRepo) InsertHistory(ctx context.Context, tx pgx.Tx, body models.PostAccHistBody) (models.AccountHistory, error) {
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
	row := tx.QueryRow(ctx, insertHistoryQuery,
		body.AccountId,
		body.LoggedTime,
		body.Balance,
	)
	if err := models.ScanAccHistory(row, &newAccHistory); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23503" {
			// Insert history that references non-existent account
			accountId := body.AccountId
			return newAccHistory, models.GetForeignKey[models.Account](accountId)
		}
		return newAccHistory, err
	}

	return newAccHistory, nil
}

// Delete the account history from database
func (r *AccountHistoryRepo) DeleteHistory(ctx context.Context, tx pgx.Tx, id int64) (models.AccountHistory, error) {
	var deletedHist models.AccountHistory

	const deleteHistQuery = `
DELETE FROM account_histories WHERE id = $1 RETURNING *;
	`
	row := tx.QueryRow(ctx, deleteHistQuery, id)
	if err := models.ScanAccHistory(row, &deletedHist); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return deletedHist, models.GetNotFound[models.AccountHistory](id)
		}
		return deletedHist, err
	}

	return deletedHist, nil
}
