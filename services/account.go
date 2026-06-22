package services

import (
	"betterov2/models"
	"betterov2/repositories"
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Account Service
type AccountService struct {
	db         *pgxpool.Pool
	accRepo    *repositories.AccountRepo
	accHisRepo *repositories.AccountHistoryRepo
	tranRepo   *repositories.TransactionRepo
}

// Generate a new account service
func NewAccountService(
	db *pgxpool.Pool, accRepo *repositories.AccountRepo, accHistRepo *repositories.AccountHistoryRepo, tranRepo *repositories.TransactionRepo,
) *AccountService {
	return &AccountService{
		db:         db,
		accRepo:    accRepo,
		accHisRepo: accHistRepo,
		tranRepo:   tranRepo,
	}
}

// ListAccounts returns the list of accounts of the user
func (s *AccountService) ListAccounts(ctx context.Context, userId int64) ([]models.Account, error) {
	accounts, err := s.accRepo.ListAccounts(ctx, s.db, userId)
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

// GetAccount returns details of account.
// If the account doesn't exist, it returns a custom not found error.
func (s *AccountService) GetAccount(ctx context.Context, id int64) (models.Account, error) {
	var account models.Account

	account, err := s.accRepo.GetAccount(ctx, s.db, id)
	if err != nil {
		return account, err
	}

	return account, nil
}

// ListAccountTransactions returns the list of transactions of account
func (s *AccountService) ListAccountTransactions(
	ctx context.Context, id int64, filter models.TransactionFilter, offset int64,
) (int, []models.Transaction, error) {
	count, transactions, err := s.accRepo.ListAccountTransactions(ctx, s.db, id, filter, offset)
	if err != nil {
		return -1, nil, err
	}
	return count, transactions, err
}

// CreateAccount inserts new account.
// If the account references a non-existent user, it returns a custom not found error.
func (s *AccountService) CreateAccount(ctx context.Context, userId int64, body models.PostAccountBody) (models.Account, error) {
	var newAccount models.Account

	// Debit account is not supposed to have credit limit and next due
	// Credit account must have credit limit and next due
	if err := body.Validate(); err != nil {
		return newAccount, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return newAccount, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	newAccount, err = s.accRepo.InsertAccount(ctx, tx, userId, body)
	if err != nil {
		return newAccount, err
	}

	// Insert the account balance history at the time of creation
	// We are guaranteed only after account is created will we be able to take actions on transactions.
	// No hypothetical race conditions between acc and tran creation.
	accHistBody := models.PostAccHistBody{
		AccountId:  newAccount.ID,
		LoggedTime: newAccount.CreatedAt,
		Balance:    newAccount.Balance,
	}
	insertedHistory, err := s.accHisRepo.InsertHistory(ctx, tx, accHistBody)
	if err != nil {
		return newAccount, err
	}
	// Log the account's history
	log.Printf("History for account %d, balance %f on %s\n",
		insertedHistory.AccountId,
		insertedHistory.Balance,
		insertedHistory.LoggedTime,
	)

	if err = tx.Commit(ctx); err != nil {
		return newAccount, err
	}

	return newAccount, nil
}

// UpdateAccount modifies the acount's details.
// If balance changes, it insert a new transaction for this account that reflects the balance change.
//
// For credit card,
//   - If balance increases, transaction is considered an Others-type expense.
//   - If balance decreases, transaction is considered an income.
//
// For debit card,
//   - If balance increases, transaction is considered an income.
//   - If balance decreases, transaction is considered an Others-type expense.
//
// NOTE: The feature is to keep ledging consistency, but user won't likely update an account's balance.
// Most of the time, they will create a descriptive transaction though.
func (s *AccountService) UpdateAccount(ctx context.Context, id int64, body models.PutAccountBody) (models.Account, error) {
	var updatedAcc models.Account

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return updatedAcc, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Get the current account data, which be used later
	previousAcc, err := s.accRepo.GetAccount(ctx, tx, id)
	if err != nil {
		return updatedAcc, err
	}
	// Debit account can't have credit limit and next due
	// Credit account must have credit limit and next due
	if err = body.Validate(previousAcc.Type); err != nil {
		return updatedAcc, err
	}

	// Update the account
	updatedAcc, err = s.accRepo.UpdateAccount(ctx, tx, id, body)
	if err != nil {
		return updatedAcc, err
	}

	// Create the transaction reflecting the balance change (reconcile transaction)
	change := updatedAcc.Balance - previousAcc.Balance
	if change != 0 {
		var description, category string
		if change > 0 {
			description = fmt.Sprintf("Balance increases %f", round(math.Abs(change)))
			category = If(updatedAcc.Type == "Credit", "Others", "Income")
		} else {
			description = fmt.Sprintf("Balance descreases %f", round(math.Abs(change)))
			category = If(updatedAcc.Type == "Credit", "Income", "Others")
		}

		tranBody := models.PostTransactionBody{
			AccountID:       id,
			Merchant:        updatedAcc.Institution,
			TranDescription: description,
			Category:        category,
			Amount:          math.Abs(change),
			CreatedAt:       time.Now(),
		}
		tran, err := s.tranRepo.InsertTransaction(ctx, tx, tranBody)
		if err != nil {
			return updatedAcc, err
		}
		log.Printf("Transaction %s has been inserted\n", tran.TranDescription)
	}

	if err = tx.Commit(ctx); err != nil {
		return updatedAcc, err
	}

	return updatedAcc, nil
}

// DeleteAccount deletes the account, its transactions and account history
func (s *AccountService) DeleteAccount(ctx context.Context, id int64) error {
	_, err := s.accRepo.DeleteAccount(ctx, s.db, id)
	if err != nil {
		return err
	}

	return nil
}

// ListHistories returns the list of balance histories of the account
func (s *AccountService) ListHistories(ctx context.Context, id int64) ([]models.AccountHistory, error) {
	histories, err := s.accHisRepo.ListHistories(ctx, s.db, id)
	if err != nil {
		return nil, err
	}

	return histories, nil
}
