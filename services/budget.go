package services

import (
	"betterov2/models"
	"betterov2/repositories"
	"context"
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Account Service
type BudgetService struct {
	db          *pgxpool.Pool
	budgetRepo  *repositories.BudgetRepo
	summaryRepo *repositories.SummaryRepo
}

// Generate a new account service
func NewBudgetService(db *pgxpool.Pool, budgetRepo *repositories.BudgetRepo, summaryRepo *repositories.SummaryRepo) *BudgetService {
	return &BudgetService{
		db:          db,
		budgetRepo:  budgetRepo,
		summaryRepo: summaryRepo,
	}
}

// GetBudgetPlan returns the analysis of user's spending behavior and budget plan
func (s *BudgetService) GetBudgetAnalysis(ctx context.Context, userId int64, intervalType models.IntervalType) (models.BudgetResponse, error) {
	budgetPlan, err := s.budgetRepo.GetBudgetPlan(ctx, s.db, userId, intervalType)
	if err != nil {
		return models.BudgetResponse{}, err
	}

	budgetResponse := models.BudgetResponse{
		ID:              budgetPlan.ID,
		IntervalType:    budgetPlan.IntervalType,
		RecurringIncome: budgetPlan.RecurringIncome,
		ExpensePortion:  budgetPlan.ExpensePortion,
		BudgetComposition: models.BudgetComposition{
			Goal: budgetPlan.CategoryPortion,
			Real: make(map[models.TransactionCategory]float64),
		},
		Progress: make(map[models.TransactionCategory]*models.CategoryProgress),
	}

	// Get the interval based on the interval type
	start, end, err := getCurrentPeriod(intervalType)
	log.Printf("Start date of this %s: %s", intervalType, start)
	log.Printf("End date of this %s: %s", intervalType, end)

	if err != nil {
		return models.BudgetResponse{}, err
	}
	categoryToAmount, err := s.summaryRepo.GetCategoryToAmount(ctx, s.db, models.UserObj, userId, start, end)
	if err != nil {
		return models.BudgetResponse{}, err
	}

	// Compute the real composition of the budget.
	// TODO: Logic is the same as summary composition percentage, so write the reusable code
	var totalExpense = 0.0
	for _, amount := range categoryToAmount {
		totalExpense += amount
	}

	for category, amount := range categoryToAmount {
		if totalExpense != 0.0 {
			percent := round(amount * 100 / totalExpense)
			budgetResponse.BudgetComposition.Real[category] = round(percent)
		} else {
			budgetResponse.BudgetComposition.Real[category] = 0.0
		}
	}

	// Compute the budget progress
	for category, amount := range categoryToAmount {
		budget := budgetPlan.RecurringIncome * budgetPlan.ExpensePortion * budgetPlan.CategoryPortion[category] / 10000
		percentage := 0.0
		if budget != 0 {
			percentage = round(amount * 100 / budget)
		}

		budgetResponse.Progress[category] = &models.CategoryProgress{
			Current:    amount,
			Budget:     budget,
			Percentage: percentage,
		}
	}

	return budgetResponse, nil
}

// getCurrentPeriod returns the time range for the current period of the given interval type:
// month, bi_week, week
func getCurrentPeriod(intervalType models.IntervalType) (time.Time, time.Time, error) {
	now := time.Now()

	switch intervalType {
	case models.Month:
		return startOfISOMonth(now), endOfISOMonth(now), nil
	case models.BiWeek:
		end := endOfISOWeek(now)
		return end.AddDate(0, 0, -13), end, nil
	case models.Week:
		return startOfISOWeek(now), endOfISOWeek(now), nil
	default:
		return time.Time{}, time.Time{}, fmt.Errorf("Invalid interval type")
	}
}

func startOfISOMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

func endOfISOMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month()+1, 0, 23, 59, 59, 999000000, t.Location())
}

func startOfISOWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}

	startDay := t.Day() - (weekday - 1)
	return time.Date(t.Year(), t.Month(), startDay, 0, 0, 0, 0, t.Location())
}

func endOfISOWeek(t time.Time) time.Time {
	start := startOfISOWeek(t)
	return time.Date(start.Year(), start.Month(), start.Day()+6, 23, 59, 59, 999000000, start.Location())
}

// CreateBudgetPlan creates and returns the new budget plan
func (s *BudgetService) CreateBudgetPlan(ctx context.Context, userId int64, body models.PostBudgetPlanBody) (models.BudgetPlan, error) {
	err := validatePortion(body.GenericBudgetPlanBody)
	if err != nil {
		return models.BudgetPlan{}, err
	}

	newBudgetPlan, err := s.budgetRepo.InsertBudgetPlan(ctx, s.db, userId, body)
	if err != nil {
		return models.BudgetPlan{}, err
	}

	return newBudgetPlan, nil
}

// UpdateBudgetPlan updates and returns the budget plan by interval type
func (s *BudgetService) UpdateBudgetPlan(
	ctx context.Context,
	userId int64,
	intervalType models.IntervalType,
	body models.PutBudgetPlanBody,
) (models.BudgetPlan, error) {
	err := validatePortion(body.GenericBudgetPlanBody)
	if err != nil {
		return models.BudgetPlan{}, err
	}

	updatedBudgetPlan, err := s.budgetRepo.UpdateBudgetPlan(ctx, s.db, userId, intervalType, body)
	if err != nil {
		return models.BudgetPlan{}, err
	}

	return updatedBudgetPlan, nil
}

// validatePortion validates the budget plan body.
//
//   - Expense portion must be valid percentage (0 - 100%)
//
//   - Each category must be of the valid transaction categories
//
//   - Each of category portion must be valid percentage
//
//   - All of category portions must add up to 100%
func validatePortion(body models.GenericBudgetPlanBody) error {
	if body.ExpensePortion < 0 || body.ExpensePortion > 100 {
		// NOTE: This can lowkey be handled at the datbase level
		return models.ErrInvalidExpensePercentage
	}

	var totalPercent float64 = 0
	for category, percent := range body.CategoryPortion {
		if !slices.Contains(models.TransactionCategories, category) {
			return models.ErrInvalidCategory
		}
		if percent < 0 || percent > 100 {
			return models.ErrInvalidCategoryPercentage
		}

		totalPercent += percent
	}
	if totalPercent != 100.00 {
		return models.ErrCategoryPercentagesNotAddUp
	}

	return nil
}

// DeleteBudgetPlan deletes and returns the budget plan by interval type
func (s *BudgetService) DeleteBudgetPlan(ctx context.Context, userId int64, intervalType models.IntervalType) (models.BudgetPlan, error) {
	deletedBudgetPlan, err := s.budgetRepo.DeleteBudgetPlan(ctx, s.db, userId, intervalType)
	if err != nil {
		return models.BudgetPlan{}, err
	}

	return deletedBudgetPlan, nil
}
