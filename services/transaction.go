package services

import (
	"betterov2/models"
	"betterov2/repositories"
	"context"

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
func (s *TransactionService) ListSuggestions(
	ctx context.Context, userId int64, q string,
) ([]models.Suggestion, error) {
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
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			panic(err)
		}
	}()

	// Insert the new transaction data
	inserted, err := s.tranRepo.InsertTransaction(ctx, tx, body)
	if err != nil {
		return newTransaction, err
	}

	// Update the account balance data
	netChange := ternary(body.Category == "Income", -body.Amount, body.Amount)
	accountId := int64(body.AccountID)
	if err = s.accRepo.UpdateAccountBalance(ctx, tx, accountId, netChange); err != nil {
		return newTransaction, err
	}

	// Get the created transaction with nested account
	newTransaction, err = s.tranRepo.GetTransaction(ctx, tx, int64(inserted.ID))
	if err != nil {
		return newTransaction, err
	}

	if err = tx.Commit(ctx); err != nil {
		return newTransaction, err
	}

	return newTransaction, nil
}

// UpateTransaction updates the transaction's info and update the account's balance.
//
// Net change to be updated is computed as follows:
//
// # net change = current effect - previous effect
//
//   - previous effect: amount if the transaction with previous info was deleted
//   - current effect: amount if the transaction with updated info was inserted
func (s *TransactionService) UpdateTransaction(ctx context.Context, id int64, body models.PutTransactionBody) (models.Transaction, error) {
	var updatedTransaction models.Transaction

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return updatedTransaction, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			panic(err)
		}
	}()

	// Get previous account before updating
	prev, err := s.tranRepo.GetTransaction(ctx, tx, id)
	if err != nil {
		return updatedTransaction, err
	}

	// Update the transaction
	updated, err := s.tranRepo.UpdateTransaction(ctx, tx, id, body)
	if err != nil {
		return updatedTransaction, err
	}

	// Compute the amount to update the account balance (if balance change)
	if prev.Amount != updated.Amount {
		prevChange := ternary(prev.Category == "Income", -prev.Amount, prev.Amount)
		currChange := ternary(updated.Category == "Income", -updated.Amount, updated.Amount)
		accountId := int64(updated.AccountId)
		err = s.accRepo.UpdateAccountBalance(ctx, tx, accountId, currChange-prevChange)
		if err != nil {
			return updatedTransaction, err
		}
	}

	// Get the updated transaction with nested account
	updatedTransaction, err = s.tranRepo.GetTransaction(ctx, tx, id)
	if err != nil {
		return updatedTransaction, err
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
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			panic(err)
		}
	}()

	deleted, err := s.tranRepo.DeleteTransaction(ctx, tx, id)
	if err != nil {
		return err
	}

	// Reverse the effect of creating the transaction
	netChange := ternary(deleted.Category == "Income", deleted.Amount, -deleted.Amount)
	accountId := int64(deleted.AccountId)
	if err = s.accRepo.UpdateAccountBalance(ctx, tx, accountId, netChange); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}
