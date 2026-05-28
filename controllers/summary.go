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
	service *services.SummaryService
}

func NewSummaryController(s *services.SummaryService) *SummaryController {
	return &SummaryController{
		service: s,
	}
}

// GET: /summary
func (s *SummaryController) GetSummary(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
	defer cancel()

	var dates [2]time.Time
	var err error
	for i, end := range [2]string{"start", "end"} {
		dates[i], err = time.Parse("2006-01-02", c.Query(end))
		if err != nil {
			respondError(c, http.StatusBadRequest, err)
			return
		}
	}

	basic, err := s.service.GetBasicAnalysis(ctx, UserID, dates[0], dates[1])
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	daily, err := s.service.GetDateToAmount(ctx, "user", UserID, dates[0], dates[1])
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	composition, err := s.service.GetCompositionMap(ctx, "user", UserID, dates[0], dates[1])
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	change, err := s.service.GetChangeMap(ctx, "user", UserID, dates[0], dates[1])
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
