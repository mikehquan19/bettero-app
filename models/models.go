package models

import (
	"time"

	"github.com/google/uuid"
)

type Money int64

type Response struct {
	Err  string `json:"error"`
	Data any    `json:"data"`
}

type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	FirstName string    `json:"first_name" db:"first_name"`
	LastName  string    `json:"last_name" db:"last_name"`
	Username  string    `json:"username" db:"username"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password_hash"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
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
	ID                     uuid.UUID   `json:"id" db:"id"`
	UserID                 uuid.UUID   `json:"user_id" db:"user_id"`
	AccountNumber          int64       `json:"account_number" db:"account_number"`
	Name                   string      `json:"name" db:"name"`
	Institution            string      `json:"institution" db:"institution"`
	Type                   AccountType `json:"type" db:"type"`
	BalanceCents           Money       `json:"balance_cents" db:"balance_cents"`
	CreditLimitCents       *Money      `json:"credit_limit_cents,omitempty" db:"credit_limit_cents"`
	NextDue                *time.Time  `json:"next_due,omitempty" db:"next_due"`
	DiscrepancyFlagged     bool        `json:"discrepancy_flagged" db:"discrepancy_flagged"`
	DiscrepancyAmountCents Money       `json:"discrepancy_amount_cents" db:"discrepancy_amount_cents"`
	CreatedAt              time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time   `json:"updated_at" db:"updated_at"`
}

type AccountBody struct {
	AccountNumber    int64      `json:"account_number"`
	Name             string     `json:"name"`
	Institution      string     `json:"institution"`
	BalanceCents     Money      `json:"balance_cents"`
	CreditLimitCents *Money     `json:"credit_limit_cents"`
	NextDue          *time.Time `json:"next_due"`
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
	ID           uuid.UUID `json:"id" db:"id"`
	AccountID    uuid.UUID `json:"account_id" db:"account_id"`
	LoggedTime   time.Time `json:"logged_time" db:"logged_time"`
	BalanceCents Money     `json:"balance_cents" db:"balance_cents"`
}

type PostAccHistBody struct {
	AccountID    uuid.UUID `json:"account_id" db:"account_id"`
	LoggedTime   time.Time `json:"logged_time" db:"logged_time"`
	BalanceCents Money     `json:"balance_cents" db:"balance_cents"`
}

type PaginatedResponse[T any] struct {
	Total  int `json:"total"`
	Offset int `json:"offset"`
	Data   []T `json:"data"`
}

type TransactionCategory string

const (
	Housing      TransactionCategory = "housing"
	Automobile   TransactionCategory = "automobile"
	Medical      TransactionCategory = "medical"
	Subscription TransactionCategory = "subscription"
	Grocery      TransactionCategory = "grocery"
	Dining       TransactionCategory = "dining"
	Shopping     TransactionCategory = "shopping"
	Gas          TransactionCategory = "gas"
	Others       TransactionCategory = "others"
	Income       TransactionCategory = "income"
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
	ID          uuid.UUID           `json:"id" db:"id"`
	Account     NestedAccount       `json:"account" db:"account"`
	Merchant    string              `json:"merchant" db:"merchant"`
	Description string              `json:"description" db:"description"`
	Category    TransactionCategory `json:"category" db:"category"`
	AmountCents Money               `json:"amount_cents" db:"amount_cents"`
	CreatedAt   time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at" db:"updated_at"`
}

// A shortened version of account that is used in nested transaction
type NestedAccount struct {
	ID            uuid.UUID   `json:"id"`
	AccountNumber int64       `json:"account_number"`
	Name          string      `json:"name"`
	Institution   string      `json:"institution"`
	Type          AccountType `json:"type"`
}

type PostTransactionBody struct {
	AccountID   uuid.UUID           `json:"account_id"`
	Merchant    string              `json:"merchant"`
	Description string              `json:"description"`
	Category    TransactionCategory `json:"category"`
	AmountCents Money               `json:"amount_cents"`
	CreatedAt   time.Time           `json:"created_at"`
}

