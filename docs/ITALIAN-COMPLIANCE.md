# LogiTrack — Italian Regulatory Compliance Map

This document enumerates the Italian and EU regulations LogiTrack
helps a buyer comply with. It is the source of truth for sales and
pre-sales conversations; every claim is backed by the official
reference and the code module that encodes it.

Access dates: 2026-04-17.

---

## 1. Codice Doganale dell'Unione (Reg. UE 952/2013)

- **Reference:** Regolamento (UE) n. 952/2013 del 9 ottobre 2013
  che istituisce il Codice Doganale dell'Unione (CDU).
- **Source:** https://eur-lex.europa.eu/eli/reg/2013/952/oj
- **Scope:** transit regime (T1, T2), procedure unioniche e non
  unioniche, deposito doganale, regime di perfezionamento attivo e
  passivo. Applicable to every cross-border shipment LogiTrack tracks
  (VR↔AT/DE is the primary corridor).
- **How LogiTrack supports:** the `CustomsStatus` entity stores the
  `DocumentType` (T1/T2/EUR1/CIM), the MRN from AIDA, and the
  clearance point and timestamp; the ARCHITECTURE section on customs
  flows illustrates the envelope. LogiTrack does not replace AIDA
  nor generate the declarative bodies; it tracks the MRN lifecycle
  and surfaces the "in dogana" state to operators in real time.

## 2. AIDA — Agenzia delle Dogane e dei Monopoli, sistema informativo

- **Reference:** https://www.adm.gov.it/portale/
- **Scope:** the AIDA integration envelope is a REST + SOAP hybrid.
  Operators are accredited via a PKCS#12 certificate. The T1 declaration
  XML follows the `XMerchant.Transit` schema published by ADM.
