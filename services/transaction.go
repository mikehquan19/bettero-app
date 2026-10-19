package services

import (
	"betterov2/models"
	"betterov2/repositories"
	"context"
	"log"
	"time"

	"github.com/google/uuid"
)

// Transaction Service
type TransactionService struct {
	db              repositories.DBTX
	unitOfWork      repositories.UnitOfWork
	transactionRepo TransactionRepository
	accountRepo     AccountRepository
}

// Generate a new transaction service
func NewTransactionService(
	database repositories.DBTX,
	unitOfWork repositories.UnitOfWork,
	transactionRepo TransactionRepository,
	accountRepo AccountRepository,
) *TransactionService {
	return &TransactionService{
		db:              database,
		unitOfWork:      unitOfWork,
		transactionRepo: transactionRepo,
		accountRepo:     accountRepo,
	}
}

// FilterTransactions returns list of transactions of category between 2 dates
func (s *TransactionService) FilterTransactions(
	ctx context.Context,
	userId uuid.UUID,
	filter models.TransactionFilter,
	offset int,
) (int, []models.Transaction, error) {
	var count int
	var transactions []models.Transaction

	if err := s.unitOfWork.Do(ctx, func(tx repositories.DBTX) error {
		var uowErr error

		count, uowErr = s.transactionRepo.CountTransactions(ctx, tx, userId, filter)
		if uowErr != nil {
			return uowErr
		}

		transactions, uowErr = s.transactionRepo.FilterTransactions(ctx, tx, userId, filter, offset)
		return uowErr
	}); err != nil {
		return -1, nil, err
	}

	return count, transactions, nil
}

// ListSuggestions returns the list of transaction description
func (s *TransactionService) ListSuggestions(ctx context.Context, userId uuid.UUID, q string) ([]models.Suggestion, error) {
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
func (s *TransactionService) CreateTransaction(ctx context.Context, userId uuid.UUID, body models.PostTransactionBody) (models.Transaction, error) {
	if err := validateTransactionTime(body.CreatedAt); err != nil {
		return models.Transaction{}, err
	}

	var newTransaction models.Transaction

	if err := s.unitOfWork.Do(ctx, func(tx repositories.DBTX) error {
		var uowErr error

		// Insert the new transaction data
		newTransaction, uowErr = s.transactionRepo.InsertTransaction(ctx, tx, userId, body)
		if uowErr != nil {
			return uowErr
		}

		// Update the account balance data
		netChange := newTransaction.AmountCents
		if newTransaction.Category == models.Income {
			netChange = -newTransaction.AmountCents
		}

		accountId := newTransaction.Account.ID

		balance, uowErr := s.accountRepo.UpdateAccountBalance(ctx, tx, accountId, netChange)
		if uowErr != nil {
			return uowErr
		}
		log.Printf("Balance changes to: %d cents\n", balance)

		return nil
	}); err != nil {
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
func (s *TransactionService) UpdateTransaction(ctx context.Context, userId, id uuid.UUID, body models.PutTransactionBody) (models.Transaction, error) {
	var updatedTransaction models.Transaction

	if err := s.unitOfWork.Do(ctx, func(tx repositories.DBTX) error {
		var uowErr error

		// Get previous account before updating
		previousData, uowErr := s.transactionRepo.GetTransaction(ctx, tx, userId, id)
		if uowErr != nil {
			return uowErr
		}

		if uowErr = validateTransactionTime(previousData.CreatedAt); uowErr != nil {
			return uowErr
		}

		// Update the transaction
		updatedTransaction, uowErr = s.transactionRepo.UpdateTransaction(ctx, tx, userId, id, body)
		if uowErr != nil {
			return uowErr
		}

		// Compute the amount to update the account balance (if balance change)
		if previousData.AmountCents != updatedTransaction.AmountCents || previousData.Category != updatedTransaction.Category {
			prevChange := previousData.AmountCents
			if previousData.Category == models.Income {
				prevChange = -previousData.AmountCents
			}

			currChange := updatedTransaction.AmountCents
			if updatedTransaction.Category == models.Income {
				currChange = -updatedTransaction.AmountCents
			}

			accountId := updatedTransaction.Account.ID

			balance, uowErr := s.accountRepo.UpdateAccountBalance(ctx, tx, accountId, currChange-prevChange)
			if uowErr != nil {
				return uowErr
			}
			log.Printf("Balance changes to: %d cents\n", balance)
		}

		return nil
	}); err != nil {
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
func (s *TransactionService) DeleteTransaction(ctx context.Context, userId, id uuid.UUID) error {
	var deletedTransaction models.Transaction

	if err := s.unitOfWork.Do(ctx, func(tx repositories.DBTX) error {
		var uowErr error

		deletedTransaction, uowErr = s.transactionRepo.DeleteTransaction(ctx, tx, userId, id)
		if uowErr != nil {
			return uowErr
		}

		if uowErr = validateTransactionTime(deletedTransaction.CreatedAt); uowErr != nil {
			return uowErr
		}

		// Reverse the effect of creating the transaction
		netChange := -deletedTransaction.AmountCents
		if deletedTransaction.Category == models.Income {
			netChange = deletedTransaction.AmountCents
		}

		accountId := deletedTransaction.Account.ID

		balance, uowErr := s.accountRepo.UpdateAccountBalance(ctx, tx, accountId, netChange)
		if uowErr != nil {
			return uowErr
		}
		log.Printf("Balance changes to: %d cents\n", balance)

		return nil
	}); err != nil {
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
