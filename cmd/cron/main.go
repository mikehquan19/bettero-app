package main

import (
	"betterov2/repositories"
	"betterov2/services"
	"betterov2/setup"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata"

	"github.com/robfig/cron/v3"
)

const (
	jobAll      = "all"
	jobLedger   = "ledger"
	jobCleanup  = "cleanup"
	jobDueDates = "due-dates"

	modeSchedule = "schedule"
	modeOnce     = "once"
)

type scheduledJob struct {
	name     string
	schedule string
	run      func() error
}

func main() {
	log.SetPrefix("[CRON] ")
	log.SetFlags(log.Ldate | log.Ltime | log.LUTC)

	if err := run(); err != nil {
		log.Printf("Cron run failed: %v", err)
		os.Exit(1)
	}
}

func run() error {
	db := setup.ConnectDB()
	defer db.Close()

	accountRepo := repositories.NewAccountRepo()
	accountHistoryRepo := repositories.NewAccountHistoryRepo()
	transactionRepo := repositories.NewTransactionRepo()
	uow := repositories.NewDBUnitOfWork(db)

	cronService := services.NewCronService(
		db,
		uow,
		accountRepo,
		accountHistoryRepo,
		transactionRepo,
	)

	job := envOrDefault("CRON_JOB", jobAll)
	mode := envOrDefault("CRON_MODE", modeSchedule)

	switch mode {
	case modeSchedule:
		return runScheduler(cronService, job)
	case modeOnce:
		return runOnce(cronService, job)
	default:
		return fmt.Errorf("unknown CRON_MODE %q; expected schedule or once", mode)
	}
}

func runOnce(cronService *services.CronService, job string) error {
	switch job {
	case jobAll:
		return runAll(cronService)
	case jobLedger:
		return runJob(jobLedger, cronService.ValidateAccountsAndUpdateHistory)
	case jobCleanup:
		return runJob(jobCleanup, cronService.DeleteOutdatedTransactions)
	case jobDueDates:
		return runJob(jobDueDates, cronService.MoveAccountsDueDate)
	default:
		return fmt.Errorf("unknown CRON_JOB %q; expected all, ledger, cleanup, or due-dates", job)
	}
}

func runScheduler(cronService *services.CronService, selectedJob string) error {
	timezone := envOrDefault("CRON_TIMEZONE", "UTC")
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return fmt.Errorf("load CRON_TIMEZONE %q: %w", timezone, err)
	}

	jobs := []scheduledJob{
		{
			name:     jobLedger,
			schedule: envOrDefault("CRON_LEDGER_SCHEDULE", "0 1 * * *"),
			run:      cronService.ValidateAccountsAndUpdateHistory,
		},
		{
			name:     jobCleanup,
			schedule: envOrDefault("CRON_CLEANUP_SCHEDULE", "0 2 * * *"),
			run:      cronService.DeleteOutdatedTransactions,
		},
		{
			name:     jobDueDates,
			schedule: envOrDefault("CRON_DUE_DATES_SCHEDULE", "0 3 * * *"),
			run:      cronService.MoveAccountsDueDate,
		},
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := cron.PrintfLogger(log.Default())
	scheduler := cron.New(
		cron.WithLocation(location),
		cron.WithChain(
			cron.Recover(logger),
			cron.SkipIfStillRunning(logger),
		),
	)

	registeredJobs := 0
	for _, job := range jobs {
		if selectedJob != jobAll && selectedJob != job.name {
			continue
		}

		if _, err := scheduler.AddFunc(job.schedule, func() {
			if err := runJob(job.name, job.run); err != nil {
				log.Printf("Job failed: %v", err)
			}
		}); err != nil {
			return fmt.Errorf("schedule %s job with %q: %w", job.name, job.schedule, err)
		}
		registeredJobs++
		log.Printf("Scheduled %s job with %q in %s", job.name, job.schedule, location)
	}
	if registeredJobs == 0 {
		return fmt.Errorf("unknown CRON_JOB %q; expected all, ledger, cleanup, or due-dates", selectedJob)
	}

	scheduler.Start()
	log.Printf("Scheduler started in %s; waiting for jobs", location)
	<-ctx.Done()
	<-scheduler.Stop().Done()
	log.Printf("Scheduler stopped")
	return nil
}

func runAll(cronService *services.CronService) error {
	var errs []error

	if err := runJob(jobDueDates, cronService.MoveAccountsDueDate); err != nil {
		errs = append(errs, err)
	}

	ledgerErr := runJob(jobLedger, cronService.ValidateAccountsAndUpdateHistory)
	if ledgerErr != nil {
		errs = append(errs, ledgerErr)
		log.Printf("Skipping cleanup because ledger maintenance failed")
	} else if err := runJob(jobCleanup, cronService.DeleteOutdatedTransactions); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func runJob(name string, job func() error) error {
	startedAt := time.Now()
	log.Printf("Starting %s job", name)

	if err := job(); err != nil {
		return fmt.Errorf("%s job: %w", name, err)
	}

	log.Printf("Finished %s job in %s", name, time.Since(startedAt).Round(time.Millisecond))
	return nil
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
