package rifiuti

import (
	"errors"
	"testing"
	"time"
)

func TestTrasportatoreCanCarry(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)
	future := now.Add(365 * 24 * time.Hour)
	past := now.Add(-24 * time.Hour)

	cases := []struct {
		name    string
		cat     AlboCategoria
		expiry  time.Time
		cer     CERCode
		when    time.Time
		wantErr error
	}{
		{"Cat 4 valid + non-pericoloso", AlboCat4, future, "150106", now, nil},
		{"Cat 4 + pericoloso rejected", AlboCat4, future, "130205*", now, ErrAlboCategoriaInsufficient},
		{"Cat 5 + pericoloso accepted", AlboCat5, future, "130205*", now, nil},
		{"Cat 5 + non-pericoloso accepted (superset)", AlboCat5, future, "150106", now, nil},
		{"Cat 6 cross-border + pericoloso", AlboCat6, future, "130205*", now, nil},
		{"Cat 2-bis own waste pericoloso", AlboCat2Bis, future, "130205*", now, nil},
		{"Expired Albo rejected", AlboCat5, past, "130205*", now, ErrAlboScaduto},
		{"Cat 1 (urbani) rejected for speciali", AlboCat1, future, "150106", now, ErrAlboCategoriaInsufficient},
		{"Cat 8 (intermediario) rejected for transport", AlboCat8, future, "150106", now, ErrAlboCategoriaInsufficient},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tr := &Trasportatore{
				AlboCategoria: tc.cat,
				AlboScadenza:  tc.expiry,
			}
			err := tr.CanCarry(tc.cer, tc.when)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("CanCarry(%q, %v) = %v, want %v", tc.cer, tc.when, err, tc.wantErr)
			}
		})
	}
}

func TestDestinatarioCanReceive(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)
	future := now.Add(365 * 24 * time.Hour)
	past := now.Add(-24 * time.Hour)

	dest := &Destinatario{
		AutorizzazioneScadenza: future,
		CERAutorizzati:         []CERCode{"150106", "130205*"},
		OperazioniAutorizzate:  []ImpiantoOperazione{OperazioneR3, OperazioneR13, OperazioneD9},
	}

	cases := []struct {
		name    string
		cer     CERCode
		op      ImpiantoOperazione
		when    time.Time
		dest    *Destinatario
		wantErr error
	}{
		{"authorised CER + R3", "150106", OperazioneR3, now, dest, nil},
		{"authorised pericoloso + D9", "130205*", OperazioneD9, now, dest, nil},
		{"unauthorised CER", "070101*", OperazioneR3, now, dest, ErrDestinatarioNotAuthorisedForCER},
		{"unauthorised operation", "150106", OperazioneR1, now, dest, ErrDestinatarioNotAuthorisedForOperation},
		{
			"expired autorizzazione",
			"150106", OperazioneR3, now,
			&Destinatario{
				AutorizzazioneScadenza: past,
				CERAutorizzati:         []CERCode{"150106"},
				OperazioniAutorizzate:  []ImpiantoOperazione{OperazioneR3},
			},
			ErrDestinatarioAutorizzazioneScaduta,
		},
		{
			"CER normalised across whitespace",
			"15 01 06", OperazioneR3, now, dest, nil,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.dest.CanReceive(tc.cer, tc.op, tc.when)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("CanReceive(%q, %q) = %v, want %v", tc.cer, tc.op, err, tc.wantErr)
			}
		})
	}
}
