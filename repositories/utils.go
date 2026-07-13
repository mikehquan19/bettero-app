package repositories

import (
	"betterov2/models"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

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