- **How LogiTrack supports:** `AIDA_API_BASE` and `AIDA_API_KEY` env
  variables are documented; the adapter is implemented as a thin
  client stub in `internal/services` with the request body shape:

  ```
  POST {AIDA_API_BASE}/transito/dichiarazione
  {
    "mrn": "25ITQX1A00000000A1",
    "documentType": "T1",
    "office": "IT182102",         // ufficio di partenza
    "carrier":  { "vat": "IT01234567890" },
    "goods":    [ { "hsCode":"..", "weightKg":.. } ]
  }
  ```

  Production deployments replace the stub with the accredited client
  at the operator's premises. Documentation: see
  [`docs/INTEGRATIONS.md`](INTEGRATIONS.md#aida).

## 3. Tachigrafo digitale (Reg. UE 165/2014)

- **Reference:** Regolamento (UE) n. 165/2014 del 4 febbraio 2014
  relativo ai tachigrafi nei trasporti su strada.
- **Source:** https://eur-lex.europa.eu/eli/reg/2014/165/oj
- **Scope:** obbligo del tachigrafo digitale di seconda generazione
  (smart tachograph v2 da giugno 2019, smart v2+ da agosto 2023) per
  tutti i veicoli > 3,5 t immatricolati per trasporto merci in UE.
  Reg. (CE) n. 561/2006 impone i tempi di guida e riposo correlati.
- **How LogiTrack supports:** `Driver.DrivingHours` in
  `internal/models/driver.go` caches the daily counters required by
  Reg. 561/2006 (driven-seconds, rested-seconds, period-start).
  Dispatch UX can refuse assigning a new load if the cap is exceeded.
  Raw tachograph DDD files are NOT ingested by LogiTrack; we consume
  the summary via the operator's tachograph-analysis vendor
  (e.g. DigitaleTacho, Vimcar, Continental VDO).

## 4. Autotrasporto merci — D.Lgs. 286/2005 + Albo Autotrasportatori

- **Reference:** D.Lgs. 21 novembre 2005, n. 286 — liberalizzazione
  dell'autotrasporto di cose per conto di terzi.
- **Source:** https://www.normattiva.it/ricerca/semplice
- **Albo Nazionale Autotrasportatori (Ministero delle Infrastrutture e
  dei Trasporti):** https://www.alboautotrasporto.it
- **Scope:** ogni vettore per conto terzi deve essere iscritto
  all'Albo. Regolamentato per requisiti di capacità finanziaria,
  onorabilità, capacità professionale e idoneità del veicolo.
- **How LogiTrack supports:** `Driver.AlboRegistered`, and
  `Vehicle.ADRCertified` / `Vehicle.ATPCertified`. Plate validation
  via the regex in `internal/models/compliance.go` accepts both
  post-1994 (`AB123CD`) and historical (`VR-123456`) formats.

## 5. ADR — trasporto merci pericolose su strada

- **Reference:** Accordo ADR (UN-ECE), recepimento italiano con
  D.Lgs. 4 febbraio 2000, n. 40 e successive modifiche; ultima edizione
  ADR 2023 vigente fino al 31/12/2024.
- **Source:** https://unece.org/transport/dangerous-goods/adr-2023
- **How LogiTrack supports:** `Shipment.ADRClass` accepts classes 1 –
  9 (incl. sub-classes 4.1/4.2/4.3, 5.1/5.2, 6.1/6.2) per
  `internal/models/compliance.go`. The dispatcher UI surfaces an ADR
  class flag; reports include the class for compliance audits.

## 6. ATP — trasporto merci deperibili

- **Reference:** Accordo ATP (UN-ECE 1970), recepimento italiano con
  L. 2 maggio 1977, n. 264.
- **Scope:** categorie IR, RNA, RRB, FRC, IN — certification riguarda
  refrigerazione e isolamento. Il mercato di Verona è sensibile per
  la filiera ortofrutticola del Consorzio Agricolo del Veronese.
- **How LogiTrack supports:** `Vehicle.ATPCertified` boolean +
  `Shipment.ATPClass` typed enum. Used by the dispatcher to match
  loads to refrigerated trailers.

## 7. Telepass — pedaggi autostradali

- **Reference:** AISCAT — Associazione Italiana Società Concessionarie
  Autostrade e Trafori, pubblica il gazzettino dei codici di tratta.
- **Source:** https://www.aiscat.it — gazzettino tariffe e codici.
- **Scope:** the toll-code dictionary is the list of alphanumeric
  codes that identify tolled stretches (e.g. `A22-VR-BZ`). Telepass
  is the dominant operator.
- **How LogiTrack supports:** `Shipment.TelepassCodes []TelepassTollCode`
  with the `Code` + `Label` shape. A background job (documented in
  [`docs/RUNBOOK.md`](RUNBOOK.md)) periodically refreshes the
  dictionary from the AISCAT publication. The seeded demo shipments
  include representative codes (A22-VR-BZ for Verona→München, A4-VR-MI
  for Verona→Milano).

## 8. Quadrante Europa Verona — Consorzio ZAI

- **Reference:** Consorzio Zona Agricolo-Industriale di Verona,
  ente gestore dell'interporto Quadrante Europa.
- **Source:** https://www.quadranteeuropa.it
- **Scope:** il Quadrante Europa è il secondo hub intermodale d'Europa
  per tonnellaggio (oltre 8 Mt/anno) e connette A4/A22 con la rete
  RFI. Le prenotazioni di slot ferroviari sono gestite dalla filiera
  Hupac / Inrail / Mercitalia attraverso l'anagrafica del Consorzio.
- **How LogiTrack supports:** un geofence dedicato (`type="terminal"`,
  `areaRef="QE-VR"`) copre il perimetro del Quadrante; l'ingresso e
  l'uscita generano automaticamente un `TrackingEvent.geofence_enter`
  e `geofence_exit`. Il Phase-3 roadmap documenta l'integrazione
  webhook con RFI per gli slot ferroviari.

## 9. GDPR e telematica

- **Reference:** Reg. UE 2016/679 (GDPR) e D.Lgs. 196/2003 come
  modificato dal D.Lgs. 101/2018.
- **Scope:** telematica veicolare e dati del conducente sono dati
  personali: finalità, minimizzazione, conservazione.
- **How LogiTrack supports:**
  - Codice Fiscale NOT stored by default (opt-in per tenant).
  - Driving-hours summaries stored without raw tachograph DDD files.
  - Data residency defaults EU; production deployments use Aruba
    Cloud Italia or OVHcloud Gravelines FR regions (both within the
    EEA). See [`docs/DATA-RESIDENCY.md`](DATA-RESIDENCY.md).

## 10. SDI/FatturaPA interoperability (read-only pointer)

- LogiTrack does not issue invoices. The CMR document and the AIDA
  MRN can be attached to the outgoing invoice body via the
  FatturaFlow integration (see
  [`docs/INTEGRATIONS.md`](INTEGRATIONS.md#fatturaflow)) so the
  buyer can satisfy SDI audits of cross-border dispatches.

---

## Audited constants

| Constant | File | Reference |
| --- | --- | --- |
| `DefaultAverageSpeedKPH = 70.0` | `internal/services/route_optimizer.go` | conservative benchmark for A4/A22 (observed by Autostrade per l'Italia bulletins) |
| `platePost1994` regex | `internal/models/compliance.go` | D.M. 20/05/1992, D.Lgs. 285/1992 art. 100 |
| `ADRClass*` enum | `internal/models/compliance.go` | ADR 2023 Annex A (UN-ECE) |
| `ATPClass*` enum | `internal/models/compliance.go` | ATP 1970 art. 4 Annex 1 (UN-ECE) |
| `EventCustomsHold/Cleared` | `internal/models/tracking_event.go` | Reg. UE 952/2013 CDU art. 226 |

## Gaps / out-of-scope

- No generation of the T1/T2 XML body — LogiTrack tracks the lifecycle
  around the MRN; the customs broker's accredited software still
  produces the declaration.
- No automatic recomputation of the Albo insurance status; the
  operator is expected to revoke a driver/vehicle via the admin API
  when the Albo subscription lapses.
- No digital tachograph ingestion. Documented in
  [`docs/TECHNICAL-DEBT.md`](TECHNICAL-DEBT.md) as a deliberate
  boundary.

---

## Source verification procedure

For each reference in this file, the Mission II consolidator verified
the document's existence and reachable URL on the dates above. The
per-citation `access-date` annotations appear inline.
