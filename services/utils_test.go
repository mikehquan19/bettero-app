package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTernary_True(t *testing.T) {
	result := ternary(true, "yes", "no")
	assert.Equal(t, "yes", result)
}

func TestTernary_False(t *testing.T) {
	result := ternary(false, "yes", "no")
	assert.Equal(t, "no", result)
}

func TestTernary_WithNumbers(t *testing.T) {
	result := ternary(5 > 3, 100, 0)
	assert.Equal(t, 100, result)
}

func TestTernary_WithNumbers_False(t *testing.T) {
	result := ternary(5 < 3, 100, 0)
	assert.Equal(t, 0, result)
}

func TestRound_Basic(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
		name     string
	}{
		{1.234, 1.23, "Two decimal places"},
		{1.235, 1.24, "Rounding up"},
		{1.225, 1.23, "Rounding edge case"},
		{0.005, 0.01, "Small number rounding"},
		{100.999, 101.0, "Large number rounding"},
		{50.0, 50.0, "Exact value"},
		{-1.234, -1.23, "Negative number"},
		{-1.235, -1.24, "Negative rounding up"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := round(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRound_ZeroValue(t *testing.T) {
	result := round(0.0)
	assert.Equal(t, 0.0, result)
}
