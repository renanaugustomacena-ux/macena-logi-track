# LogiTrack Rifiuti — Pitch One-Pager

> Vertical pitch for trasportatori e intermediari di rifiuti speciali
> nel corridoio Verona Sud / Mantova / Brescia. Companion to the
> logistics one-pager (`PITCH.md`), which targets a different ICP.

## Il problema

Il **15 settembre 2026** il FIR cartaceo non è più valido. Ogni
trasportatore di rifiuti speciali, ogni intermediario, ogni gestore
di impianto in Italia deve già oggi essere iscritto al RENTRI
(art. 188-bis D.Lgs. 152/2006 + D.M. MASE 4 aprile 2023, n. 59) e
trasmettere FIR digitali firmati XAdES e movimenti di carico/scarico
in formato XML allo schema `rentri-formulario-1.0.xsd`. Le
sanzioni sono concrete:

- **€ 1.600 – 10.000** per ogni FIR errato o non trasmesso.
- **€ 4.000 – 20.000** (non pericolosi) o **€ 10.000 – 30.000**
  (pericolosi) per registro carico/scarico incompleto.
- **Arresto 3-12 mesi o ammenda € 2.600 – 26.000** se trasporti
  rifiuti senza iscrizione Albo (art. 256 c. 1).
- **Sospensione patente di guida + sospensione iscrizione Albo**
  per recidive (DL 116/2025).

I gestionali rifiuti del mercato (TeamSystem Waste 360, Passepartout,
Ambiente.it, Rifiutoo, Winwaste, Modular SVFOR02, Ecofacile)
risolvono parte del problema ma:

1. Tengono i listini riservati e gli iter commerciali sono lunghi.
2. Non espongono API pubbliche per integrazione TMS/ERP — lock-in.
3. La UX mobile per l'autista è quasi sempre un afterthought.
4. La verifica in tempo reale dell'iscrizione Albo della
   controparte non è quasi mai nativa.
5. La conservazione a norma AgID raramente è inclusa nel canone.

## La soluzione

LogiTrack Rifiuti è il modulo verticale del kit LogiTrack costruito
sopra la stessa piattaforma multi-tenant Go + MongoDB + Redis che
serve già il modulo logistico. Aggiunge:

1. **Modello di dominio completo** Produttore / Trasportatore /
   Destinatario / FIR / Registro cronologico / Albo, con codici
   EER (catalogo Decisione UE 2014/955), classi di pericolo
   HP1–HP15 (Reg. UE 1357/2014), e mapping ADR per i rifiuti
   pericolosi.
2. **Macchina a stati FIR completa** (`draft → vidimato →
   consegnato_trasportatore → in_transito → consegnato_destinatario →
   chiuso`, più rami `respinto` e `annullato`), con guard su ogni
   transizione e watchdog dei 90 giorni per la copia produttore
   (art. 188-bis c. 4).
3. **Verifica Albo + autorizzazione impianto** in tempo reale prima
   della firma del FIR: una `Trasportatore.CanCarry(cer, oggi)`
   blocca il movimento se l'iscrizione è scaduta o la categoria non
   copre il codice; una `Destinatario.CanReceive(cer, op, oggi)`
   blocca se l'impianto non è autorizzato per quel CER + operazione.
4. **Adapter RENTRI a doppio binario**: in produzione parte uno
   `QueuedStub` deterministico, idempotency-key-protected, che
   accetta FIR e movimenti, accumula la coda e consente
   l'esercizio completo della piattaforma. Quando il certificato
   digitale RENTRI del cliente è attivo, basta un cambio di
   costruttore in `cmd/server/main.go` e l'adapter HTTP live
   sostituisce lo stub. Sandbox `demoapi.rentri.gov.it` configurata
   come default in non-produzione.
5. **Catena di custodia firmata SHA-256 + audit-log multi-tenant**
   già platform-grade (eredità del modulo logistico).
6. **Tracciamento GPS dei mezzi** (Viasat, Octo, Geotab) e
   geofencing già integrati: il pannello dispatcher vede in tempo
   reale dove è il mezzo, quale FIR sta trasportando, e quanto
   manca all'arrivo.

## Target customer

- Trasportatore rifiuti speciali iscritto Albo cat. **4** e/o
  **5**, con **5–30 mezzi**.
- Intermediario senza detenzione iscritto Albo cat. **8**.
- Cooperative o consorzi di micro-trasportatori che condividono
  back-office.
- Cluster Verona Sud / Mantova / Brescia / Lago di Garda;
  estensione naturale al Trentino e all'asse del Brennero per i
  trasporti transfrontalieri (cat. 6 + Reg. UE 1013/2006).

## Design partner: FRO S.r.l. — Mozzecane (VR)

Primo target di discovery: **FRO S.r.l.** (gruppo Effevi Rottami)
in Via Quartieri snc, 37060 Mozzecane (VR), telefono
**045 6340188**. Famiglia attiva nel settore dal 1954, ISO
14001:2015, attività di trasporto pericolosi + non pericolosi +
intermediazione + container + scrap yard ferroviario interno. Sito
<https://www.frotrasporti.com>. È il candidato design-partner per
il modulo Rifiuti, con backup su una short-list di altri 19
trasportatori entro 25 km dal centro.

## Implementation timeline

- **Settimana 1** — onboarding anagrafiche Produttori /
  Trasportatori / Destinatari, sincronizzazione delle iscrizioni
  Albo del cliente (catalogo curato; verifica giornaliera in
  background).
- **Settimana 2** — ingestione FIR esistenti via import CSV /
  scansione retroattiva, taratura della macchina a stati con i
  flussi reali del cliente, attivazione watchdog 90 giorni.
- **Settimana 3** — collegamento di una telematica veicolare
  (Viasat) e attivazione del cruscotto dispatcher.
- **Settimana 4** — trasmissione di un FIR pilota end-to-end via
  `QueuedStub`; preparazione delegazione SPID/CIE/CNS per il
  certificato RENTRI di produzione.
- **Settimana 6** — cutover sull'adapter HTTP live RENTRI appena
  il certificato è disponibile; conservazione a norma AgID via
  conservatore accreditato selezionato in onboarding.

## Cosa NON facciamo

- Non sostituiamo il consulente DGSA né l'analista chimico per
  l'attribuzione delle caratteristiche di pericolo HP.
- Non emettiamo direttamente fatture: ci integriamo con il
  gestionale fiscale del cliente, non lo rimpiazziamo.
- Non gestiamo il ciclo dei rifiuti urbani — Albo cat. 1 è fuori
  scope per l'ICP iniziale.

## Contatto

- Renan Augusto Macena — Mozzecane (VR), Italia
- Email: renanaugustomacena@gmail.com
- Demo 30 min in caffè a Mozzecane o in videochiamata.
