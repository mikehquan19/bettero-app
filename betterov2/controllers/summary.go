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

	var summary models.FinancialSummary
	var err error

	var dates [4]time.Time
	for i, date := range [4]string{"prev_start", "prev_end", "curr_start", "curr_end"} {
		dates[i], err = time.Parse("2006-01-02", c.Query(date))
		if err != nil {
			respondError(c, http.StatusBadRequest, err)
			return
		}
	}

	summary.Basic, err = services.GetBasicAnalysis(ctx, UserID, dates[2], dates[3])
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	previous, err := services.GetCategoryToAmount(ctx, UserID, dates[0], dates[1])
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	current, err := services.GetCategoryToAmount(ctx, UserID, dates[2], dates[3])
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	// Computer the daily expense
	summary.Daily, err = services.GetDateToAmount(ctx, UserID, dates[2], dates[3])
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	// Compute the composition percentage
	summary.Composition = make(map[string]float64)
	if summary.Basic.TotalExpense != 0 {
		for category, amount := range current {
			summary.Composition[category] = amount * 100 / summary.Basic.TotalExpense
			summary.Composition[category] = round(summary.Composition[category])
		}
	}

	// Compute the change percentage
	summary.Change = make(map[string]*float64)
	for category, prevAmount := range previous {
		if prevAmount != 0 {
			change := round((current[category] - prevAmount) * 100 / prevAmount)
			summary.Change[category] = &change
		}
	}

	respondSuccess(c, http.StatusOK, summary)
}
