package controllers

import (
	"betterov2/models"
	"betterov2/services"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GET: /summary
func GetSummary(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	dates := make(map[string]time.Time)
	var err error
	for _, endType := range [4]string{
		"prev_start", "prev_end", "curr_start", "curr_end",
	} {
		dates[endType], err = time.Parse("2006-01-02", c.Query(endType))
		if err != nil {
			respondError(c, http.StatusBadRequest, err)
			return
		}
	}

	basic, err := services.GetBasicAnalysis(ctx, UserID, dates)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
	}

	daily, err := services.GetDateToAmount(ctx, UserID, dates)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
	}

	composition, err := services.GetCompositionMap(ctx, UserID, dates, basic.TotalExpense)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
	}

	change, err := services.GetChangeMap(ctx, UserID, dates)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
	}

	summary := models.FinancialSummary{
		Basic:       basic,
		Daily:       daily,
		Change:      change,
		Composition: composition,
	}

	respondSuccess(c, http.StatusOK, summary)
}
