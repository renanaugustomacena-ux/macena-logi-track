# LogiTrack Rifiuti — Pitch One-Pager

> One-pager per trasportatori e intermediari di rifiuti speciali nel
> corridoio Verona Sud / Mantova / Brescia. Companion alla one-pager
> logistica ([`PITCH.md`](PITCH.md)) che si rivolge a un ICP diverso.
>
> Modello di ingaggio: software gestionale dedicato per cliente,
> installato su infrastruttura intestata al cliente. NON un servizio
> cloud condiviso. Dettagli commerciali in
> [`FREELANCER-COMMERCIAL-MODEL.md`](FREELANCER-COMMERCIAL-MODEL.md).

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

LogiTrack Rifiuti è il modulo verticale per il trasporto di rifiuti
speciali, costruito sulla stessa piattaforma tecnologica del modulo
logistico. Quando ti ingaggio sviluppo per la tua azienda una
applicazione gestionale dedicata, brandizzata, installata su un
server intestato a te (Aruba IT, infrastruttura on-premise, oppure
cloud cliente) — non su un cloud condiviso con altre aziende. Cosa
aggiunge il modulo rifiuti rispetto a quello logistico:

1. **Modello di dominio** Produttore / Trasportatore /
   Destinatario / FIR / Albo, con classi di pericolo HP1–HP15
   (Reg. UE 1357/2014) e mapping ADR per i rifiuti pericolosi.
   Onestà sul perimetro: il registro cronologico carico/scarico e
   il catalogo EER completo (Decisione UE 2014/955) **non sono
   ancora nel kit** — oggi c'è la validazione del formato dei
   codici EER — e vengono sviluppati nelle prime settimane
   dell'ingaggio.
2. **Ciclo di vita FIR digitale completo** (bozza → vidimato →
   consegnato al trasportatore → in transito → consegnato al
   destinatario → chiuso, con rami "respinto" e "annullato"). Ogni
   transizione di stato è validata; watchdog automatico a 90 giorni
   per la copia produttore (art. 188-bis c. 4 TUA).
3. **Verifica Albo + autorizzazione impianto** in tempo reale prima
   della firma del FIR: una `Trasportatore.CanCarry(cer, oggi)`
   blocca il movimento se l'iscrizione è scaduta o la categoria non
   copre il codice; una `Destinatario.CanReceive(cer, op, oggi)`
   blocca se l'impianto non è autorizzato per quel CER + operazione.
4. **Integrazione RENTRI a doppio binario**: nei primi giorni il
   software opera in modalità "coda interna" — accumula i FIR
   pronti per la trasmissione, garantisce idempotenza in caso di
   ritrasmissione, ti permette di esercitare la piattaforma sui
   tuoi dati reali. L'adapter HTTP verso la sandbox RENTRI **non è
   ancora scritto**: si sviluppa durante l'ingaggio sopra
   l'interfaccia già pronta della coda interna, e il passaggio alla
   trasmissione live avviene quando arriva il certificato digitale
   RENTRI intestato alla tua impresa.
5. **Catena di custodia SHA-256** implementata e testata a livello
   di piattaforma (eredità del modulo logistico); il collegamento
   degli eventi operativi (carico, sigillo, consegna) alla catena
   si completa durante l'ingaggio.
6. **Tracciamento GPS dei mezzi** tramite endpoint generico di
   waypoint già funzionante, con pannello dispatcher in tempo
   reale. Gli adapter per le telematiche commerciali (Viasat,
   Octo, Geotab) **non sono inclusi nel kit**: si sviluppano
   durante l'ingaggio su specifica del fornitore scelto.

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

Engagement standard: **6–8 settimane** dalla firma del preventivo
alla consegna delle chiavi dell'applicazione operativa.

- **Settimana 1** — onboarding anagrafiche Produttori /
  Trasportatori / Destinatari e caricamento delle iscrizioni Albo
  del cliente. La verifica di scadenza e categoria alla creazione
  del FIR è già attiva nel kit; il monitoraggio giornaliero in
  background si sviluppa durante l'ingaggio.
- **Settimana 2** — ingestione FIR esistenti via import CSV /
  scansione retroattiva, taratura della macchina a stati con i
  flussi reali del cliente, attivazione watchdog 90 giorni.
- **Settimana 3** — collegamento opzionale di una telematica
  veicolare (Viasat o equivalente) e attivazione del cruscotto
  dispatcher.
- **Settimana 4** — trasmissione di un FIR pilota end-to-end in
  modalità coda interna; preparazione della delegazione SPID/CIE/CNS
  per il certificato RENTRI di produzione.
- **Settimana 6** — passaggio alla trasmissione RENTRI live appena
  il certificato è disponibile; attivazione della conservazione a
  norma AgID via conservatore accreditato selezionato in onboarding.
- **Settimana 7-8** — handover, formazione operatori, transizione
  al retainer mensile.

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
