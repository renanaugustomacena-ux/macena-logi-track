package rifiuti

import (
	"errors"
	"regexp"
	"strings"
)

// CERCode represents a six-digit waste classification code from the
// European List of Waste (Decisione 2014/955/UE which amended
// 2000/532/CE). The code is colloquially called "CER" in Italy
// (Catalogo Europeo dei Rifiuti) and "EER" (European Waste Code) in
// EU documentation; the two terms are synonymous and both refer to
// the same six-digit identifier of the form "AABBCC" where AA is the
// chapter (01..20), BB the sub-chapter, and CC the specific waste.
//
// A trailing asterisk ("AABBCC*") denotes a "rifiuto pericoloso"
// (hazardous waste). Hazardous waste triggers ADR road-transport
// rules, mandates Albo Categoria 5 enrolment for the carrier, and
// requires a higher fine bracket on misclassification (D.Lgs.
// 152/2006 art. 258, comma 4).
//
// References:
//   - Decisione 2014/955/UE (catalogo EER consolidato).
//   - D.Lgs. 152/2006 art. 184 (classificazione + Allegato D).
//   - Reg. UE 1357/2014 (caratteristiche di pericolo HP1..HP15).
type CERCode string

// cerPattern matches six digits optionally followed by a single
// asterisk. We accept whitespace and dots as visual separators and
// strip them before validation so operators can paste codes copied
// from PDFs ("01 03 04*") without the platform rejecting them.
var cerPattern = regexp.MustCompile(`^[0-9]{6}\*?$`)

// ErrInvalidCERCode is returned when the input does not match the
// six-digit (with optional asterisk) shape required by the European
// catalogue.
var ErrInvalidCERCode = errors.New("invalid CER/EER waste code")

// NormaliseCER strips whitespace, dots and dashes and upper-cases the
// resulting string. The output is always either "AABBCC" or
// "AABBCC*".
func NormaliseCER(c string) string {
	c = strings.TrimSpace(c)
	c = strings.ReplaceAll(c, " ", "")
	c = strings.ReplaceAll(c, ".", "")
	c = strings.ReplaceAll(c, "-", "")
	return strings.ToUpper(c)
}

// ValidateCER returns nil iff the input is a syntactically valid CER
// code. Membership in the catalogue is not enforced here because the
// catalogue is large and is updated by Commission decisions; the
// repository layer maintains a synchronised copy and rejects unknown
// codes. The contract of this function is purely shape validation.
func ValidateCER(c string) error {
	if !cerPattern.MatchString(NormaliseCER(c)) {
		return ErrInvalidCERCode
	}
	return nil
}

// IsCERPericoloso reports whether the code carries the hazardous
// asterisk marker. Whitespace and case are normalised first.
func IsCERPericoloso(c string) bool {
	return strings.HasSuffix(NormaliseCER(c), "*")
}

// CERChapter returns the two-digit chapter prefix ("01".."20") of a
// CER code, or empty string if the code is invalid. The chapter is
// the most useful coarse grouping for operator dashboards (chapter
// 17 = construction & demolition, chapter 19 = waste from waste
// management facilities, etc.).
func CERChapter(c string) string {
	n := NormaliseCER(c)
	if cerPattern.MatchString(n) {
		return n[:2]
	}
	return ""
}

// AlboCategoria enumerates the categories of the Albo Nazionale
// Gestori Ambientali enrolment that are relevant to a waste
// transporter. The categoria determines what waste types the carrier
// is legally allowed to move, and must be present, in date, and
// matched to the CER code on every FIR.
//
// References:
//   - D.M. 120/2014 (regolamento Albo).
//   - D.Lgs. 152/2006 art. 212 (iscrizione Albo).
type AlboCategoria string

