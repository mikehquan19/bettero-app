package services

import (
	"math"
)

// ternary operator since it's not built-in in Go
func If[T any](condition bool, trueVal, falseVal T) T {
	if condition {
		return trueVal
	}
	return falseVal
}

// round rounds a float64 to 2 decimal places
func round(x float64) float64 {
	return math.Round(x*100) / 100
}
