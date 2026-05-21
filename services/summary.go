package services

import (
	"betterov2/models"
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SummaryService
type SummaryService struct {
	database *pgxpool.Pool
}

// Generate a new summary service
func NewSummaryService(database *pgxpool.Pool) *SummaryService {
	return &SummaryService{
		database: database,
	}
}

// GetBasicAnalysis returns the basic aggregation of the user's account between 2 dates
func (s *SummaryService) GetBasicAnalysis(
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
	analysisRow := s.database.QueryRow(ctx, getAnalysisQuery, userId, start, end)
	if err := models.ScanAnalysis(analysisRow, &analysis); err != nil {
		return analysis, err
	}

	return analysis, nil
}

// GetCompositionMap returns the map from category of expense to its percentage vs total expense.
// If the total expense is 0, then each category takes up 0% of the total expense.
// It uses the total expense computed by GetBasicAnalysis for efficiency.
func (s *SummaryService) GetCompositionMap(
	ctx context.Context, userId int64, start time.Time, end time.Time,
) (map[string]float64, error) {
	var compositionMap = make(map[string]float64)

	categoryToAmount, err := getCategoryToAmount(ctx, s.database, userId, start, end)
	if err != nil {
		return compositionMap, err
	}

	// Constant time since the number of categories is fixed
	var totalExpense = 0.0
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

// GetChangeMap returns the map from category to the expense change percentage
// from previous period to current period.
// If the previous expense is 0 then it's null.
func (s *SummaryService) GetChangeMap(
	ctx context.Context, userId int64, start time.Time, end time.Time,
) (map[string]*float64, error) {
	var changeMap = make(map[string]*float64)

	current, err := getCategoryToAmount(ctx, s.database, userId, start, end)
	if err != nil {
		return changeMap, err
	}

	prevStart, prevEnd := getPrevInterval(start, end)
	previous, err := getCategoryToAmount(ctx, s.database, userId, prevStart, prevEnd)
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
func (s *SummaryService) GetDateToAmount(
	ctx context.Context, userId int64, start time.Time, end time.Time,
) (map[string]float64, error) {
	var dateToAmount = make(map[string]float64)

	for date := start; !date.After(end); date = date.AddDate(0, 0, 1) {
		dateToAmount[date.Format("2006-01-02")] = 0.0
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
	groupByDateRows, err := s.database.Query(ctx, getDateToAmtQuery, userId, start, end)
	if err != nil {
		return dateToAmount, err
	}

	var date time.Time
	var amount float64
	cmdTag, err := pgx.ForEachRow(
		groupByDateRows,
		[]any{&date, &amount},
		func() error {
			dateToAmount[date.Format("2006-01-02")] = amount
			return nil
		},
	)
	if cmdTag.RowsAffected() == 0 {
		// There is something wrong with the aggregation
		return dateToAmount, fmt.Errorf("error fetching date to amount")
	}

	return dateToAmount, err
}

// getPrevInterval gets the previous period of the 2 dates
func getPrevInterval(start time.Time, end time.Time) (time.Time, time.Time) {
	if start.Day() == 1 && start.AddDate(0, 1, 0).Equal(end) {
		return start.AddDate(0, -1, 0), end.AddDate(0, -1, 0)
	}
	dur := end.Sub(start)
	return start.Add(-dur), end.Add(-dur)
}

// Helper function: getCategoryToAmount returns the map from categoryto total expense
func getCategoryToAmount(
	ctx context.Context, db *pgxpool.Pool, userId int64, start time.Time, end time.Time,
) (map[string]float64, error) {
	var categoryToAmount = make(map[string]float64)

	// Since there are only 10 categories, this operates in constant time
	for _, category := range []string{
		"Housing", "Automobile", "Medical", "Subscription", "Grocery", "Dining", "Shopping", "Gas", "Others",
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
	groupByCategoryRows, err := db.Query(ctx, getCatToAmtQuery, userId, start, end)
	if err != nil {
		return categoryToAmount, err
	}

	var category string
	var amount float64
	cmdTag, err := pgx.ForEachRow(
		groupByCategoryRows,
		[]any{&category, &amount},
		func() error {
			categoryToAmount[category] = amount
			return nil
		},
	)
	if cmdTag.RowsAffected() == 0 {
		// There is something wrong with the aggregation (dangerous)
		return categoryToAmount, fmt.Errorf("error fetching category to amount")
	}

	return categoryToAmount, err
}
