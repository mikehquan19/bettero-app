package models

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
	ID                 int64      `json:"id" db:"id"`
	UserID             int64      `json:"user_id" db:"user_id"`
	AccNumber          int64      `json:"acc_number" db:"acc_number"`
	AccName            string     `json:"acc_name" db:"acc_name"`
	Institution        string     `json:"institution" db:"institution"`
	Type               string     `json:"type" db:"type"`
	Balance            float64    `json:"balance" db:"balance"`
	CreditLimit        *float64   `json:"credit_limit,omitempty" db:"credit_limit"`
	NextDue            *time.Time `json:"next_due,omitempty" db:"next_due"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
	DiscrepancyFlagged bool       `json:"discrepancy_flagged" db:"discrepancy_flagged"`
	DiscrepancyAmount  float64    `json:"discrepancy_amount" db:"discrepancy_amount"`
}

type PostAccountBody struct {
	AccNumber   int64      `json:"acc_number"`
	AccName     string     `json:"acc_name"`
	Institution string     `json:"institution"`
	Type        string     `json:"type"`
	Balance     float64    `json:"balance"`
	CreditLimit *float64   `json:"credit_limit"`
	NextDue     *time.Time `json:"next_due"`
}

func (p PostAccountBody) Validate() error {
	if p.Type == "Debit" && (p.NextDue != nil || p.CreditLimit != nil) {
		return ErrDebitCardWithCreditInfo
	}
	if p.Type == "Credit" && (p.NextDue == nil || p.CreditLimit == nil) {
		return ErrCreditCardWithoutCreditInfo
	}

	return nil
}

// User is not allowed to update the type of the account
type PutAccountBody struct {
	AccNumber   int64    `json:"acc_number"`
	AccName     string   `json:"acc_name"`
	Institution string   `json:"institution"`
	Balance     float64  `json:"balance"`
	CreditLimit *float64 `json:"credit_limit"`
	NextDue     *string  `json:"next_due"`
}

func (p PutAccountBody) Validate(accType string) error {
	if accType == "Debit" && (p.NextDue != nil || p.CreditLimit != nil) {
		return ErrDebitCardWithCreditInfo
	}
	if accType == "Credit" && (p.NextDue == nil || p.CreditLimit == nil) {
		return ErrCreditCardWithoutCreditInfo
	}

	return nil
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
		&acc.DiscrepancyFlagged,
		&acc.DiscrepancyAmount,
	)
	return err
}

type AccountHistory struct {
	ID         int64     `json:"id" db:"id"`
	AccountId  int64     `json:"account_id" db:"account_id"`
	LoggedTime time.Time `json:"logged_time" db:"logged_time"`
	Balance    float64   `json:"balance" db:"balance"`
}

type PostAccHistBody struct {
	AccountId  int64     `json:"account_id" db:"account_id"`
	LoggedTime time.Time `json:"logged_time" db:"logged_time"`
	Balance    float64   `json:"balance" db:"balance"`
}

// ScanAccHistory parses the returned db row into account history struct and destinations
func ScanAccHistory(accHistRow pgx.Row, accHist *AccountHistory) error {
	err := accHistRow.Scan(
		&accHist.ID,
		&accHist.AccountId,
		&accHist.LoggedTime,
		&accHist.Balance,
	)
	return err
}

type PaginatedResponse[T any] struct {
	Total  int `json:"total"`
	Offset int `json:"offset"`
	Data   []T `json:"data"`
}

type Transaction struct {
	ID              int64         `json:"id" db:"id"`
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

type PostTransactionBody struct {
	AccountID       int64     `json:"account_id"`
	Merchant        string    `json:"merchant"`
	TranDescription string    `json:"tran_description"`
	Category        string    `json:"category"`
	Amount          float64   `json:"amount"`
	CreatedAt       time.Time `json:"created_at"`
}

// User is not allowed to update account's Id
type PutTransactionBody struct {
	Merchant        string    `json:"merchant"`
	TranDescription string    `json:"tran_description"`
	Category        string    `json:"category"`
	Amount          float64   `json:"amount"`
	CreatedAt       time.Time `json:"created_at"`
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

type TransactionFilter struct {
	Category        string
	Merchant        string
	TranDescription string
	CreatedAtFrom   *time.Time
	CreatedAtTo     *time.Time
}

type Bill struct {
	ID          int64         `json:"id" db:"id"`
	Account     NestedAccount `json:"account" db:"account"`
	Merchant    string        `json:"merchant" db:"merchant"`
	Description string        `json:"description" db:"description"`
	Category    string        `json:"category" db:"category"`
	Amount      float64       `json:"amount" db:"amount"`
	DueDate     time.Time     `json:"due_date" db:"due_date"`
}

type BillBody struct {
	AccountID   int64     `json:"account_id"`
	Merchant    string    `json:"merchant"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Amount      float64   `json:"amount"`
	DueDate     time.Time `json:"due_date"`
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

