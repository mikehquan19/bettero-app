package repositories

import (
	"betterov2/models"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type BudgetRepo struct{}

func NewBudgetRepo() *BudgetRepo {
	return &BudgetRepo{}
}

// GetBudgetPlan returns the raw budget plan with given type from user
func (r *BudgetRepo) GetBudgetPlan(
	ctx context.Context,
	dbtx DBTX,
	userId int64,
	intervalType models.IntervalType,
) (models.BudgetPlan, error) {
	var budgetPlan models.BudgetPlan

	const getBudgetSQL = `SELECT * FROM budget_plans WHERE user_id = $1 AND interval_type = $2;`
	row := dbtx.QueryRow(ctx, getBudgetSQL, userId, intervalType)
	if err := scanBudgetPlan(row, &budgetPlan); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.BudgetPlan{}, models.ErrNotFound
		}
		return models.BudgetPlan{}, err
	}

	return budgetPlan, nil
}

// InsertBudgetPlan inserts to the database and returns the new budget plan
func (r *BudgetRepo) InsertBudgetPlan(
	ctx context.Context,
	dbtx DBTX,
	userId int64,
	body models.PostBudgetPlanBody,
) (models.BudgetPlan, error) {
	var newBudgetPlan models.BudgetPlan

	const insertBudgetSQL = `
	INSERT INTO budget_plans (
		user_id, 
		interval_type, 
		recurring_income, 
		expense_portion, 
		category_portion
	) 
	VALUES ($1, $2, $3, $4, $5)
	RETURNING *;`

	row := dbtx.QueryRow(ctx, insertBudgetSQL,
		userId,
		body.IntervalType,
		body.RecurringIncome,
		body.ExpensePortion,
		body.CategoryPortion,
	)
	if err := scanBudgetPlan(row, &newBudgetPlan); err != nil {
		if isForeignKeyViolation(err) {
			return models.BudgetPlan{}, models.ErrForeignKey
		}
		return models.BudgetPlan{}, err
	}

	return newBudgetPlan, nil
}

// UpdateBudgetPlan updates and returns the budget plan by interval type
func (r *BudgetRepo) UpdateBudgetPlan(
	ctx context.Context,
	dbtx DBTX,
	userId int64,
	intervalType models.IntervalType,
	body models.PutBudgetPlanBody,
) (models.BudgetPlan, error) {
	var updatedBudgetPlan models.BudgetPlan

	const updateBudgetPlanSQL = `
	UPDATE budget_plans
	SET recurring_income = $3, 
		expense_portion = $4, 
		category_portion = $5,
		updated_at = NOW()
	WHERE user_id = $1 AND interval_type = $2
	RETURNING *;`

	row := dbtx.QueryRow(ctx, updateBudgetPlanSQL,
		userId,
		intervalType,
		body.RecurringIncome,
		body.ExpensePortion,
		body.CategoryPortion,
	)
	if err := scanBudgetPlan(row, &updatedBudgetPlan); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.BudgetPlan{}, models.ErrNotFound
		}
		return models.BudgetPlan{}, err
	}

	return updatedBudgetPlan, nil
}

// DeleteBudgetPlan deletes and returns the budget plan
func (r *BudgetRepo) DeleteBudgetPlan(
	ctx context.Context,
	dbtx DBTX,
	userId int64,
	intervalType models.IntervalType,
) (models.BudgetPlan, error) {
	var deletedBudgetPlan models.BudgetPlan

	const deleteBudgetPlanSQL = `
	DELETE FROM budget_plans 
	WHERE user_id = $1 AND interval_type = $2 
	RETURNING *;`

	row := dbtx.QueryRow(ctx, deleteBudgetPlanSQL, userId, intervalType)
	if err := scanBudgetPlan(row, &deletedBudgetPlan); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.BudgetPlan{}, models.ErrNotFound
		}
		return models.BudgetPlan{}, err
	}
	return deletedBudgetPlan, nil
}

func scanBudgetPlan(budgetPlanRow pgx.Row, budgetPlan *models.BudgetPlan) error {
	return budgetPlanRow.Scan(
		&budgetPlan.ID,
		&budgetPlan.UserID,
		&budgetPlan.IntervalType,
		&budgetPlan.RecurringIncome,
		&budgetPlan.ExpensePortion,
		&budgetPlan.CategoryPortion,
		&budgetPlan.CreatedAt,
		&budgetPlan.UpdatedAt,
	)
}
