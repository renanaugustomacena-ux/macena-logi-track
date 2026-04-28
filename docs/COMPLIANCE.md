# LogiTrack — Compliance One-Pager

A single page a prospect's legal counsel can read before signing the
DPA / contratto d'opera. The full regulatory map is in
[`ITALIAN-COMPLIANCE.md`](ITALIAN-COMPLIANCE.md).

LogiTrack is **not a SaaS**. There is no shared multi-tenant cloud.
Each customer receives a dedicated application installed on
infrastructure intestated to the customer (a VPS in the customer's
name, the customer's on-premise infrastructure, or the customer's
private cloud). The professional (Renan Augusto Macena) is engaged
under a **contratto d'opera** ex art. 2222 c.c. — the data controller
remains the customer, full stop. The professional accesses customer
data only as needed for the agreed maintenance scope (monthly
retainer), and is registered as Responsabile del Trattamento ex
GDPR Art. 28 in the customer's data-processor register.

## What LogiTrack helps the customer satisfy

- **GDPR Art. 30 (registro dei trattamenti)** — the kit's `audit_log`
  collection records every state-changing API call (who, what, when,
  source IP, correlation id). Customer can export it to feed their
  registry.
- **GDPR Art. 32 (misure di sicurezza)** — HS256 JWT with alg-pinned
  validation, tenant-scoping at the repository layer, distroless
  container image, secret management via env / orchestrator, HSTS
  with preload, CSP, rate-limit per IP with eviction, redirect-block
  on outbound HTTP. Production refuses to boot with weak/empty
  secrets, unauthenticated MongoDB / Redis URIs, or empty demo identity
  password.
- **D.Lgs. 152/2006 art. 188-bis + 190 + 193 (rifiuti speciali)** —
  the rifiuti module enforces FIR contenuto minimo, registro carico/scarico
  retention 3 anni (D.Lgs. 116/2020 reform), and the FIR state
  machine matches the lifecycle prescribed by D.M. MASE 04/04/2023
  n. 59. Sanctions table in [`MODULE-RIFIUTI.md`](MODULE-RIFIUTI.md) §6.
- **Catena di custodia tamper-evidente** — every custody record
  embeds a SHA-256 of the previous record's canonical form
  (UTC-truncated to ms for BSON round-trip stability). Any later
  modification of an earlier record breaks the chain on
  `services.VerifyChain`.
- **Reg. UE 952/2013 (CDU)** — shipment carries the customs MRN
  envelope; the AIDA adapter is a per-customer build, not a kit
  feature.
- **Accordo ADR 2025 + Reg. UE 1357/2014** — `compliance.go` in the
  rifiuti module enforces ADR class + UN number + HP marker on
  pericolosi FIR.

## What the customer still owns

- Accreditation with AIDA (certificato PKCS#12, SPID/CNS).
- The RENTRI certificato digitale (issued by RENTRI to the
  trasportatore after SPID/CIE/CNS-bound Entratel delegation).
- Conservazione a norma AgID via accreditato of choice
  (https://www.agid.gov.it/it/piattaforme/conservazione).
- Physical chain-of-custody signing devices (sigilli, biometric
  driver ID).
- The contratto law conformance (CMR convention, contratto di
  trasporto art. 1678 c.c., D.Lgs. 286/2005 for autotrasporto, etc.).
- Financial reporting / fatturazione elettronica (the customer keeps
  their existing gestionale fiscale; LogiTrack does not replace it).

## DPA / contratto d'opera clauses

- Standard Contractual Clauses (EU 2021/914) for any sub-processor
  outside EEA — by default the kit has none. Customer hosting choice
  determines this entirely.
- 72 h breach notification per GDPR Art. 33 — the freelancer notifies
  the customer; the customer (titolare del trattamento) is the one
  who notifies the Garante.
- Data return in open formats (JSONL, CSV, GeoJSON) within 30 days of
  contract termination. The professional commits to delivering a
  full database archive and any customer-specific source code at the
  end of the engagement.
- The customer holds an exclusive perpetual license on the personalised
  application; the source code of the customer's deployment is
  deposited on a private Git repository intestated to the customer.
  End of contract = customer keeps the application, the source code,
  the data, the deployment. The professional retains no copy of
  customer data after the agreed transition window (60 days), and
  retains rights only over the generic technology components of the
  shared base platform (security libraries, audit pattern, EER
  validator, etc.).

## Audit surface

- [`SECURITY.md`](SECURITY.md) — threat model + controls + self-assessment.
- [`ITALIAN-COMPLIANCE.md`](ITALIAN-COMPLIANCE.md) — Normattiva +
  EUR-Lex + AgID + AdE references.
- [`MODULE-RIFIUTI.md`](MODULE-RIFIUTI.md) — rifiuti regulatory map +
  sanctions table.
- [`ARCHITECTURE.md`](ARCHITECTURE.md) — data model, sequence
  diagrams, audit trail mechanics.
