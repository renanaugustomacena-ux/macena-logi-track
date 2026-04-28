package logistics

import (
	"errors"
	"regexp"
	"strings"
)

// Italian plate validators.
//
// The post-1994 format (e.g. "AB123CD") was introduced by the
// Decreto Ministeriale 27 aprile 1994 and was operational from
// 18 February 1994 onward. The DM 27/04/1994 art. 2 specifies the
// alphabet for the four letter positions and explicitly excludes
// the letters I, O, Q and U because they are visually confusable
// with the digits 1, 0, 0, and the letter V respectively. Historical
// plates — still legal on "vintage" vehicles registered before
// 1994 — follow the older province-prefixed shape (e.g. "VR-123456",
// "MI-ABC-123"). We accept both because carriers still operate
// rebuilt Fiat 242 vans for specialised routes.
//
// References:
//   - D.Lgs. 30 aprile 1992, n. 285 (Codice della Strada), art. 100
//     — general plate visibility / readability requirements.
//   - DPR 16 dicembre 1992, n. 495 (Regolamento di esecuzione del
//     Codice della Strada) — implementing regulation.
//   - Decreto Ministeriale 27 aprile 1994 — formato delle targhe
//     post-1994 e alfabeto ammesso (esclude I, O, Q, U).
//   - ACI circolare 17/2003.
//
// Historical plates use a province prefix (MI, VR, TO, ...) followed by
// a hyphen or space and a sequence of 4–6 digits/letters. We require
// the explicit separator so shorter inputs like "AB12CD" do not match.
//
// The post-1994 character class is split into [A-H] ∪ [J-N] ∪ [P] ∪
// [R-T] ∪ [V-Z]. Letters excluded: I (between H and J), O (between N
// and P), Q (between P and R), U (between T and V).
var (
	platePost1994 = regexp.MustCompile(`^[A-HJ-NPR-TV-Z]{2}[0-9]{3}[A-HJ-NPR-TV-Z]{2}$`)
	plateHistoric = regexp.MustCompile(`^[A-Z]{2}[-\s][0-9A-Z]{4,6}$`)
)

// ErrInvalidItalianPlate is returned when a plate fails both
// post-1994 and historical validation.
var ErrInvalidItalianPlate = errors.New("invalid Italian plate format")

// NormalisePlate upper-cases the input and collapses all internal
// whitespace. It preserves a single explicit separator (space or dash)
// between province prefix and serial so the historical validator can
// tell them apart from modern plates.
func NormalisePlate(p string) string {
	p = strings.ToUpper(strings.TrimSpace(p))
	// Collapse repeated internal whitespace but keep a single separator.
	for strings.Contains(p, "  ") {
		p = strings.ReplaceAll(p, "  ", " ")
	}
	return p
}

// ValidatePlate returns nil iff the plate matches either the modern
// (AB123CD) or a historical Italian format (XX-NNNNNN).
//
// The historical form is only accepted when the separator is present
// so a post-1994 plate like "AB123CD" is never mis-classified as
// "AB" + "123CD".
func ValidatePlate(p string) error {
	n := NormalisePlate(p)
	// Modern match: the post-1994 regex ignores separators, so first
	// collapse them.
	compact := strings.ReplaceAll(n, " ", "")
	compact = strings.ReplaceAll(compact, "-", "")
	if platePost1994.MatchString(compact) {
		return nil
	}
	if plateHistoric.MatchString(n) {
		return nil
	}
	return ErrInvalidItalianPlate
}

// ADRClass enumerates the ADR (Accord européen relatif au transport
// international des marchandises Dangereuses par Route) classes as
// defined by Annex A of the ADR 2023 agreement (UNECE). Italy ratified
// the agreement via L. 12 agosto 1962, n. 1839.
type ADRClass string

const (
	ADRClass1   ADRClass = "1"   // explosives
	ADRClass2   ADRClass = "2"   // gases
	ADRClass3   ADRClass = "3"   // flammable liquids
	ADRClass4_1 ADRClass = "4.1" // flammable solids
	ADRClass4_2 ADRClass = "4.2" // spontaneously combustible
	ADRClass4_3 ADRClass = "4.3" // dangerous when wet
	ADRClass5_1 ADRClass = "5.1" // oxidizers
	ADRClass5_2 ADRClass = "5.2" // organic peroxides
	ADRClass6_1 ADRClass = "6.1" // toxic substances
	ADRClass6_2 ADRClass = "6.2" // infectious substances
	ADRClass7   ADRClass = "7"   // radioactive
	ADRClass8   ADRClass = "8"   // corrosives
	ADRClass9   ADRClass = "9"   // miscellaneous
)

// ATPClass enumerates the ATP (Accord sur les Transports Perissables)
// vehicle categories per the 1970 UN agreement. Italy ratified via
// L. 2 maggio 1977, n. 264. The class drives refrigeration certification
// requirements for perishable food movements in and out of the Verona
// Ortofrutticolo market.
type ATPClass string

const (
	ATPIR  ATPClass = "IR"  // insulated — reinforced
	ATPRNA ATPClass = "RNA" // refrigerated — normal — class A
	ATPRRB ATPClass = "RRB" // refrigerated — reinforced — class B
	ATPFRC ATPClass = "FRC" // mechanically refrigerated — class C
	ATPIN  ATPClass = "IN"  // insulated — normal
)

// TelepassTollCode — a short code from the AISCAT dictionary used on
// invoices and CMR documents to reference the tolled stretch. We store
// just the code and a display label; the full gazetted list is
// periodically synchronised via a background job (see docs/RUNBOOK.md).
type TelepassTollCode struct {
	Code  string `bson:"code" json:"code"`   // e.g. "A22-VR-BZ"
	Label string `bson:"label" json:"label"` // e.g. "A22 Verona Nord - Bolzano Sud"
}
