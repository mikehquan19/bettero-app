package services

import (
	"betterov2/models"
	"betterov2/repositories"
	"context"
	"fmt"
	"time"

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
	start, end time.Time,
) (models.BasicAnalysis, error) {
	var analysis models.BasicAnalysis

	analysis, err := s.summaryRepo.GetBasicAnalysis(ctx, s.db, userId, start, end)
	if err != nil {
		return analysis, err
	}

	return analysis, nil
}

// GetDailyMap returns the map from the date to expense
func (s *SummaryService) GetDailyMap(
	ctx context.Context,
	objectType string,
	id int64,
	start, end time.Time,
) (map[string]float64, error) {
	var dailyMap map[string]float64

	dailyMap, err := s.summaryRepo.GetDateToAmount(ctx, s.db, objectType, id, start, end)
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
	objectType string,
	id int64,
	start, end time.Time,
) (map[string]float64, error) {
	var compositionMap = make(map[string]float64)

	categoryToAmount, err := s.summaryRepo.GetCategoryToAmount(ctx, s.db, objectType, id, start, end)
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
	objectType string,
	id int64,
	intervalType string,
	start, end time.Time,
) (map[string]*float64, error) {
	var changeMap = make(map[string]*float64)

	current, err := s.summaryRepo.GetCategoryToAmount(ctx, s.db, objectType, id, start, end)
	if err != nil {
		return changeMap, err
	}

	prevStart, prevEnd, err := getPrevInterval(intervalType, start, end)
	if err != nil {
		return changeMap, err
	}
	previous, err := s.summaryRepo.GetCategoryToAmount(ctx, s.db, objectType, id, *prevStart, *prevEnd)
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

// getPrevInterval gets the previous period of the 2 dates
func getPrevInterval(intervalType string, start, end time.Time) (*time.Time, *time.Time, error) {
	var prevStart, prevEnd time.Time
	switch intervalType {
	case "MONTH":
		prevStart = start.AddDate(0, -1, 0)
		prevEnd = time.Date(
			prevStart.Year(),
			prevStart.Month()+1,
			0,
			0, 0, 0, 0, // From hour -> nsec
			prevStart.Location(),
		)

	case "WEEK":
		prevStart = start.AddDate(0, 0, -7)
		prevEnd = end.AddDate(0, 0, -7)

	case "BIWEEK":
		prevStart = start.AddDate(0, 0, -14)
		prevEnd = end.AddDate(0, 0, -14)

	default:
		return nil, nil, fmt.Errorf("Invalid interval type")
	}

	return &prevStart, &prevEnd, nil
}
