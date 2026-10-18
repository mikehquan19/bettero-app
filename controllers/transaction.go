package controllers

import (
	"betterov2/models"
	"betterov2/services"
	"context"
	"errors"
	"fmt"

	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TransactionController struct {
	tranService *services.TransactionService
}

func NewTransactionController(s *services.TransactionService) *TransactionController {
	return &TransactionController{
		tranService: s,
	}
}

// GET: /autocomplete?q=
//
// Return the list of descriptions suggested from keyword
func (t *TransactionController) SearchTransactions(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	userID, err := authenticatedUserID(c)
	if err != nil {
		abortAuthentication(c, err)
		return
	}

	if c.Query("q") == "" {
		respondError(c, http.StatusBadRequest, fmt.Errorf("q must be specified"))
		return
	}

	suggestions, err := t.tranService.ListSuggestions(ctx, userID, c.Query("q"))
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	respondSuccess(c, http.StatusOK, suggestions)
}

// GET: /transactions
//
// Return the paginated list of latest transactions of the user
func (t *TransactionController) GetTransactions(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	userID, err := authenticatedUserID(c)
	if err != nil {
		abortAuthentication(c, err)
		return
	}

	offset, err := getOffset(c)
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	if (c.Query("start") != "") != (c.Query("end") != "") {
		// Enforce both ends of the date parameters
		respondError(c, http.StatusBadRequest, fmt.Errorf("both start, & end must be specified"))
		return
	}
	// Either both dates are defined or nil
	dates := map[string]*time.Time{"start": nil, "end": nil}
	for end := range dates {
		if c.Query(end) != "" {
			date, err := time.Parse("2006-01-02", c.Query(end))
			if err != nil {
				respondError(c, http.StatusBadRequest, err)
				return
			}
			dates[end] = &date
		}
	}

	filter := models.TransactionFilter{
		Category:        models.TransactionCategory(c.Query("category")),
		TranDescription: c.Query("description"),
		Merchant:        c.Query("merchant"),
		CreatedAtFrom:   dates["start"],
		CreatedAtTo:     dates["end"],
	}

	// Filter transactions based on category, and dates
	total, transactions, err := t.tranService.FilterTransactions(ctx, userID, filter, offset)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	paginatedResponse := models.PaginatedResponse[models.Transaction]{
		Total:  total,
		Offset: offset,
		Data:   transactions,
	}
	respondSuccess(c, http.StatusOK, paginatedResponse)
}

// POST: /transactions/
//
// Create the new transaction of the account
func (t *TransactionController) PostTransaction(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	userID, err := authenticatedUserID(c)
	if err != nil {
		abortAuthentication(c, err)
		return
	}

	var body models.PostTransactionBody
	if err := c.ShouldBindJSON(&body); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	newTransaction, err := t.tranService.CreateTransaction(ctx, userID, body)
	if err != nil {
		if errors.Is(err, models.ErrForeignKey) {
			respondError(c, http.StatusNotFound, err)
			return
		}
		if errors.Is(err, models.ErrTransactionTooOld) {
			respondError(c, http.StatusBadRequest, err)
			return
		}
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondSuccess(c, http.StatusCreated, newTransaction)
}

// PUT: /transactions/:id
//
// Update the info of the transaction with the given id
func (t *TransactionController) PutTransaction(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	userID, err := authenticatedUserID(c)
	if err != nil {
		abortAuthentication(c, err)
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	var body models.PutTransactionBody
	if err = c.ShouldBindJSON(&body); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	updatedTran, err := t.tranService.UpdateTransaction(ctx, userID, id, body)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			respondError(c, http.StatusNotFound, err)
			return
		}
		if errors.Is(err, models.ErrTransactionTooOld) {
			respondError(c, http.StatusBadRequest, err)
			return
		}
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondSuccess(c, http.StatusAccepted, updatedTran)
}

// DELETE: /transactions/:id
//
// Delete the transaction with given ID
func (t *TransactionController) DeleteTransaction(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	userID, err := authenticatedUserID(c)
	if err != nil {
		abortAuthentication(c, err)
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	if err := t.tranService.DeleteTransaction(ctx, userID, id); err != nil {
		if errors.Is(err, models.ErrNotFound) {
			respondError(c, http.StatusNotFound, err)
			return
		}
		if errors.Is(err, models.ErrTransactionTooOld) {
			respondError(c, http.StatusBadRequest, err)
			return
		}
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondSuccess(c, http.StatusAccepted, "Transaction deleted!")
}
