package rifiuti

import (
	"errors"
	"testing"
)

func TestValidateCER(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{"valid non-pericoloso", "150106", nil},
		{"valid pericoloso", "130205*", nil},
		{"valid with spaces", "13 02 05*", nil},
		{"valid with dots", "13.02.05", nil},
		{"too short", "12345", ErrInvalidCERCode},
		{"too long", "1234567", ErrInvalidCERCode},
		{"non-numeric", "ABCDEF", ErrInvalidCERCode},
		{"trailing letters", "130205X", ErrInvalidCERCode},
		{"empty", "", ErrInvalidCERCode},
		{"only asterisk", "*", ErrInvalidCERCode},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateCER(tc.input)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("ValidateCER(%q) = %v, want %v", tc.input, err, tc.wantErr)
			}
		})
	}
}

func TestIsCERPericoloso(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  bool
	}{
		{"130205*", true},
		{"13 02 05*", true},
		{"150106", false},
		{"15.01.06", false},
		{"", false},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			got := IsCERPericoloso(tc.input)
			if got != tc.want {
				t.Fatalf("IsCERPericoloso(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestCERChapter(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  string
	}{
		{"170504", "17"}, // construction
		{"190101", "19"}, // waste from waste management
		{"130205*", "13"},
		{"01 03 04*", "01"},
		{"invalid", ""},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			got := CERChapter(tc.input)
			if got != tc.want {
				t.Fatalf("CERChapter(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestValidateAlboCategoria(t *testing.T) {
	t.Parallel()
	tests := []struct {
		c       AlboCategoria
		wantErr error
	}{
		{AlboCat1, nil},
		{AlboCat2Bis, nil},
		{AlboCat4, nil},
		{AlboCat5, nil},
		{AlboCat6, nil},
		{AlboCat8, nil},
		{AlboCat9, nil},
		{AlboCat10, nil},
		{"3", ErrInvalidAlboCategoria},
		{"", ErrInvalidAlboCategoria},
		{"99", ErrInvalidAlboCategoria},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(string(tc.c), func(t *testing.T) {
			t.Parallel()
			err := ValidateAlboCategoria(tc.c)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("ValidateAlboCategoria(%q) = %v, want %v", tc.c, err, tc.wantErr)
			}
		})
	}
}
