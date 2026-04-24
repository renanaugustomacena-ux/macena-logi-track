# LogiTrack — Compliance One-Pager

Purpose: a single page a prospect's legal counsel can read before
signing a DPA. See [`ITALIAN-COMPLIANCE.md`](ITALIAN-COMPLIANCE.md)
for the full regulatory map.

## What LogiTrack helps you satisfy

- **GDPR Art. 30** (registro dei trattamenti): LogiTrack provides a
  pre-populated processor registry template in `docs/DATA-RESIDENCY.md`.
- **GDPR Art. 32** (misure di sicurezza): HS256 JWT, tenant scoping,
  TLS 1.2+ in production, secret management via environment, audit
  log (`audit_log` collection), chain-of-custody hash linkage.
- **Reg. UE 952/2013 CDU** (tracking MRN per spedizione cross-border).
- **Reg. UE 165/2014** (cached driver hours summary for Reg. 561/2006
  compliance checks).
- **D.Lgs. 286/2005 + Albo Autotrasportatori** (flag `alboRegistered`
  on each driver).
- **Accordo ADR 2023** (class marking on shipments + vehicle
  certification flag).
- **Accordo ATP 1970** (class marking on refrigerated loads + vehicle
  certification flag).
- **GDPR Art. 88** (labour): driver data minimisation, optional CF
  ingestion, no permanent storage of raw tachograph DDD.

## What the buyer still owns

- Accreditation with AIDA (certificato PKCS#12, SPID/CNS).
- Physical chain-of-custody signing devices (seals, biometric
  driver ID).
- Biometric or RFID authentication of drivers at handover.
- The underlying contract law conformance (CMR convention and
  successive bilateral agreements).
- Financial reporting / fatturazione elettronica (see FatturaFlow).

## DPA clauses we sign

- Standard Contractual Clauses (EU 2021/914) for any sub-processor
  outside EEA (none by default — all sub-processors are EU-resident).
- 72 h breach notification to Data Controller.
- Right to audit on-premise by the Data Controller with 30 days'
  notice (no more than once per year unless for cause).
- Data return in open formats (JSONL, CSV, GeoJSON) within 30 days
  of contract termination.

## Audit surface

- `docs/SECURITY.md` — threat model + controls.
- `docs/SECURITY-SELF-ASSESSMENT.md` — CIS Controls v8 + AgID Misure
  Minime self-score.
- `docs/DATA-RESIDENCY.md` — processor register + region map.
- `docs/ITALIAN-COMPLIANCE.md` — normattiva references.
