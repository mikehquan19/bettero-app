package services

import (
	"betterov2/models"
	"betterov2/repositories"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Transaction Service
type TransactionService struct {
	db       *pgxpool.Pool
	tranRepo *repositories.TransactionRepo
	accRepo  *repositories.AccountRepo
}

// Generate a new transaction service
func NewTransactionService(
	database *pgxpool.Pool, tranRepo *repositories.TransactionRepo, accRepo *repositories.AccountRepo,
) *TransactionService {
	return &TransactionService{
		db:       database,
		tranRepo: tranRepo,
		accRepo:  accRepo,
	}
}

// FilterTransactions returns list of transactions of category between 2 dates
func (s *TransactionService) FilterTransactions(
	ctx context.Context, userId int64, tranFilter models.TransactionFilter, offset int,
) (int, []models.Transaction, error) {
	count, transactions, err := s.tranRepo.FilterTransactions(ctx, s.db, userId, tranFilter, offset)
	if err != nil {
		return -1, nil, err
	}
	return count, transactions, err
}

// ListSuggestions returns the list of transaction description
func (s *TransactionService) ListSuggestions(ctx context.Context, userId int64, q string) ([]models.Suggestion, error) {
	suggestions, err := s.tranRepo.ListSuggestions(ctx, s.db, userId, q)
	if err != nil {
		return suggestions, nil
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
	var newTransaction models.Transaction

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return newTransaction, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if body.CreatedAt.Before(time.Now().AddDate(0, 0, -14)) {
		// Transaction older than 2 weeks ago can't be created
		return newTransaction, models.ErrTransactionTooOld
	}

	// Insert the new transaction data
	newTransaction, err = s.tranRepo.InsertTransaction(ctx, tx, body)
	if err != nil {
		return newTransaction, err
	}

	// Update the account balance data
	netChange := If(newTransaction.Category == "Income", -newTransaction.Amount, newTransaction.Amount)
	balance, err := s.accRepo.UpdateAccountBalance(ctx, tx, newTransaction.Account.Id, netChange)
	if err != nil {
		return newTransaction, err
	}
	log.Printf("Balance changes to: %f", balance)

	if err = tx.Commit(ctx); err != nil {
		return newTransaction, err
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
	var updatedTransaction models.Transaction

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return updatedTransaction, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Get previous account before updating
	previous, err := s.tranRepo.GetTransaction(ctx, tx, id)
	if err != nil {
		return updatedTransaction, err
	}

	if previous.CreatedAt.Before(time.Now().AddDate(0, 0, -14)) {
		return updatedTransaction, models.ErrTransactionTooOld
	}

	// Update the transaction
	updatedTransaction, err = s.tranRepo.UpdateTransaction(ctx, tx, id, body)
	if err != nil {
		return updatedTransaction, err
	}

	// Compute the amount to update the account balance (if balance change)
	if previous.Amount != updatedTransaction.Amount {
		prevChange := If(previous.Category == "Income", -previous.Amount, previous.Amount)
		currChange := If(updatedTransaction.Category == "Income", -updatedTransaction.Amount, updatedTransaction.Amount)
		netChange := currChange - prevChange
		balance, err := s.accRepo.UpdateAccountBalance(ctx, tx, updatedTransaction.Account.Id, netChange)
		if err != nil {
			return updatedTransaction, err
		}
		log.Printf("Balance changes to: %f", balance)
	}

	if err = tx.Commit(ctx); err != nil {
		return updatedTransaction, err
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

	deleted, err := s.tranRepo.DeleteTransaction(ctx, tx, id)
	if err != nil {
		return err
	}

	if deleted.CreatedAt.Before(time.Now().AddDate(0, 0, -14)) {
		return models.ErrTransactionTooOld
	}

	if deleted.ID != id {
		return fmt.Errorf("expected to delete transaction %d, deleted %d", id, deleted.ID)
	}

	// Reverse the effect of creating the transaction
	netChange := If(deleted.Category == "Income", deleted.Amount, -deleted.Amount)
	balance, err := s.accRepo.UpdateAccountBalance(ctx, tx, deleted.Account.Id, netChange)
	if err != nil {
		return err
	}
	log.Printf("Balance changes to: %f", balance)

	if err = tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}
