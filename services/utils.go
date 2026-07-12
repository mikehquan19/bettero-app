package services

import (
	"betterov2/models"
	"fmt"
	"math"
	"time"
)

// round rounds a float64 to 2 decimal places
func round(x float64) float64 {
	return math.Round(x*100) / 100
}

// getCurrentPeriod returns the time range for the current period of the given interval type.
// The returned start and end times represent the current month, biweek, or week.
func getCurrentPeriod(intervalType models.IntervalType) (time.Time, time.Time, error) {
	now := time.Now()

	switch intervalType {
	case models.Month:
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		end := time.Date(now.Year(), now.Month()+1, 0, 23, 59, 59, 999000000, now.Location())
		return start, end, nil
	case models.BiWeek:
		end := endOfISOWeek(now)
		return end.AddDate(0, 0, -13), end, nil
	case models.Week:
		return startOfISOWeek(now), endOfISOWeek(now), nil
	default:
		return time.Time{}, time.Time{}, fmt.Errorf("Invalid interval type")
	}
}

// startOfISOWeek returns the Monday of the ISO week containing the given time.
func startOfISOWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}

	startDay := t.Day() - (weekday - 1)
	return time.Date(t.Year(), t.Month(), startDay, 0, 0, 0, 0, t.Location())
}

// endOfISOWeek returns the Sunday of the ISO week containing the given time.
func endOfISOWeek(t time.Time) time.Time {
	start := startOfISOWeek(t)
	return time.Date(start.Year(), start.Month(), start.Day()+6, 23, 59, 59, 999000000, start.Location())
}
