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
// If update fails, retrying will be on the accounts that haven't been updated on the previous try, because accounts that have will not be queried.
//
// Frequency: daily
func (c *CronService) MoveAccountsDueDate() error {
	ctx := context.Background()
	pastDueAccounts, err := c.accRepo.ListAccountsByStatus(ctx, c.db, repositories.PastDue)
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

// ValidateAccountsAndUpdateHistory validates every account, updates its discrepancy state,
// and creates a history when the latest one is at least 14 days old.
//
// Histories older than six months are deleted after a successful validation.
//
// TODO: Address the limitation that if one account fails, the retry will re-validate
// the successfully validated accounts.
//
// Frequency: daily, before DeleteOutdatedTransactions.
func (c *CronService) ValidateAccountsAndUpdateHistory() error {
	ctx := context.Background()
	allAccounts, err := c.accRepo.ListAccountsByStatus(ctx, c.db, repositories.All)
	if err != nil {
		return err
	}

	var errs []error
	for _, account := range allAccounts {
		if err = c.validateAndUpdateHistory(ctx, account.ID); err != nil {
			errs = append(errs, fmt.Errorf("account %s: %w", account.ID, err))
			log.Printf("Account %s failed ledger maintenance: %v\n", account.ID, err)
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	log.Println("Done validating accounts and updating histories!")
	return nil
}

// validateAndUpdateHistory processes one account atomically so the history
// always represents a balance that passed validation.
func (c *CronService) validateAndUpdateHistory(ctx context.Context, accountID uuid.UUID) error {
	accCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return c.unitOfWork.Do(accCtx, func(tx repositories.DBTX) error {
		// Validate the account
		account, uowErr := c.accRepo.LockAccount(accCtx, tx, accountID)
		if uowErr != nil {
			return uowErr
		}
		latestHist, uowErr := c.accHistRepo.GetLatestHistory(accCtx, tx, account.ID)
		if uowErr != nil {
			return uowErr
		}
		transactionSum, uowErr := c.tranRepo.GetTransactionSum(
			accCtx,
			tx,
			account.ID,
			latestHist.LoggedTime,
		)
		if uowErr != nil {
			return uowErr
		}

		discrepancy := calculateDiscrepancy(account, latestHist, transactionSum)
		updatedAccount, uowErr := c.accRepo.SetAccountDiscrepancy(accCtx, tx, account.ID, discrepancy)
		if uowErr != nil {
			return uowErr
		}
		if discrepancy != 0 {
			// If the account is in consistent state, skip creating the history checkpoint
			log.Printf("Account %s flagged with discrepancy %d cents; history skipped\n", updatedAccount.ID, discrepancy)
			return nil
		} else {
			log.Printf("Account %s is OK!\n", account.ID)
		}

		// Create an account history checkpoint every 14 days.
		now := time.Now().UTC()
		if shouldCreateHistory(latestHist, now) {
			insertedHistory, uowErr := c.accHistRepo.InsertHistory(accCtx, tx, models.PostAccHistBody{
				AccountID:    account.ID,
				LoggedTime:   now,
				BalanceCents: account.BalanceCents,
			})
			if uowErr != nil {
				return uowErr
			}
			log.Printf("New history %s created for account %s\n", insertedHistory.ID, account.ID)
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
	if account.Type == models.Credit {
		return account.BalanceCents - latestHist.BalanceCents - transactionSum
	}
	return account.BalanceCents - latestHist.BalanceCents + transactionSum
}

func shouldCreateHistory(latestHist models.AccountHistory, now time.Time) bool {
	return !now.Before(latestHist.LoggedTime.AddDate(0, 0, 14))
}

// Frequency: daily, after ValidateAccountsAndUpdateHistory.
// DeleteOutdatedTransactions deletes transactions older than six months for all accounts.
func (c *CronService) DeleteOutdatedTransactions() error {
	ctx := context.Background()
	accounts, err := c.accRepo.ListAccountsByStatus(ctx, c.db, repositories.All)
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
