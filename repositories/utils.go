package repositories

import (
	"betterov2/models"
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBTX represents a database connection that execute SQL queries.
// It is implemented by both pgxpool.Pool and pgx.Tx
type DBTX interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type UnitOfWork interface {
	Do(ctx context.Context, work func(DBTX) error) error
}

type DBUnitOfWork struct {
	pool *pgxpool.Pool
}

func NewDBUnitOfWork(pool *pgxpool.Pool) *DBUnitOfWork {
	return &DBUnitOfWork{
		pool: pool,
	}
}

func (u *DBUnitOfWork) Do(ctx context.Context, work func(DBTX) error) error {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if err := work(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// buildTransactionFilter gets the SQL query to filter transaction from the TransactionFilter.
// Returns the SQL, the arguments to be passed to pgx
func buildTransactionFilter(condition string, value any, filter models.TransactionFilter) (string, []any) {
	conditions := []string{condition}
	args := []any{value}
	index := 1

	if filter.Category != "" {
		index++
		conditions = append(conditions, fmt.Sprintf("t.category = $%d", index))
		args = append(args, filter.Category)
	}

	if filter.Merchant != "" {
		index++
		conditions = append(conditions, fmt.Sprintf("t.merchant = $%d", index))
		args = append(args, filter.Merchant)
	}

	if filter.TranDescription != "" {
		index++
		conditions = append(conditions, fmt.Sprintf("t.description = $%d", index))
		args = append(args, filter.TranDescription)
	}

	if filter.CreatedAtFrom != nil {
		index++
		conditions = append(conditions, fmt.Sprintf("t.created_at >= $%d", index))
		args = append(args, *filter.CreatedAtFrom)
	}

	if filter.CreatedAtTo != nil {
		index++
		conditions = append(conditions, fmt.Sprintf("t.created_at < $%d", index))
		args = append(args, *filter.CreatedAtTo)
	}

	sql := strings.Join(conditions, " AND ")
	return sql, args
}

func isForeignKeyViolation(err error) bool {
	pgErr, ok := err.(*pgconn.PgError)
	return ok && pgErr.Code == "23503"
}
