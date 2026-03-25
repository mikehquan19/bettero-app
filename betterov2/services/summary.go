package services

import (
	"betterov2/models"
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

// GetBasicAnalysis
func GetBasicAnalysis(
	ctx context.Context, userId int64, startDate time.Time, endDate time.Time,
) (models.BasicAnalysis, error) {
	var analysis models.BasicAnalysis

	analysisRow := database.QueryRow(ctx, `
	WITH total_balance AS (
		SELECT SUM(a.balance) AS total_balance 
		FROM accounts a 
		WHERE a.user_id = $1
	),
	total_amount_due AS (
		SELECT SUM(a.balance) AS total_amount_due 
		FROM accounts a 
		WHERE a.user_id = $1
	),
	total_income AS (
		SELECT SUM(t.amount) AS total_income
		FROM transactions t
		JOIN accounts a ON t.account_id = a.id
		WHERE
			a.user_id = $1 AND
			t.category <> 'Income' AND 
			t.created_at BETWEEN $2 AND $3
	),
	total_expense AS (
		SELECT SUM(t.amount) AS total_expense
		FROM transactions t
		JOIN accounts a ON t.account_id = a.id
		WHERE
			a.user_id = $1 AND
			t.category = 'Income' AND 
			t.created_at BETWEEN $2 AND $3
	)
	SELECT
		b.total_balance, a.total_amount_due, i.total_income, e.total_expense
	FROM total_balance b, total_amount_due a, total_income i, total_expense e
	`,
		userId, startDate, endDate,
	)
	if err := models.ScanAnalysis(analysisRow, &analysis); err != nil {
		return analysis, err
	}

	return analysis, nil
}

// GetDateToAmount returns the map from date to total expense
func GetDateToAmount(
	ctx context.Context, userId int64, startDate time.Time, endDate time.Time,
) (map[string]float64, error) {
	dateToAmount := make(map[string]float64)
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateToAmount[d.Format("2006-01-02")] = 0.0
	}

	groupByDateRows, err := database.Query(ctx, `
	SELECT
		t.created_at::date AS date,
		SUM(t.amount) AS total_amount
	FROM transactions t 
	JOIN accounts a ON t.account_id = a.id
	WHERE
		a.user_id = $1 AND
		t.category <> 'Income' AND t.created_at BETWEEN $2 AND $3
	GROUP BY date;
	`,
		userId, startDate, endDate,
	)
	if err != nil {
		return dateToAmount, err
	}

	var (
		createdDate time.Time
		totalAmount float64
	)
	pgx.ForEachRow(groupByDateRows, []any{&createdDate, &totalAmount}, func() error {
		dateToAmount[createdDate.Format("2006-01-02")] = totalAmount
		return nil
	})

	return dateToAmount, err
}

// GetCategoryToAmount returns the map from categoryto total expense
func GetCategoryToAmount(
	ctx context.Context, userId int64, startDate time.Time, endDate time.Time,
) (map[string]float64, error) {

	categoryToAmount := make(map[string]float64)
	for _, category := range []string{
		"Income", "Housing", "Automobile", "Medical", "Subscription",
		"Grocery", "Dining", "Shopping", "Gas", "Others",
	} {
		categoryToAmount[category] = 0.0
	}

	groupByCategoryRows, err := database.Query(ctx, `
	SELECT
		t.category,
		SUM(t.amount) AS total_amount
	FROM transactions t 
	JOIN accounts a ON t.account_id = a.id
	WHERE 
		a.user_id = $1 AND
		t.category <> 'Income' AND (t.created_at BETWEEN $2 AND $3)
	GROUP BY t.category;
	`,
		userId, startDate, endDate,
	)
	if err != nil {
		return categoryToAmount, err
	}

	var (
		category    string
		totalAmount float64
	)
	pgx.ForEachRow(groupByCategoryRows, []any{&category, &totalAmount}, func() error {
		categoryToAmount[category] = totalAmount
		return nil
	})
	return categoryToAmount, err
}
