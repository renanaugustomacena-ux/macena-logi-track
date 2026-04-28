package rifiuti

import (
	"errors"
	"time"

	"github.com/logitrack/backend/internal/modules/logistics"
)

// FIRState enumerates the lifecycle states of a Formulario di
// Identificazione del Rifiuto under RENTRI. The paper FIR (D.M.
// 145/1998) had three vidimazioni and four manual signatures; the
// digital RENTRI FIR replaces those with a state machine that the
// platform persists and audits.
//
// State transitions are guarded by FIR.Validate and the
// AdvanceState helper so the entity can never be driven into a
// contradictory state by a misbehaving carrier integration.
//
// References:
//   - D.Lgs. 152/2006 art. 193 (formulario, contenuti minimi).
//   - D.M. 4 aprile 2023, n. 59 (RENTRI: registro elettronico,
//     vidimazione digitale, schema XML del formulario).
//   - Circolare MASE / Albo Gestori Ambientali sulle scadenze di
//     attivazione RENTRI (rollout per fasce dimensionali).
type FIRState string

const (
	// FIRDraft — created in the platform, not yet vidimated by
	// RENTRI. Editable. Cannot accompany a movement.
	FIRDraft FIRState = "draft"
	// FIRVidimato — RENTRI has issued the vidimazione digitale and
	// the unique number. Signed by the produttore. Frozen content.
	FIRVidimato FIRState = "vidimato"
	// FIRConsegnatoTrasportatore — the trasportatore has signed the
	// FIR upon load (T+0 of the journey).
	FIRConsegnatoTrasportatore FIRState = "consegnato_trasportatore"
	// FIRInTransito — the consignment is on the road. GPS waypoints
	// are appended through the logistics module's Waypoint type.
	FIRInTransito FIRState = "in_transito"
	// FIRConsegnatoDestinatario — the destinatario has signed
	// receipt with the actual weight and any deviations from the
	// declared content. The producer's "copia produttore" must be
	// returned within 90 days.
	FIRConsegnatoDestinatario FIRState = "consegnato_destinatario"
	// FIRRespinto — the destinatario refused all or part of the
	// load. Triggers a return-leg FIR with reverse parties.
	FIRRespinto FIRState = "respinto"
	// FIRChiuso — the producer has received the signed copy back
	// and the chain of custody is closed. Append-only after this
	// point.
	FIRChiuso FIRState = "chiuso"
	// FIRAnnullato — voided before movement. Permitted only from
	// FIRDraft or FIRVidimato.
	FIRAnnullato FIRState = "annullato"
)

