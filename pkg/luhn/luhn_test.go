package luhn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		orderNum  string
		wantValid bool
	}{
		{"valid from TZ", "12345678903", true},
		{"valid short", "79927398713", true},
		{"valid long", "4242424242424242", true},
		{"invalid checksum", "79927398714", false},
		{"empty", "", false},
		{"zero", "0", false},
		{"zeros", "00", false},
		{"not number", "test", false},
		{"mixed", "f21fad", false},
		{"single digit", "1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantValid, Validate(tt.orderNum))
		})
	}
}
