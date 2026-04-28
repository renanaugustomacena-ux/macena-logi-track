# LogiTrack Rifiuti — Pricing & Licensing

> Vertical pricing for the rifiuti speciali module (RENTRI-ready).
> The horizontal logistics module pricing is in `PRICING.md`.
> Tiers are independent: a customer that already runs LogiTrack
> Logistics can add LogiTrack Rifiuti à la carte.

All prices in EUR, IVA esclusa, monthly billing, no multi-year
lock-in unless explicitly negotiated.

## Tier matrix

| Capability | Starter | Pro | Multi-deposito |
| --- | --- | --- | --- |
| Canone mensile | **€ 249** | **€ 749** | **€ 1.490** |
| Setup una-tantum (incl. onboarding catalogo Albo + impianti) | € 490 | € 1.290 | € 2.490 |
| FIR / mese (fair use) | 200 | 1.500 | 5.000 |
| Mezzi tracciati | 5 | 30 | 80 |
| Sedi operative | 1 | 1 | 5 |
| Utenti operatori | 3 | 12 | 30 |
| RENTRI sandbox | sì | sì | sì |
| RENTRI live (post-certificato) | add-on | incluso | incluso |
| Verifica Albo controparte (background) | settimanale | giornaliera | giornaliera |
| Watchdog 90 giorni copia produttore | sì | sì | sì |
| Watchdog scadenza Albo + autorizzazione impianto | sì | sì | sì |
| Conservazione a norma AgID | add-on | incluso (conservatore selezionato in onboarding) | incluso |
| Telematica veicolare (Viasat / Octo / Geotab) | 1 integrazione | tutte | tutte + custom |
| Cruscotto dispatcher mobile (PWA) | sì | sì | sì |
| ADR helper (UN, classe, codice tunnel) | sì | sì | sì |
| Export MUD annuale | manuale | semi-automatico | automatico + revisore |
| SLA | best-effort 8×5 | 99,5 % business hours | 99,9 % 24×7 + penali |
| Regione dati | Aruba IT | Aruba IT | Aruba IT o customer-private |
| Supporto | email 8×5 | email + telefono 12×6 | CSM dedicato |

## Cosa è incluso "out of the box"

- Modello dati produttore / trasportatore / destinatario / FIR /
  registro cronologico, validatori CER, classi pericolo HP1–HP15,
  enum operazioni R/D dagli Allegati B + C TUA.
- Macchina a stati FIR completa con guard `AdvanceState`.
- Adapter `rentri.QueuedStub` di default; cutover all'adapter
  HTTP live richiede solo il certificato RENTRI del cliente.
- Sandbox `demoapi.rentri.gov.it` + `demobackoffice.rentri.gov.it`
  preconfigurate.
- Catena di custodia firmata SHA-256, audit-log multi-tenant
  (eredità del modulo logistico).
- 100 % dei modelli di documento (FIR, registro, MUD) basati su
  schemi pubblici verificati 2026-04-28.

## Add-on a giornata

- Connettore custom (gestionale fiscale, ERP, EDI legacy):
  **€ 800 / giornata** T&M.
- Onboarding aggiuntivo (sede operativa supplementare oltre il
  tier): **€ 290 / sede**.
- Formazione in sede a Mozzecane / Verona: **€ 1.200 / giornata**
  fino a 8 partecipanti.
- Consulenza pre-RENTRI (analisi anagrafica, mapping CER →
  impianti destino, taratura categoria Albo): **€ 1.500 / settimana
  uomo**, prima settimana al 50 %.
- Consulente DGSA esterno (per cliente che non ne ha uno):
  partnership con consulente terzo locale, costo passante.

## Trial

- **45 giorni di pilota** sul tier Pro, completo. Setup
  riconosciuto al 50 % se la commessa parte entro 60 giorni.
- Nessuna carta di credito; lettera di intenti firmata che
  delimita lo scope del pilota e i KPI di accettazione.

## Self-hosted

- Self-hosted disponibile per il tier Multi-deposito a partire da
  **+ € 4.900 / anno**: include manifesti Kubernetes, immagini
  patchate trimestralmente, supporto al deploy per 5 giornate
  uomo.
- Hybrid (telematica on-prem + dashboard SaaS) è il default
  consigliato per i clienti con vincoli di residenza dati spinti.

## Billing

- Fatturazione mensile via FatturaPA (SDI).
- Termini di pagamento: 30 gg fattura fine mese.
- Tolleranza: un mese di ritardo + diffida formale, sospensione
  servizio a 45 gg.

## Discount strutturali

- Cooperativa o consorzio con N ≥ 5 trasportatori: **−15 %** sul
  canone Pro per ogni iscritto, fatturazione centralizzata.
- Camera di Commercio / sezione regionale Albo che reseller del
  servizio: trattativa dedicata.
- Consulente ambientale che porta il suo portafoglio clienti come
  reseller: **20 %** revenue share sul primo anno per ogni
  cliente attivato.

## Note legali

LogiTrack non è un soggetto autorizzato a sostituire né il DGSA
né il consulente Albo del cliente. Il software prepara, valida,
firma e trasmette i documenti ma la responsabilità regolatoria
resta in capo al titolare dell'impresa, come previsto dall'art.
188-bis D.Lgs. 152/2006.
