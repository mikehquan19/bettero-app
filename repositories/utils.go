package repositories

import (
	"betterov2/models"
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// DBTX represents a database connection that can execute queries
// and start transactions. It is implemented by both pgxpool.Pool and pgx.Tx,
// allowing repositories and services to operate on
// either a direct database connection or an active transaction.
type DBTX interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// buildTransactionFilter gets the SQL query to filter transaction from the TransactionFilter.
// Returns the SQL, the arguments to be passed to pgx
func buildTransactionFilter(initCond string, value int64, filter models.TransactionFilter) (string, []any) {
	conditions := []string{initCond}
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
		conditions = append(conditions, fmt.Sprintf("t.tran_description = $%d", index))
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
