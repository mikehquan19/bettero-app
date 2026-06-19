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
	ctx context.Context, userId int64, start, end time.Time,
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
	FROM total_balance b, total_amount_due a, total_income i, total_expense e
	`
	row := s.database.QueryRow(ctx, getAnalysisQuery, userId, start, end)
	if err := models.ScanAnalysis(row, &analysis); err != nil {
		return analysis, err
	}

	return analysis, nil
}

// GetCompositionMap returns the map from category of expense to its percentage vs total expense.
// If the total expense is 0, then each category takes up 0% of the total expense.
// It uses the total expense computed by GetBasicAnalysis for efficiency.
func (s *SummaryService) GetCompositionMap(
	ctx context.Context, objectType string, id int64, start, end time.Time,
) (map[string]float64, error) {
	var compositionMap = make(map[string]float64)

	categoryToAmount, err := getCategoryToAmount(ctx, s.database, objectType, id, start, end)
	if err != nil {
		return compositionMap, err
	}

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
	ctx context.Context, objectType string, id int64, intervalType string, start, end time.Time,
) (map[string]*float64, error) {
	var changeMap = make(map[string]*float64)

	current, err := getCategoryToAmount(ctx, s.database, objectType, id, start, end)
	if err != nil {
		return changeMap, err
	}

	prevStart, prevEnd := getPrevInterval(intervalType, start, end)
	previous, err := getCategoryToAmount(ctx, s.database, objectType, id, prevStart, prevEnd)
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
	ctx context.Context, objectType string, id int64, start, end time.Time,
) (map[string]float64, error) {
	var dateToAmount = make(map[string]float64)

	for date := start; !date.After(end); date = date.AddDate(0, 0, 1) {
		dateToAmount[date.Format("2006-01-02")] = 0.0
	}

	var objectTable, objectFilter string
	switch objectType {
	case "user":
		// For user, join accounts table to get user ID
		objectTable = "transactions t JOIN accounts a ON t.account_id = a.id"
		objectFilter = "a.user_id = $1"
	case "account":
		// Otherwise, account's ID is already in the transactions table
		objectTable = "transactions t"
		objectFilter = "t.account_id = $1"
	default:
		return nil, fmt.Errorf("invalid object type to filter, user or account")
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
	GROUP BY date;
	`, objectTable, objectFilter)
	rows, err := s.database.Query(ctx, getDateToAmtQuery, id, start, end)
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

// getPrevInterval gets the previous period of the 2 dates
func getPrevInterval(intervalType string, start, end time.Time) (time.Time, time.Time) {
	switch intervalType {
	case "MONTH":
		return start.AddDate(0, -1, 0), end.AddDate(0, -1, 0)

	case "WEEK":
		return start.AddDate(0, 0, -7), end.AddDate(0, 0, -7)

	case "BIWEEK":
		return start.AddDate(0, 0, -14), end.AddDate(0, 0, -14)

	default:
		panic(fmt.Errorf("Invalid inverval type"))
	}
}

// Helper function: getCategoryToAmount returns the map from categoryto total expense
func getCategoryToAmount(
	ctx context.Context, db *pgxpool.Pool, objectType string, id int64, start, end time.Time,
) (map[string]float64, error) {
	var categoryToAmount = make(map[string]float64)

	// There are 10 categories
	for _, category := range []string{
		"Housing", "Automobile", "Medical", "Subscription", "Grocery", "Dining", "Shopping", "Gas", "Others",
	} {
		categoryToAmount[category] = 0.0
	}

	var objectTable, objectFilter string
	switch objectType {
	case "user":
		// For user, join accounts table to get user ID
		objectTable = "transactions t JOIN accounts a ON t.account_id = a.id"
		objectFilter = "a.user_id = $1"
	case "account":
		objectTable = "transactions t"
		objectFilter = "t.account_id = $1"
	default:
		return nil, fmt.Errorf("invalid object type to filter, user or account")
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
	GROUP BY t.category;
	`, objectTable, objectFilter)
	rows, err := db.Query(ctx, getCatToAmtQuery, id, start, end)
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
