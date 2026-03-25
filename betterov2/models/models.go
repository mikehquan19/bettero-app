package models

import (
	"time"

	"github.com/jackc/pgx/v5"
)

type Response struct {
	Err  string `json:"error"`
	Data any    `json:"data"`
}

type User struct {
	ID           int64     `json:"id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	UserPassword string    `json:"user_password"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Financial account (card) of the user. There are 2 types of financial accounts:
//
// For debit card:
//   - Balance represents the amount of money the user HAS.
//   - Debit account doesn't have credit limit and next due date.
//
// For credit card:
//   - Balance represents the amount of money the users OWES.
//   - Credit limit and next due date is required.
//   - Balance is allowed to exceed credit limit, but there will be a message notifying in app
//   - Next due date will be updated monthly if today is past the current due date.
type Account struct {
	ID          int64      `json:"id" db:"id"`
	UserID      int64      `json:"user_id" db:"user_id"`
	AccNumber   int64      `json:"acc_number" db:"acc_number"`
	AccName     string     `json:"acc_name" db:"acc_name"`
	Institution string     `json:"institution" db:"institution"`
	Type        string     `json:"type" db:"type"`
	Balance     float64    `json:"balance" db:"balance"`
	CreditLimit *float64   `json:"credit_limit,omitempty" db:"credit_limit"`
	NextDue     *time.Time `json:"next_due,omitempty" db:"next_due"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

type AccountBody struct {
	AccNumber   int64    `json:"acc_number"`
	AccName     string   `json:"acc_name"`
	Institution string   `json:"institution"`
	Type        string   `json:"type"`
	Balance     float64  `json:"balance"`
	CreditLimit *float64 `json:"credit_limit"`
	NextDue     *string  `json:"next_due"`
}

// ScanAccount parses the returned db row into account struct and destinations
func ScanAccount(accRow pgx.Row, acc *Account) error {
	err := accRow.Scan(
		&acc.ID,
		&acc.UserID,
		&acc.AccNumber,
		&acc.AccName,
		&acc.Institution,
		&acc.Type,
		&acc.Balance,
		&acc.CreditLimit,
		&acc.NextDue,
		&acc.CreatedAt,
		&acc.UpdatedAt,
	)
	return err
}

type Transaction struct {
	ID              int           `json:"id" db:"id"`
	Account         NestedAccount `json:"account" db:"account"`
	Merchant        string        `json:"merchant" db:"merchant"`
	TranDescription string        `json:"tran_description" db:"tran_description"`
	Category        string        `json:"category" db:"category"`
	Amount          float64       `json:"amount" db:"amount"`
	CreatedAt       time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at" db:"updated_at"`
}

// A shortened version of account that is used in nested transaction
type NestedAccount struct {
	Id          int64  `json:"id" db:"id"`
	AccNumber   int64  `json:"acc_number" db:"acc_number"`
	AccName     string `json:"acc_name" db:"acc_name"`
	Institution string `json:"institution" db:"institution"`
	Type        string `json:"type" db:"type"`
}

type TransactionBody struct {
	AccountID       int     `json:"account_id"`
	Merchant        string  `json:"merchant"`
	TranDescription string  `json:"tran_description"`
	Category        string  `json:"category"`
	Amount          float64 `json:"amount"`
	CreatedAt       string  `json:"created_at"`
}

// ScanTransaction parses the returned row into transaction and destinations
func ScanTransaction(tranRow pgx.Row, tran *Transaction) error {
	err := tranRow.Scan(
		&tran.ID,
		&tran.Account,
		&tran.Merchant,
		&tran.TranDescription,
		&tran.Category,
		&tran.Amount,
		&tran.CreatedAt,
		&tran.UpdatedAt,
	)
	return err
}

type Bill struct {
	ID          int           `json:"id" db:"id"`
	Account     NestedAccount `json:"account" db:"account"`
	Merchant    string        `json:"merchant" db:"merchant"`
	Description string        `json:"description" db:"description"`
	Category    string        `json:"category" db:"category"`
	Amount      float64       `json:"amount" db:"amount"`
	DueDate     time.Time     `json:"due_date" db:"due_date"`
}

type BillBody struct {
	AccountID   int     `json:"account_id"`
	Merchant    string  `json:"merchant"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Amount      float64 `json:"amount"`
	DueDate     string  `json:"due_date"`
}

// ScanBill parses the returned row into bill
func ScanBill(billRow pgx.Row, bill *Bill) error {
	err := billRow.Scan(
		&bill.ID,
		&bill.Account,
		&bill.Merchant,
		&bill.Description,
		&bill.Category,
		&bill.Amount,
		&bill.DueDate,
	)
	return err
}

type BasicAnalysis struct {
	TotalBalance   float64 `json:"total_balance"`
	TotalAmountDue float64 `json:"total_amount_due"`
	TotalIncome    float64 `json:"total_income"`
	TotalExpense   float64 `json:"total_expense"`
}

func ScanAnalysis(analysisRow pgx.Row, analysis *BasicAnalysis) error {
	err := analysisRow.Scan(
		&analysis.TotalBalance,
		&analysis.TotalAmountDue,
		&analysis.TotalIncome,
		&analysis.TotalExpense,
	)
	return err
}

type FinancialSummary struct {
	Basic       BasicAnalysis       `json:"basic"`
	Daily       map[string]float64  `json:"daily"`
	Change      map[string]*float64 `json:"change"`
	Composition map[string]float64  `json:"composition"`
}
