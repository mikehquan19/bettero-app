package controllers

import (
	"betterov2/models"
	"betterov2/services"
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

var UserID int64 = 1

// GET: /accounts
//
// Return the list of debit and credit accounts of the user
func GetAccounts(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	accounts, err := services.ListAccounts(ctx, UserID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
	}

	respondSuccess(c, http.StatusOK, accounts)
}

// GET: /accounts/:id
//
// Return the detail the account
func GetAccount(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	account, err := services.GetAccount(ctx, int64(id))
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			respondError(c, http.StatusNotFound, err)
		} else {
			respondError(c, http.StatusInternalServerError, err)
		}
		return
	}

	respondSuccess(c, http.StatusOK, account)
}

// POST: /accounts
//
// Create new account of the user
func PostAccounts(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var body models.PostAccountBody

	if err := c.ShouldBindJSON(&body); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	newAccount, err := services.CreateAccount(ctx, UserID, body)
	if err != nil {
		if errors.Is(err, models.ErrInvalidAccountBody) {
			respondError(c, http.StatusBadRequest, err)
		} else if errors.Is(err, models.ErrForeignKey) {
			respondError(c, http.StatusNotFound, err)
		} else {
			respondError(c, http.StatusInternalServerError, err)
		}
		return
	}

	respondSuccess(c, http.StatusCreated, newAccount)
}

// PUT: /accounts/:id
//
// Update the information of the account (including the amount).
func PutAccount(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	var body models.PutAccountBody
	if err := c.ShouldBindJSON(&body); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	updatedAccount, err := services.UpdateAccount(ctx, int64(id), body)
	if err != nil {
		if errors.Is(err, models.ErrInvalidAccountBody) {
			respondError(c, http.StatusBadRequest, err)
		} else if errors.Is(err, models.ErrNotFound) {
			respondError(c, http.StatusNotFound, err)
		} else {
			respondError(c, http.StatusInternalServerError, err)
		}
		return
	}

	respondSuccess(c, http.StatusAccepted, updatedAccount)
}

// DELETE: /account/:id
//
// Delete the account with a given ID
func DeleteAccount(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	if err = services.DeleteAccount(ctx, int64(id)); err != nil {
		if errors.Is(err, models.ErrNotFound) {
			respondError(c, http.StatusNotFound, err)
		} else {
			respondError(c, http.StatusInternalServerError, err)
		}
		return
	}

	respondSuccess(c, http.StatusAccepted, "Account deleted!")
}

// GET: /accounts/:id/transactions
//
// Return the paginated list of transactions of an account
func GetAccountTransactions(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}
	offset, err := getOffset(c)
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	total, transactions, err := services.ListAccountTransactions(ctx, int64(id), int64(offset))
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
