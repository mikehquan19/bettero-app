package repositories

import (
	"betterov2/models"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type BudgetRepo struct{}

func NewBudgetRepo() *BudgetRepo {
	return &BudgetRepo{}
}

// GetBudgetPlan returns the raw budget plan with given type from user
func (r *BudgetRepo) GetBudgetPlan(
	ctx context.Context,
	db models.DBTX,
	userId int64,
	intervalType models.IntervalType,
) (models.BudgetPlan, error) {
	var budgetPlan models.BudgetPlan

	const getBudgetQuery = `SELECT * FROM budget_plans WHERE user_id = $1 AND interval_type = $2;`
	row := db.QueryRow(ctx, getBudgetQuery, userId, intervalType)
	if err := models.ScanBudgetPlan(row, &budgetPlan); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Budget doesn't exist
			budgetNotFound := fmt.Errorf("Budget of %s not found, %w", intervalType, models.ErrNotFound)
			return budgetPlan, budgetNotFound
		}
	}

	return budgetPlan, nil
}
