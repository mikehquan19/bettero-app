package controllers

import (
	"betterov2/models"
	"betterov2/services"
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type BudgetController struct {
	budgetService *services.BudgetService
}

func NewBudgetController(s *services.BudgetService) *BudgetController {
	return &BudgetController{budgetService: s}
}

// GET: /budget_plan/:intervalType
func (b *BudgetController) GetBudgetPlan(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	intervalType := models.IntervalType(c.Param("intervalType"))

	budgetResponse, err := b.budgetService.GetBudgetAnalysis(ctx, UserID, intervalType)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			respondError(c, http.StatusNotFound, err)
		} else {
			respondError(c, http.StatusInternalServerError, err)
		}
		return
	}

	respondSuccess(c, http.StatusOK, budgetResponse)
}

// POST: /budget_plan
func (b *BudgetController) PostBudgetPlan(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var body models.PostBudgetPlanBody
	if err := c.ShouldBindJSON(&body); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	newBudgetPlan, err := b.budgetService.CreateBudgetPlan(ctx, UserID, body)
	if err != nil {
		if errors.Is(err, models.ErrInvalidBudgetBody) {
			respondError(c, http.StatusBadRequest, err)
		} else if errors.Is(err, models.ErrForeignKey) {
			respondError(c, http.StatusNotFound, err)
		} else {
			respondError(c, http.StatusInternalServerError, err)
		}
		return
	}

	respondSuccess(c, http.StatusCreated, newBudgetPlan)
}

// PUT: /budget_plan/:intervalType
func (b *BudgetController) PutBudgetPlan(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	intervalType := models.IntervalType(c.Param("intervalType"))

	var body models.PutBudgetPlanBody
	if err := c.ShouldBindJSON(&body); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	updatedBudgetPlan, err := b.budgetService.UpdateBudgetPlan(ctx, UserID, intervalType, body)
	if err != nil {
		if errors.Is(err, models.ErrInvalidBudgetBody) {
			respondError(c, http.StatusBadRequest, err)
		} else if errors.Is(err, models.ErrNotFound) {
			respondError(c, http.StatusNotFound, err)
		} else {
			respondError(c, http.StatusInternalServerError, err)
		}
		return
	}

	respondSuccess(c, http.StatusAccepted, updatedBudgetPlan)
}

// DELETE: /budget_plan/:intervalType
func (b *BudgetController) DeleteBudgetPlan(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	intervalType := models.IntervalType(c.Param("intervalType"))

	_, err := b.budgetService.DeleteBudgetPlan(ctx, UserID, intervalType)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			respondError(c, http.StatusNotFound, err)
		} else {
			respondError(c, http.StatusInternalServerError, err)
		}
		return
	}

	respondSuccess(c, http.StatusAccepted, "Budget plan deleted!")
}
