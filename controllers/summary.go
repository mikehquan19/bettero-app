package controllers

import (
	"betterov2/models"
	"betterov2/services"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type SummaryController struct {
	summaryService *services.SummaryService
}

func NewSummaryController(s *services.SummaryService) *SummaryController {
	return &SummaryController{
		summaryService: s,
	}
}

// GET: /summary
func (s *SummaryController) GetSummary(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
	defer cancel()

	dates, err := getSummaryDates(c)
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	basic, err := s.summaryService.GetBasicAnalysis(ctx, UserID, dates)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	daily, err := s.summaryService.GetDailyMap(ctx, models.UserObj, UserID, dates)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	composition, err := s.summaryService.GetCompositionMap(ctx, models.UserObj, UserID, dates)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	change, err := s.summaryService.GetChangeMap(ctx, models.UserObj, UserID, dates)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	summary := models.FinancialSummary{
		Basic:       basic,
		Daily:       daily,
		Change:      change,
		Composition: composition,
	}

	respondSuccess(c, http.StatusOK, summary)
}

// getSummaryDates parses the summary dates from query params
func getSummaryDates(c *gin.Context) (models.SummaryDates, error) {
	var dates models.SummaryDates
	var err error

	for _, field := range []struct {
		dest *time.Time
		name string
	}{
		{dest: &dates.CurrStart, name: "curr_start"},
		{dest: &dates.CurrEnd, name: "curr_end"},
		{dest: &dates.PrevStart, name: "prev_start"},
		{dest: &dates.PrevEnd, name: "prev_end"},
	} {
		*field.dest, err = time.Parse("2006-01-02", c.Query(field.name))
		if err != nil {
			return models.SummaryDates{}, err
		}
	}

	return dates, nil
}
