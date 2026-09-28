package services

import (
	"betterov2/models"
	"betterov2/repositories"
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CronService struct {
	db          *pgxpool.Pool
	accRepo     repositories.AccountRepo
	accHistRepo repositories.AccountHistoryRepo
	tranRepo    repositories.TransactionRepo
}

// Initialize the new cron service
func NewCronService(
	db *pgxpool.Pool,
	accRepo repositories.AccountRepo,
	accHistRepo repositories.AccountHistoryRepo,
	tranRepo repositories.TransactionRepo,
) *CronService {
	return &CronService{
		db:          db,
		accRepo:     accRepo,
		accHistRepo: accHistRepo,
		tranRepo:    tranRepo,
	}
}

// MoveAccountsDueDate updates next due date for accounts whose due date is past today.
// If update fails, retrying will be on the accounts that haven't been updated on the previous try.
// It's because accounts that have will not be queried.
func (c *CronService) MoveAccountsDueDate() error {
	ctx := context.Background()
	pastDueAccounts, err := c.accRepo.FilterAccounts(ctx, c.db, repositories.PastDue)
	if err != nil {
		return err
	}

	// For each batch, bulk update the due date of the accounts of that batch
	index := 1
	for batch := range slices.Chunk(pastDueAccounts, 50) {
		accountIDs := make([]uuid.UUID, 0, len(batch))
		for _, account := range batch {
			accountIDs = append(accountIDs, account.ID)
		}

		// Bulk update the list of account IDs
		batchCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		numUpdated, err := c.accRepo.MoveAccountsDueDate(batchCtx, c.db, accountIDs)
		cancel()
		if err != nil {
			return err
		}
		log.Printf("%d accounts of batch %d have been updated!\n", numUpdated, index)
		index++
	}

	log.Println("Update done!")
	return nil
}

// ValidateAccountsAndInsertHistory validates all the account balances.
//
// TODO: Address the limitation that if one account fails, the retry will re-validate
// the successfully validated accounts.
func (c *CronService) ValidateAccounts() error {
	ctx := context.Background()
	accounts, err := c.accRepo.FilterAccounts(ctx, c.db, repositories.All)
	if err != nil {
		return err
	}

	var errs []error
	for _, account := range accounts {
		if err = c.validate(ctx, account); err != nil {
			errs = append(errs, fmt.Errorf("account %s: %w", account.ID, err))
			log.Printf("Account %s failed validating: %v\n", account.ID, err)
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	log.Println("Done validating accounts!")
	return nil
}

// Process each individual account for validation
func (c *CronService) validate(ctx context.Context, account models.Account) error {
	accCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	tx, err := c.db.Begin(accCtx)
	if err != nil {
		return err
	}
	defer tx.Rollback(accCtx) //nolint:errcheck

	// Validate the account and flag it if there's discrepancy
	// Get the latest balance history and sum of all transactions from the time
	latestHist, err := c.accHistRepo.GetLatestHistory(accCtx, tx, account.ID)
	if err != nil {
		return err
	}
	sum, err := c.tranRepo.GetTransactionSum(accCtx, tx, account.ID, latestHist.LoggedTime)
	if err != nil {
		return err
	}

	discrepancy := calculateDiscrepancy(account, latestHist, sum)
	if discrepancy != 0 {
		flaggedAccount, err := c.accRepo.FlagAccount(accCtx, tx, account.ID, discrepancy)
		if err != nil {
			return err
		}
		log.Printf("Account %s has been flagged!\n", flaggedAccount.ID)
	} else {
		log.Printf("Account %s is OK!\n", account.ID)
	}

	if err = tx.Commit(accCtx); err != nil {
		return err
	}

	return nil
}

// UpdateHistory will create the new history of account whose latest history is past 2 weeks,
// and delete the history that is older than 6 months ago.
func (c *CronService) UpdateHistory() error {
	ctx := context.Background()
	unflaggedAccounts, err := c.accRepo.FilterAccounts(ctx, c.db, repositories.Unflagged)
	if err != nil {
		return err
	}

	var errs []error
	for _, account := range unflaggedAccounts {
		if err = c.updateHistory(ctx, account, time.Now().UTC()); err != nil {
			log.Printf("Error while executing account %s: %v\n", account.ID, err)
			errs = append(errs, fmt.Errorf("account %s: %w", account.ID, err))
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	log.Printf("Done creating new history!\n")
	return nil
}

// Process each individual account when updating the balance history
func (c *CronService) updateHistory(ctx context.Context, account models.Account, now time.Time) error {
	accCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	tx, err := c.db.Begin(accCtx)
	if err != nil {
		return err
	}
	defer tx.Rollback(accCtx) //nolint:errcheck

	latestHist, err := c.accHistRepo.GetLatestHistory(accCtx, tx, account.ID)
	if err != nil {
		return err
	}

	// Create the account history every month
	if shouldCreateHistory(latestHist, now) {
		_, err := c.accHistRepo.InsertHistory(accCtx, tx, models.PostAccHistBody{
			AccountID:    account.ID,
			LoggedTime:   now,
			BalanceCents: account.BalanceCents,
		})
		if err != nil {
			return err
		}
		log.Printf("New history created for account %s", account.ID)
	} else {
		log.Printf("No history created for account %s", account.ID)
	}

	// Delete outdated histories of account
	numDeleted, err := c.accHistRepo.DeleteOutdatedHistories(accCtx, tx, account.ID)
	if err != nil {
		return err
	}
	log.Printf("Deleted %d histories of account %s", numDeleted, account.ID)

	if err = tx.Commit(accCtx); err != nil {
		return err
	}

	return nil
}

func calculateDiscrepancy(
	account models.Account,
	latestHist models.AccountHistory,
	transactionSum models.Money,
) models.Money {
	if account.Type == models.Debit {
		return account.BalanceCents - latestHist.BalanceCents + transactionSum
	}
	return account.BalanceCents - latestHist.BalanceCents - transactionSum
}

func shouldCreateHistory(latestHist models.AccountHistory, now time.Time) bool {
	return latestHist.LoggedTime.Before(now.AddDate(0, -1, 0))
}

// Delete all the outdated transactions (transactions that are 6 months old) of
// all the accounts
func (c *CronService) DeleteOutdatedTransactions() error {
	ctx := context.Background()
	accounts, err := c.accRepo.FilterAccounts(ctx, c.db, repositories.All)
	if err != nil {
		return err
	}

	for _, account := range accounts {
		accCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		numDeleted, err := c.tranRepo.DeleteOutdatedTransactions(accCtx, c.db, account.ID)
		cancel()
		if err != nil {
			log.Printf("Error while executing account %s\n", account.ID)
			return err
		}
		log.Printf("Deleted %d outdated transactions of account %s", numDeleted, account.ID)
	}

	log.Printf("Done deleting transactions!")
	return nil
}