// FIR is the per-movement document that accompanies a load of
// rifiuti speciali from produttore to destinatario. The struct
// carries the data fields RENTRI's XML schema mandates plus the
// LogiTrack-specific telemetry and chain-of-custody fields.
//
// The FIR composes — does NOT duplicate — the logistics.Shipment.
// Every FIR has exactly one underlying logistics.Shipment whose
// Waypoints, Driver and Vehicle are reused. The FIR adds the
// regulatory metadata (CER, weight, RENTRI number, signatures,
// state machine) that wine, freight forwarders, and other
// verticals do not need.
type FIR struct {
	ID            string `bson:"_id,omitempty" json:"id"`
	TenantID      string `bson:"tenant_id" json:"tenantId"`
	ShipmentID    string `bson:"shipment_id" json:"shipmentId"`

	// NumeroRENTRI is the unique identifier issued by RENTRI on
	// vidimazione. Format is published in D.M. 59/2023 Allegato 2;
	// shape is alphanumeric, ~16-20 chars, validated against the
	// RENTRI client at ingestion.
	NumeroRENTRI string `bson:"numero_rentri,omitempty" json:"numeroRentri,omitempty"`

	State          FIRState        `bson:"state" json:"state"`
	ProduttoreID   string          `bson:"produttore_id" json:"produttoreId"`
	TrasportatoreID string         `bson:"trasportatore_id" json:"trasportatoreId"`
	DestinatarioID string          `bson:"destinatario_id" json:"destinatarioId"`

	CER            CERCode         `bson:"cer" json:"cer"`
	DescrizioneRifiuto string      `bson:"descrizione_rifiuto" json:"descrizioneRifiuto"`
	StatoFisico    StatoFisico     `bson:"stato_fisico" json:"statoFisico"`
	Caratteristiche []HPClass      `bson:"caratteristiche,omitempty" json:"caratteristiche,omitempty"`

	// Quantita are kept in grams to avoid float drift on weight
	// reconciliation between produttore-declared and
	// destinatario-accepted values. The platform converts to
	// kilograms / quintali / tonnellate at presentation time.
	QuantitaDichiarataGrammi int64 `bson:"quantita_dichiarata_grammi" json:"quantitaDichiarataGrammi"`
	QuantitaAccettataGrammi  int64 `bson:"quantita_accettata_grammi,omitempty" json:"quantitaAccettataGrammi,omitempty"`

	OperazioneDestino ImpiantoOperazione `bson:"operazione_destino" json:"operazioneDestino"`

	// ADR fields — only populated when CER is pericoloso.
	ADRClass    logistics.ADRClass `bson:"adr_class,omitempty" json:"adrClass,omitempty"`
	NumeroONU   string             `bson:"numero_onu,omitempty" json:"numeroOnu,omitempty"`
	GruppoImballaggio string       `bson:"gruppo_imballaggio,omitempty" json:"gruppoImballaggio,omitempty"`

	// Vidimazione + signatures.
	VidimatoAt              time.Time `bson:"vidimato_at,omitempty" json:"vidimatoAt,omitempty"`
	FirmaProduttoreAt       time.Time `bson:"firma_produttore_at,omitempty" json:"firmaProduttoreAt,omitempty"`
	FirmaTrasportatoreAt    time.Time `bson:"firma_trasportatore_at,omitempty" json:"firmaTrasportatoreAt,omitempty"`
	FirmaDestinatarioAt     time.Time `bson:"firma_destinatario_at,omitempty" json:"firmaDestinatarioAt,omitempty"`
	CopiaProduttoreReturnedAt time.Time `bson:"copia_produttore_returned_at,omitempty" json:"copiaProduttoreReturnedAt,omitempty"`

	// MotivazioneAnnullamento / MotivazioneRespingimento are free
	// text required by RENTRI when the FIR enters one of those
	// terminal states.
	MotivazioneAnnullamento string `bson:"motivazione_annullamento,omitempty" json:"motivazioneAnnullamento,omitempty"`
	MotivazioneRespingimento string `bson:"motivazione_respingimento,omitempty" json:"motivazioneRespingimento,omitempty"`

	CreatedAt time.Time `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time `bson:"updated_at" json:"updatedAt"`
}

// StatoFisico classifies the physical state of the waste — required
// by RENTRI XML schema (campo "Stato Fisico"). Values mirror the
// MUD (Modello Unico di Dichiarazione ambientale) catalogue.
type StatoFisico string

const (
	StatoFisicoSolidoPolverulento  StatoFisico = "solido_polverulento"
	StatoFisicoSolidoNonPolverulento StatoFisico = "solido_non_polverulento"
	StatoFisicoFangosoPalabile     StatoFisico = "fangoso_palabile"
	StatoFisicoLiquido             StatoFisico = "liquido"
	StatoFisicoAeriforme           StatoFisico = "aeriforme"
	StatoFisicoVischiosoSciropposo StatoFisico = "vischioso_sciropposo"
)

// HPClass enumerates the hazard properties HP1..HP15 from
// Reg. UE 1357/2014 (and HP9 amendment 2017/997). Required on the
// FIR for any rifiuto pericoloso.
type HPClass string

const (
	HP1  HPClass = "HP1"  // esplosivo
	HP2  HPClass = "HP2"  // comburente
	HP3  HPClass = "HP3"  // infiammabile
	HP4  HPClass = "HP4"  // irritante
	HP5  HPClass = "HP5"  // tossicità specifica per organi bersaglio
	HP6  HPClass = "HP6"  // tossicità acuta
	HP7  HPClass = "HP7"  // cancerogeno
	HP8  HPClass = "HP8"  // corrosivo
	HP9  HPClass = "HP9"  // infettivo
	HP10 HPClass = "HP10" // tossico per la riproduzione
	HP11 HPClass = "HP11" // mutageno
	HP12 HPClass = "HP12" // liberazione di gas a tossicità acuta
	HP13 HPClass = "HP13" // sensibilizzante
	HP14 HPClass = "HP14" // ecotossico
	HP15 HPClass = "HP15" // capace di sviluppare proprietà pericolose
)

// FIR error sentinels. Kept as exported variables so the HTTP layer
// can map each to a stable RFC 7807 problem type.
var (
	ErrFIRMissingProduttore     = errors.New("FIR missing produttore reference")
	ErrFIRMissingTrasportatore  = errors.New("FIR missing trasportatore reference")
	ErrFIRMissingDestinatario   = errors.New("FIR missing destinatario reference")
	ErrFIRMissingCER            = errors.New("FIR missing CER code")
	ErrFIRInvalidQuantita       = errors.New("FIR quantita must be positive")
	ErrFIRMissingOperazione     = errors.New("FIR missing operazione destino")
	ErrFIRPericolosoMissingADR  = errors.New("rifiuto pericoloso requires ADR class + numero ONU")
	ErrFIRPericolosoMissingHP   = errors.New("rifiuto pericoloso requires at least one HP class")
	ErrFIRInvalidStateTransition = errors.New("invalid FIR state transition")
	ErrFIRCopiaProduttoreOverdue = errors.New("FIR copia produttore overdue (>90 days)")
)

// Validate runs the lightweight, no-I/O invariant checks on the FIR
// that must hold before any persistence. Cross-entity checks
// (Albo / autorizzazione validity against the trasportatore /
// destinatario records) live in the service layer because they
// require repository lookups.
func (f *FIR) Validate() error {
	if f.ProduttoreID == "" {
		return ErrFIRMissingProduttore
	}
	if f.TrasportatoreID == "" {
		return ErrFIRMissingTrasportatore
	}
	if f.DestinatarioID == "" {
		return ErrFIRMissingDestinatario
	}
	if f.CER == "" {
		return ErrFIRMissingCER
	}
	if err := ValidateCER(string(f.CER)); err != nil {
		return err
	}
	if f.QuantitaDichiarataGrammi <= 0 {
		return ErrFIRInvalidQuantita
	}
	if f.OperazioneDestino == "" {
		return ErrFIRMissingOperazione
	}
	if IsCERPericoloso(string(f.CER)) {
		if f.ADRClass == "" || f.NumeroONU == "" {
			return ErrFIRPericolosoMissingADR
		}
		if len(f.Caratteristiche) == 0 {
			return ErrFIRPericolosoMissingHP
		}
	}
	if f.State == "" {
		f.State = FIRDraft
	}
	return nil
}

// firTransitions encodes the legal state-machine edges. Any
// transition not listed here is rejected by AdvanceState. Putting
// the edges in a map makes it trivial to audit + extend (e.g. add
// a "scaduto" timeout edge in a future commit).
var firTransitions = map[FIRState]map[FIRState]bool{
	FIRDraft: {
		FIRVidimato:  true,
		FIRAnnullato: true,
	},
	FIRVidimato: {
		FIRConsegnatoTrasportatore: true,
		FIRAnnullato:               true,
	},
	FIRConsegnatoTrasportatore: {
		FIRInTransito: true,
	},
	FIRInTransito: {
		FIRConsegnatoDestinatario: true,
		FIRRespinto:               true,
	},
	FIRConsegnatoDestinatario: {
		FIRChiuso: true,
	},
	FIRRespinto: {
		// A rejected FIR is closed only after a return-leg FIR is
		// issued; the accept-edge stays in the service layer
		// because it depends on linking the new FIR id.
		FIRChiuso: true,
	},
	FIRChiuso:    {}, // terminal
	FIRAnnullato: {}, // terminal
}

// AdvanceState attempts the supplied transition. Returns
// ErrFIRInvalidStateTransition if the edge is not in the legal map.
// On success, mutates State and refreshes UpdatedAt to the supplied
// instant.
func (f *FIR) AdvanceState(to FIRState, at time.Time) error {
	allowed, ok := firTransitions[f.State]
	if !ok || !allowed[to] {
		return ErrFIRInvalidStateTransition
	}
	f.State = to
	f.UpdatedAt = at
	return nil
}

// CopiaProduttoreOverdue reports whether the producer's signed copy
// has not been returned within the 90-day window from the
// destinatario's signature, per D.Lgs. 152/2006 art. 188-bis comma
// 4. The 90-day clock is statutory: missing it forces the producer
// to denounce the lost copy to the provincia.
func (f *FIR) CopiaProduttoreOverdue(now time.Time) bool {
	if f.FirmaDestinatarioAt.IsZero() {
		return false
	}
	if !f.CopiaProduttoreReturnedAt.IsZero() {
		return false
	}
	deadline := f.FirmaDestinatarioAt.Add(90 * 24 * time.Hour)
	return now.After(deadline)
}
