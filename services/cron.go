package services

import (
	"betterov2/models"
	"betterov2/repositories"
	"context"
	"log"
	"slices"
	"time"

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
func (cs *CronService) MoveAccountsDueDate() error {
	ctx := context.Background()
	pastDueAccounts, err := cs.accRepo.ListAllAccounts(ctx, cs.db, repositories.PastDue)
	if err != nil {
		return err
	}

	// Iterate the list of accounts in batch of 50.
	// For each batch, bulk update the due date of the accounts of that batch
	batchSize := 50
	index := 1
	for batch := range slices.Chunk(pastDueAccounts, batchSize) {
		batchCtx, cancel := context.WithTimeout(ctx, 30*time.Second)

		accountIDs := make([]int64, 0, len(batch))
		for _, account := range batch {
			accountIDs = append(accountIDs, account.ID)
		}

		// Bulk update the list of account IDs
		numUpdated, err := cs.accRepo.MoveAccountsDueDate(batchCtx, cs.db, accountIDs)
		// Cancel the batch
		cancel()
		if err != nil {
			return err
		}
		log.Printf("%d accounts of batch %d have been updated!\n", numUpdated, index)
		index++
	}

	log.Printf("Update done!\n")
	return nil
}

// ValidateAccountsAndInsertHistory validates all the account balances.
//
// TODO: Address the limitation that if one account fails, the retry will re-validate
// the successfully validated accounts.
func (cs *CronService) ValidateAccounts() error {
	ctx := context.Background()
	accounts, err := cs.accRepo.ListAllAccounts(ctx, cs.db, repositories.All)
	if err != nil {
		return err
	}

	for _, account := range accounts {
		err = cs.validate(ctx, account)
		if err != nil {
			log.Printf("Account %d failed validating\n", account.ID)
			return err
		}
	}

	log.Printf("Done validating accounts!\n")
	return nil
}

// Process each individual account for validation
func (cs *CronService) validate(ctx context.Context, account models.Account) error {
	// Each validating is only limited to 90 seconds
	accCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	tx, err := cs.db.Begin(accCtx)
	if err != nil {
		return err
	}
	defer tx.Rollback(accCtx) //nolint:errcheck

	// Validate the account and flag it if there's discrepancy
	// Get the latest balance history and sum of all transactions from the time
	latestHist, err := cs.accHistRepo.GetLatestHistory(accCtx, tx, account.ID)
	if err != nil {
		return err
	}
	sum, err := cs.tranRepo.GetTransactionSum(accCtx, tx, account.ID, latestHist.LoggedTime)
	if err != nil {
		return err
	}

	// Latest balance + sum of transactions = the current balance.
	discrepancy := account.Balance - If(
		account.Type == "Debit",
		latestHist.Balance-sum,
		latestHist.Balance+sum,
	)

	if discrepancy != 0 {
		// Flag the account because there is discrepancy here
		flaggedAcc, err := cs.accRepo.FlagAccount(accCtx, tx, account.ID, discrepancy)
		if err != nil {
			return err
		}
		log.Printf("Account %d has been flagged!\n", flaggedAcc.ID)
	} else {
		log.Printf("Account %d is OK!\n", account.ID)
	}

	if err = tx.Commit(accCtx); err != nil {
		return err
	}

	return nil
}

// UpdateHistory will create the new history of account whose latest history is past 2 weeks,
// and delete the history that is older than 6 months ago.
func (cs *CronService) UpdateHistory() error {
	ctx := context.Background()
	unflaggedAccounts, err := cs.accRepo.ListAllAccounts(ctx, cs.db, repositories.Unflagged)
	if err != nil {
		return err
	}

	for _, account := range unflaggedAccounts {
		err = cs.updateHistory(ctx, account)
		if err != nil {
			log.Printf("Error while executing account %d\n", account.ID)
			return err
		}
	}

	log.Printf("Done creating new history!\n")
	return nil
}

// Process each individual account when updating the balance history
func (cs *CronService) updateHistory(ctx context.Context, account models.Account) error {
	accCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	tx, err := cs.db.Begin(accCtx)
	if err != nil {
		return err
	}
	defer tx.Rollback(accCtx) //nolint:errcheck

	latestHist, err := cs.accHistRepo.GetLatestHistory(accCtx, tx, account.ID)
	if err != nil {
		return err
	}

	// Create the account history every month
	if latestHist.LoggedTime.Before(time.Now().AddDate(0, -1, 0)) {
		body := models.PostAccHistBody{
			AccountId:  account.ID,
			LoggedTime: time.Now(),
			Balance:    account.Balance,
		}
		_, err := cs.accHistRepo.InsertHistory(accCtx, tx, body)
		if err != nil {
			return err
		}
		log.Printf("New history created for account %d", account.ID)
	} else {
		log.Printf("No history created for account %d", account.ID)
	}

	// Delete outdated histories of account
	numDeleted, err := cs.accHistRepo.DeleteOutdatedHistories(accCtx, tx, account.ID)
	if err != nil {
		return err
	}
	log.Printf("Deleted %d histories of account %d", numDeleted, account.ID)

	if err = tx.Commit(accCtx); err != nil {
		return err
	}

	return nil
}

// Delete all the outdated transactions (transactions that are 6 months old) of
// all the accounts
func (cs *CronService) DeleteOutdatedTransactions() error {
	ctx := context.Background()
	accounts, err := cs.accRepo.ListAllAccounts(ctx, cs.db, repositories.All)
	if err != nil {
		return err
	}

	for _, account := range accounts {
		accCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
		numDeleted, err := cs.tranRepo.DeleteOutdatedTransactions(accCtx, cs.db, account.ID)
		// Cancel the context immediately after this
		cancel()
		if err != nil {
			log.Printf("Error while executing account %d\n", account.ID)
			return err
		}
		log.Printf("Deleted %d outdated transactions of account %d", numDeleted, account.ID)
	}

	log.Printf("Done deleting transactions!")
	return nil
}
