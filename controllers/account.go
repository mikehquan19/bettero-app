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

var UserID int64 = 1

type AccountController struct {
	service *services.AccountService
}

func NewAccountController(s *services.AccountService) *AccountController {
	return &AccountController{
		service: s,
	}
}

// GET: /accounts
//
// Return the list of debit and credit accounts of the user
func (a *AccountController) GetAccounts(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	accounts, err := a.service.ListAccounts(ctx, UserID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
	}

	respondSuccess(c, http.StatusOK, accounts)
}

// GET: /accounts/:id
//
// Return the detail the account
func (a *AccountController) GetAccount(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	account, err := a.service.GetAccount(ctx, int64(id))
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
func (a *AccountController) PostAccounts(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var body models.PostAccountBody

	if err := c.ShouldBindJSON(&body); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	newAccount, err := a.service.CreateAccount(ctx, UserID, body)
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
func (a *AccountController) PutAccount(c *gin.Context) {
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

	updatedAccount, err := a.service.UpdateAccount(ctx, int64(id), body)
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
func (a *AccountController) DeleteAccount(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	if err = a.service.DeleteAccount(ctx, int64(id)); err != nil {
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
func (a *AccountController) GetAccountTransactions(c *gin.Context) {
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

	total, transactions, err := a.service.ListAccountTransactions(ctx, int64(id), filter, int64(offset))
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
