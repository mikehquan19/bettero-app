package services

import (
	"betterov2/models"
	"betterov2/repositories"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestValidatePortion(t *testing.T) {
	validCategories := map[models.TransactionCategory]float64{
		models.Grocery: 60,
		models.Dining:  40,
	}

	tests := []struct {
		name        string
		body        models.GenericBudgetPlanBody
		expectedErr error
	}{
		{
			name: "valid portions",
			body: models.GenericBudgetPlanBody{
				ExpensePortion:  50,
				CategoryPortion: validCategories,
			},
		},
		{
			name: "invalid expense percentage",
			body: models.GenericBudgetPlanBody{
				ExpensePortion:  101,
				CategoryPortion: validCategories,
			},
			expectedErr: models.ErrInvalidExpensePercentage,
		},
		{
			name: "invalid category",
			body: models.GenericBudgetPlanBody{
				ExpensePortion: 50,
				CategoryPortion: map[models.TransactionCategory]float64{
					"fake": 100,
				},
			},
			expectedErr: models.ErrInvalidCategory,
		},
		{
			name: "invalid category percentage",
			body: models.GenericBudgetPlanBody{
				ExpensePortion: 50,
				CategoryPortion: map[models.TransactionCategory]float64{
					models.Grocery: 101,
				},
			},
			expectedErr: models.ErrInvalidCategoryPercentage,
		},
		{
			name: "category percentages do not total one hundred",
			body: models.GenericBudgetPlanBody{
				ExpensePortion: 50,
				CategoryPortion: map[models.TransactionCategory]float64{
					models.Grocery: 80,
				},
			},
			expectedErr: models.ErrCategoryPercentagesNotAddUp,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePortion(tt.body)
			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("validatePortion() error = %v, expected %v", err, tt.expectedErr)
			}
		})
	}
}

func TestBudgetServiceGetBudgetAnalysis(t *testing.T) {
	userID := uuid.New()
	plan := models.BudgetPlan{
		ID:                   uuid.New(),
		UserID:               userID,
		IntervalType:         string(models.Month),
		RecurringIncomeCents: 100000,
		ExpensePortion:       50,
		CategoryPortion: map[models.TransactionCategory]float64{
			models.Grocery: 60,
			models.Dining:  40,
		},
	}
	tests := []struct {
		name            string
		amounts         map[models.TransactionCategory]models.Money
		expectedGrocery models.CategoryProgress
		expectedDining  models.CategoryProgress
	}{
		{
			name: "computes composition and progress",
			amounts: map[models.TransactionCategory]models.Money{
				models.Grocery: 15000,
				models.Dining:  5000,
			},
			expectedGrocery: models.CategoryProgress{
				CurrentCents: 15000,
				BudgetCents:  30000,
				Percentage:   50,
			},
			expectedDining: models.CategoryProgress{
				CurrentCents: 5000,
				BudgetCents:  20000,
				Percentage:   25,
			},
		},
		{
			name: "handles zero spending",
			amounts: map[models.TransactionCategory]models.Money{
				models.Grocery: 0,
				models.Dining:  0,
			},
			expectedGrocery: models.CategoryProgress{
				BudgetCents: 30000,
			},
			expectedDining: models.CategoryProgress{
				BudgetCents: 20000,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			budgetRepo := &fakeBudgetRepository{
				getBudgetPlan: func(context.Context, repositories.DBTX, uuid.UUID, models.IntervalType) (models.BudgetPlan, error) {
					return plan, nil
				},
			}
			summaryRepo := &fakeSummaryRepository{
				getCategoryToAmount: func(context.Context, repositories.DBTX, models.ObjectType, uuid.UUID, time.Time, time.Time) (map[models.TransactionCategory]models.Money, error) {
					return tt.amounts, nil
				},
			}
			service := NewBudgetService(nil, budgetRepo, summaryRepo)

			analysis, err := service.GetBudgetAnalysis(context.Background(), userID, models.Month)

			if err != nil {
				t.Fatalf("GetBudgetAnalysis() error = %v", err)
			}
			if *analysis.Progress[models.Grocery] != tt.expectedGrocery {
				t.Errorf("grocery progress = %#v, expected %#v", *analysis.Progress[models.Grocery], tt.expectedGrocery)
			}
			if *analysis.Progress[models.Dining] != tt.expectedDining {
				t.Errorf("dining progress = %#v, expected %#v", *analysis.Progress[models.Dining], tt.expectedDining)
			}
		})
	}
}

func TestGetCurrentPeriod(t *testing.T) {
	tests := []struct {
		name        string
		interval    models.IntervalType
		expectedErr bool
	}{
		{
			name:     "month",
			interval: models.Month,
		},
		{
			name:     "bi-week",
			interval: models.BiWeek,
		},
		{
			name:     "week",
			interval: models.Week,
		},
		{
			name:        "invalid",
			interval:    "invalid",
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := getCurrentPeriod(tt.interval)
			if (err != nil) != tt.expectedErr {
				t.Fatalf("getCurrentPeriod() error = %v, expectedErr %v", err, tt.expectedErr)
			}
			if !tt.expectedErr && start.After(end) {
				t.Errorf("start %s is after end %s", start, end)
			}
		})
	}
}
