package services

import (
	"betterov2/models"
	"betterov2/repositories"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SummaryService
type SummaryService struct {
	db          *pgxpool.Pool
	summaryRepo *repositories.SummaryRepo
}

// Generate a new summary service
func NewSummaryService(db *pgxpool.Pool, summaryRepo *repositories.SummaryRepo) *SummaryService {
	return &SummaryService{
		db:          db,
		summaryRepo: summaryRepo,
	}
}

// GetBasicAnalysis returns the basic aggregation of the user's account between 2 dates
func (s *SummaryService) GetBasicAnalysis(
	ctx context.Context,
	userId int64,
	dates models.SummaryDates,
) (models.BasicAnalysis, error) {
	var analysis models.BasicAnalysis

	start := dates.CurrStart
	end := dates.CurrEnd

	analysis, err := s.summaryRepo.GetBasicAnalysis(ctx, s.db, userId, start, end)
	if err != nil {
		return analysis, err
	}

	return analysis, nil
}

// GetDailyMap returns the map from the date to expense
func (s *SummaryService) GetDailyMap(
	ctx context.Context,
	objType models.ObjectType,
	objId int64,
	dates models.SummaryDates,
) (map[string]float64, error) {
	var dailyMap map[string]float64

	start := dates.CurrStart
	end := dates.CurrEnd

	dailyMap, err := s.summaryRepo.GetDateToAmount(ctx, s.db, objType, objId, start, end)
	if err != nil {
		return dailyMap, err
	}

	return dailyMap, nil
}

// GetCompositionMap returns the map from category of expense to its percentage vs total expense.
// If the total expense is 0, then each category takes up 0% of the total expense.
// It uses the total expense computed by GetBasicAnalysis for efficiency.
func (s *SummaryService) GetCompositionMap(
	ctx context.Context,
	objType models.ObjectType,
	objId int64,
	dates models.SummaryDates,
) (map[string]float64, error) {
	var compositionMap = make(map[string]float64)

	start := dates.CurrStart
	end := dates.CurrEnd

	categoryToAmount, err := s.summaryRepo.GetCategoryToAmount(ctx, s.db, objType, objId, start, end)
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
	ctx context.Context,
	objType models.ObjectType,
	objId int64,
	dates models.SummaryDates,
) (map[string]*float64, error) {
	var changeMap = make(map[string]*float64)

	start := dates.CurrStart
	end := dates.CurrEnd

	current, err := s.summaryRepo.GetCategoryToAmount(ctx, s.db, objType, objId, start, end)
	if err != nil {
		return changeMap, err
	}

	prevStart := dates.PrevStart
	prevEnd := dates.PrevEnd

	previous, err := s.summaryRepo.GetCategoryToAmount(ctx, s.db, objType, objId, prevStart, prevEnd)
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
