package repositories

import (
	"betterov2/models"
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type SummaryRepo struct{}

func NewSummaryRepo() *SummaryRepo {
	return &SummaryRepo{}
}

// GetBasicAnalysis returns the basic aggregation of the user's account between 2 dates
func (s *SummaryRepo) GetBasicAnalysis(
	ctx context.Context,
	db models.DBTX,
	userId int64,
	start, end time.Time,
) (models.BasicAnalysis, error) {
	var analysis models.BasicAnalysis

	getAnalysisQuery := `
	WITH total_balance AS (
		SELECT COALESCE(SUM(a.balance), 0) AS total_balance 
		FROM accounts a 
		WHERE a.user_id = $1 and a.type = 'Debit'
	),
	total_amount_due AS (
		SELECT COALESCE(SUM(a.balance), 0) AS total_amount_due 
		FROM accounts a 
		WHERE a.user_id = $1 and a.type = 'Credit'
	),
	total_income AS (
		SELECT COALESCE(SUM(t.amount), 0) AS total_income 
		FROM transactions t
		JOIN accounts a ON t.account_id = a.id
		WHERE
			a.user_id = $1 AND
			t.category = 'Income' AND 
			t.created_at >= $2 AND t.created_at < $3
	),
	total_expense AS (
		SELECT COALESCE(SUM(t.amount), 0) AS total_expense
		FROM transactions t
		JOIN accounts a ON t.account_id = a.id
		WHERE
			a.user_id = $1 AND
			t.category <> 'Income' AND 
			t.created_at >= $2 AND t.created_at < $3
	)
	SELECT 
		b.total_balance, 
		a.total_amount_due, 
		i.total_income, 
		e.total_expense
	FROM 
		total_balance b, 
		total_amount_due a, 
		total_income i, 
		total_expense e;`
	row := db.QueryRow(ctx, getAnalysisQuery, userId, start, end)
	if err := models.ScanAnalysis(row, &analysis); err != nil {
		return analysis, err
	}

	return analysis, nil
}

// GetDateToAmount returns the map from date to total expense of the user or account
func (s *SummaryRepo) GetDateToAmount(
	ctx context.Context,
	db models.DBTX,
	objType models.ObjectType,
	objId int64,
	start, end time.Time,
) (map[string]float64, error) {
	var dateToAmount = make(map[string]float64)

	for date := start; !date.After(end); date = date.AddDate(0, 0, 1) {
		dateToAmount[date.Format("2006-01-02")] = 0.0
	}

	var table, filter string
	switch objType {
	case models.UserObj:
		// For user, join accounts table to get user ID
		table = "transactions t JOIN accounts a ON t.account_id = a.id"
		filter = "a.user_id = $1"
	case models.AccountObj:
		// Otherwise, account's ID is already in the transactions table
		table = "transactions t"
		filter = "t.account_id = $1"
	}

	getDateToAmtQuery := fmt.Sprintf(`
	SELECT
		t.created_at::date AS date, 
		SUM(t.amount) AS amount
	FROM %s
	WHERE
		%s AND
		t.category <> 'Income' AND 
		t.created_at >= $2 AND t.created_at < $3
	GROUP BY date;`, table, filter)

	rows, err := db.Query(ctx, getDateToAmtQuery, objId, start, end)
	if err != nil {
		return dateToAmount, err
	}

	var date time.Time
	var amount float64
	_, err = pgx.ForEachRow(rows, []any{&date, &amount}, func() error {
		dateToAmount[date.Format("2006-01-02")] = amount
		return nil
	})

	return dateToAmount, err
}

// getCategoryToAmount returns the map from category to total expense of the obj
func (s *SummaryRepo) GetCategoryToAmount(
	ctx context.Context,
	db models.DBTX,
	objType models.ObjectType,
	objId int64,
	start, end time.Time,
) (map[string]float64, error) {
	var categoryToAmount = make(map[string]float64)

	// There are 10 categories
	for _, category := range []string{
		"Housing", "Automobile", "Medical", "Subscription", "Grocery", "Dining", "Shopping", "Gas", "Others",
	} {
		categoryToAmount[category] = 0.0
	}

	var table, filter string
	switch objType {
	case models.UserObj:
		// For user, join accounts table to get user ID
		table = "transactions t JOIN accounts a ON t.account_id = a.id"
		filter = "a.user_id = $1"
	case models.AccountObj:
		table = "transactions t"
		filter = "t.account_id = $1"
	}

	getCatToAmtQuery := fmt.Sprintf(`
	SELECT
		t.category, 
		SUM(t.amount) AS amount
	FROM %s
	WHERE
		%s AND
		t.category <> 'Income' AND 
		t.created_at >= $2 AND t.created_at < $3
	GROUP BY t.category;`, table, filter)

	rows, err := db.Query(ctx, getCatToAmtQuery, objId, start, end)
	if err != nil {
		return categoryToAmount, err
	}

	var category string
	var amount float64
	_, err = pgx.ForEachRow(rows, []any{&category, &amount}, func() error {
		categoryToAmount[category] = amount
		return nil
	})

	return categoryToAmount, err
}
