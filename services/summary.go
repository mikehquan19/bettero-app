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
	analysis, err := s.summaryRepo.GetBasicAnalysis(ctx, s.db, userId, dates.CurrStart, dates.CurrStart)
	if err != nil {
		return models.BasicAnalysis{}, err
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
	dailyMap, err := s.summaryRepo.GetDateToAmount(ctx, s.db, objType, objId, dates.CurrStart, dates.CurrEnd)
	if err != nil {
		return nil, err
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
) (map[models.TransactionCategory]float64, error) {
	var compositionMap = make(map[models.TransactionCategory]float64)

	categoryToAmount, err := s.summaryRepo.GetCategoryToAmount(ctx, s.db, objType, objId, dates.CurrStart, dates.CurrEnd)
	if err != nil {
		return nil, err
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

	return compositionMap, nil
}

// GetChangeMap returns the map from category to the expense change percentage
// from previous period to current period.
// If the previous expense is 0 then it's null.
func (s *SummaryService) GetChangeMap(
	ctx context.Context,
	objType models.ObjectType,
	objId int64,
	dates models.SummaryDates,
) (map[models.TransactionCategory]*float64, error) {
	var changeMap = make(map[models.TransactionCategory]*float64)

	current, err := s.summaryRepo.GetCategoryToAmount(ctx, s.db, objType, objId, dates.CurrStart, dates.CurrEnd)
	if err != nil {
		return nil, err
	}

	previous, err := s.summaryRepo.GetCategoryToAmount(ctx, s.db, objType, objId, dates.PrevStart, dates.PrevEnd)
	if err != nil {
		return nil, err
	}

	for category, amount := range previous {
		if amount != 0 {
			percent := round((current[category] - amount) * 100 / amount)
			changeMap[category] = &percent
		} else {
			changeMap[category] = nil
		}
	}

	return changeMap, nil
}
