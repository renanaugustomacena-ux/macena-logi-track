package rifiuti

import (
	"errors"
	"testing"
	"time"

	"github.com/logitrack/backend/internal/modules/logistics"
)

func newFIR() *FIR {
	return &FIR{
		ProduttoreID:             "p-1",
		TrasportatoreID:          "t-1",
		DestinatarioID:           "d-1",
		CER:                      "150106",
		StatoFisico:              StatoFisicoSolidoNonPolverulento,
		QuantitaDichiarataGrammi: 12_000_000, // 12 t
		OperazioneDestino:        OperazioneR3,
	}
}

func TestFIRValidate(t *testing.T) {
	t.Parallel()

	t.Run("happy path defaults state", func(t *testing.T) {
		t.Parallel()
		f := newFIR()
		if err := f.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if f.State != FIRDraft {
			t.Fatalf("expected default state %q, got %q", FIRDraft, f.State)
		}
	})

	t.Run("missing produttore rejected", func(t *testing.T) {
		t.Parallel()
		f := newFIR()
		f.ProduttoreID = ""
		if err := f.Validate(); !errors.Is(err, ErrFIRMissingProduttore) {
			t.Fatalf("got %v, want ErrFIRMissingProduttore", err)
		}
	})

	t.Run("invalid CER rejected", func(t *testing.T) {
		t.Parallel()
		f := newFIR()
		f.CER = "bogus"
		if err := f.Validate(); !errors.Is(err, ErrInvalidCERCode) {
			t.Fatalf("got %v, want ErrInvalidCERCode", err)
		}
	})

	t.Run("zero quantita rejected", func(t *testing.T) {
		t.Parallel()
		f := newFIR()
		f.QuantitaDichiarataGrammi = 0
		if err := f.Validate(); !errors.Is(err, ErrFIRInvalidQuantita) {
			t.Fatalf("got %v, want ErrFIRInvalidQuantita", err)
		}
	})

	t.Run("pericoloso missing ADR rejected", func(t *testing.T) {
		t.Parallel()
		f := newFIR()
		f.CER = "130205*"
		f.Caratteristiche = []HPClass{HP3}
		if err := f.Validate(); !errors.Is(err, ErrFIRPericolosoMissingADR) {
			t.Fatalf("got %v, want ErrFIRPericolosoMissingADR", err)
		}
	})

	t.Run("pericoloso missing HP rejected", func(t *testing.T) {
		t.Parallel()
		f := newFIR()
		f.CER = "130205*"
		f.ADRClass = logistics.ADRClass3
		f.NumeroONU = "UN1202"
		if err := f.Validate(); !errors.Is(err, ErrFIRPericolosoMissingHP) {
			t.Fatalf("got %v, want ErrFIRPericolosoMissingHP", err)
		}
	})

	t.Run("pericoloso fully populated", func(t *testing.T) {
		t.Parallel()
		f := newFIR()
		f.CER = "130205*"
		f.ADRClass = logistics.ADRClass3
		f.NumeroONU = "UN1202"
		f.Caratteristiche = []HPClass{HP3, HP14}
		if err := f.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestFIRAdvanceState(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)

	type step struct {
		to      FIRState
		wantErr error
	}
	cases := []struct {
		name  string
		start FIRState
		steps []step
	}{
		{
			"happy path full lifecycle",
			FIRDraft,
			[]step{
				{FIRVidimato, nil},
				{FIRConsegnatoTrasportatore, nil},
				{FIRInTransito, nil},
				{FIRConsegnatoDestinatario, nil},
				{FIRChiuso, nil},
			},
		},
		{
			"draft annullato direct",
			FIRDraft,
			[]step{{FIRAnnullato, nil}},
		},
		{
			"vidimato annullato direct",
			FIRVidimato,
			[]step{{FIRAnnullato, nil}},
		},
		{
			"intransit respinto then chiuso",
			FIRInTransito,
			[]step{
				{FIRRespinto, nil},
				{FIRChiuso, nil},
			},
		},
		{
			"draft cannot skip to in_transito",
			FIRDraft,
			[]step{{FIRInTransito, ErrFIRInvalidStateTransition}},
		},
		{
			"chiuso is terminal",
			FIRChiuso,
			[]step{{FIRDraft, ErrFIRInvalidStateTransition}},
		},
		{
			"annullato is terminal",
			FIRAnnullato,
			[]step{{FIRVidimato, ErrFIRInvalidStateTransition}},
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			f := newFIR()
			f.State = tc.start
			for i, s := range tc.steps {
				err := f.AdvanceState(s.to, now.Add(time.Duration(i)*time.Hour))
				if !errors.Is(err, s.wantErr) {
					t.Fatalf("step %d AdvanceState(%q from %q) = %v, want %v", i, s.to, f.State, err, s.wantErr)
				}
			}
		})
	}
}

func TestFIRCopiaProduttoreOverdue(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)

	t.Run("not signed yet -> not overdue", func(t *testing.T) {
		t.Parallel()
		f := newFIR()
		if f.CopiaProduttoreOverdue(now) {
			t.Fatalf("unsigned FIR cannot be overdue")
		}
	})

	t.Run("returned -> not overdue", func(t *testing.T) {
		t.Parallel()
		f := newFIR()
		f.FirmaDestinatarioAt = now.Add(-100 * 24 * time.Hour)
		f.CopiaProduttoreReturnedAt = now.Add(-1 * 24 * time.Hour)
		if f.CopiaProduttoreOverdue(now) {
			t.Fatalf("returned copy cannot be overdue")
		}
	})

	t.Run("inside 90 days -> not overdue", func(t *testing.T) {
		t.Parallel()
		f := newFIR()
		f.FirmaDestinatarioAt = now.Add(-89 * 24 * time.Hour)
		if f.CopiaProduttoreOverdue(now) {
			t.Fatalf("89 days is inside the 90-day window")
		}
	})

	t.Run("past 90 days -> overdue", func(t *testing.T) {
		t.Parallel()
		f := newFIR()
		f.FirmaDestinatarioAt = now.Add(-91 * 24 * time.Hour)
		if !f.CopiaProduttoreOverdue(now) {
			t.Fatalf("91 days exceeds the 90-day statutory window")
		}
	})
}
