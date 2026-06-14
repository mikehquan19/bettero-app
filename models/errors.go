package models

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

var (
	ErrNotFound          = errors.New("resource not found")
	ErrForeignKey        = errors.New("references a non-existent resource")
	ErrTransactionTooOld = errors.New("transaction too old to be taken actions on")
)

func GetNotFound[T any](id int64) error {
	resourceType := strings.ToLower(reflect.TypeFor[T]().String())
	return fmt.Errorf("%s %d not found, %w", resourceType, id, ErrNotFound)
}

func GetForeignKey[T any](id int64) error {
	resourceType := strings.ToLower(reflect.TypeFor[T]().String())
	return fmt.Errorf("%s %d not found, %w", resourceType, id, ErrForeignKey)
}

var ErrInvalidAccountBody = errors.New("invalid account body")
var ErrDebitCardWithCreditInfo = fmt.Errorf("debit card with due date or credit limit, %w", ErrInvalidAccountBody)
var ErrCreditCardWithoutCreditInfo = fmt.Errorf("credit card without due date or credit limit, %w", ErrInvalidAccountBody)
