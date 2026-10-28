package services

import (
	"betterov2/models"
	"betterov2/repositories"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSummaryServiceGetCompositionMap(t *testing.T) {
	tests := []struct {
		name     string
		amounts  map[models.TransactionCategory]models.Money
		expected map[models.TransactionCategory]float64
	}{
		{
			name: "calculates percentages",
			amounts: map[models.TransactionCategory]models.Money{
				models.Grocery: 300,
				models.Dining:  100,
			},
			expected: map[models.TransactionCategory]float64{
				models.Grocery: 75,
				models.Dining:  25,
			},
		},
		{
			name: "zero total produces zero percentages",
			amounts: map[models.TransactionCategory]models.Money{
				models.Grocery: 0,
				models.Dining:  0,
			},
			expected: map[models.TransactionCategory]float64{
				models.Grocery: 0,
				models.Dining:  0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeSummaryRepository{
				getCategoryToAmount: func(context.Context, repositories.DBTX, models.ObjectType, uuid.UUID, time.Time, time.Time) (map[models.TransactionCategory]models.Money, error) {
					return tt.amounts, nil
				},
			}
			service := NewSummaryService(nil, repo)

			composition, err := service.GetCompositionMap(context.Background(), models.UserObj, uuid.New(), models.SummaryDates{})

			if err != nil {
				t.Fatalf("GetCompositionMap() error = %v", err)
			}
			for category, expected := range tt.expected {
				if composition[category] != expected {
					t.Errorf("composition[%s] = %v, expected %v", category, composition[category], expected)
				}
			}
		})
	}
}

func TestSummaryServiceGetChangeMap(t *testing.T) {
	tests := []struct {
		name              string
		current           map[models.TransactionCategory]models.Money
		previous          map[models.TransactionCategory]models.Money
		expectedGrocery   *float64
		expectedDiningNil bool
	}{
		{
			name: "calculates change and preserves nil baseline",
			current: map[models.TransactionCategory]models.Money{
				models.Grocery: 150,
				models.Dining:  50,
			},
			previous: map[models.TransactionCategory]models.Money{
				models.Grocery: 100,
				models.Dining:  0,
			},
			expectedGrocery:   floatPointer(50),
			expectedDiningNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := 0
			repo := &fakeSummaryRepository{
				getCategoryToAmount: func(context.Context, repositories.DBTX, models.ObjectType, uuid.UUID, time.Time, time.Time) (map[models.TransactionCategory]models.Money, error) {
					call++
					if call == 1 {
						return tt.current, nil
					}
					return tt.previous, nil
				},
			}
			service := NewSummaryService(nil, repo)

			change, err := service.GetChangeMap(context.Background(), models.UserObj, uuid.New(), models.SummaryDates{})

			if err != nil {
				t.Fatalf("GetChangeMap() error = %v", err)
			}
			if change[models.Grocery] == nil || *change[models.Grocery] != *tt.expectedGrocery {
				t.Errorf("grocery change = %v, expected %v", change[models.Grocery], tt.expectedGrocery)
			}
			if (change[models.Dining] == nil) != tt.expectedDiningNil {
				t.Errorf("dining nil = %v, expected %v", change[models.Dining] == nil, tt.expectedDiningNil)
			}
		})
	}
}

func floatPointer(value float64) *float64 {
	return &value
}
