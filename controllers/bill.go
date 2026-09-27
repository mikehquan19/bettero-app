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
	"github.com/google/uuid"
)

type BillController struct {
	billService *services.BillService
}

func NewBillController(s *services.BillService) *BillController {
	return &BillController{
		billService: s,
	}
}

// GET: /bills
//
// Return the list of bills of the user
func (t *BillController) GetBills(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	bills, err := t.billService.ListBills(ctx, UserID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondSuccess(c, http.StatusOK, bills)
}

// POST: /bills
//
// Create the new bill of the account
func (t *BillController) PostBill(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var body models.BillBody
	if err := c.ShouldBindJSON(&body); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	newBill, err := t.billService.CreateBill(ctx, body)
	if err != nil {
		if errors.Is(err, models.ErrForeignKey) {
			respondError(c, http.StatusNotFound, err)
			return
		}
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondSuccess(c, http.StatusCreated, newBill)
}

// PUT: /bills/:id
//
// Update the bill's details with the given Id
func (t *BillController) PutBill(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	var body models.BillBody
	if err = c.ShouldBindJSON(&body); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	updatedBill, err := t.billService.UpdateBill(ctx, id, body)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			respondError(c, http.StatusNotFound, err)
			return
		}
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondSuccess(c, http.StatusAccepted, updatedBill)
}

// DELETE: /bill/:id
//
// Delete the bill with given ID
func (t *BillController) DeleteBill(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	var pay, recurring bool
	// Specify if the user pays the bill or just deletes
	if c.Query("pay") != "" {
		pay, err = strconv.ParseBool(c.Query("pay"))
		if err != nil {
			respondError(c, http.StatusBadRequest, err)
			return
		}
	}
	// Specify if the user want the bill to recur
	if c.Query("recurring") != "" {
		recurring, err = strconv.ParseBool(c.Query("recurring"))
		if err != nil {
			respondError(c, http.StatusBadRequest, err)
			return
		}
	}

	if err := t.billService.DeleteBill(ctx, id, pay, recurring); err != nil {
		if errors.Is(err, models.ErrNotFound) {
			respondError(c, http.StatusNotFound, err)
			return
		}
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondSuccess(c, http.StatusAccepted, "Bill deleted!")
}
