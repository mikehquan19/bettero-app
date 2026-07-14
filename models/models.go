package models

import (
	"time"
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

type AccountType string

const (
	Debit  AccountType = "Debit"
	Credit AccountType = "Credit"
)

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
	ID                 int64       `json:"id" db:"id"`
	UserID             int64       `json:"user_id" db:"user_id"`
	AccNumber          int64       `json:"acc_number" db:"acc_number"`
	AccName            string      `json:"acc_name" db:"acc_name"`
	Institution        string      `json:"institution" db:"institution"`
	Type               AccountType `json:"type" db:"type"`
	Balance            float64     `json:"balance" db:"balance"`
	CreditLimit        *float64    `json:"credit_limit,omitempty" db:"credit_limit"`
	NextDue            *time.Time  `json:"next_due,omitempty" db:"next_due"`
	CreatedAt          time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time   `json:"updated_at" db:"updated_at"`
	DiscrepancyFlagged bool        `json:"discrepancy_flagged" db:"discrepancy_flagged"`
	DiscrepancyAmount  float64     `json:"discrepancy_amount" db:"discrepancy_amount"`
}

type AccountBody struct {
	AccNumber   int64      `json:"acc_number"`
	AccName     string     `json:"acc_name"`
	Institution string     `json:"institution"`
	Balance     float64    `json:"balance"`
	CreditLimit *float64   `json:"credit_limit"`
	NextDue     *time.Time `json:"next_due"`
}

type PostAccountBody struct {
	AccountBody
	Type AccountType `json:"type"`
}

// User is not allowed to update the type of the account
type PutAccountBody struct {
	AccountBody
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

type PaginatedResponse[T any] struct {
	Total  int `json:"total"`
	Offset int `json:"offset"`
	Data   []T `json:"data"`
}

type TransactionCategory string

const (
	Housing      TransactionCategory = "Housing"
	Automobile   TransactionCategory = "Automobile"
	Medical      TransactionCategory = "Medical"
	Subscription TransactionCategory = "Subscription"
	Grocery      TransactionCategory = "Grocery"
	Dining       TransactionCategory = "Dining"
	Shopping     TransactionCategory = "Shopping"
	Gas          TransactionCategory = "Gas"
	Others       TransactionCategory = "Others"
	Income       TransactionCategory = "Income"
)

var TransactionCategories = []TransactionCategory{
	Housing,
	Automobile,
	Medical,
	Subscription,
	Grocery,
	Dining,
	Shopping,
	Gas,
	Others,
}

type Transaction struct {
	ID              int64               `json:"id" db:"id"`
	Account         NestedAccount       `json:"account" db:"account"`
	Merchant        string              `json:"merchant" db:"merchant"`
	TranDescription string              `json:"tran_description" db:"tran_description"`
	Category        TransactionCategory `json:"category" db:"category"`
	Amount          float64             `json:"amount" db:"amount"`
	CreatedAt       time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at" db:"updated_at"`
}

// A shortened version of account that is used in nested transaction
type NestedAccount struct {
	Id          int64       `json:"id" db:"id"`
	AccNumber   int64       `json:"acc_number" db:"acc_number"`
	AccName     string      `json:"acc_name" db:"acc_name"`
	Institution string      `json:"institution" db:"institution"`
	Type        AccountType `json:"type" db:"type"`
}

type PostTransactionBody struct {
	AccountID       int64               `json:"account_id"`
	Merchant        string              `json:"merchant"`
	TranDescription string              `json:"tran_description"`
	Category        TransactionCategory `json:"category"`
	Amount          float64             `json:"amount"`
	CreatedAt       time.Time           `json:"created_at"`
}

// User is not allowed to update account's Id
type PutTransactionBody struct {
	Merchant        string              `json:"merchant"`
	TranDescription string              `json:"tran_description"`
	Category        TransactionCategory `json:"category"`
	Amount          float64             `json:"amount"`
	CreatedAt       time.Time           `json:"created_at"`
}

type TransactionFilter struct {
	Category        TransactionCategory
	Merchant        string
	TranDescription string
	CreatedAtFrom   *time.Time
	CreatedAtTo     *time.Time
}

type Bill struct {
	ID          int64               `json:"id" db:"id"`
	Account     NestedAccount       `json:"account" db:"account"`
	Merchant    string              `json:"merchant" db:"merchant"`
	Description string              `json:"description" db:"description"`
	Category    TransactionCategory `json:"category" db:"category"`
	Amount      float64             `json:"amount" db:"amount"`
	DueDate     time.Time           `json:"due_date" db:"due_date"`
}

type BillBody struct {
	AccountID   int64               `json:"account_id"`
	Merchant    string              `json:"merchant"`
	Description string              `json:"description"`
	Category    TransactionCategory `json:"category"`
	Amount      float64             `json:"amount"`
	DueDate     time.Time           `json:"due_date"`
}

type BasicAnalysis struct {
	TotalBalance   float64 `json:"total_balance"`
	TotalAmountDue float64 `json:"total_amount_due"`
	TotalIncome    float64 `json:"total_income"`
	TotalExpense   float64 `json:"total_expense"`
}

type FinancialSummary struct {
	Basic       BasicAnalysis                    `json:"basic"`
	Daily       map[string]float64               `json:"daily"`
	Change      map[TransactionCategory]*float64 `json:"change"`
	Composition map[TransactionCategory]float64  `json:"composition"`
}

// Account financial summary doesn't need the basic info
type AccountFinancialSummary struct {
	Daily       map[string]float64               `json:"daily"`
	Change      map[TransactionCategory]*float64 `json:"change"`
	Composition map[TransactionCategory]float64  `json:"composition"`
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
	Goal map[TransactionCategory]float64 `json:"goal"` // Goal composition, which is the category portion
	Real map[TransactionCategory]float64 `json:"real"` // Real composition, computed from GetCompositionMap logic
}

// A detailed analysis of user's spending analysis and budget's info
type BudgetResponse struct {
	ID                int64                                     `json:"id"`
	IntervalType      string                                    `json:"interval_type"`
	RecurringIncome   float64                                   `json:"recurring_income"`
	ExpensePortion    float64                                   `json:"expense_portion"`
	BudgetComposition BudgetComposition                         `json:"budget_composition"`
	Progress          map[TransactionCategory]*CategoryProgress `json:"progress"`
	CreatedAt         time.Time                                 `json:"created_at"`
	UpdatedAt         time.Time                                 `json:"updated_at"`
}

type BudgetPlan struct {
	ID              int64                           `json:"id" db:"id"`
	UserID          int64                           `json:"user_id" db:"user_id"`
	IntervalType    string                          `json:"interval_type" db:"interval_type"`
	RecurringIncome float64                         `json:"recurring_income" db:"recurring_income"`
	ExpensePortion  float64                         `json:"expense_portion" db:"expense_portion"`
	CategoryPortion map[TransactionCategory]float64 `json:"category_portion" db:"category_portion"`
	CreatedAt       time.Time                       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time                       `json:"updated_at" db:"updated_at"`
}

type GenericBudgetPlanBody struct {
	RecurringIncome float64                         `json:"recurring_income"`
	ExpensePortion  float64                         `json:"expense_portion"`
	CategoryPortion map[TransactionCategory]float64 `json:"category_portion"`
}

type PostBudgetPlanBody struct {
	IntervalType string `json:"interval_type"`
	GenericBudgetPlanBody
}

// BudgetPlanbody does not allow for updating the interval type
type PutBudgetPlanBody struct {
	GenericBudgetPlanBody
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
