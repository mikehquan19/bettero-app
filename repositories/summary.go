package repositories

import (
	"betterov2/models"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type SummaryRepo struct{}

func NewSummaryRepo() *SummaryRepo {
	return &SummaryRepo{}
}

// GetBasicAnalysis returns the basic aggregation of the user's account between 2 dates
func (s *SummaryRepo) GetBasicAnalysis(
	ctx context.Context,
	dbtx DBTX,
	userId uuid.UUID,
	start, end time.Time,
) (models.BasicAnalysis, error) {
	var analysis models.BasicAnalysis

	getBasicAnalysisSQL := `
	WITH total_balance AS (
		SELECT COALESCE(SUM(a.balance_cents), 0) AS total_balance
		FROM accounts a 
		WHERE a.user_id = $1 and a.type = 'Debit'
	),
	total_amount_due AS (
		SELECT COALESCE(SUM(a.balance_cents), 0) AS total_amount_due
		FROM accounts a 
		WHERE a.user_id = $1 and a.type = 'Credit'
	),
	total_income AS (
		SELECT COALESCE(SUM(t.amount_cents), 0) AS total_income
		FROM transactions t
		JOIN accounts a ON t.account_id = a.id
		WHERE
			a.user_id = $1 AND
			t.category = 'income' AND
			t.created_at >= $2 AND t.created_at < $3
	),
	total_expense AS (
		SELECT COALESCE(SUM(t.amount_cents), 0) AS total_expense
		FROM transactions t
		JOIN accounts a ON t.account_id = a.id
		WHERE
			a.user_id = $1 AND
			t.category <> 'income' AND
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
	row := dbtx.QueryRow(ctx, getBasicAnalysisSQL, userId, start, end)
	if err := scanAnalysis(row, &analysis); err != nil {
		return models.BasicAnalysis{}, err
	}

	return analysis, nil
}

// GetDateToAmount returns the map from date to total expense of the user or account
func (s *SummaryRepo) GetDateToAmount(
	ctx context.Context,
	dbtx DBTX,
	objType models.ObjectType,
	objId uuid.UUID,
	start, end time.Time,
) (map[string]models.Money, error) {
	var dateToAmount = make(map[string]models.Money)

	for date := start; !date.After(end); date = date.AddDate(0, 0, 1) {
		dateToAmount[date.Format("2006-01-02")] = 0
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

	getDateToAmounttSQL := fmt.Sprintf(`
	SELECT
		t.created_at::date AS date, 
		SUM(t.amount_cents) AS amount
	FROM %s
	WHERE
		%s AND
		t.category <> 'income' AND
		t.created_at >= $2 AND t.created_at < $3
	GROUP BY date;`, table, filter)

	rows, err := dbtx.Query(ctx, getDateToAmounttSQL, objId, start, end)
	if err != nil {
		return nil, err
	}

	var date time.Time
	var amount models.Money
	_, err = pgx.ForEachRow(rows, []any{&date, &amount}, func() error {
		dateToAmount[date.Format("2006-01-02")] = amount
		return nil
	})
	if err != nil {
		return nil, err
	}

	return dateToAmount, nil
}

// getCategoryToAmount returns the map from category to total expense of the obj
func (s *SummaryRepo) GetCategoryToAmount(
	ctx context.Context,
	dbtx DBTX,
	objType models.ObjectType,
	objId uuid.UUID,
	start, end time.Time,
) (map[models.TransactionCategory]models.Money, error) {
	var categoryToAmount = make(map[models.TransactionCategory]models.Money)

	// There are 10 categories
	for _, category := range models.TransactionCategories {
		categoryToAmount[category] = 0
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

	getCategoryToAmountSQL := fmt.Sprintf(`
	SELECT
		t.category, 
		SUM(t.amount_cents) AS amount
	FROM %s
	WHERE
		%s AND
		t.category <> 'income' AND
		t.created_at >= $2 AND t.created_at < $3
	GROUP BY t.category;`, table, filter)

	rows, err := dbtx.Query(ctx, getCategoryToAmountSQL, objId, start, end)
	if err != nil {
		return nil, err
	}

	var category models.TransactionCategory
	var amount models.Money
	_, err = pgx.ForEachRow(rows, []any{&category, &amount}, func() error {
		categoryToAmount[category] = amount
		return nil
	})
	if err != nil {
		return nil, err
	}

	return categoryToAmount, nil
}

func scanAnalysis(analysisRow pgx.Row, analysis *models.BasicAnalysis) error {
	return analysisRow.Scan(
		&analysis.TotalBalanceCents,
		&analysis.TotalAmountDueCents,
		&analysis.TotalIncomeCents,
		&analysis.TotalExpenseCents,
	)
}
