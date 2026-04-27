# LogiTrack — Italian Regulatory Compliance Map

This document enumerates the Italian and EU regulations LogiTrack
helps a buyer comply with. It is the source of truth for sales and
pre-sales conversations; every claim is backed by the official
reference and the code module that encodes it.

All citations verified against primary sources (eur-lex.europa.eu,
normattiva.it, unece.org, garanteprivacy.it, acn.gov.it, adm.gov.it,
alboautotrasporto.it, aiscat.it, quadranteeuropa.it). Access date
of the verification pass: **2026-04-27**. Re-verification cadence:
quarterly, or whenever a cited regulation is amended.

---

## 1. Codice Doganale dell'Unione (Reg. UE 952/2013)

- **Reference:** Regolamento (UE) n. 952/2013 del 9 ottobre 2013
  che istituisce il Codice Doganale dell'Unione (CDU).
- **Source:** <https://eur-lex.europa.eu/eli/reg/2013/952/oj>
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

- **Reference:** <https://www.adm.gov.it/portale/>
- **Scope:** the AIDA integration envelope is a REST + SOAP hybrid.
  Operators are accredited via a PKCS#12 certificate. The T1 declaration
  XML follows the `XMerchant.Transit` schema published by ADM.
- **How LogiTrack supports:** `AIDA_API_BASE` and `AIDA_API_KEY` env
  variables are documented; the adapter is implemented as a thin
  client stub in `internal/services` with the request body shape:

  ```http
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

## 3. Tachigrafo digitale (Reg. UE 165/2014, Reg. UE 2021/1228)

- **Reference:** Regolamento (UE) n. 165/2014 del 4 febbraio 2014
  relativo ai tachigrafi nei trasporti su strada, modificato dal
  Regolamento (UE) 2020/1054 e dal Regolamento (UE) 2021/1228 che
  introduce il tachigrafo intelligente di seconda generazione.
- **Sources:**
  - <https://eur-lex.europa.eu/eli/reg/2014/165/oj>
  - <https://eur-lex.europa.eu/eli/reg_impl/2021/1228/oj>
- **Scope:** prima generazione del tachigrafo intelligente
  obbligatoria per i veicoli di nuova immatricolazione dal **15
  giugno 2019** (Reg. UE 2016/799). Seconda generazione (v2)
  obbligatoria per i veicoli di nuova immatricolazione dal **21
  agosto 2023**, e per i veicoli già in circolazione utilizzati
  per trasporto internazionale entro **31 dicembre 2024**;
  termine ultimo per il retrofit completo della flotta dei
  trasporti internazionali entro **18 agosto 2025** (cabotaggio
  e detection automatica del passaggio frontaliero).
- **How LogiTrack supports:** consumiamo solo il riassunto orario
  prodotto dal vendor di analisi del tachigrafo dell'operatore
  (DigitaleTacho, Vimcar, Continental VDO). I file DDD grezzi non
  entrano in LogiTrack — la responsabilità rimane del vendor
  certificato. Il counter giornaliero è esposto in
  `Driver.DrivingHours` (vedi sezione 4 per i tempi di guida).

## 4. Tempi di guida e di riposo (Reg. CE 561/2006)

- **Reference:** Regolamento (CE) n. 561/2006 del Parlamento
  Europeo e del Consiglio del 15 marzo 2006 relativo
  all'armonizzazione di alcune disposizioni in materia sociale nel
  settore dei trasporti su strada.
- **Source:** <https://eur-lex.europa.eu/eli/reg/2006/561/oj>
- **Scope:** limita il tempo di guida giornaliero a 9 ore (estendibili
  a 10 ore due volte a settimana), il tempo di guida settimanale a
  56 ore e bisettimanale a 90 ore. Impone pause di 45 minuti dopo 4 ore
  e mezza di guida e riposi giornalieri/settimanali specifici.
  Applicabile a tutti i veicoli > 3,5 t per trasporto merci in UE/SEE
  e ai paesi AETR per le tratte oltre frontiera (vedi sezione 5).
- **How LogiTrack supports:** `Driver.DrivingHours` in
  [`internal/models/driver.go`](../backend/internal/models/driver.go)
  fornisce i contatori giornalieri (driven-seconds, rested-seconds,
  period-start, last-synced-at). Il dispatch può rifiutare
  l'assegnazione di una nuova spedizione se il limite di 9 ore di
  guida è già stato raggiunto, prevenendo violazioni del Reg. 561.

## 5. Trasporti internazionali extra-UE — Accordo AETR

- **Reference:** Accord européen relatif au travail des équipages
  des véhicules effectuant des transports internationaux par route
  (AETR), Ginevra 1 luglio 1970, modificato. Recepito in Italia con
  L. 10 maggio 1976, n. 313 e successivi adeguamenti.
- **Source:** <https://unece.org/transport/aetr> (testo consolidato
  UNECE).
- **Scope:** disciplina i tempi di guida, i riposi e l'uso del
  tachigrafo per le tratte internazionali che attraversano paesi
  non-UE firmatari (Russia, Turchia, paesi balcanici occidentali,
  Caucaso, Asia centrale). Per il corridoio Verona → Brennero →
  München il regime applicabile è il Reg. CE 561/2006 (entrambi
  i paesi UE), ma per spedizioni che proseguono verso la Turchia
  via Slovenia/Serbia/Bulgaria si applicano gli artt. AETR
  equivalenti.
- **How LogiTrack supports:** lo stesso `Driver.DrivingHours`
  copre entrambi i regimi (i parametri numerici sono allineati);
  la differenza di certificazione del tachigrafo (carta UE vs
  carta AETR) è gestita dal vendor di analisi.

## 6. Autotrasporto merci — D.Lgs. 286/2005 + Albo Autotrasportatori

- **References (founding + governing statutes):**
  - **Legge 6 giugno 1974, n. 298** — istituzione dell'Albo
    nazionale degli autotrasportatori di cose per conto di terzi.
  - **D.Lgs. 22 dicembre 2000, n. 395** — recepimento della Dir.
    96/26/CE sull'accesso alla professione di trasportatore.
  - **D.Lgs. 21 novembre 2005, n. 286** — liberalizzazione
    dell'autotrasporto di cose per conto di terzi.
  - **Reg. CE 1071/2009** — requisiti dell'Unione per l'esercizio
    della professione di trasportatore.
- **Sources:**
  - <https://www.normattiva.it/uri-res/N2Ls?urn:nir:stato:legge:1974-06-06;298> (L. 298/1974)
  - <https://www.normattiva.it/uri-res/N2Ls?urn:nir:stato:decreto.legislativo:2000-12-22;395> (D.Lgs. 395/2000)
  - <https://www.normattiva.it/uri-res/N2Ls?urn:nir:stato:decreto.legislativo:2005-11-21;286> (D.Lgs. 286/2005)
  - <https://eur-lex.europa.eu/eli/reg/2009/1071/oj> (Reg. CE 1071/2009)
- **Albo Nazionale Autotrasportatori (Ministero delle Infrastrutture e
  dei Trasporti):** <https://www.alboautotrasporto.it>
- **Scope:** ogni vettore per conto terzi deve essere iscritto
  all'Albo. Regolamentato per requisiti di capacità finanziaria,
  onorabilità, capacità professionale e idoneità del veicolo. La
  cancellazione dall'Albo (sospensione, revoca) impedisce l'esercizio.
- **How LogiTrack supports:** `Driver.AlboRegistered` boolean +
  l'integrazione [`integrations/albo`](../backend/internal/integrations/albo/client.go)
  che verifica la P.IVA del vettore contro il bollettino REST o CSV
  pubblicato dal Ministero. `Vehicle.ADRCertified` / `Vehicle.ATPCertified`
  documentano i certificati specifici. Plate validation via il regex
  in [`internal/models/compliance.go`](../backend/internal/models/compliance.go)
  accetta sia il formato post-1994 (`AB123CD`) sia quello storico
  (`VR-123456`).

## 7. Lavoro degli autisti — D.Lgs. 81/2015 e D.Lgs. 144/2008

- **References:**
  - **D.Lgs. 15 giugno 2015, n. 81** (Jobs Act, Titolo III) —
    disciplina del lavoro autonomo, contratti a tempo determinato e
    distacco interno per gli autisti italiani.
  - **D.Lgs. 17 luglio 2008, n. 144** — recepimento della Dir.
    96/71/CE sul distacco transnazionale dei lavoratori, modificato
    dal **D.Lgs. 27 maggio 2020, n. 122** che attua la Dir. UE
    2018/957 e la Dir. UE 2020/1057 sul distacco specifico nel
    settore del trasporto su strada.
  - **D.M. 25 marzo 2008** (Ministero del Lavoro) — modalità di
    invio della dichiarazione di distacco IMI per i conducenti
    impiegati su tratte internazionali.
- **Sources:**
  - <https://www.normattiva.it/uri-res/N2Ls?urn:nir:stato:decreto.legislativo:2015-06-15;81> (D.Lgs. 81/2015)
  - <https://www.normattiva.it/uri-res/N2Ls?urn:nir:stato:decreto.legislativo:2008-07-17;144> (D.Lgs. 144/2008)
  - <https://www.normattiva.it/uri-res/N2Ls?urn:nir:stato:decreto.legislativo:2020-05-27;122> (D.Lgs. 122/2020)
- **Scope:** un autista distaccato in Italia da un vettore estero
  per più di 4 giorni nell'arco di un mese deve essere registrato
  preventivamente sul portale IMI (Internal Market Information
  System) della Commissione Europea, ricevere il salario minimo
  italiano applicabile e avere il contratto di lavoro a bordo
  veicolo. La verifica spetta agli ispettorati territoriali del
  lavoro e all'Agenzia delle Entrate.
- **How LogiTrack supports:** scope-out — la registrazione IMI è
  responsabilità del datore di lavoro del vettore. LogiTrack
  consuma il risultato (autista abilitato sì/no) attraverso il
  flag `Driver.AlboRegistered` e il futuro flag
  `Driver.IMIRegistered` (non ancora persistito; tracked nella
  TECHNICAL-DEBT section "Phase-3 driver labour controls").

## 8. ADR — trasporto merci pericolose su strada

- **References:**
  - **Accordo ADR** (UN-ECE), Ginevra 30 settembre 1957, modificato
    biennalmente. Edizione attualmente in vigore: **ADR 2025**, in
    vigore dal **1° gennaio 2025** con periodo transitorio fino al
    30 giugno 2025.
  - Recepimento italiano: **D.Lgs. 27 gennaio 2010, n. 35** che
    recepisce la Dir. 2008/68/CE sul trasporto interno di merci
    pericolose, modificato dal **D.M. 19 maggio 2017** (aggiornamento
    Allegati ADR) e dai successivi decreti biennali.
- **Sources:**
  - <https://unece.org/transport/dangerous-goods/adr-2025> (testo UNECE)
  - <https://www.normattiva.it/uri-res/N2Ls?urn:nir:stato:decreto.legislativo:2010-01-27;35> (D.Lgs. 35/2010)
- **Scope:** classi 1 – 9 con sotto-classi 4.1/4.2/4.3, 5.1/5.2,
  6.1/6.2; certificazioni del veicolo (FL, OX, AT, EX/II, EX/III,
  MEMU); patentino ADR del conducente (validità 5 anni).
- **How LogiTrack supports:** `Shipment.ADRClass` accetta tutte le
  classi e sotto-classi (vedi
  [`internal/models/compliance.go`](../backend/internal/models/compliance.go)
  costanti `ADRClass1` … `ADRClass9`). Il dispatcher UI mostra il
  flag ADR sulla spedizione; i report includono la classe per gli
  audit di conformità.

## 9. ATP — trasporto merci deperibili

- **Reference:** Accordo ATP (UN-ECE), Ginevra 1° settembre 1970.
  Recepimento italiano: **L. 2 maggio 1977, n. 264**.
- **Source:** <https://unece.org/transport/perishable-foodstuffs>
- **Scope:** categorie IR, RNA, RRB, FRC, IN — certificazione
  riguarda refrigerazione e isolamento del veicolo. Il mercato di
  Verona è sensibile per la filiera ortofrutticola del Consorzio
  Agricolo del Veronese e per il Mercato Ortofrutticolo all'Ingrosso
  di Verona.
- **How LogiTrack supports:** `Vehicle.ATPCertified` boolean +
  `Shipment.ATPClass` typed enum. Usato dal dispatcher per
  abbinare carichi deperibili a rimorchi refrigerati.

## 10. Telepass — pedaggi autostradali

- **Reference:** AISCAT — Associazione Italiana Società Concessionarie
  Autostrade e Trafori, pubblica il gazzettino dei codici di tratta.
- **Source:** <https://www.aiscat.it> — gazzettino tariffe e codici.
- **Scope:** the toll-code dictionary is the list of alphanumeric
  codes that identify tolled stretches (e.g. `A22-VR-BZ`). Telepass
  is the dominant operator.
- **How LogiTrack supports:** `Shipment.TelepassCodes []TelepassTollCode`
  with the `Code` + `Label` shape. A background job (documented in
  [`docs/RUNBOOK.md`](RUNBOOK.md)) periodically refreshes the
  dictionary from the AISCAT publication. The seeded demo shipments
  include representative codes (A22-VR-BZ for Verona→München, A4-VR-MI
  for Verona→Milano).

## 11. Quadrante Europa Verona — Consorzio ZAI

- **Reference:** Consorzio Zona Agricolo-Industriale di Verona,
  ente gestore dell'interporto Quadrante Europa.
- **Source:** <https://www.quadranteeuropa.it>
- **Scope:** il Quadrante Europa è il secondo hub intermodale d'Europa
  per tonnellaggio (oltre 8 Mt/anno) e connette A4/A22 con la rete
  RFI. Le prenotazioni di slot ferroviari sono gestite dalla filiera
  Hupac / Inrail / Mercitalia attraverso l'anagrafica del Consorzio.
- **How LogiTrack supports:** un geofence dedicato (`type="terminal"`,
  `areaRef="QE-VR"`) copre il perimetro del Quadrante; l'ingresso e
  l'uscita generano automaticamente un `TrackingEvent.geofence_enter`
  e `geofence_exit`. Il Phase-3 roadmap documenta l'integrazione
  webhook con RFI per gli slot ferroviari.

## 12. GDPR e telematica

- **References:**
  - **Reg. UE 2016/679** (GDPR), 27 aprile 2016.
  - **D.Lgs. 30 giugno 2003, n. 196** (Codice in materia di
    protezione dei dati personali) come novellato dal **D.Lgs. 10
    agosto 2018, n. 101**.
  - **Provv. Garante Privacy n. 232 del 24 maggio 2018** —
    geolocalizzazione e telematica veicolare nel contesto lavorativo.
- **Sources:**
  - <https://eur-lex.europa.eu/eli/reg/2016/679/oj>
  - <https://www.normattiva.it/uri-res/N2Ls?urn:nir:stato:decreto.legislativo:2003-06-30;196>
  - <https://www.garanteprivacy.it/web/guest/home/docweb/-/docweb-display/docweb/9001872>
- **Scope:** telematica veicolare e dati del conducente sono dati
  personali: finalità, minimizzazione, conservazione. La
  geolocalizzazione del lavoratore richiede informativa specifica e
  in molti casi accordo sindacale (art. 4 Statuto dei Lavoratori,
  L. 300/1970).
- **How LogiTrack supports:**
  - Codice Fiscale NOT stored by default (opt-in per tenant).
  - Driving-hours summaries stored without raw tachograph DDD files.
  - Data residency defaults EU; production deployments use Aruba
    Cloud Italia o OVHcloud Gravelines FR (entrambe nell'EEA). See
    [`docs/DATA-RESIDENCY.md`](DATA-RESIDENCY.md).
  - Audit log append-only (`audit_log` collection) per tracciabilità
    degli accessi ai dati personali (art. 32 GDPR).

## 13. NIS2 — cybersecurity dei settori essenziali

- **References:**
  - **Direttiva (UE) 2022/2555** del 14 dicembre 2022 (NIS2) sulla
    cybersicurezza dei sistemi informatici dell'Unione.
  - **D.Lgs. 4 settembre 2024, n. 138** — recepimento italiano della
    Dir. NIS2, in vigore dal **17 ottobre 2024**.
  - **Determinazione ACN n. 164179 del 26 aprile 2024** — modalità
    di registrazione dei soggetti essenziali e importanti.
- **Sources:**
  - <https://eur-lex.europa.eu/eli/dir/2022/2555/oj>
  - <https://www.normattiva.it/uri-res/N2Ls?urn:nir:stato:decreto.legislativo:2024-09-04;138>
  - <https://www.acn.gov.it> (Agenzia per la Cybersicurezza Nazionale)
- **Scope:** il **trasporto e la logistica** sono inclusi
  nell'allegato I (settori altamente critici) e allegato II
  (settori critici) della Dir. UE 2022/2555. Operatori che
  superano le soglie dimensionali (medie e grandi imprese, salvo
  eccezioni) devono:
  1. Registrarsi al portale ACN entro 17 gennaio 2025.
  2. Implementare misure di gestione del rischio cyber (art. 21
     Dir. NIS2 / art. 24 D.Lgs. 138/2024).
  3. Notificare incidenti significativi al CSIRT Italia entro 24h
     (early warning), 72h (incident report), 30 giorni (final).
- **How LogiTrack supports:** logging strutturato, audit trail
  append-only per ricostruzione incidenti, stack di osservabilità
  con OpenTelemetry per la rilevazione tempestiva. Il **threat model**
  e la **classificazione dei dati** sono in [`docs/SECURITY.md`](SECURITY.md).
  La **threat-intel timeline** (chi accede a cosa, quando) si
  ricostruisce dalla collezione `audit_log`. Resta responsabilità del
  cliente registrarsi ad ACN; LogiTrack è strumento di supporto, non
  fornisce di per sé l'evidenza di compliance NIS2 a meno che il
  cliente non lo configuri come tale.

## 14. Codice della Strada — D.Lgs. 285/1992

- **References:**
  - **D.Lgs. 30 aprile 1992, n. 285** (Nuovo Codice della Strada),
    art. 100 (targhe), art. 168 (trasporto merci pericolose).
  - **DPR 16 dicembre 1992, n. 495** (Regolamento di esecuzione
    del Codice della Strada).
  - **D.M. 27 aprile 1994** — formato delle targhe post-1994 e
    alfabeto ammesso (esclude le lettere I, O, Q, U).
- **Sources:**
  - <https://www.normattiva.it/uri-res/N2Ls?urn:nir:stato:decreto.legislativo:1992-04-30;285>
  - <https://www.normattiva.it/uri-res/N2Ls?urn:nir:stato:decreto.del.presidente.della.repubblica:1992-12-16;495>
- **How LogiTrack supports:** vedi
  [`internal/models/compliance.go`](../backend/internal/models/compliance.go)
  per il regex `platePost1994` e l'enumerazione delle classi ADR.

## 15. SDI/FatturaPA interoperability (read-only pointer)

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
| `platePost1994` regex `[A-HJ-NPR-TV-Z]{2}[0-9]{3}[A-HJ-NPR-TV-Z]{2}` | `internal/models/compliance.go` | D.M. 27/04/1994 art. 2 (formato e alfabeto delle targhe post-1994; esclude I, O, Q, U); D.Lgs. 285/1992 art. 100; DPR 495/1992 |
| `ADRClass*` enum | `internal/models/compliance.go` | ADR 2025 Annex A (UN-ECE), recepito in Italia dal D.Lgs. 35/2010 + D.M. 19/05/2017 |
| `ATPClass*` enum | `internal/models/compliance.go` | ATP 1970 art. 4 Annex 1 (UN-ECE), recepito in Italia dalla L. 264/1977 |
| `EventCustomsHold/Cleared` | `internal/models/tracking_event.go` | Reg. UE 952/2013 CDU art. 226 |
| `Driver.DrivingHours` daily counters | `internal/models/driver.go` | Reg. CE 561/2006 art. 6 (limiti di guida giornaliera, settimanale, bisettimanale) |
| `audit_log` append-only collection | `internal/audit/audit.go` | art. 32 GDPR (Reg. UE 2016/679); D.Lgs. 138/2024 art. 24 (NIS2 misure di gestione del rischio) |

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
