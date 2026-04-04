package models

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("resource not found")
var ErrForeignKey = errors.New("references a non-existent resource")

var ErrInvalidAccountBody = errors.New("invalid account body")
var ErrDebitCardWithCreditInfo = fmt.Errorf("debit card with due date or credit limit, %w", ErrInvalidAccountBody)
var ErrCreditCardWithoutCreditInfo = fmt.Errorf("credit card without due date or credit limit, %w", ErrInvalidAccountBody)
