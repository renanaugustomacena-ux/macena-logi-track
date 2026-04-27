package tests

import (
	"testing"

	"github.com/logitrack/backend/internal/models"
)

// TestValidatePlate exercises the post-1994 Italian plate format and
// the historical province-prefixed form. The post-1994 alphabet must
// reject I, O, Q and U per DM 27/04/1994 art. 2 — the four letters
// excluded for visual disambiguation against digits and other letters.
func TestValidatePlate(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want bool
	}{
		// Modern post-1994 — accepted forms.
		{"modern_uppercase", "AB123CD", true},
		{"modern_lowercase_with_spaces", "ab 123 cd", true},
		{"modern_demo_plate_FA123LT", "FA123LT", true},
		{"modern_separators_collapsed", "FA-123-LT", true},

		// Historical province-prefixed — accepted forms.
		{"historic_dash_VR_6digits", "VR-123456", true},
		{"historic_space_VR_6digits", "VR 123456", true},
		{"historic_dash_ZZ_alnum", "ZZ-9ABC", true},

		// Forbidden letters per DM 27/04/1994 art. 2: I, O, Q, U.
		{"forbidden_I_first_position", "IB123CD", false},
		{"forbidden_O_first_position", "OB123CD", false},
		{"forbidden_Q_first_position", "QB123CD", false},
		{"forbidden_U_first_position", "UB123CD", false},
		{"forbidden_I_second_position", "AI123CD", false},
		{"forbidden_O_third_position", "AB123OD", false},
		{"forbidden_Q_third_position", "AB123QD", false},
		{"forbidden_U_third_position", "AB123UD", false}, // the U bug
		{"forbidden_U_fourth_position", "AB123CU", false},
		{"forbidden_AU123CD_smoke_regression", "AU123CD", false}, // confirmed live
		{"forbidden_all_four_letters", "I0123OO", false},

		// Malformed.
		{"too_short", "AB12CD", false},
		{"all_digits", "1234567", false},
		{"empty", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := models.ValidatePlate(tc.in) == nil
			if got != tc.want {
				t.Errorf("ValidatePlate(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}
