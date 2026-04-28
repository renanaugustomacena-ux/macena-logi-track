# LogiTrack — Module 2: rifiuti speciali

> Vertical: SME waste-transport operators ("trasportatori di
> rifiuti speciali") integrating with the Italian RENTRI national
> digital traceability system. Module status: production-grade
> entities + queued-stub RENTRI client + race-tested. Live RENTRI
> submission deferred until the design-partner trasportatore
> provides the SPID/CIE/CNS-bound Entratel delegation that unlocks
> the API certificate.

## 1. Why this vertical

RENTRI is a hard regulatory deadline that forces every Italian
waste producer, transporter, intermediary and treatment-facility
operator to abandon paper FIR + Carica/Scarico registers and adopt
a server-validated XML workflow. The transition timeline (per
**D.M. MASE 4 aprile 2023, n. 59** + **Legge 199/2025**) is:

| Window | Period | Actors | Threshold |
|---|---|---|---|
| 1 | 15/12/2024 → 13/02/2025 | Trasportatori, intermediari, gestori, produttori &gt; 50 dipendenti | n/a — universal for carriers |
| 2 | 15/06/2025 → 14/08/2025 | Produttori 11–50 dipendenti | 11 ≤ N ≤ 50 |
| 3 | 15/12/2025 → 13/02/2026 | Produttori ≤ 10 dipendenti, soli pericolosi | ≤ 10 |

Hard end-of-paper dates:
- **13/02/2026** — operative deadline for switching to the digital
  FIR.
- **15/09/2026** — paper FIR is fully decommissioned.

Every transporter must therefore have either (a) a vendor
implementation of RENTRI, (b) a custom integration, or (c) a
fallback to a service bureau. The incumbent vendor landscape
(TeamSystem Waste 360, Passepartout, Ambiente.it, Rifiutoo,
Winwaste, Modular SVFOR02, Ecofacile) keeps pricing private
behind sales-led quotes; the only public listing surfaced during
the 2026-04-28 research pass was Modular SVFOR02 at
**€ 1.311 + IVA setup + € 245 / semestre** ongoing.

Public-API integration with TMS/ERP layers is essentially absent
across the incumbent set, mobile UX for autisti is generally an
afterthought, and AgID-conforming conservation is rarely included
in the base canone. These are the gaps LogiTrack fills.

## 2. First design-partner target

**FRO S.r.l.** — transport arm of Effevi Rottami group.

| Field | Value |
|---|---|
| Address | Via Quartieri snc, 37060 Mozzecane (VR) |
| Distance from Mozzecane town centre | ~ 0 km |
| Telephone | 045 6340188 |
| Website | https://www.frotrasporti.com (sister: https://www.effevirottami.com) |
| P.IVA | 03728630231 |
| Family-run since | 1954 |
| Quality cert | ISO 14001:2015 (sister) |
| Activities (publicly verified, Albo iscrizione UNVERIFIED) | ADR pericolosi + non-pericolosi transport, intermediation, container service, EoW ferrosi/inox/ghisa/alluminio, RAEE, batterie |
| Industries served | Siderurgico, industriale, urbano (recupero) |
| Strategic detail | 2 km internal rail spur; cranes 45 t; MEWPs 22-24 m |

First-call goals: validate Albo categoria (likely 4 + 5 + 8 +
4-bis), parco mezzi, dipendenti, current FIR workflow, RENTRI
readiness, willingness to participate as design partner.

Backup candidates within 25 km (full agent report archived in the
working notes for the 2026-04-28 pivot session): Ferramenta
Villafranca Rottami (Mozzecane), Boccagni S.n.c. (Villafranca,
recupero), Eco Gest (Valeggio sul Mincio, intermediario
transfrontaliero, Cat 8), Ecoservizi S.n.c. (Castelnuovo del
Garda, RAEE+pericolosi+toner), Veneta Recuperi (Sona), Sun Oil
Italiana (Sona, oli esausti), Ecodent (Villafranca, sanitario
dentale), Cimaf Goito (MN), Agrofert (Isola della Scala,
agro/ortofrutta).

## 3. Italian regulatory anchors

Documented in detail in `docs/ITALIAN-COMPLIANCE.md` § "Rifiuti
speciali". The summary anchors are:

| Surface | Source | Implementation |
|---|---|---|
| Tracciabilità rifiuti | D.Lgs. 152/2006 art. 188-bis | `rentri.Client` interface, `QueuedStub` default |
| RENTRI regolamento | D.M. MASE 04/04/2023 n. 59 (GU 126/2023) | xFIR XML payload + lifecycle state machine |
| Specifiche tecniche xFIR | Decreto direttoriale MASE n. 251 del 19/12/2023 + spec v1.0 del 10/02/2025 | `rentri/client.go` request/response types |
| FIR contenuti minimi | D.Lgs. 152/2006 art. 193 | `FIR` struct + `Validate` |
| Registro carico/scarico | D.Lgs. 152/2006 art. 190 (3 anni post D.Lgs. 116/2020) | `RegistroEntry` + `RetentionYears = 3` |
| Catalogo EER | Decisione 2014/955/UE | `CERCode` validator + chapter helper |
| Caratteristiche di pericolo | Reg. UE 1357/2014 + 2017/997 | `HPClass` enum |
| Albo Gestori Ambientali | D.M. 120/2014 + D.Lgs. 152/2006 art. 212 | `AlboCategoria`, `AlboClasse`, `Trasportatore.CanCarry` |
| Operazioni R/D | D.Lgs. 152/2006 Allegati B + C (recepimento Dir. 2008/98/CE) | `ImpiantoOperazione` enum + `Destinatario.CanReceive` |
| ADR rifiuti pericolosi | Accordo ADR 2025 + D.Lgs. 35/2010 | reuse of `logistics.ADRClass` + UN number + HP marker on FIR |
| Conservazione documentale | CAD D.Lgs. 82/2005 + Linee Guida AgID | external conservatore accreditato (deferred to deployment) |
| Sanzioni | D.Lgs. 152/2006 art. 256 + 258 | encoded in service-layer policy (future commit) |

## 4. xFIR lifecycle (state machine in `rifiuti/fir.go`)

```text
draft ──► vidimato ──► consegnato_trasportatore ──► in_transito ──┬──► consegnato_destinatario ──► chiuso
  │           │                                                    │
  │           └──► annullato (terminal)                             └──► respinto ──► chiuso
  └──► annullato (terminal)
```

The `firTransitions` map encodes every legal edge; `AdvanceState`
rejects everything else. The 90-day "copia produttore" return
window from D.Lgs. 152/2006 art. 188-bis comma 4 is checked by
`FIR.CopiaProduttoreOverdue` so the operator dashboard can warn
the producer before they need to denounce a missing copy to the
provincia.

## 5. Auth + sandbox

| Environment | Base URL | Notes |
|---|---|---|
| Sandbox API | `https://demoapi.rentri.gov.it` | OpenAPI v1.0 at `/docs/dati-registri/v1.0` |
| Sandbox portal | `https://demobackoffice.rentri.gov.it` | SPID/CIE/CNS login |
| Production API | `https://api.rentri.gov.it` | Requires RENTRI-issued certificato digitale |

The exact production auth scheme (mTLS vs OAuth2
client-credentials with a JWT signed by the RENTRI certificate)
is not currently published in machine-readable form. The
`rentri.Client` interface in `internal/modules/rifiuti/rentri/client.go`
is auth-agnostic; switching schemes is a one-file change.

## 6. Sanctions (D.Lgs. 152/2006 art. 256 + 258)

| Violation | Non-pericolosi | Pericolosi |
|---|---|---|
| Omessa/irregolare iscrizione RENTRI | € 500 – 2.000 | € 1.000 – 3.000 |
| FIR mancante / errato / non trasmesso | € 1.600 – 10.000 | € 1.600 – 10.000 + reclusione (art. 483 c.p.) |
| Omessa tenuta registro C/S | € 4.000 – 20.000 | € 10.000 – 30.000 |
| Trasporto senza Albo (art. 256 c.1) | arresto 3-12 mesi o ammenda € 2.600 – 26.000 | arresto 6 mesi - 2 anni e ammenda € 2.600 – 26.000 |

Riduzione 1/3 ex art. 258 c. 10 if regularised entro 60 giorni
dalla scadenza. **DL 116/2025** added optional sospensione
patente + sospensione iscrizione Albo on recidive.

## 7. What the module ships today (commit-level inventory)

- `doc.go` — module contract, regulatory anchors.
- `compliance.go` — `CERCode` + validators, `IsCERPericoloso`,
  `CERChapter`, `AlboCategoria` (1, 2-bis, 4, 5, 6, 8, 9, 10),
  `AlboClasse` A–F with documented tonnage bands,
  `ImpiantoOperazione` R + D codes from Allegati B/C TUA.
- `party.go` — `Produttore`, `Trasportatore`, `Destinatario`,
  `Trasportatore.CanCarry` (cer, time) → enforces categoria
  scope + Albo expiry, `Destinatario.CanReceive` (cer,
  operazione, time) → enforces autorizzazione scope + expiry.
- `fir.go` — `FIR` aggregate composing `logistics.Shipment`,
  `FIRState` enum, `firTransitions` legal-edge map,
  `AdvanceState`, `Validate` with pericoloso-specific ADR + HP
  guards, 90-day `CopiaProduttoreOverdue` check.
- `registro.go` — `RegistroEntry` append-only line, `Validate`,
  `RetentionYears = 3` (D.Lgs. 116/2020 reform).
- `rentri/` sub-package — `Client` interface + sandbox endpoints
  + `QueuedStub` adapter (idempotency-keyed, deterministic numero
  generation, FIFO drain pattern, race-tested under 16-worker
  contention).

Tests: 5 test files, every assertion table-driven, full state
machine covered including terminal-state rejections,
`-race -count=1` clean inside the standard build container.

## 8. What the module deliberately does NOT ship yet

- **Live RENTRI HTTP adapter**. The `Client` interface +
  `QueuedStub` are wired and exercised under race; the live HTTP
  adapter materialises the day a design-partner customer provides
  the SPID/CIE/CNS-bound Entratel delegation that unlocks the API
  certificate. Switching to live is a one-file change in
  `cmd/server/main.go`.
- **Conservazione AgID**. The list of accreditati is at
  https://www.agid.gov.it/it/piattaforme/conservazione; selecting
  the conservatore is a per-fork deployment decision, not a kit
  feature.
- **Sanction-policy advisor**. A reasonable v2 follow-on for the
  operator dashboard; not part of the kit core because the
  responsibility legally sits on the impresa, not on the software.
- **xFIR XSD encoder**. The current `VidimaFIR` handler emits a
  placeholder `<formulario><cer>…</cer></formulario>` payload built
  via `encoding/xml` (so user-controlled fields are escaped). The
  full RENTRI v1.0 XSD-driven encoder lands together with the live
  HTTP adapter.
- **MUD annual export**. Will be added as a per-customer feature
  during the engagement: the MUD form is stable but the customer's
  CER/operazione mapping is not.

These are deliberately deferred: shipping them ahead of customer
feedback would re-introduce the "build something nobody asked for"
failure pattern the kit doctrine was designed to break.

## 9. What the module DOES ship today (HTTP + frontend)

- REST endpoints under `/api/v1/rifiuti/*`:
  - Anagrafiche: `POST/GET produttori|trasportatori|destinatari`.
  - FIR: `POST /fir`, `GET /fir`, `GET /fir/:id`,
    `POST /fir/:id/transition`, `POST /fir/:id/vidima`.
  - Inline validator: `GET /cer/:code`.
- All endpoints are JWT-authenticated and tenant-scoped.
- Vue 3 view at `/rifiuti` with party drop-downs, inline CER
  validator, ADR-conditional fields, FIR list with state badges.
- Repository accessors mirror the logistics pattern (Insert/Get/List/Update),
  unique index on `(tenant_id, codice_fiscale)` for anagrafiche and
  on `(tenant_id, numero_rentri)` sparse for FIR.
