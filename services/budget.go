package services

import (
	"betterov2/models"
	"betterov2/repositories"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Account Service
type BudgetService struct {
	db         *pgxpool.Pool
	budgetRepo *repositories.BudgetRepo
}

// Generate a new account service
func NewBudgetService(db *pgxpool.Pool, budgetRepo *repositories.BudgetRepo) *BudgetService {
	return &BudgetService{
		db:         db,
		budgetRepo: budgetRepo,
	}
}

func (s *BudgetService) GetBudgetPlan(ctx context.Context, userId int64, intervalType models.IntervalType) (models.BudgetResponse, error) {
	return models.BudgetResponse{}, nil
}
