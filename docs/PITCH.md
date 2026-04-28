# LogiTrack Logistics — Pitch One-Pager

> Door-to-door / cold-call pitch for trasportatori, spedizionieri e
> piccole flotte nel Quadrante Europa, lungo il Brennero (A22) e
> sull'A4 Verona–Milano. Companion to the rifiuti pitch in
> [`PITCH-RIFIUTI.md`](PITCH-RIFIUTI.md).

## Il problema

Le piccole flotte (5–30 mezzi) e gli spedizionieri PMI in Veneto
combattono ogni giorno con tre punti dolenti:

1. **Visibilità in tempo reale frammentata** — la telematica del
   produttore (Viasat, Octo, Geotab) parla solo al portale del
   produttore. Il dispatcher salta tra 3 schermate per tracciare un
   carico.
2. **Catena di custodia su carta** — CMR cartaceo, foto su WhatsApp,
   firma del destinatario in PDF spedito via mail. Quando il cliente
   contesta, mancano sempre i pezzi.
3. **Software gestionale generico che non parla con la cabina del
   camion** — l'autista non vede il piano carichi, il responsabile non
   vede dove è il mezzo, l'amministrazione non vede il giorno-uomo
   speso per ogni cliente.

I gestionali commerciali (Tesi, GTS, AS-Software, Mertel) coprono
una parte ma:

1. Sono SaaS multi-tenant con canone mensile per posto.
2. La UI mobile per l'autista è generica.
3. Le integrazioni sono spesso limitate al portale del produttore
   telematico unico, lock-in inclusa.
4. Il vendor sa più di te dei tuoi dati.

## La soluzione

LogiTrack è un **kit di software**: un freelancer (Renan Augusto
Macena, Mozzecane VR) prende la base Go + MongoDB + Redis + Vue 3,
**la fa fork sul tuo nome**, la personalizza per i tuoi flussi, la
deploya sul tuo VPS (o sul tuo Kubernetes, o su un VPS Aruba IT
intestato a te), e ti consegna le chiavi.

Cosa porta il kit "out of the box":

1. **Mappa live multi-mezzo** con marker che si muovono in tempo
   reale, polilinea OSRM (self-hosted opzionale) o straight-line di
   fallback, geofence in/out per cantieri/depositi/terminal.
2. **Dashboard dispatcher unificato** — una sola schermata, filtri per
   stato/vettore/data, timeline eventi per spedizione, ETA aggiornato
   sull'ultima velocità misurata.
3. **Catena di custodia firmata SHA-256** — append-only, ogni
   transizione (creato, caricato, sigillato, in transito,
   consegnato, eccezione) genera un record immutabile col prev-hash
   del precedente. Tampering rilevabile in audit.
4. **Ingest webhook telematica** generico (Viasat, Octo, Geotab e
   simili). L'integrazione con la specifica fonte è un adapter di
   poche centinaia di righe da scrivere durante l'engagement.
5. **WebSocket live** per il dashboard, autenticazione JWT solida,
   rate-limit sia all'handshake che per connessione, idle disconnect
   a 5 minuti, slow-consumer drop.
6. **Scaffolding multi-tenant** — anche se il kit lo deployi per UN
   cliente, lo scoping per `tenant_id` è già nel repository layer.
   Aggiungere un secondo magazzino o una seconda ragione sociale è
   gratis.

## Target customer

- Trasportatore o spedizioniere PMI con **5–30 mezzi**.
- Cluster Verona Sud / Mantova / Brescia / asse Brennero, attività
  prevalente sull'A4, A22 e Quadrante Europa.
- Cooperative o consorzi di micro-trasportatori che condividono
  back-office.
- Imprese famigliari con dispatcher unico e bisogno di una
  dashboard pulita da mostrare al cliente committente estero.

## Cosa NON è LogiTrack

- **Non è un SaaS.** Niente canone mensile per utente. Niente cloud
  condiviso. Niente "i tuoi dati sui nostri server".
- **Non è una sostituzione del gestionale fiscale.** Si integra col
  tuo gestionale esistente via export CSV (e eventualmente API REST
  se il vendor le espone).
- **Non è un TMS enterprise.** Niente moduli che non hai chiesto.
  Niente bloat da listino. Quello che non serve, non c'è.
- **Non è un servizio doganale.** AIDA, FERTRAM, Telepass: se ti
  servono, si integrano dopo l'engagement iniziale come modulo
  separato, non come "feature di listino" del kit.

## Come si ingaggia

1. **Discovery 1h gratis** — caffè a Mozzecane o videochiamata. Capire
   parco mezzi, flussi, fonte telematica, gestionale fiscale, dolori
   reali.
2. **Preventivo a corpo** — progetto da 4 a 8 settimane,
   personalizzazione + onboarding + handover. Vedi
   [`FREELANCER-COMMERCIAL-MODEL.md`](FREELANCER-COMMERCIAL-MODEL.md).
3. **Setup deployment** — VPS Aruba (IT) o Hetzner (DE/FI) intestato
   al cliente, oppure on-premise se hai già un'infrastruttura.
4. **Retainer mensile opzionale** — per aggiornamenti normativi,
   piccoli interventi, monitor health. Mai SLA di prodotto: SLA di
   freelancer (8×5 best-effort, response in 24h lavorative).

## Contatto

- **Renan Augusto Macena** — Mozzecane (VR), Italia
- Email: renanaugustomacena@gmail.com
- Demo 30 min in caffè a Mozzecane o in videochiamata.
