package controllers

import (
	"betterov2/models"
	"betterov2/services"
	"context"
	"errors"
	"fmt"

	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// GET: /transactions
//
// Return the paginated list of latest transactions of the user
func GetTransactions(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	offset, err := getOffset(c)
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	if (c.Query("start") != "") != (c.Query("end") != "") {
		// Enforce both ends of the date parameters
		respondError(c, http.StatusBadRequest, fmt.Errorf("Both start, & end must be specified"))
		return
	}
	// Either both dates are defined or nil
	var dates [2]*time.Time
	for i, end := range [2]string{"start", "end"} {
		if c.Query(end) != "" {
			date, err := time.Parse("2006-01-02", c.Query(end))
			if err != nil {
				respondError(c, http.StatusBadRequest, err)
				return
			}
			dates[i] = &date
		}
	}
	filter := models.TransactionFilter{
		Category:        c.Query("category"),
		TranDescription: c.Query("description"),
		CreatedAtFrom:   dates[0],
		CreatedAtTo:     dates[1],
	}

	// Filter transactions based on category, and dates
	total, transactions, err := services.FilterTransactions(ctx, UserID, filter, offset)
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
func PostTransaction(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var body models.PostTransactionBody
	if err := c.ShouldBindJSON(&body); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	newTransaction, err := services.CreateTransaction(ctx, body)
	if err != nil {
		if errors.Is(err, models.ErrForeignKey) {
			respondError(c, http.StatusNotFound, err)
		} else {
			respondError(c, http.StatusInternalServerError, err)
		}
		return
	}

	respondSuccess(c, http.StatusCreated, newTransaction)
}

// PUT: /transactions/:id
//
// Update the info of the transaction with the given id
func PutTransaction(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	var body models.PutTransactionBody
	if err = c.ShouldBindJSON(&body); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	updatedTran, err := services.UpdateTransaction(ctx, int64(id), body)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			respondError(c, http.StatusNotFound, err)
		} else {
			respondError(c, http.StatusInternalServerError, err)
		}
		return
	}

	respondSuccess(c, http.StatusAccepted, updatedTran)
}

// DELETE: /transactions/:id
//
// Delete the transaction with given ID
func DeleteTransaction(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	if err := services.DeleteTransaction(ctx, int64(id)); err != nil {
		if errors.Is(err, models.ErrNotFound) {
			respondError(c, http.StatusNotFound, err)
		} else {
			respondError(c, http.StatusInternalServerError, err)
		}
		return
	}

	respondSuccess(c, http.StatusAccepted, "Transaction deleted!")
}
