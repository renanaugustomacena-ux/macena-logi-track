package rifiuti

import (
	"errors"
	"testing"
	"time"
)

func TestRegistroEntryValidate(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)

	base := func() *RegistroEntry {
		return &RegistroEntry{
			OperatoreID:       "op-1",
			OperatoreRuolo:    RuoloTrasportatore,
			NumeroProgressivo: 1,
			DataOperazione:    now,
			Operazione:        RegistroCarico,
			CER:               "150106",
			QuantitaGrammi:    50_000_000,
			StatoFisico:       StatoFisicoSolidoNonPolverulento,
		}
	}

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		if err := base().Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("missing operatore", func(t *testing.T) {
		t.Parallel()
		r := base()
		r.OperatoreID = ""
		if err := r.Validate(); !errors.Is(err, ErrRegistroMissingOperatore) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("missing ruolo", func(t *testing.T) {
		t.Parallel()
		r := base()
		r.OperatoreRuolo = ""
		if err := r.Validate(); !errors.Is(err, ErrRegistroMissingRuolo) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("zero progressivo", func(t *testing.T) {
		t.Parallel()
		r := base()
		r.NumeroProgressivo = 0
		if err := r.Validate(); !errors.Is(err, ErrRegistroInvalidProgressivo) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("zero quantita", func(t *testing.T) {
		t.Parallel()
		r := base()
		r.QuantitaGrammi = 0
		if err := r.Validate(); !errors.Is(err, ErrRegistroInvalidQuantita) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("invalid CER", func(t *testing.T) {
		t.Parallel()
		r := base()
		r.CER = "bogus"
		if err := r.Validate(); !errors.Is(err, ErrInvalidCERCode) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("invalid operazione", func(t *testing.T) {
		t.Parallel()
		r := base()
		r.Operazione = "bogus"
		if err := r.Validate(); !errors.Is(err, ErrRegistroInvalidOperazione) {
			t.Fatalf("got %v", err)
		}
	})
}

func TestRetentionYearsMatchesD116(t *testing.T) {
	t.Parallel()
	if RetentionYears != 3 {
		t.Fatalf("RetentionYears must be 3 per D.Lgs. 116/2020 (TUA art. 190 c.4 reform); got %d", RetentionYears)
	}
}
