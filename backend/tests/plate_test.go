package tests

import (
	"testing"

	"github.com/logitrack/backend/internal/models"
)

func TestValidatePlate(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"AB123CD", true},
		{"ab 123 cd", true},
		{"FA123LT", true},
		{"VR-123456", true},
		{"VR 123456", true},
		{"ZZ-9ABC", true},
		{"I0123OO", false}, // forbidden letter I/O/Q
		{"AB12CD", false},  // too short
		{"1234567", false}, // all digits
		{"", false},
	}
	for _, tc := range cases {
		got := models.ValidatePlate(tc.in) == nil
		if got != tc.want {
			t.Errorf("ValidatePlate(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}
