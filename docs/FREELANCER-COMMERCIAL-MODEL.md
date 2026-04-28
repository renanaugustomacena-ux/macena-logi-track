# LogiTrack — Freelancer Commercial Model

How to price, contract, invoice and collect on a LogiTrack engagement
as a sole-trader freelancer in Italy. The model deliberately rejects
the SaaS-canone-mensile-per-posto playbook of TeamSystem / Passepartout
/ Tesi: the freelancer cannot safely offer 24/7 multi-tenant uptime,
and the customer base in Italian PMI prefers "il mio software, sul mio
server" over "il loro abbonamento, sul loro cloud".

## 1. The contract: contratto d'opera ex art. 2222 c.c.

LogiTrack engagements ship under a **contratto d'opera** (codice
civile art. 2222 — la prestazione d'opera).

Why not contratto di consulenza?
- Consulenza is hourly time-and-materials with no committed
  deliverable. That suits an open-ended advisor, not a freelancer
  shipping a working system.

Why not contratto di appalto (art. 1655 c.c.)?
- Appalto requires "organizzazione dei mezzi necessari e gestione a
  proprio rischio" beyond what a single freelancer can credibly
  claim. Plus, INPS reclassification risk if a single freelancer
  appalts to a single committente for > 8 months.

Why contratto d'opera ex art. 2222?
- Defines a specific opera (the customer's deployed LogiTrack fork +
  the personalisation) for a specific compenso, with the freelancer
  performing it personally. Clean fit. Income classified as
  "redditi di lavoro autonomo" (TUIR art. 53), regime dei minimi
  forfettario fino a € 85.000 di ricavi/anno (Legge 197/2022 art. 1
  c. 54).

### Minimum clauses every contract must include

1. **Identificazione delle parti**: ragione sociale + P.IVA del
   committente, dati anagrafici + P.IVA del freelancer.
2. **Oggetto**: "personalizzazione e deployment di LogiTrack
   versione X.Y.Z, vertical <logistics|rifiuti>, secondo lo scope
   definito in Allegato A".
3. **Allegato A — scope tecnico**: lista di funzionalità incluse,
   integrazioni custom, deliverable di handover. NON un elenco di
   "tutto quello che c'è nel kit" ma quello che sarà specifico per
   il cliente.
4. **Compenso e termini**: importo netto IVA esclusa,
   tre stati di avanzamento (33% alla firma, 33% al go-live di
   staging, 34% al go-live di produzione + handover).
5. **Tempi**: data inizio, data prevista go-live, milestone
   intermedi.
6. **Proprietà intellettuale e licenze** (vedi §11 sotto per il
   modello completo): il committente riceve una **licenza esclusiva,
   perpetua e illimitata d'uso** dell'applicazione personalizzata, con
   pieno accesso al codice sorgente del proprio deployment depositato
   su un repository Git privato a lui intestato. Il freelancer
   mantiene la proprietà dei **componenti tecnologici generici della
   piattaforma di base** (librerie, pattern di sicurezza, modelli
   generici) e il diritto di riutilizzarli per altri progetti. Se il
   committente desidera esclusiva sull'intera piattaforma di base
   (nessun altro cliente in grado di usarla), il prezzo si triplica.
