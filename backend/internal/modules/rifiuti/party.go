package rifiuti

import (
	"errors"
	"time"
)

// Produttore is the waste producer (the industrial site that
// generates the rifiuto). On every FIR the Produttore is the
// upstream party in the chain: their codice fiscale, partita IVA
// and codice unità locale identify the physical site of origin
// for inspection purposes.
//
// In the Italian regulatory model the Produttore retains liability
// for the waste from cradle to certified disposal/recovery — the
// trasportatore and destinatario operate on the producer's behalf.
// This is why the FIR's "copia produttore" must be returned to them
// signed by the destinatario within ninety days, otherwise the
// producer must denounce the missing return to the provincia
// (D.Lgs. 152/2006 art. 188-bis comma 4).
type Produttore struct {
	ID                  string    `bson:"_id,omitempty" json:"id"`
	TenantID            string    `bson:"tenant_id" json:"tenantId"`
	RagioneSociale      string    `bson:"ragione_sociale" json:"ragioneSociale"`
	CodiceFiscale       string    `bson:"codice_fiscale" json:"codiceFiscale"`
	PartitaIVA          string    `bson:"partita_iva" json:"partitaIva"`
	CodiceUnitaLocale   string    `bson:"codice_unita_locale" json:"codiceUnitaLocale"`
	Indirizzo           string    `bson:"indirizzo" json:"indirizzo"`
	CAP                 string    `bson:"cap" json:"cap"`
	Comune              string    `bson:"comune" json:"comune"`
	Provincia           string    `bson:"provincia" json:"provincia"`
	AttivitaCodiceATECO string    `bson:"ateco" json:"ateco"`
	ContattoEmail       string    `bson:"contatto_email,omitempty" json:"contattoEmail,omitempty"`
	ContattoTelefono    string    `bson:"contatto_telefono,omitempty" json:"contattoTelefono,omitempty"`
	CreatedAt           time.Time `bson:"created_at" json:"createdAt"`
	UpdatedAt           time.Time `bson:"updated_at" json:"updatedAt"`
}

// Trasportatore is the carrier moving the rifiuto from the
// Produttore to the Destinatario. Their Albo Nazionale Gestori
// Ambientali enrolment (categoria + classe + numero iscrizione +
// scadenza) is the most operationally critical field on the entity
// because an expired or out-of-class iscrizione invalidates every
// FIR they sign.
//
// When the same legal entity also owns the LogiTrack tenant
// (typical: the trasportatore is our customer), the TenantID
// matches the global platform tenant; when the trasportatore is a
// third party hauling for our customer (un sub-vettore), the
// TenantID is still the customer's, and the Trasportatore record
// carries the sub-vettore's distinct codice fiscale + Albo data.
type Trasportatore struct {
	ID                   string        `bson:"_id,omitempty" json:"id"`
	TenantID             string        `bson:"tenant_id" json:"tenantId"`
	RagioneSociale       string        `bson:"ragione_sociale" json:"ragioneSociale"`
	CodiceFiscale        string        `bson:"codice_fiscale" json:"codiceFiscale"`
	PartitaIVA           string        `bson:"partita_iva" json:"partitaIva"`
	AlboCategoria        AlboCategoria `bson:"albo_categoria" json:"alboCategoria"`
	AlboClasse           AlboClasse    `bson:"albo_classe" json:"alboClasse"`
	AlboNumeroIscrizione string        `bson:"albo_numero_iscrizione" json:"alboNumeroIscrizione"`
	AlboScadenza         time.Time     `bson:"albo_scadenza" json:"alboScadenza"`
	SedeLegale           string        `bson:"sede_legale" json:"sedeLegale"`
	CreatedAt            time.Time     `bson:"created_at" json:"createdAt"`
	UpdatedAt            time.Time     `bson:"updated_at" json:"updatedAt"`
}

