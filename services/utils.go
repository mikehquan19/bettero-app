package services

import (
	"betterov2/models"
	"fmt"
	"math"
	"reflect"
	"strings"
	"time"
)

// ternary operator since it's not built-in in Go
func ternary[T any](condition bool, trueVal, falseVal T) T {
	if condition {
		return trueVal
	}
	return falseVal
}

// round rounds a float64 to 2 decimal places
func round(x float64) float64 {
	return math.Round(x*100) / 100
}

// buildFilterSQL gets the SQL query to filter transaction from the TransactionFilter.
//
// For this to work, model TransactionFilter needs to have tags: "db" and "operator". Keep that
// in mind whenever we modify the model.
//
// Returns the SQL, the arguments to be passed to pgx
func buildFilterSQL(
	initComd string, value int64, tranFilter models.TransactionFilter,
) (string, []any) {
	conditions := []string{initComd}
	args := []any{value}
	index := 2

	v := reflect.ValueOf(tranFilter)
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
