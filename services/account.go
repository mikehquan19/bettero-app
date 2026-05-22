package services

import (
	"betterov2/models"
	"betterov2/repositories"
	"context"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Account Service
type AccountService struct {
	db       *pgxpool.Pool
	accRepo  *repositories.AccountRepo
	tranRepo *repositories.TransactionRepo
}

// Generate a new account service
func NewAccountService(
	db *pgxpool.Pool, accRepo *repositories.AccountRepo, tranRepo *repositories.TransactionRepo,
) *AccountService {
	return &AccountService{
		db:       db,
		accRepo:  accRepo,
		tranRepo: tranRepo,
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

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return account, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			panic(err)
		}
	}()

	account, err = s.accRepo.GetAccount(ctx, tx, id)
	if err != nil {
		return account, err
	}

	if err = tx.Commit(ctx); err != nil {
		return account, err
	}

	return account, nil
}

// ListAccountTransactions returns the list of transactions of account
func (s *AccountService) ListAccountTransactions(
	ctx context.Context, id int64, tranFilter models.TransactionFilter, offset int64,
) (int, []models.Transaction, error) {
	count, transactions, err := s.accRepo.ListTransactions(ctx, s.db, id, tranFilter, offset)
	if err != nil {
		return -1, nil, err
	}
	return count, transactions, err
}

// CreateAccount inserts new account.
// If the account references a non-existent user, it returns a custom not found error.
func (s *AccountService) CreateAccount(
	ctx context.Context, userId int64, body models.PostAccountBody,
) (models.Account, error) {
	var newAccount models.Account

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return newAccount, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			panic(err)
		}
	}()

	newAccount, err = s.accRepo.InsertAccount(ctx, tx, userId, body)
	if err != nil {
		return newAccount, err
	}

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
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			panic(err)
		}
	}()

	// Get the previous account data
	prevAccData, err := s.accRepo.GetAccount(ctx, tx, id)
	if err != nil {
		return updatedAcc, err
	}
	if err = body.Validate(prevAccData.Type); err != nil {
		return updatedAcc, err
	}

	// Update the account
	updatedAcc, err = s.accRepo.UpdateAccount(ctx, tx, id, body)
	if err != nil {
		return updatedAcc, err
	}

	// Create the transaction reflecting the balance change, aka reconile transaction
	change := updatedAcc.Balance - prevAccData.Balance
	if change != 0 {
		var description, category string
		if change > 0 {
			description = fmt.Sprintf("Balance increases %f", round(math.Abs(change)))
			category = ternary(updatedAcc.Type == "Credit", "Others", "Income")
		} else {
			description = fmt.Sprintf("Balance descreases %f", round(math.Abs(change)))
			category = ternary(updatedAcc.Type == "Credit", "Income", "Others")
		}

		body := models.PostTransactionBody{
			AccountID:       int(id),
			Merchant:        updatedAcc.Institution,
			TranDescription: description,
			Category:        category,
			Amount:          math.Abs(change),
			CreatedAt:       time.Now(),
		}
		_, err = s.tranRepo.InsertTransaction(ctx, tx, body)
		if err != nil {
			return updatedAcc, err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return updatedAcc, err
	}

	return updatedAcc, nil
}

// DeleteAccount deletes the account and all of its transactions.
func (s *AccountService) DeleteAccount(ctx context.Context, id int64) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			panic(err)
		}
	}()

	_, err = s.accRepo.DeleteAccount(ctx, tx, id)
	if err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}
