package models

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound          = errors.New("resource not found")
	ErrForeignKey        = errors.New("references a non-existent resource")
	ErrTransactionTooOld = errors.New("transaction too old to be taken actions on")
)

var (
	ErrInvalidAccountBody          = errors.New("invalid account body")
	ErrDebitCardWithCreditInfo     = fmt.Errorf("%w, debit card with due date or credit limit", ErrInvalidAccountBody)
	ErrCreditCardWithoutCreditInfo = fmt.Errorf("%w, credit card without due date or credit limit", ErrInvalidAccountBody)
)

var (
	ErrInvalidBudgetBody           = errors.New("invalid budget body")
	ErrInvalidCategory             = fmt.Errorf("%w, category is invalid", ErrInvalidBudgetBody)
	ErrInvalidExpensePercentage    = fmt.Errorf("%w, expense portion is invalid percentage", ErrInvalidBudgetBody)
	ErrInvalidCategoryPercentage   = fmt.Errorf("%w, category portion is invalid percentage", ErrInvalidBudgetBody)
	ErrCategoryPercentagesNotAddUp = fmt.Errorf("%w, category portions not adding up to 100 percent", ErrInvalidBudgetBody)
)