7. **Riservatezza**: NDA reciproca. Definire perimetro (dati clienti
   del committente trattati durante l'engagement).
8. **GDPR + Trattamento Dati**: il freelancer è "Responsabile del
   Trattamento" ex art. 28 GDPR durante la fase di sviluppo e
   l'eventuale retainer; il committente è il "Titolare". Allegato
   B — DPA con clausole standard (vedi
   [`COMPLIANCE.md`](COMPLIANCE.md)).
9. **Risoluzione**: clausola di recesso a 30 giorni con saldo del
   lavoro fatto. Penali simmetriche: cliente in mora > 60 gg →
   sospensione dei lavori; freelancer in mora > 30 gg sui milestone
   → riduzione 5% per ogni 7 gg di ritardo non giustificato.
10. **Foro competente**: Verona (o foro del committente, da
    negoziare).

## 2. Pricing matrix (orientative, IVA esclusa)

Pricing is a starting point per type of engagement. Final price
emerges from the discovery call and depends on actual customisation
scope.

### One-off engagement

| Vertical | Scope | Settimane di lavoro | Compenso netto |
| --- | --- | --- | --- |
| Logistics, base | 1 telematica integration + dashboard branded + 5-10 mezzi seed | 4-6 | € 14.000 – 22.000 |
| Logistics, complex | + integrazione gestionale fiscale (CSV o API) + portale committenti + custom dashboard | 6-10 | € 24.000 – 40.000 |
| Rifiuti, base | Anagrafiche + FIR + RENTRI sandbox + dashboard | 5-7 | € 18.000 – 26.000 |
| Rifiuti, complex | + integrazione gestionale rifiuti esistente + MUD export semi-automatico + RENTRI live | 7-12 | € 28.000 – 48.000 |
| Multi-vertical | Logistics + rifiuti combined | 10-16 | € 38.000 – 65.000 |

Discounts:
- **First design-partner del modulo**: −20% to −30%.
- **Cooperativa o consorzio (≥ 3 imprese parallele)**: −15% per impresa.
- **Pagamento integrale all'inizio (no milestone)**: −5%.

Premium:
- **Esclusiva sul kit** (nessun altro cliente futuro): ×3.
- **Hosting on-premise con vincoli speciali** (kerberos, mTLS
  customer-specific, air-gapped): +25% to +50% sul base.

### Retainer mensile (after go-live)

Three tiers; cap on freelancer commitments per tier so the freelancer
doesn't oversell time.

| Tier | Compenso mensile | Inclusioni | Cap freelancer |
| --- | --- | --- | --- |
| Light | € 350 | Patch security + regulatory critiche; revisione mensile log + backup | 4 ore/mese |
| Standard | € 750 | Tutto Light + 1 piccola feature ogni mese (≤ 4 ore) + supporto email/telefono 8×5 | 12 ore/mese |
| Active | € 1.500 | Tutto Standard + 2 feature/mese + intervento same-day per critical | 24 ore/mese |

Tutto oltre il cap: € 80/ora T&M, in giornate intere o mezze giornate
(no frazione minore di 4 ore).

Tutti i tier escludono:
- Sviluppo di moduli verticali nuovi (= nuovo contratto d'opera).
- Migrazioni infrastrutturali (cambio hosting, K8s onboarding).
- Recupero dati post-incidente per sinistri non causati dal codice
  (es. ransomware sul VPS del cliente).

## 3. Italian fiscal reality

### Regime forfettario (RF1)

- Soglia: ricavi annui ≤ € 85.000 (Legge 197/2022).
- Coefficiente di redditività per "altre attività professionali"
  (codice ATECO 62.01.00 sviluppo software): **78%**.
- IVA: esente (Legge 190/2014 art. 1 c. 54). Niente liquidazione IVA,
  niente IVA in fattura, niente split-payment.
- IRPEF sostitutiva: **5%** primi 5 anni se nuova partita IVA con i
  requisiti, **15%** dopo.
- Contributi INPS Gestione Separata (NON forfettario):
  **26,07%** (2026) sul reddito imponibile (ricavi × 78%) ridotto del
  35% di franchigia se gestione separata + senza altra previdenza
  obbligatoria.

Esempio numerico (ricavi 2026 = € 60.000, primo anno con i requisiti):
- Reddito imponibile = 60.000 × 78% = 46.800
- Imposta sostitutiva = 46.800 × 5% = **2.340**
- INPS Gestione Separata = 46.800 × 26,07% × (1 − 0,35) = ~**7.928**
- Netto in tasca ≈ 60.000 − 2.340 − 7.928 = **49.732 €**

### Quando uscire dal forfettario

- Quando ricavi annui superano € 85.000 (passa a regime ordinario IVA).
- Quando si vuole dedurre spese reali (hardware, software pro,
  formazione) > 22% dei ricavi.
- Quando si assumono dipendenti / altri freelancer in modo strutturato
  (oltre € 20.000/anno di costi del personale).

Il forfettario è ottimale per il freelancer LogiTrack tipico: 2-5
clienti l'anno, ricavi tra € 40.000 e € 80.000, spese reali contenute.

### FatturaPA via SDI

Tutte le fatture, anche in regime forfettario, devono essere
elettroniche dal 01/01/2024 (Legge 178/2020 art. 1 c. 1108-1110, ed
estensione forfettari Legge 197/2022).

Il freelancer ha bisogno di:
- Codice destinatario / PEC del committente per il routing SDI.
- Software / portale FatturaPA: Aruba FatturaElettronica € 9/mese,
  o portale gratuito Agenzia Entrate (Fatture e Corrispettivi).
- Conservazione decennale: il portale AdE conserva, oppure
  conservatore accreditato AgID a pagamento.

## 4. Cash-flow management

I committenti PMI italiane pagano spesso a 60-90 giorni nonostante
il contratto preveda 30. Pianifica come segue:

- **Acconto 33% alla firma**: incassato entro 7-15 giorni dalla
  firma. Anchor del cash-flow.
- **Milestone go-live staging (33%)**: pagabile entro 30 giorni dal
  go-live. Realisticamente arriva a 45-60.
- **Saldo (34%)**: pagabile entro 30 giorni dal go-live di
  produzione + handover firmato. Realisticamente 60-90.
- **Retainer**: fattura emessa il 1° del mese, pagamento entro 30
  giorni. Realisticamente arriva a 45-60. Incasso costante una
  volta a regime.

Bagaglio cuscinetto consigliato: **3-6 mesi di costi fissi** in
liquidità per attraversare ritardi, periodi senza nuovi clienti,
tempi morti tra un engagement e l'altro.

## 5. Non si fa: red lines commerciali

- **Niente lavoro a "share di equity" o "% dei ricavi futuri"**.
  Pagamento in fiat, sui milestone, sempre. Equity in piccola PMI
  italiana = illiquidità eterna.
- **Niente "ti faccio una demo gratis con i tuoi dati"**. La discovery
  è gratis. La demo del kit funzionante è gratis. La personalizzazione
  con dati del cliente parte solo a contratto firmato + acconto
  incassato.
- **Niente "ti consegno tutto quando hai finito di pagare"**. Antitetico
  al modello kit: il committente prende possesso del fork in stati
  di avanzamento; al saldo finale è già operativo.
- **Niente clausole di esclusiva non pagate**. Se il committente vuole
  che tu non lavori per concorrenti, paga il premium ×3 sul base.
- **Niente prestazioni illimitate**. Ogni tier di retainer ha un cap
  ore esplicito. Oltre, T&M.
- **Niente fatturazione a un'altra ragione sociale per "ottimizzazione
  fiscale"**. Il committente paga te alla tua P.IVA, su una fattura
  intestata a lui. Se chiede triangolazioni, è una bandiera rossa
  fiscale (fatture per operazioni inesistenti, art. 8 D.Lgs.
  74/2000) che non vale alcun cliente.
- **Niente accesso permanente alle credenziali di produzione del
  cliente dopo l'handover**. Lo scope del retainer prevede accesso
  on-demand con audit log, non ssh permanente.

## 6. Quando rifiutare un'opportunità

- **Cliente con cultura di "il software deve essere gratis"**. Lascia
  perdere.
- **Cliente che chiede SLA di prodotto (99,9%, response in 1h)**.
  Non sei una NOC. Spiega il limite del modello freelancer e, se
  insistono, indirizza verso TeamSystem / Modular.
- **Cliente con dati a regime di esportazione sensibile** (dual-use
  Reg. UE 2021/821, dati di cliente PA classificati). La compliance
  è sopra il livello di un freelancer; serve studio professionale.
- **Cliente che vuole comprare il kit "completo, una volta, per
  sempre"**. Spiega che il kit non è un prodotto. Se insistono per
  un'acquisizione, prezzo: minimo 5× il preventivo standard, e il
  freelancer perde il diritto di usare il kit per altri clienti.
  Quasi sempre non vale la pena.
- **Cliente che chiede di rivendere il kit a terzi sotto il proprio
  brand**. Stessa risposta: prezzo licenza enterprise, o no.

## 7. Numeri di riferimento per il primo anno

Anno 1 ipotizzato:
- 3 clienti completati (preventivo medio € 25.000) = € 75.000 ricavi
  da progetti.
- 2 clienti già in retainer Standard da metà anno (€ 750 × 6 mesi × 2)
  = € 9.000 ricavi da retainer.
- **Totale Anno 1 ≈ € 84.000** (sopra soglia forfettaria — pianifica
  l'uscita dal regime, oppure modula il volume per restare sotto).

Costi diretti tipici:
- Software / hosting personale (laptop, IDE, dominio, VPS demo):
  ~€ 2.000/anno.
- Formazione + viaggi: ~€ 1.500/anno.
- Commercialista + consulenze: ~€ 2.000/anno.
- Marketing (poco; principalmente referral): ~€ 500/anno.
- **Totale costi ~ € 6.000**.

Con il regime forfettario, i costi non sono deducibili (il 22% di
forfait copre già le spese). Quindi i costi effettivi escono dal
netto in tasca, riducendolo, ma non riducono l'imposta. È un
trade-off che vale la pena solo se le spese reali sono < 22% dei
ricavi.

## 11. Modello di licenza — proprietà di cosa, di chi

Da spiegare al cliente in linguaggio chiaro, da pinnare nel contratto
in clausola dedicata. Tre livelli:

### 11.1 La piattaforma tecnologica di base — di proprietà del freelancer

Le componenti tecnologiche generiche e riutilizzabili tra clienti
diversi:

- Librerie tecniche (autenticazione, audit log, repository pattern,
  catena di custodia firmata, validatori EER/CER, plate validator
  italiano, integrazione RENTRI, ecc.).
- Pattern di sicurezza (alg-pinning JWT, rate-limit con eviction,
  SSRF guards, hardening produzione).
- Architettura della piattaforma (Go + MongoDB + Redis + Vue 3,
  separazione modulare, contratti tra layer).

Restano di **proprietà esclusiva del freelancer**, che le riutilizza
in ogni nuovo progetto. Esattamente come uno studio di architettura
riutilizza i propri schemi strutturali standard. Il committente NON
acquisisce diritti su queste componenti generiche.

### 11.2 L'applicazione personalizzata — licenza esclusiva perpetua al committente

Quello che il freelancer sviluppa **specificamente per il committente**:

- Branding, palette colori, logo, layout adattato.
- Adapter di integrazione con la specifica telematica del committente.
- Personalizzazione dei flussi operativi (regole di stato, workflow
  approvativi, custom validation).
- Configurazione delle anagrafiche, dei ruoli operatore, dei report.
- Eventuali moduli verticali commissionati ad hoc.

Il committente riceve una **licenza esclusiva, perpetua, irrevocabile
e illimitata** d'uso dell'applicazione personalizzata, con accesso al
codice sorgente del proprio deployment depositato su un repository Git
privato a lui intestato. Esclusiva significa che il freelancer non
distribuirà la stessa applicazione personalizzata (con lo stesso
brand, gli stessi adapter, gli stessi flussi specifici) ad altri
clienti. La licenza copre uso interno illimitato, modifica del codice,
ma NON la rivendita o sublicenza a terzi sotto il marchio LogiTrack
senza accordo separato.

Conseguenza pratica per il committente: se domani il freelancer
sparisce o sceglie un altro tecnico, l'applicazione resta funzionante
sul server, il codice sorgente è disponibile, un altro sviluppatore Go
può prenderla in mano e mantenerla. Niente lock-in vendor.

### 11.3 I dati operativi — proprietà esclusiva del committente

Tutto ciò che il committente carica o produce attraverso
l'applicazione (anagrafiche dei suoi clienti finali, FIR, posizioni
mezzi, audit log, configurazioni operative) è **al 100% del
committente**, sempre. Il freelancer non ne acquisisce alcun diritto
e ne accede solo nell'ambito della manutenzione concordata, registrato
nel registro dei trattamenti GDPR Art. 30 come Responsabile del
Trattamento.

A fine contratto: export completo in formati aperti (JSON, CSV,
GeoJSON) consegnato entro 30 giorni; il freelancer cancella ogni copia
locale entro 60 giorni dalla consegna dell'export, con dichiarazione
sostitutiva di atto di notorietà.

### 11.4 Esclusiva totale (opzione premium ×3)

Il committente che desidera anche **l'esclusiva sulla piattaforma di
base** (i.e. il freelancer non potrà riutilizzare la piattaforma per
altri clienti del medesimo settore o area geografica) paga il
preventivo base moltiplicato per tre. È un'opzione raramente
consigliata: la piattaforma di base trae beneficio dal mantenimento
condiviso (patch normative, fix di sicurezza, ottimizzazioni)
distribuito su più clienti.

### 11.5 Da pinnare nel preventivo — formula sintetica

> "L'applicazione personalizzata è in licenza esclusiva, perpetua e
> irrevocabile alla committente. Il codice sorgente del deployment
> della committente è depositato su un repository Git privato a lei
> intestato. La committente è titolare esclusiva dei dati operativi.
> Il fornitore mantiene la proprietà dei componenti tecnologici
> generici della piattaforma di base e il diritto di riutilizzarli
> per altri progetti, senza alcun riferimento al brand o ai dati
> della committente."

Questa formula è molto vicina a come operano gli studi di
architettura, gli avvocati associati e le software house italiane
serie. È difendibile in udienza, accettabile dal commercialista del
cliente, e non richiede consulenza legale costosa per essere
spiegata.
