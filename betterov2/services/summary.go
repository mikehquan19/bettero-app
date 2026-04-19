package services

import (
	"betterov2/models"
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

// GetBasicAnalysis returns the basic aggregation of the user's account between 2 dates
func GetBasicAnalysis(
	ctx context.Context, userId int64, start time.Time, end time.Time,
) (models.BasicAnalysis, error) {
	var analysis models.BasicAnalysis

	getAnalysisQuery := `
	WITH total_balance AS (
		SELECT SUM(a.balance) AS total_balance 
		FROM accounts a 
		WHERE a.user_id = $1 and a.type = 'Debit'
	),
	total_amount_due AS (
		SELECT SUM(a.balance) AS total_amount_due 
		FROM accounts a 
		WHERE a.user_id = $1 and a.type = 'Credit'
	),
	total_income AS (
		SELECT SUM(t.amount) AS total_income 
		FROM transactions t
		JOIN accounts a ON t.account_id = a.id
		WHERE
			a.user_id = $1 AND
			t.category = 'Income' AND 
			t.created_at >= $2 AND t.created_at < $3
	),
	total_expense AS (
		SELECT SUM(t.amount) AS total_expense
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
	FROM total_balance b, total_amount_due a, total_income i, total_expense e
	`
	analysisRow := database.QueryRow(ctx, getAnalysisQuery, userId, start, end)
	if err := models.ScanAnalysis(analysisRow, &analysis); err != nil {
		return analysis, err
	}

	return analysis, nil
}

// GetCompositionMap returns the map from category of expense to its percentage vs total expense.
// If the total expense is 0, then each category takes up 0% of the total expense.
// It uses the total expense computed by GetBasicAnalysis for efficiency.
func GetCompositionMap(
	ctx context.Context, userId int64, start time.Time, end time.Time,
) (map[string]float64, error) {
	var compositionMap = make(map[string]float64)

	categoryToAmount, err := GetCategoryToAmount(ctx, userId, start, end)
	if err != nil {
		return compositionMap, err
	}

	// Constant time since the number of categories is fixed
	var totalExpense float64 = 0.0
	for _, amount := range categoryToAmount {
		totalExpense += amount
	}

	for category, amount := range categoryToAmount {
		if totalExpense != 0.0 {
			percent := round(amount * 100 / totalExpense)
			compositionMap[category] = round(percent)
		} else {
			compositionMap[category] = 0.0
		}
	}

	return compositionMap, err
}

// getPrevInterval gets the previous period of the 2 dates
func getPrevInterval(start time.Time, end time.Time) (time.Time, time.Time) {
	if start.Day() == 1 && start.AddDate(0, 1, 0).Equal(end) {
		return start.AddDate(0, -1, 0), end.AddDate(0, -1, 0)
	}
	d := end.Sub(start)
	return start.Add(-d), end.Add(-d)
}

// GetChangeMap returns the map from category to the expense change percentage
// previous period -> current period.
// If the previous expense is 0 then it's null.
func GetChangeMap(
	ctx context.Context, userId int64, start time.Time, end time.Time,
) (map[string]*float64, error) {
	var changeMap = make(map[string]*float64)

	current, err := GetCategoryToAmount(ctx, userId, start, end)
	if err != nil {
		return changeMap, err
	}

	prevStart, prevEnd := getPrevInterval(start, end)
	previous, err := GetCategoryToAmount(ctx, userId, prevStart, prevEnd)
	if err != nil {
		return changeMap, err
	}

	for category, amount := range previous {
		if amount != 0 {
			percent := round((current[category] - amount) * 100 / amount)
			changeMap[category] = &percent
		} else {
			changeMap[category] = nil
		}
	}

	return changeMap, err
}

// GetDateToAmount returns the map from date to total expense
func GetDateToAmount(
	ctx context.Context, userId int64, start time.Time, end time.Time,
) (map[string]float64, error) {
	var dateToAmount = make(map[string]float64)

	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dateToAmount[d.Format("2006-01-02")] = 0.0
	}

	getDateToAmtQuery := `
	SELECT
		t.created_at::date AS date, 
		SUM(t.amount) AS amount
	FROM transactions t 
	JOIN accounts a ON t.account_id = a.id
	WHERE
		a.user_id = $1 AND
		t.category <> 'Income' AND 
		t.created_at >= $2 AND t.created_at < $3
	GROUP BY date;
	`
	groupByDateRows, err := database.Query(ctx, getDateToAmtQuery, userId, start, end)
	if err != nil {
		return dateToAmount, err
	}

	var (
		date   time.Time
		amount float64
	)
	pgx.ForEachRow(groupByDateRows, []any{&date, &amount}, func() error {
		dateToAmount[date.Format("2006-01-02")] = amount
		return nil
	})

	return dateToAmount, err
}

// GetCategoryToAmount returns the map from categoryto total expense
func GetCategoryToAmount(
	ctx context.Context, userId int64, start time.Time, end time.Time,
) (map[string]float64, error) {
	var categoryToAmount = make(map[string]float64)

	// Since there are only 10 categories, this operates in constant time
	for _, category := range []string{
		"Housing", "Automobile", "Medical", "Subscription",
		"Grocery", "Dining", "Shopping", "Gas", "Others",
	} {
		categoryToAmount[category] = 0.0
	}

	getCatToAmtQuery := `
	SELECT
		t.category, 
		SUM(t.amount) AS amount
	FROM transactions t 
	JOIN accounts a ON t.account_id = a.id
	WHERE
		a.user_id = $1 AND
		t.category <> 'Income' AND 
		t.created_at >= $2 AND t.created_at < $3
	GROUP BY t.category;
	`
	groupByCategoryRows, err := database.Query(ctx, getCatToAmtQuery, userId, start, end)
	if err != nil {
		return categoryToAmount, err
	}

	var (
		category string
		amount   float64
	)
	pgx.ForEachRow(groupByCategoryRows, []any{&category, &amount}, func() error {
		categoryToAmount[category] = amount
		return nil
	})

	return categoryToAmount, err
}
