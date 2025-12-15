package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateLuhn(t *testing.T) {
	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{"valid number 1", "79927398713", true},
		{"valid number 2", "4532015112830366", true},
		{"valid number 3", "12345678903", true},
		{"invalid number 1", "12345678901", false},
		{"invalid number 2", "1234567890", false},
		{"empty string", "", false},
		{"non-numeric", "123abc456", false},
		{"single digit valid", "0", true},
		{"single digit invalid", "1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateLuhn(tt.number)
			assert.Equal(t, tt.want, got)
		})
	}
}
