package services

import (
	"betterov2/models"
	"betterov2/repositories"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

// Account Service
type AccountService struct {
	db                 repositories.DBTX
	unitOfWork         repositories.UnitOfWork
	accountRepo        AccountRepository
	accountHistoryRepo AccountHistoryRepository
	transactionRepo    TransactionRepository
}

// Generate a new account service
func NewAccountService(
	db repositories.DBTX,
	unitOfWork repositories.UnitOfWork,
	accountRepo AccountRepository,
	accountHistoryRepo AccountHistoryRepository,
	transactionRepo TransactionRepository,
) *AccountService {
	return &AccountService{
		db:                 db,
		unitOfWork:         unitOfWork,
		accountRepo:        accountRepo,
		accountHistoryRepo: accountHistoryRepo,
		transactionRepo:    transactionRepo,
	}
}

// ListAccounts returns the list of accounts of the user
func (s *AccountService) ListAccounts(ctx context.Context, userId uuid.UUID) ([]models.Account, error) {
	accounts, err := s.accountRepo.ListAccounts(ctx, s.db, userId)
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

// GetAccount returns details of account.
func (s *AccountService) GetAccount(ctx context.Context, userId, id uuid.UUID) (models.Account, error) {
	account, err := s.accountRepo.GetAccount(ctx, s.db, userId, id)
	if err != nil {
		// If the account doesn't exist, it returns a custom not found error.
		return models.Account{}, err
	}

	return account, nil
}

// ListAccountTransactions returns the list of transactions of account
func (s *AccountService) ListAccountTransactions(
	ctx context.Context,
	userId, id uuid.UUID,
	filter models.TransactionFilter,
	offset int64,
) (int, []models.Transaction, error) {
	var count int
	var transactions []models.Transaction

	// Return the combination of count and list of transactions from the filter
	if err := s.unitOfWork.Do(ctx, func(tx repositories.DBTX) error {
		var uowErr error

		count, uowErr = s.accountRepo.CountAccTransactions(ctx, tx, userId, id, filter)
		if uowErr != nil {
			return uowErr
		}

		transactions, uowErr = s.accountRepo.ListAccTransactions(ctx, tx, userId, id, filter, offset)
		return uowErr
	}); err != nil {
		return -1, nil, err
	}
	return count, transactions, nil
}

// ListHistories returns the list of balance histories of the account
func (s *AccountService) ListAccountHistories(ctx context.Context, userId, id uuid.UUID) ([]models.AccountHistory, error) {
	histories, err := s.accountHistoryRepo.ListHistories(ctx, s.db, userId, id)
	if err != nil {
		return nil, err
	}

	return histories, nil
}

// CreateAccount inserts new account.
func (s *AccountService) CreateAccount(ctx context.Context, userId uuid.UUID, body models.PostAccountBody) (models.Account, error) {
	if err := validateAccount(body.Type, body.AccountBody); err != nil {
		return models.Account{}, err
	}

	var newAccount models.Account

	if err := s.unitOfWork.Do(ctx, func(tx repositories.DBTX) error {
		var uowErr error

		newAccount, uowErr = s.accountRepo.InsertAccount(ctx, tx, userId, body)
		if uowErr != nil {
			// The account references a non-existent user, return 404
			return uowErr
		}

		// Insert the account balance history at the time of creation.
		// Guaranteed that only after account is created can we take actions on transactions.
		postHistBody := models.PostAccHistBody{
			AccountID:    newAccount.ID,
			LoggedTime:   newAccount.CreatedAt,
			BalanceCents: newAccount.BalanceCents,
		}
		insertedHistory, uowErr := s.accountHistoryRepo.InsertHistory(ctx, tx, postHistBody)
		if uowErr != nil {
			return uowErr
		}

		log.Printf("History for account %s, balance %d cents on %s\n",
			insertedHistory.AccountID,
			insertedHistory.BalanceCents,
			insertedHistory.LoggedTime,
		)

		return nil
	}); err != nil {
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
func (s *AccountService) UpdateAccount(ctx context.Context, userId, id uuid.UUID, body models.PutAccountBody) (models.Account, error) {
	var updatedAccount models.Account

	if err := s.unitOfWork.Do(ctx, func(tx repositories.DBTX) error {
		var uowErr error

		// Get the current account data, which be used later
		previousData, uowErr := s.accountRepo.GetAccount(ctx, tx, userId, id)
		if uowErr != nil {
			return uowErr
		}
		if uowErr = validateAccount(previousData.Type, body.AccountBody); uowErr != nil {
			return uowErr
		}

		// Update the account
		updatedAccount, uowErr = s.accountRepo.UpdateAccount(ctx, tx, userId, id, body)
		if uowErr != nil {
			return uowErr
		}

		// Create the transaction reflecting the balance change (reconcile transaction)
		change := updatedAccount.BalanceCents - previousData.BalanceCents
		if change == 0 {
			return nil
		}
		var description string
		var category models.TransactionCategory

		amountCents := change
		if amountCents < 0 {
			amountCents = -amountCents
		}
		displayedAmount := float64(amountCents) / 100
		if change > 0 {
			description = fmt.Sprintf("Balance reconciled (+%.2f)", displayedAmount)

			category = models.Income
			if updatedAccount.Type == models.Credit {
				category = models.Others
			} else {
				category = models.Income
			}
		} else {
			description = fmt.Sprintf("Balance reconciled (-%.2f)", displayedAmount)

			category = models.Others
			if updatedAccount.Type == models.Credit {
				category = models.Income
			} else {
				category = models.Others
			}
		}

		postTranBody := models.PostTransactionBody{
			AccountID:   id,
			Merchant:    updatedAccount.Institution,
			Description: description,
			Category:    category,
			AmountCents: amountCents,
			CreatedAt:   time.Now(),
		}
		transaction, uowErr := s.transactionRepo.InsertTransaction(ctx, tx, userId, postTranBody)
		if uowErr != nil {
			return uowErr
		}
		log.Printf("Transaction %s has been inserted\n", transaction.ID)

		return nil
	}); err != nil {
		return models.Account{}, err
	}

	return updatedAccount, nil
}

// validate validates the account body
//
//   - Debit account can't have credit limit and next due
//
//   - Credit account must have credit limit and next due
func validateAccount(aType models.AccountType, body models.AccountBody) error {
	if aType == models.Debit && (body.NextDue != nil || body.CreditLimitCents != nil) {
		return models.ErrDebitCardWithCreditInfo
	}

	if aType == models.Credit && (body.NextDue == nil || body.CreditLimitCents == nil) {
		return models.ErrCreditCardWithoutCreditInfo
	}

	return nil
}

// DeleteAccount deletes the account, its transactions and account history
func (s *AccountService) DeleteAccount(ctx context.Context, userId, id uuid.UUID) error {
	deletedAccount, err := s.accountRepo.DeleteAccount(ctx, s.db, userId, id)
	if err != nil {
		return err
	}
	log.Printf("Account %s has been deleted\n", deletedAccount.ID)

	return nil
}
