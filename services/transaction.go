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
	var created models.Transaction

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return created, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if body.CreatedAt.Before(time.Now().AddDate(0, 0, -14)) {
		// Transaction older than 2 weeks ago can't be created
		return created, models.ErrTransactionTooOld
	}

	// Insert the new transaction data
	created, err = s.tranRepo.InsertTransaction(ctx, tx, body)
	if err != nil {
		return created, err
	}

	// Update the account balance data
	netChange := If(created.Category == "Income", -created.Amount, created.Amount)
	balance, err := s.accRepo.UpdateAccountBalance(ctx, tx, created.Account.Id, netChange)
	if err != nil {
		return created, err
	}
	log.Printf("Balance changes to: %f\n", balance)

	if err = tx.Commit(ctx); err != nil {
		return created, err
	}

	return created, nil
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
	var updated models.Transaction

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return updated, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Get previous account before updating
	previous, err := s.tranRepo.GetTransaction(ctx, tx, id)
	if err != nil {
		return updated, err
	}

	if previous.CreatedAt.Before(time.Now().AddDate(0, 0, -14)) {
		return updated, models.ErrTransactionTooOld
	}

	// Update the transaction
	updated, err = s.tranRepo.UpdateTransaction(ctx, tx, id, body)
	if err != nil {
		return updated, err
	}

	// Compute the amount to update the account balance (if balance change)
	if previous.Amount != updated.Amount {
		prevChange := If(previous.Category == "Income", -previous.Amount, previous.Amount)
		currChange := If(updated.Category == "Income", -updated.Amount, updated.Amount)
		netChange := currChange - prevChange
		balance, err := s.accRepo.UpdateAccountBalance(ctx, tx, updated.Account.Id, netChange)
		if err != nil {
			return updated, err
		}
		log.Printf("Balance changes to: %f\n", balance)
	}

	if err = tx.Commit(ctx); err != nil {
		return updated, err
	}

	return updated, nil
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

	// Reverse the effect of creating the transaction
	netChange := If(deleted.Category == "Income", deleted.Amount, -deleted.Amount)
	balance, err := s.accRepo.UpdateAccountBalance(ctx, tx, deleted.Account.Id, netChange)
	if err != nil {
		return err
	}
	log.Printf("Balance changes to: %f\n", balance)

	if err = tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}
