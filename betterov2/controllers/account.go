package controllers

import (
	"betterov2/models"
	"betterov2/services"
	"context"
	"fmt"
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
		if err.Error() == fmt.Sprintf("Account %d not found", id) {
			respondError(c, http.StatusBadRequest, err)
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

	var (
		body    models.AccountBody
		nextDue *time.Time
	)
	if err := c.ShouldBindJSON(&body); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}
	if body.NextDue != nil && *body.NextDue != "" {
		t, err := time.Parse("2006-01-02", *body.NextDue)
		if err != nil {
			respondError(c, http.StatusBadRequest, err)
			return
		}
		nextDue = &t
	}

	newAccount, err := services.CreateAccount(ctx, UserID, body, nextDue)
	if err != nil {
		if err.Error() == fmt.Sprintf("User %d not found.", UserID) {
			respondError(c, http.StatusBadRequest, err)
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

	var body models.AccountBody
	var nextDue *time.Time
	if err := c.ShouldBindJSON(&body); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}
	if body.NextDue != nil && *body.NextDue != "" {
		parsed, err := time.Parse("2006-01-02", *body.NextDue)
		if err != nil {
			respondError(c, http.StatusBadRequest, err)
			return
		}
		nextDue = &parsed
	}

	updatedAccount, err := services.UpdateAccount(ctx, int64(id), body, nextDue)
	if err != nil {
		if err.Error() == fmt.Sprintf("Account %d not found", id) {
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
		if err.Error() == fmt.Sprintf("Account %d not found", id) {
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

	transactions, err := services.ListAccountTransactions(ctx, int64(id), int64(offset))
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
	}

	respondSuccess(c, http.StatusOK, transactions)
}
