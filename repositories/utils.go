package repositories

import (
	"betterov2/models"
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

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

// buildFilterSQL gets the SQL query to filter transaction from the TransactionFilter.
// For this to work, model TransactionFilter needs to have tags: "db" and "operator". Keep that
// in mind whenever we modify the model.
//
// Returns the SQL, the arguments to be passed to pgx
func buildDynamicFilter(
	initComd string, value int64, filter models.TransactionFilter,
) (string, []any) {
	conditions := []string{initComd}
	args := []any{value}
	index := 2

	v := reflect.ValueOf(filter)
	for field, value := range v.Fields() {
		if value.Kind() == reflect.String && value.String() == "" {
			continue
		}
		if value.Type() == reflect.TypeFor[*time.Time]() {
			if ptr, ok := value.Interface().(*time.Time); !ok || ptr == nil {
				continue
			}
			value = value.Elem()
		}
		column := field.Tag.Get("db")
		operator := field.Tag.Get("operator")
		conditions = append(
			conditions,
			fmt.Sprintf("t.%s %s $%d", column, operator, index),
		)
		args = append(args, value.Interface())
		index++
	}
	filterSQL := "WHERE " + strings.Join(conditions, " AND ")

	return filterSQL, args
}
