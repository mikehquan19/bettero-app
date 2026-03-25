package services

import (
	"betterov2/setup"
	"math"

	"github.com/jackc/pgx/v5/pgxpool"
)

var database *pgxpool.Pool = setup.ConnectDB()
var userID int64 = 1

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