// Destinatario is the impianto that receives the rifiuto and
// performs (or further forwards for) a recovery or disposal
// operation. Their autorizzazione provinciale or AIA (autorizzazione
// integrata ambientale) determines which CER codes they may accept
// and which ImpiantoOperazione they may perform on each.
//
// In practice the destinatario population is small (a few hundred
// active impianti in Veneto + Lombardia covers the entire
// addressable corridor for our first ICP) so the platform can ship
// a curated catalogue on day one.
type Destinatario struct {
	ID                     string               `bson:"_id,omitempty" json:"id"`
	TenantID               string               `bson:"tenant_id" json:"tenantId"`
	RagioneSociale         string               `bson:"ragione_sociale" json:"ragioneSociale"`
	CodiceFiscale          string               `bson:"codice_fiscale" json:"codiceFiscale"`
	PartitaIVA             string               `bson:"partita_iva" json:"partitaIva"`
	Indirizzo              string               `bson:"indirizzo" json:"indirizzo"`
	Comune                 string               `bson:"comune" json:"comune"`
	Provincia              string               `bson:"provincia" json:"provincia"`
	AutorizzazioneNumero   string               `bson:"autorizzazione_numero" json:"autorizzazioneNumero"`
	AutorizzazioneScadenza time.Time            `bson:"autorizzazione_scadenza" json:"autorizzazioneScadenza"`
	OperazioniAutorizzate  []ImpiantoOperazione `bson:"operazioni_autorizzate" json:"operazioniAutorizzate"`
	CERAutorizzati         []CERCode            `bson:"cer_autorizzati" json:"cerAutorizzati"`
	CreatedAt              time.Time            `bson:"created_at" json:"createdAt"`
	UpdatedAt              time.Time            `bson:"updated_at" json:"updatedAt"`
}

// ErrAlboScaduto is returned when the carrier's Albo enrolment is
// past its expiry date at the time of FIR validation. A FIR signed
// with an expired Albo is administratively void.
var ErrAlboScaduto = errors.New("Albo Gestori Ambientali iscrizione scaduta")

// ErrAlboCategoriaInsufficient is returned when the carrier's Albo
// categoria does not cover the waste class on the FIR — most
// commonly a Cat 4-only carrier attempting to sign a FIR for a CER
// pericoloso, which requires Cat 5.
var ErrAlboCategoriaInsufficient = errors.New("Albo categoria does not cover waste class")

// ErrDestinatarioNotAuthorisedForCER is returned when the
// destinatario's autorizzazione does not include the CER on the FIR.
var ErrDestinatarioNotAuthorisedForCER = errors.New("destinatario not authorised for this CER code")

// ErrDestinatarioNotAuthorisedForOperation is returned when the
// declared ImpiantoOperazione is not in the destinatario's
// autorizzazione.
var ErrDestinatarioNotAuthorisedForOperation = errors.New("destinatario not authorised for this operation code")

// CanCarry reports whether the trasportatore's current Albo
// enrolment lets them legally move the supplied CER code at the
// supplied moment. The check is purely temporal + categoria-level;
// classe (tonnage band) is enforced separately at FIR-level once
// the load weight is known, and Cat 2-bis (own-waste-only) requires
// an additional service-layer check that producer == carrier.
//
// Cat 5 covers pericolosi (and conto-terzi non-pericolosi as a
// superset). Cat 4 covers conto-terzi non-pericolosi only. Cat 6
// covers cross-border movements regardless of pericolosità (Reg. UE
// 1013/2006); we accept it on the categoria-level check and leave
// the cross-border-leg verification to the FIR service. Cat 2-bis
// covers own-waste only — non-pericolosi senza limite, pericolosi
// solo entro 30 kg/litri al giorno.
func (t *Trasportatore) CanCarry(cer CERCode, at time.Time) error {
	if at.After(t.AlboScadenza) {
		return ErrAlboScaduto
	}
	if IsCERPericoloso(string(cer)) {
		switch t.AlboCategoria {
		case AlboCat5, AlboCat6, AlboCat2Bis:
			return nil
		}
		return ErrAlboCategoriaInsufficient
	}
	switch t.AlboCategoria {
	case AlboCat4, AlboCat5, AlboCat6, AlboCat2Bis:
		return nil
	}
	return ErrAlboCategoriaInsufficient
}

// CanReceive reports whether the destinatario is currently
// authorised to accept the supplied CER code under the supplied
// recovery/disposal operation.
func (d *Destinatario) CanReceive(cer CERCode, op ImpiantoOperazione, at time.Time) error {
	if at.After(d.AutorizzazioneScadenza) {
		return ErrDestinatarioAutorizzazioneScaduta
	}
	cerOK := false
	for _, c := range d.CERAutorizzati {
		if NormaliseCER(string(c)) == NormaliseCER(string(cer)) {
			cerOK = true
			break
		}
	}
	if !cerOK {
		return ErrDestinatarioNotAuthorisedForCER
	}
	for _, o := range d.OperazioniAutorizzate {
		if o == op {
			return nil
		}
	}
	return ErrDestinatarioNotAuthorisedForOperation
}

// ErrDestinatarioAutorizzazioneScaduta is returned when the
// destinatario's autorizzazione has expired at the time of the
// movement.
var ErrDestinatarioAutorizzazioneScaduta = errors.New("destinatario autorizzazione scaduta")
