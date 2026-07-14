package services

import (
	"betterov2/models"
	"betterov2/repositories"
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Transaction Service
type TransactionService struct {
	db              *pgxpool.Pool
	transactionRepo *repositories.TransactionRepo
	accountRepo     *repositories.AccountRepo
}

// Generate a new transaction service
func NewTransactionService(
	database *pgxpool.Pool,
	transactionRepo *repositories.TransactionRepo,
	accountRepo *repositories.AccountRepo,
) *TransactionService {
	return &TransactionService{
		db:              database,
		transactionRepo: transactionRepo,
		accountRepo:     accountRepo,
	}
}

// FilterTransactions returns list of transactions of category between 2 dates
func (s *TransactionService) FilterTransactions(
	ctx context.Context,
	userId int64,
	filter models.TransactionFilter,
	offset int,
) (int, []models.Transaction, error) {
	count, transactions, err := s.transactionRepo.FilterTransactions(ctx, s.db, userId, filter, offset)
	if err != nil {
		return -1, nil, err
	}
	return count, transactions, err
}

// ListSuggestions returns the list of transaction description
func (s *TransactionService) ListSuggestions(ctx context.Context, userId int64, q string) ([]models.Suggestion, error) {
	suggestions, err := s.transactionRepo.ListSuggestions(ctx, s.db, userId, q)
	if err != nil {
		return nil, err
	}
	return suggestions, nil
}

// CreateTransaction inserts transaction into the database, and update the account's balance.
//
// For an income transaction, by transaction amount (positive by default):
//   - Debit card's balance will increase (income check, tax refund, allowance, etc)
//   - Credit card's balance will decrease (monthly payment, refund, etc.)
//
// For an expense transaction, by transaction amount:
//   - Debit card's balance will decrease
//   - Credit card's balance will increase
func (s *TransactionService) CreateTransaction(ctx context.Context, body models.PostTransactionBody) (models.Transaction, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return models.Transaction{}, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if err := validateTransactionTime(body.CreatedAt); err != nil {
		return models.Transaction{}, err
	}

	// Insert the new transaction data
	newTransaction, err := s.transactionRepo.InsertTransaction(ctx, tx, body)
	if err != nil {
		return models.Transaction{}, err
	}

	// Update the account balance data
	netChange := newTransaction.Amount
	if newTransaction.Category == models.Income {
		netChange = -newTransaction.Amount
	}

	accountId := newTransaction.Account.Id

	balance, err := s.accountRepo.UpdateAccountBalance(ctx, tx, accountId, netChange)
	if err != nil {
		return models.Transaction{}, err
	}
	log.Printf("Balance changes to: %f\n", balance)

	if err = tx.Commit(ctx); err != nil {
		return models.Transaction{}, err
	}

	return newTransaction, nil
}

// UpateTransaction updates the transaction's info and update the account's balance.
//
// Net change to be updatedTransaction is computed as follows:
//
// # net change = current effect - previous effect
//
//   - previous effect: amount if the transaction with previous info was deleted
//   - current effect: amount if the transaction with updatedTransaction info was inserted
func (s *TransactionService) UpdateTransaction(ctx context.Context, id int64, body models.PutTransactionBody) (models.Transaction, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return models.Transaction{}, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Get previous account before updating
	previousData, err := s.transactionRepo.GetTransaction(ctx, tx, id)
	if err != nil {
		return models.Transaction{}, err
	}

	if err := validateTransactionTime(previousData.CreatedAt); err != nil {
		return models.Transaction{}, err
	}

	// Update the transaction
	updatedTransaction, err := s.transactionRepo.UpdateTransaction(ctx, tx, id, body)
	if err != nil {
		return models.Transaction{}, err
	}

	// Compute the amount to update the account balance (if balance change)
	if previousData.Amount != updatedTransaction.Amount {
		prevChange := previousData.Amount
		if previousData.Category == models.Income {
			prevChange = -previousData.Amount
		}

		currChange := updatedTransaction.Amount
		if updatedTransaction.Category == models.Income {
			currChange = -updatedTransaction.Amount
		}

		accountId := updatedTransaction.Account.Id

		balance, err := s.accountRepo.UpdateAccountBalance(ctx, tx, accountId, currChange-prevChange)
		if err != nil {
			return models.Transaction{}, err
		}
		log.Printf("Balance changes to: %f\n", balance)
	}

	if err = tx.Commit(ctx); err != nil {
		return models.Transaction{}, err
	}

	return updatedTransaction, nil
}

// DeleteTransaction deletes the transaction from database and update balance.
//
// For an income transaction,
//   - Debit card's balance will decrease (income check, tax refund, allowance, etc)
//   - Credit card's balance will increase (monthly payment, refund, etc.)
//
// For an expense transaction,
//   - Debit card's balance will increase
//   - Credit card's balance will decrease
func (s *TransactionService) DeleteTransaction(ctx context.Context, id int64) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	deletedTransaction, err := s.transactionRepo.DeleteTransaction(ctx, tx, id)
	if err != nil {
		return err
	}

	if err := validateTransactionTime(deletedTransaction.CreatedAt); err != nil {
		return err
	}

	// Reverse the effect of creating the transaction
	netChange := -deletedTransaction.Amount
	if deletedTransaction.Category == models.Income {
		netChange = deletedTransaction.Amount
	}

	accountId := deletedTransaction.Account.Id

	balance, err := s.accountRepo.UpdateAccountBalance(ctx, tx, accountId, netChange)
	if err != nil {
		return err
	}
	log.Printf("Balance changes to: %f\n", balance)

	if err = tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

// validateTransactionTime validates the transaction.
// Transaction older than 2 weeks ago can't be created
func validateTransactionTime(createdAt time.Time) error {
	if createdAt.Before(time.Now().AddDate(0, 0, -13)) {
		return models.ErrTransactionTooOld
	}
	return nil
}
