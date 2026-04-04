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

// GET: /bills
//
// Return the list of bills of the user
func GetBills(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	bills, err := services.ListBills(ctx, UserID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondSuccess(c, http.StatusOK, bills)
}

// POST: /bills
//
// Create the new bill of the account
func PostBill(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var body models.BillBody
	if err := c.ShouldBindJSON(&body); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	newBill, err := services.CreateBill(ctx, body)
	if err != nil {
		if errors.Is(err, models.ErrForeignKey) {
			respondError(c, http.StatusNotFound, err)
		} else {
			respondError(c, http.StatusInternalServerError, err)
		}
		return
	}

	respondSuccess(c, http.StatusCreated, newBill)
}

// PUT: /bills/:id
//
// Update the bill's details with the given Id
func PutBill(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	var body models.BillBody
	if err = c.ShouldBindJSON(&body); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	updatedBill, err := services.UpdateBill(ctx, int64(id), body)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			respondError(c, http.StatusNotFound, err)
		} else {
			respondError(c, http.StatusInternalServerError, err)
		}
		return
	}

	respondSuccess(c, http.StatusAccepted, updatedBill)
}

// DELETE: /bill/:id
//
// Delete the bill with given ID
func DeleteBill(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	var pay, recurring bool
	if c.Query("pay") != "" {
		recurring, err = strconv.ParseBool(c.Query("recurring"))
		if err != nil {
			respondError(c, http.StatusBadRequest, err)
		}
	}
	if c.Query("recurring") != "" {
		recurring, err = strconv.ParseBool(c.Query("recurring"))
		if err != nil {
			respondError(c, http.StatusBadRequest, err)
		}
	}

	if err := services.DeleteBill(ctx, int64(id), pay, recurring); err != nil {
		if errors.Is(err, models.ErrNotFound) {
			respondError(c, http.StatusNotFound, err)
		} else {
			respondError(c, http.StatusInternalServerError, err)
		}
		return
	}

	respondSuccess(c, http.StatusAccepted, "Bill deleted!")
}
