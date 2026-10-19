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
)

type CronService struct {
	db          repositories.DBTX
	unitOfWork  repositories.UnitOfWork
	accRepo     CronAccountRepository
	accHistRepo CronAccountHistoryRepository
	tranRepo    CronTransactionRepository
}

// Initialize the new cron service
func NewCronService(
	db repositories.DBTX,
	unitOfWork repositories.UnitOfWork,
	accRepo CronAccountRepository,
	accHistRepo CronAccountHistoryRepository,
	tranRepo CronTransactionRepository,
) *CronService {
	return &CronService{
		db:          db,
		unitOfWork:  unitOfWork,
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

	return c.unitOfWork.Do(accCtx, func(tx repositories.DBTX) error {
		// Validate the account and flag it if there's discrepancy.
		// Get the latest balance history and sum of all transactions from that time.
		latestHist, uowErr := c.accHistRepo.GetLatestHistory(accCtx, tx, account.ID)
		if uowErr != nil {
			return uowErr
		}
		sum, uowErr := c.tranRepo.GetTransactionSum(accCtx, tx, account.ID, latestHist.LoggedTime)
		if uowErr != nil {
			return uowErr
		}

		discrepancy := calculateDiscrepancy(account, latestHist, sum)
		if discrepancy != 0 {
			flaggedAccount, uowErr := c.accRepo.FlagAccount(accCtx, tx, account.ID, discrepancy)
			if uowErr != nil {
				return uowErr
			}
			log.Printf("Account %s has been flagged!\n", flaggedAccount.ID)
		} else {
			log.Printf("Account %s is OK!\n", account.ID)
		}

		return nil
	})
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

	log.Println("Done creating new history!")
	return nil
}

// Process each individual account when updating the balance history
func (c *CronService) updateHistory(ctx context.Context, account models.Account, now time.Time) error {
	accCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return c.unitOfWork.Do(accCtx, func(tx repositories.DBTX) error {
		latestHist, uowErr := c.accHistRepo.GetLatestHistory(accCtx, tx, account.ID)
		if uowErr != nil {
			return uowErr
		}

		// Create the account history every month
		if shouldCreateHistory(latestHist, now) {
			_, uowErr = c.accHistRepo.InsertHistory(accCtx, tx, models.PostAccHistBody{
				AccountID:    account.ID,
				LoggedTime:   now,
				BalanceCents: account.BalanceCents,
			})
			if uowErr != nil {
				return uowErr
			}
			log.Printf("New history created for account %s\n", account.ID)
		} else {
			log.Printf("No history created for account %s\n", account.ID)
		}

		// Delete outdated histories of account
		numDeleted, uowErr := c.accHistRepo.DeleteOutdatedHistories(accCtx, tx, account.ID)
		if uowErr != nil {
			return uowErr
		}
		log.Printf("Deleted %d histories of account %s\n", numDeleted, account.ID)

		return nil
	})
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
		log.Printf("Deleted %d outdated transactions of account %s\n", numDeleted, account.ID)
	}

	log.Println("Done deleting transactions!")
	return nil
}
