// Package wine is the second vertical module on the LogiTrack kit —
// the design-partner module for Italian cantine exporting bottled
// wine through the Brennero corridor to Germany and Austria.
//
// Status: SKELETON. The module shape is committed so the kit
// architecture is visible, but the implementations are placeholders
// for now. Real implementations land driven by the first design-
// partner cantina's actual workflow, not by guesses from a desk.
//
// Module contract (see docs/DOMAIN-MODULES.md):
//
//   - Owns vertical entities (planned): Cantina, WineLot,
//     AcciseDocument (e-AD / EMCS ARC), BottlingBatch, TemperatureLog.
//   - Wraps logistics: a wine shipment IS a logistics shipment
//     plus wine-specific dispatch metadata. Compose, do not
//     duplicate.
//   - Adds vertical integrations: EMCS (Excise Movement and
//     Control System), SIAN Vitivinicolo, ICQRF tracciabilità,
//     temperature-sensor providers.
//   - Adds Italian regulatory anchors specific to wine: DOC/DOCG
//     zone validation, accise stamp range tracking, ATP class IR
//     enforcement on bulk wine shipments, e-AD lifecycle.
//
// Roadmap is documented in docs/MODULE-WINE.md (created when the
// design-partner cantina is signed and the first concrete
// requirements land).
package wine

// Cantina is the wine-producer entity. Placeholder; real fields
// land when the design-partner cantina's onboarding clarifies what
// is needed (P.IVA, registro vigneti SIAN, codice accisa, DOC
// zone associations, contatti dispatcher, etc.).
type Cantina struct {
	ID       string
	TenantID string
	Name     string
	// TODO: add the rest once we have a design-partner cantina.
}

// WineLot is the unit of wine being moved. Placeholder.
//
// In wine logistics, the "shipment" is downstream of the lot: a
// single export pallet typically contains one or more WineLot,
// each with its own DOC/DOCG zone, vintage, varietal, accise
// stamp range and quality-control documents.
type WineLot struct {
	ID       string
	TenantID string
	// TODO: DOC/DOCG zone, vintage year, varietal, bottling date,
	// accise stamp range, certifications.
}

// AcciseDocument is the e-AD (electronic Accompanying Document)
// the wine shipment must travel under for cross-border excise
// control. The ARC is the EU-wide unique reference. Placeholder.
type AcciseDocument struct {
	ID  string
	ARC string // 21-character EU-wide reference
	// TODO: sender, consignee, products, dispatch date, EMCS state.
}