// Account financial summary doesn't need the basic info
type AccountFinancialSummary struct {
	Daily       map[string]float64  `json:"daily"`
	Change      map[string]*float64 `json:"change"`
	Composition map[string]float64  `json:"composition"`
}

type SummaryDates struct {
	CurrStart time.Time
	CurrEnd   time.Time
	PrevStart time.Time
	PrevEnd   time.Time
}

type CategoryProgress struct {
	Current    float64 `json:"current"`
	Budget     float64 `json:"budget"`
	Percentage float64 `json:"percentage"`
}

type BudgetComposition struct {
	Goal map[string]float64 `json:"goal"` // Goal composition, which is the category portion
	Real map[string]float64 `json:"real"` // Real composition, computed from GetCompositionMap logic
}

// A detailed analysis of user's spending analysis and budget's info
type BudgetResponse struct {
	ID                int64                        `json:"id"`
	IntervalType      string                       `json:"interval_type"`
	RecurringIncome   float64                      `json:"recurring_income"`
	ExpensePortion    float64                      `json:"expense_portion"`
	BudgetComposition BudgetComposition            `json:"budget_composition"`
	Progress          map[string]*CategoryProgress `json:"progress"`
	CreatedAt         time.Time                    `json:"created_at"`
	UpdatedAt         time.Time                    `json:"updated_at"`
}

type BudgetPlan struct {
	ID              int64              `json:"id" db:"id"`
	UserID          int64              `json:"user_id" db:"user_id"`
	IntervalType    string             `json:"interval_type" db:"interval_type"`
	RecurringIncome float64            `json:"recurring_income" db:"recurring_income"`
	ExpensePortion  float64            `json:"expense_portion" db:"expense_portion"`
	CategoryPortion map[string]float64 `json:"category_portion" db:"category_portion"`
	CreatedAt       time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at" db:"updated_at"`
}

type GenericBudgetPlanBody struct {
	RecurringIncome float64            `json:"recurring_income"`
	ExpensePortion  float64            `json:"expense_portion"`
	CategoryPortion map[string]float64 `json:"category_portion"`
}

type PostBudgetPlanBody struct {
	IntervalType string `json:"interval_type"`
	GenericBudgetPlanBody
}

// BudgetPlanbody does not allow for updating the interval type
type PutBudgetPlanBody struct {
	GenericBudgetPlanBody
}

func ScanBudgetPlan(budgetPlanRow pgx.Row, budgetPlan *BudgetPlan) error {
	err := budgetPlanRow.Scan(
		&budgetPlan.ID,
		&budgetPlan.UserID,
		&budgetPlan.IntervalType,
		&budgetPlan.RecurringIncome,
		&budgetPlan.ExpensePortion,
		&budgetPlan.CategoryPortion,
		&budgetPlan.CreatedAt,
		&budgetPlan.UpdatedAt,
	)
	return err
}

type ObjectType string

const (
	UserObj    ObjectType = "user"
	AccountObj ObjectType = "account"
)

type IntervalType string

const (
	Month  IntervalType = "month"
	BiWeek IntervalType = "bi_week"
	Week   IntervalType = "week"
)

// The overal result of the autocomplete search
type Suggestion struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type APIKey struct{}

// DBTX represents a database connection that can execute queries
// and start transactions. It is implemented by both pgxpool.Pool and pgx.Tx,
// allowing repositories and services to operate on
// either a direct database connection or an active transaction.
type DBTX interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}
