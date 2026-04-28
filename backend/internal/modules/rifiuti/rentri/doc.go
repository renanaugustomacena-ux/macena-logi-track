// Package rentri is the client adapter for the Italian Registro
// Elettronico Nazionale per la Tracciabilità dei Rifiuti (RENTRI),
// the national digital traceability system established by D.M. 4
// aprile 2023, n. 59 (GU Serie Generale n. 126 del 31 maggio 2023,
// in vigore dal 15 giugno 2023) under D.Lgs. 152/2006 art. 188-bis.
//
// The package exposes:
//
//   - Endpoint constants for the RENTRI sandbox (demoapi.rentri.gov.it
//     and demobackoffice.rentri.gov.it) and production
//     (api.rentri.gov.it) environments, sourced from the official
//     OpenAPI documentation.
//
//   - The Client interface that any concrete RENTRI integration must
//     implement (VidimaFIR, TrasmettiMovimento, GetFIR, ListFIRByARC).
//     The interface is small on purpose so test doubles, sandbox
//     adapters, and the eventual production client all stay swappable.
//
//   - A QueuedStub implementation that accepts every call, persists
//     the payload to a FIFO for later replay, and returns deterministic
//     stub identifiers. This is the adapter wired in production until
//     the design-partner trasportatore provides the SPID/CIE/CNS-bound
//     Entratel delegation that unlocks the live API. The stub keeps the
//     rest of the platform writing real FIR data so the day the
//     credentials land the upstream switch is one constructor change.
//
//   - The xFIR XML data shapes that mirror the schemas published by
//     the MASE Direzione Generale Economia Circolare and referenced
//     in art. 8 D.M. 59/2023:
//
//       rentri-formulario-1.0.xsd   (master)
//       rentri-common-1.0.xsd
//       rentri-enum-1.0.xsd
//       rentri-registri-1.0.xsd
//       rentri-movimenti-1.0.xsd
//       xmldsig-core-schema.xsd     (W3C XML Signature)
//
//     The Go structs in this package are the encoding-friendly mirror
//     of those schemas. Server-side validation against the canonical
//     XSD remains authoritative — the structs are convenience, not
//     replacement.
//
// The exact authentication scheme on the production API (mTLS vs
// OAuth2 client_credentials with a JWT signed by the RENTRI-issued
// certificate) is not currently published in machine-readable form;
// the Client interface is therefore deliberately auth-agnostic and
// the concrete adapter is responsible for transport-level credentials.
// Switching auth schemes when the documentation lands is a single-file
// change inside this package.
package rentri
