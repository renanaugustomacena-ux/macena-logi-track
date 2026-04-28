// Package rifiuti is the second vertical module on the LogiTrack
// kit — the design-partner module for Italian SME waste-transport
// operators (trasportatori di rifiuti speciali) operating around
// the Verona industrial corridor and feeding the RENTRI national
// digital traceability system.
//
// Module contract (formal version in docs/DOMAIN-MODULES.md):
//
//   - Owns its persistent entities: Produttore, Trasportatore,
//     Destinatario (recipient impianto), Rifiuto (waste batch),
//     FIR (Formulario Identificazione Rifiuti — the per-movement
//     digital document), RegistroEntry (carico/scarico log line),
//     AlboRegistration (Albo Nazionale Gestori Ambientali
//     enrolment), plus the typed enums (CERCode, FIRState,
//     AlboCategoria, ImpiantoOperazione).
//
//   - Exposes domain validators (ValidateCER, IsCERPericoloso,
//     ValidateAlboCategoria, FIR.Validate) and the FIR state-machine
//     guard. No I/O. Persistence and HTTP wiring live in the platform
//     layers (repository, handlers).
//
//   - Imports the cross-vertical primitives (errors, time helpers,
//     plate validator from logistics, ADR enums) but does NOT import
//     any other domain module beyond the explicit reuse of generic
//     logistics primitives that are platform-grade. The dependency
//     direction stays one-way: rifiuti → logistics, never the
//     opposite.
//
//   - Ships its own integration sub-packages for regulators that
//     only matter for waste transport (RENTRI, Albo Gestori
//     Ambientali) under internal/integrations/. Cross-vertical
//     integrations (AIDA, Telepass, Albo carrier register) stay
//     where they are.
//
// Italian regulatory anchors:
//
//   - D.Lgs. 152/2006 (Testo Unico Ambientale), Parte IV, Titolo I.
//   - D.M. 4 aprile 2023, n. 59 (regolamento RENTRI).
//   - D.Lgs. 116/2020 (recepimento direttive UE economia circolare).
//   - Decisione 2014/955/UE (catalogo europeo dei rifiuti — EER).
//   - D.M. 120/2014 (regolamento Albo Nazionale Gestori Ambientali).
//   - ADR 2023 (UNECE, ratificato L. 1839/1962) per rifiuti
//     pericolosi soggetti a trasporto strada.
//
// Status: SKELETON for the module shape; entities and validators are
// production-ready Go but the RENTRI client is queued-stub until the
// design-partner trasportatore provides Entratel/SPID delegation
// credentials and ADM/MASE enables sandbox access. The kit doctrine
// (platform / module / per-customer overlay) means swapping the stub
// for a live client is one file in one commit, not a rewrite.
package rifiuti