const (
	// AlboCat1 — raccolta e trasporto rifiuti urbani. Out of scope
	// for the SME industrial transporter that is our first ICP, but
	// included so the enum mirrors the official catalogue.
	AlboCat1 AlboCategoria = "1"
	// AlboCat2Bis — produttore iniziale che trasporta i propri
	// rifiuti. Non pericolosi senza limite, pericolosi solo entro
	// 30 kg/litri al giorno. Iscrizione decennale (vs quinquennale
	// delle altre categorie). Frequente per piccole officine che
	// portano i propri scarti al deposito.
	AlboCat2Bis AlboCategoria = "2-bis"
	// AlboCat4 — raccolta e trasporto di rifiuti speciali NON
	// pericolosi. The minimum enrolment for a carrier hauling
	// industrial waste that is not hazardous.
	AlboCat4 AlboCategoria = "4"
	// AlboCat5 — raccolta e trasporto di rifiuti speciali
	// pericolosi. Required whenever the FIR references at least one
	// CER pericoloso ("AABBCC*").
	AlboCat5 AlboCategoria = "5"
	// AlboCat6 — trasporto transfrontaliero di rifiuti su
	// territorio italiano (Reg. UE 1013/2006). Required when the
	// shipment is part of a notifica transfrontaliera leg.
	AlboCat6 AlboCategoria = "6"
	// AlboCat8 — intermediazione e commercio di rifiuti senza
	// detenzione. Brokers, not carriers proper, but some SME
	// operators hold both Cat 4/5 and Cat 8.
	AlboCat8 AlboCategoria = "8"
	// AlboCat9 — bonifica di siti contaminati. Out of scope for the
	// general carrier; included for catalogue completeness.
	AlboCat9 AlboCategoria = "9"
	// AlboCat10 — bonifica di beni contenenti amianto. Out of scope.
	AlboCat10 AlboCategoria = "10"
)

// AlboClasse enumerates the classes (A..F) of an Albo enrolment.
// Class A is the largest tonnage band, F the smallest. The class is
// recorded on the FIR alongside the categoria so an inspector can
// verify the carrier is operating within the enrolled tonnage limit.
//
// Reference: D.M. 120/2014 art. 9.
type AlboClasse string

const (
	AlboClasseA AlboClasse = "A" // ≥ 200 000 t/anno
	AlboClasseB AlboClasse = "B" // 60 000 – 200 000 t/anno
	AlboClasseC AlboClasse = "C" // 15 000 – 60 000 t/anno
	AlboClasseD AlboClasse = "D" // 6 000 – 15 000 t/anno
	AlboClasseE AlboClasse = "E" // 3 000 – 6 000 t/anno
	AlboClasseF AlboClasse = "F" // < 3 000 t/anno
)

// ImpiantoOperazione enumerates the recovery (R) and disposal (D)
// operation codes from the Allegati B and C of D.Lgs. 152/2006,
// which transpose Annex I and II of Directive 2008/98/CE on waste.
// The destination impianto declares which operation it performs on
// every incoming load, and the FIR records that declaration so the
// chain of custody includes the eventual fate of the material.
type ImpiantoOperazione string

const (
	OperazioneR1  ImpiantoOperazione = "R1"  // utilizzo come combustibile
	OperazioneR3  ImpiantoOperazione = "R3"  // riciclo sostanze organiche
	OperazioneR4  ImpiantoOperazione = "R4"  // riciclo metalli
	OperazioneR5  ImpiantoOperazione = "R5"  // riciclo altri inorganici
	OperazioneR10 ImpiantoOperazione = "R10" // spandimento sul suolo a beneficio agricolo
	OperazioneR12 ImpiantoOperazione = "R12" // scambio rifiuti per sottoporli a R1..R11
	OperazioneR13 ImpiantoOperazione = "R13" // messa in riserva (deposito preliminare)
	OperazioneD1  ImpiantoOperazione = "D1"  // deposito sul/nel suolo (discarica)
	OperazioneD9  ImpiantoOperazione = "D9"  // trattamento fisico-chimico
	OperazioneD10 ImpiantoOperazione = "D10" // incenerimento a terra
	OperazioneD13 ImpiantoOperazione = "D13" // raggruppamento preliminare
	OperazioneD14 ImpiantoOperazione = "D14" // ricondizionamento preliminare
	OperazioneD15 ImpiantoOperazione = "D15" // deposito preliminare
)

// ValidateAlboCategoria returns nil iff the supplied categoria is one
// of the well-known values in the AlboCategoria enum. Empty input is
// rejected because every carrier on a FIR must carry an enrolment.
func ValidateAlboCategoria(c AlboCategoria) error {
	switch c {
	case AlboCat1, AlboCat2Bis, AlboCat4, AlboCat5, AlboCat6, AlboCat8, AlboCat9, AlboCat10:
		return nil
	}
	return ErrInvalidAlboCategoria
}

// ErrInvalidAlboCategoria is returned by ValidateAlboCategoria for an
// unrecognised categoria.
var ErrInvalidAlboCategoria = errors.New("invalid Albo Gestori Ambientali categoria")
