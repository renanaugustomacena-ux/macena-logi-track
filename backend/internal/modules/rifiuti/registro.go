package rifiuti

import (
	"errors"
	"time"
)

// RegistroOperazione classifies the direction of an entry in the
// registro cronologico carico/scarico. The producer records "carico"
// when waste is generated and held on site, "scarico" when it leaves
// under a FIR. The carrier records "carico" on load and "scarico"
// on unload at the destinatario. The destinatario records "carico"
// on accept.
//
// Reference: D.Lgs. 152/2006 art. 190 (registro di carico e scarico).
type RegistroOperazione string

const (
	RegistroCarico  RegistroOperazione = "carico"
	RegistroScarico RegistroOperazione = "scarico"
)

// RegistroEntry is a single line in the registro cronologico. The
// registro is append-only by law: corrections happen as new entries
// of opposite sign, never as in-place edits. The platform enforces
// this by rejecting any attempt to mutate an existing entry through
// the repository layer; the database itself adds a CHECK constraint
// in the migration so the invariant survives a buggy service.
//
// The registro is per-operatore (one logical registro per Produttore,
// per Trasportatore, per Destinatario), so the OperatoreID + role
// pair partitions the table. Every entry carries the FIR id when
// the movement is one tied to a formulario; entries without a FIR
// id capture internal movements (es. movimentazione in deposito
// temporaneo) that the regulation still requires.
type RegistroEntry struct {
	ID                string             `bson:"_id,omitempty" json:"id"`
	TenantID          string             `bson:"tenant_id" json:"tenantId"`
	OperatoreID       string             `bson:"operatore_id" json:"operatoreId"`
	OperatoreRuolo    OperatoreRuolo     `bson:"operatore_ruolo" json:"operatoreRuolo"`
	NumeroProgressivo int64              `bson:"numero_progressivo" json:"numeroProgressivo"`
	DataOperazione    time.Time          `bson:"data_operazione" json:"dataOperazione"`
	Operazione        RegistroOperazione `bson:"operazione" json:"operazione"`
	CER               CERCode            `bson:"cer" json:"cer"`
	QuantitaGrammi    int64              `bson:"quantita_grammi" json:"quantitaGrammi"`
	StatoFisico       StatoFisico        `bson:"stato_fisico" json:"statoFisico"`
	FIRID             string             `bson:"fir_id,omitempty" json:"firId,omitempty"`
	NumeroRENTRI      string             `bson:"numero_rentri,omitempty" json:"numeroRentri,omitempty"`
	Annotazioni       string             `bson:"annotazioni,omitempty" json:"annotazioni,omitempty"`
	CreatedAt         time.Time          `bson:"created_at" json:"createdAt"`
}

// OperatoreRuolo identifies which role's registro the entry belongs
// to. A single legal entity may hold multiple registri (e.g. a
// trasportatore that also operates a deposito di stoccaggio is both
// trasportatore and produttore for distinct CER streams).
type OperatoreRuolo string

const (
	RuoloProduttore    OperatoreRuolo = "produttore"
	RuoloTrasportatore OperatoreRuolo = "trasportatore"
	RuoloDestinatario  OperatoreRuolo = "destinatario"
	RuoloIntermediario OperatoreRuolo = "intermediario"
)

// ErrRegistroEntryImmutable is returned when a caller attempts to
// modify or delete an existing RegistroEntry. Corrections must be
// applied as new compensating entries.
var ErrRegistroEntryImmutable = errors.New("registro cronologico entry is append-only")

// ErrRegistroProgressivoConflict is returned when two entries for
// the same (operatore_id, operatore_ruolo) carry the same
// numero_progressivo. Progressive numbering must be strictly
// monotonic per operatore-ruolo pair.
var ErrRegistroProgressivoConflict = errors.New("registro progressivo conflict")

// Validate checks the lightweight invariants for a registro entry.
// Cross-entity checks (FIR existence, CER catalogue membership) live
// in the service layer.
func (r *RegistroEntry) Validate() error {
	if r.OperatoreID == "" {
		return ErrRegistroMissingOperatore
	}
	if r.OperatoreRuolo == "" {
		return ErrRegistroMissingRuolo
	}
	if r.NumeroProgressivo <= 0 {
		return ErrRegistroInvalidProgressivo
	}
	if r.DataOperazione.IsZero() {
		return ErrRegistroMissingData
	}
	if r.Operazione != RegistroCarico && r.Operazione != RegistroScarico {
		return ErrRegistroInvalidOperazione
	}
	if err := ValidateCER(string(r.CER)); err != nil {
		return err
	}
	if r.QuantitaGrammi <= 0 {
		return ErrRegistroInvalidQuantita
	}
	if r.StatoFisico == "" {
		return ErrRegistroMissingStatoFisico
	}
	return nil
}

// Registro entry error sentinels.
var (
	ErrRegistroMissingOperatore   = errors.New("registro entry missing operatore id")
	ErrRegistroMissingRuolo       = errors.New("registro entry missing operatore ruolo")
	ErrRegistroInvalidProgressivo = errors.New("registro entry numero progressivo must be positive")
	ErrRegistroMissingData        = errors.New("registro entry missing data operazione")
	ErrRegistroInvalidOperazione  = errors.New("registro entry operazione must be carico or scarico")
	ErrRegistroInvalidQuantita    = errors.New("registro entry quantita must be positive")
	ErrRegistroMissingStatoFisico = errors.New("registro entry missing stato fisico")
)

// RetentionYears is the statutory minimum retention period for the
// registro cronologico after the D.Lgs. 116/2020 reform of D.Lgs.
// 152/2006 art. 190 comma 4: three years from the date of the last
// registration. The previous five-year rule was reduced as part of
// the broader transposition of Directive (EU) 2018/851. The RENTRI
// regulation (D.M. 4 aprile 2023, n. 59) confirms the three-year
// floor and adds the requirement that the digital registro be
// transferred to a conservation system at least once per year per
// the AgID Linee Guida ex CAD D.Lgs. 82/2005.
//
// The constant exposes the legal floor; deployments may configure a
// longer retention via the platform's data-residency settings when
// a customer wants extra margin for Carabinieri NOE inspections
// that historically look back further than the statutory minimum.
const RetentionYears = 3