// User is not allowed to update account's Id
type PutTransactionBody struct {
	Merchant    string              `json:"merchant"`
	Description string              `json:"description"`
	Category    TransactionCategory `json:"category"`
	AmountCents Money               `json:"amount_cents"`
	CreatedAt   time.Time           `json:"created_at"`
}

type TransactionFilter struct {
	Category        TransactionCategory
	Merchant        string
	TranDescription string
	CreatedAtFrom   *time.Time
	CreatedAtTo     *time.Time
}

type Bill struct {
	ID          uuid.UUID           `json:"id" db:"id"`
	Account     *NestedAccount      `json:"account,omitempty" db:"account"`
	Merchant    string              `json:"merchant" db:"merchant"`
	Description string              `json:"description" db:"description"`
	Category    TransactionCategory `json:"category" db:"category"`
	AmountCents Money               `json:"amount_cents" db:"amount_cents"`
	DueDate     time.Time           `json:"due_date" db:"due_date"`
}

type BillBody struct {
	AccountID   uuid.UUID           `json:"account_id"`
	Merchant    string              `json:"merchant"`
	Description string              `json:"description"`
	Category    TransactionCategory `json:"category"`
	AmountCents Money               `json:"amount_cents"`
	DueDate     time.Time           `json:"due_date"`
}

type BasicAnalysis struct {
	TotalBalanceCents   Money `json:"total_balance_cents"`
	TotalAmountDueCents Money `json:"total_amount_due_cents"`
	TotalIncomeCents    Money `json:"total_income_cents"`
	TotalExpenseCents   Money `json:"total_expense_cents"`
}

type FinancialSummary struct {
	Basic       BasicAnalysis                    `json:"basic"`
	Daily       map[string]Money                 `json:"daily"`
	Change      map[TransactionCategory]*float64 `json:"change"`
	Composition map[TransactionCategory]float64  `json:"composition"`
}

// Account financial summary doesn't need the basic info
type AccountFinancialSummary struct {
	Daily       map[string]Money                 `json:"daily"`
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
	CurrentCents Money   `json:"current_cents"`
	BudgetCents  Money   `json:"budget_cents"`
	Percentage   float64 `json:"percentage"`
}

type BudgetComposition struct {
	Goal map[TransactionCategory]float64 `json:"goal"` // Goal composition, which is the category portion
	Real map[TransactionCategory]float64 `json:"real"` // Real composition, computed from GetCompositionMap logic
}

// A detailed analysis of user's spending analysis and budget's info
type BudgetResponse struct {
	ID                   uuid.UUID                                 `json:"id"`
	IntervalType         string                                    `json:"interval_type"`
	RecurringIncomeCents Money                                     `json:"recurring_income_cents"`
	ExpensePortion       float64                                   `json:"expense_portion"`
	BudgetComposition    BudgetComposition                         `json:"budget_composition"`
	Progress             map[TransactionCategory]*CategoryProgress `json:"progress"`
	CreatedAt            time.Time                                 `json:"created_at"`
	UpdatedAt            time.Time                                 `json:"updated_at"`
}

type BudgetPlan struct {
	ID                   uuid.UUID                       `json:"id" db:"id"`
	UserID               uuid.UUID                       `json:"user_id" db:"user_id"`
	IntervalType         string                          `json:"interval_type" db:"interval_type"`
	RecurringIncomeCents Money                           `json:"recurring_income_cents" db:"recurring_income_cents"`
	ExpensePortion       float64                         `json:"expense_portion" db:"expense_portion"`
	CategoryPortion      map[TransactionCategory]float64 `json:"category_portion" db:"category_portion"`
	CreatedAt            time.Time                       `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time                       `json:"updated_at" db:"updated_at"`
}

type GenericBudgetPlanBody struct {
	RecurringIncomeCents Money                           `json:"recurring_income_cents"`
	ExpensePortion       float64                         `json:"expense_portion"`
	CategoryPortion      map[TransactionCategory]float64 `json:"category_portion"`
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
