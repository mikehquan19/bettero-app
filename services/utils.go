package services

import (
	"math"
)

// round rounds a float64 to 2 decimal places
func round(x float64) float64 {
	return math.Round(x*100) / 100
}
