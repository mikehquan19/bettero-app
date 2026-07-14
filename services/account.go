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
	db                 *pgxpool.Pool
	accountRepo        *repositories.AccountRepo
	accountHistoryRepo *repositories.AccountHistoryRepo
	transactionRepo    *repositories.TransactionRepo
}

// Generate a new account service
func NewAccountService(
	db *pgxpool.Pool,
	accountRepo *repositories.AccountRepo,
	accountHistoryRepo *repositories.AccountHistoryRepo,
	transactionRepo *repositories.TransactionRepo,
) *AccountService {
	return &AccountService{
		db:                 db,
		accountRepo:        accountRepo,
		accountHistoryRepo: accountHistoryRepo,
		transactionRepo:    transactionRepo,
	}
}

// ListAccounts returns the list of accounts of the user
func (s *AccountService) ListAccounts(ctx context.Context, userId int64) ([]models.Account, error) {
	accounts, err := s.accountRepo.ListAccounts(ctx, s.db, userId)
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

// GetAccount returns details of account.
// If the account doesn't exist, it returns a custom not found error.
func (s *AccountService) GetAccount(ctx context.Context, id int64) (models.Account, error) {
	account, err := s.accountRepo.GetAccount(ctx, s.db, id)
	if err != nil {
		return models.Account{}, err
	}

	return account, nil
}

// ListAccountTransactions returns the list of transactions of account
func (s *AccountService) ListAccountTransactions(
	ctx context.Context,
	id int64,
	filter models.TransactionFilter,
	offset int64,
) (int, []models.Transaction, error) {
	count, transactions, err := s.accountRepo.ListAccountTransactions(ctx, s.db, id, filter, offset)
	if err != nil {
		return -1, nil, err
	}
	return count, transactions, err
}

// ListHistories returns the list of balance histories of the account
func (s *AccountService) ListAccountHistories(ctx context.Context, id int64) ([]models.AccountHistory, error) {
	histories, err := s.accountHistoryRepo.ListHistories(ctx, s.db, id)
	if err != nil {
		return nil, err
	}

	return histories, nil
}

// CreateAccount inserts new account.
// If the account references a non-existent user, it returns a custom not found error.
func (s *AccountService) CreateAccount(ctx context.Context, userId int64, body models.PostAccountBody) (models.Account, error) {
	if err := validateAccount(body.Type, body.AccountBody); err != nil {
		return models.Account{}, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return models.Account{}, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	newAccount, err := s.accountRepo.InsertAccount(ctx, tx, userId, body)
	if err != nil {
		return models.Account{}, err
	}

	// Insert the account balance history at the time of creation
	// Guaranteed that only after account is created can we take actions on transactions.
	insertedHistory, err := s.accountHistoryRepo.InsertHistory(ctx, tx, models.PostAccHistBody{
		AccountId:  newAccount.ID,
		LoggedTime: newAccount.CreatedAt,
		Balance:    newAccount.Balance,
	})
	if err != nil {
		return models.Account{}, err
	}
	log.Printf("History for account %d, balance %f on %s\n",
		insertedHistory.AccountId,
		insertedHistory.Balance,
		insertedHistory.LoggedTime,
	)

	if err = tx.Commit(ctx); err != nil {
		return models.Account{}, err
	}

	return newAccount, nil
}

// UpdateAccount modifies the acount's details.
// If balance changes, it insert a new transaction for this account that reflects the balance change.
//
// For credit card,
//   - If balance increases, transaction is considered an Others.
//   - If balance decreases, transaction is considered an income.
//
// For debit card,
//   - If balance increases, transaction is considered an income.
//   - If balance decreases, transaction is considered an Others.
//
// NOTE: The feature is to keep ledging consistency, but user won't likely update an account's balance.
// Most of the time, they will create a descriptive transaction though.
func (s *AccountService) UpdateAccount(ctx context.Context, id int64, body models.PutAccountBody) (models.Account, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return models.Account{}, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Get the current account data, which be used later
	previousData, err := s.accountRepo.GetAccount(ctx, tx, id)
	if err != nil {
		return models.Account{}, err
	}
	if err = validateAccount(previousData.Type, body.AccountBody); err != nil {
		return models.Account{}, err
	}

	// Update the account
	updatedAccount, err := s.accountRepo.UpdateAccount(ctx, tx, id, body)
	if err != nil {
		return models.Account{}, err
	}

	// Create the transaction reflecting the balance change (reconcile transaction)
	change := updatedAccount.Balance - previousData.Balance
	if change != 0 {
		var description, category string
		if change > 0 {
			description = fmt.Sprintf("Balance reconciled (+%.2f)", math.Abs(change))
			if updatedAccount.Type == "Credit" {
				category = "Others"
			} else {
				category = "Income"
			}
		} else {
			description = fmt.Sprintf("Balance reconciled (-%.2f)", math.Abs(change))
			if updatedAccount.Type == "Credit" {
				category = "Income"
			} else {
				category = "Others"
			}
		}
		transaction, err := s.transactionRepo.InsertTransaction(ctx, tx, models.PostTransactionBody{
			AccountID:       id,
			Merchant:        updatedAccount.Institution,
			TranDescription: description,
			Category:        category,
			Amount:          math.Abs(change),
			CreatedAt:       time.Now(),
		})
		if err != nil {
			return models.Account{}, err
		}
		log.Printf("Transaction %s has been inserted\n", transaction.TranDescription)
	}

	if err = tx.Commit(ctx); err != nil {
		return models.Account{}, err
	}

	return updatedAccount, nil
}

// validate validates the account body
//
//   - Debit account can't have credit limit and next due
//
//   - Credit account must have credit limit and next due
func validateAccount(aType string, body models.AccountBody) error {
	if aType == "Debit" && (body.NextDue != nil || body.CreditLimit != nil) {
		return models.ErrDebitCardWithCreditInfo
	}

	if aType == "Credit" && (body.NextDue == nil || body.CreditLimit == nil) {
		return models.ErrCreditCardWithoutCreditInfo
	}

	return nil
}

// DeleteAccount deletes the account, its transactions and account history
func (s *AccountService) DeleteAccount(ctx context.Context, id int64) error {
	_, err := s.accountRepo.DeleteAccount(ctx, s.db, id)
	if err != nil {
		return err
	}

	return nil
}
