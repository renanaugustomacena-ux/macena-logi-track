# LogiTrack — Architettura

Documento di riferimento per l'architettura logica e fisica del kit
LogiTrack. Il kit è progettato per essere fork-ato per ogni cliente:
ciò che segue descrive la baseline upstream del kit, non un singolo
deployment cliente.

## 1. Vista d'insieme

LogiTrack combina ingestione telematica via webhook, archiviazione
documentale flessibile, pub/sub Redis per la fan-out WebSocket e una
piattaforma multi-modulo per i verticali. I verticali shipping oggi:
`logistics` (visibilità supply-chain) e `rifiuti` (RENTRI-ready
trasporto rifiuti speciali).

```
   GPS / Telematics webhook ─────┐
   Operator REST/WS calls ───────┼──> Gin handlers (auth, rate-limit)
                                 │         │
                          ┌──────▼─────────▼────────────┐
                          │  Service layer              │
                          │   - Shipment / Tracking     │
                          │   - ETA (moving avg)        │
                          │   - Route optimiser (OSRM)  │
                          │   - Rifiuti FIR + RENTRI    │
                          └──┬───────────┬──────────────┘
                             │           │
                       ┌─────▼───┐  ┌────▼───────────┐
                       │ MongoDB │  │ Redis          │
                       │(shipmts │  │(position cache,│
                       │ custody,│  │ pub/sub for WS)│
                       │ FIR,    │  └────────────────┘
                       │ audit)  │
                       └─────────┘
                             ▲
                             │ REST + WebSocket
                     ┌───────┴───────┐
                     │ Vue 3 SPA     │
                     │ Map dashboard │
                     │ Rifiuti panel │
                     └───────────────┘
```

### 1.1 Componenti principali

- **Handler layer (Gin)**. Riceve webhook telematica, chiamate REST
  REST autenticate da operatori e SPA, upgrade WebSocket. Tutti gli
  endpoint mutanti passano dal middleware `JWTAuth` + `Audit` +
  `RateLimiter`. Il middleware è in `internal/middleware/`.
- **Service layer**. `ShipmentService`, `TrackingService`,
  `ETAService`, `OSRMOptimizer`, e gli handler rifiuti orchestrano
  repository + integrazioni esterne. Definiti in `internal/services/`.
- **Repository layer**. `MongoRepository` (shipments, custody, fleet,
  rifiuti) e `RedisRepository` (position cache, pub/sub). Sotto
  `internal/repository/`. Ogni metodo prende `tenantID` esplicitamente:
  scoping multi-tenant è una proprietà del repository, non un'opzione
  del caller.
- **Module layer**. Le entità di dominio per ciascun verticale vivono
  in `internal/modules/<vertical>/`: `logistics` ospita
  `Shipment`, `Vehicle`, `Driver`, `Geofence`, `CustodyRecord`,
  `TrackingEvent`, validatori targhe + ADR/ATP; `rifiuti` ospita
  `Produttore`, `Trasportatore`, `Destinatario`, `FIR`,
  `RegistroEntry`, validatori CER + Albo, e il sub-package
  `rifiuti/rentri` con il client RENTRI.
- **Frontend SPA (Vue 3 + Vite + TypeScript)**. Composition API,
  Pinia per lo state, Vue Router 4, Leaflet per la mappa. Tre viste
  principali: `HomeView` (dashboard), `ShipmentView` (dettaglio +
  timeline + mappa), `RifiutiView` (anagrafiche + FIR).

### 1.2 Principi architetturali applicati

1. **Twelve-Factor**: configurazione via env (vedi
   [`config/config.go`](../backend/internal/config/config.go)),
   processi stateless, logging strutturato JSON su stdout, scaling
   orizzontale di default.
2. **DDD lite**: pacchetti `modules/<vertical>` ↔ `services` ↔
   `repository` ↔ `handlers` con dipendenze a senso unico
   (handlers → services → repository → modules).
3. **Module isolation**: i moduli verticali non si importano a vicenda.
   Vedi [`DOMAIN-MODULES.md`](DOMAIN-MODULES.md) per il contratto.
4. **API-first**: schema REST/WebSocket versionato in `/api/v1/` e
   documentato in [`API.md`](API.md) prima dell'implementazione.
5. **Security by design**: HS256 JWT con issuer pinnato, rate-limit
   per IP con eviction, CORS esplicito, distroless in produzione,
   redirect-block sull'http client OSRM (anti-SSRF), guardia
   produzione su segreti deboli.
6. **Observability**: OpenTelemetry OTLP exporter (opt-in via env),
   log JSON zap, `/metrics` Prometheus.

## 2. Modello dati (MongoDB)

Le collezioni primarie rispecchiano gli aggregati di dominio. Tutte
sono tenant-scoped: ogni query filtra obbligatoriamente per
`tenant_id`.

### 2.1 `shipments`

```json
{
  "_id": "5c3c…",
  "tenant_id": "ten_mozzecane_sml",
  "reference": "CMR-2026-04-0001",
  "carrier": "Autotrasporti Rossi Srl",
  "mode": "multimodal",
  "status": "in_transit",
  "consignor": { "name": "…", "vat_number": "IT…", "city": "Mozzecane", "province": "VR", "country": "IT" },
  "consignee": { "name": "Müller GmbH", "city": "München", "country": "DE" },
  "origin": { "type": "Point", "coordinates": [10.778, 45.333] },
  "destination": { "type": "Point", "coordinates": [11.5755, 48.1374] },
  "waypoints": [ /* ingested telematics events */ ],
  "current_position": { "type": "Point", "coordinates": [11.01, 45.93] },
  "etd": "2026-04-17T05:00:00Z",
  "eta": "2026-04-17T13:30:00Z",
  "tags": ["quadrante-europa", "brennero"]
}
```

### 2.2 `tracking_events`

Eventi immutabili, uno per aggiornamento posizione o transizione di
stato. Chiave `(shipment_id, sequence)`.

### 2.3 `vehicles`

Anagrafica flotta — targa (validatore italiano post-1994 +
storiche), categoria EURO, marker ADR/ATP, provider telematico.

### 2.4 `drivers`

Anagrafica autisti — patente, CQC, iscrizione Albo Autotrasportatori
(flag, da verificare manualmente o via integrazione futura).

### 2.5 `geofences`

Poligoni GeoJSON con tipo (warehouse, customs, port, depot, terminal,
rest_area, loading_bay).

### 2.6 `chain_of_custody`

Append-only, ogni record firmato SHA-256 sul JSON canonico (con
`time.UTC().Truncate(time.Millisecond)` per round-trip BSON pulito).
Ogni record contiene `prev_hash` del precedente.
`services.VerifyChain` ricomputa la catena per audit.

### 2.7 Collezioni `rifiuti_*`

`rifiuti_produttori`, `rifiuti_trasportatori`, `rifiuti_destinatari`,
`rifiuti_fir`, `rifiuti_registro`. Indici unici su
`(tenant_id, codice_fiscale)` per le anagrafiche e su
`(tenant_id, numero_rentri)` sparse per il FIR.

### 2.8 `audit_log`

Audit middleware emette un record per ogni POST/PUT/PATCH/DELETE non-5xx.
Ingestione async best-effort via `audit.Writer`: sotto pressione
estrema può perdere record (counter `audit writer dropped records`
loggato a WARN). Per uso "compliance audit" guarantee strict serve
incrementare buffer o switch a inserimento sincrono.

## 3. Diagrammi di sequenza

### 3.1 Ingest webhook telematica

```
Carrier → POST /api/v1/shipments/{id}/waypoints → Gin handler
  Gin handler → JWT auth + rate limit + audit
    → ShipmentService.RecordWaypoint
       → Mongo: $push waypoints, set current_position
       → Redis: SET logitrack:position:{id}
       → Mongo: insert tracking_events
       → Redis: PUBLISH logitrack:events:tracking
WebSocket hub ← Redis SUBSCRIBE
  hub → broadcast to matching subscribers
  subscriber → Vue SPA: marker update + timeline append
```

### 3.2 Subscription WebSocket

```
Browser → GET /api/v1/stream/tracking (Upgrade WS, JWT in subprotocol)
  Gin handler → handshake rate-limit (per-IP, evictable)
    → JWT extract + validate (HS256 only, alg-pinned)
    → Origin allow-list check
    → upgrader.Upgrade
    → hub.Register(subscriber)
Browser → {op:"subscribe", shipmentId:"…"}
  handler → set subscriber.ShipmentID
Redis event → hub.Broadcast(evt)
  matching subscribers → conn.Write {type:"event", event:{…}}
Browser keepalive → {op:"ping"} → server {type:"pong"}
```

### 3.3 Append chain-of-custody

```
Operator → POST /api/v1/shipments/{id}/custody (handover)
  Gin handler → JWT auth, role gating
    → ShipmentService.AppendCustody
       → Mongo: find latest sequence + hash for tenant + shipment
       → compute prev_hash = last.hash
       → compute hash = SHA256(canonical JSON sans hash field, time UTC ms)
       → Mongo: insert chain_of_custody
    → return record
```

### 3.4 RENTRI vidimazione FIR

```
Operator → POST /api/v1/rifiuti/fir/{id}/vidima
  Gin handler → JWT auth, audit
    → repo.GetFIR(tenantID, id)
    → if state != FIRDraft → 422
    → idempotency = "FIR-" + sha256(tenantID:firID)[:8]
    → xfir = encoding/xml encode placeholder (CER escaped)
    → rentri.Client.VidimaFIR(IdempotencyKey, XFIRPayload)
       → QueuedStub default: deterministic numero, persist locally
       → live HTTP adapter (when cert available): POST sandbox/prod RENTRI
    → f.NumeroRENTRI = resp.NumeroRENTRI
    → f.AdvanceState(FIRVidimato)
    → repo.UpdateFIR
  → 200 with { fir, numero_rentri, vidimato_at, qr_code_payload }
```

## 4. Integrazioni effettivamente wired oggi

Il kit volutamente NON include adapter per AIDA, FERTRAM/RFI,
Telepass o Albo Autotrasportatori. Sono integrazioni reali che si
aggiungono **per cliente** quando il cliente le chiede e ne paga
l'integrazione. Dettagli in [`INTEGRATIONS.md`](INTEGRATIONS.md).

Quello che il kit fornisce:

- **Telematics ingestion** generico via
  `POST /api/v1/shipments/{id}/waypoints`. L'adapter per il provider
  specifico è una piccola unit di codice da scrivere durante
  l'engagement.
- **OSRM** opt-in (env `OSRM_BASE_URL`); fallback straight-line +
  70 km/h se non configurato.
- **RENTRI** via `rentri.Client` interface +
  `rentri.QueuedStub` default. L'adapter HTTP live è un
  one-file-change quando il certificato del cliente è attivo.

## 5. Architettura di deploy

### 5.1 Sviluppo (docker-compose)

| Servizio | Immagine | Porta host | Volume |
| --- | --- | --- | --- |
| logitrack-backend | build ./backend | 8080 | — |
| logitrack-frontend | build ./frontend | 5174 | — |
| logitrack-mongodb | mongo:7 | 127.0.0.1:27017 | logitrack-mongo-data |
| logitrack-redis | redis:7-alpine | 127.0.0.1:6380 | — |
| logitrack-simulator | build ./backend (Dockerfile.simulator) | — | profile: demo |

Mongo e Redis sono bind-loopback con auth obbligatoria. Vedi
[`../docker-compose.yml`](../docker-compose.yml).

### 5.2 Produzione (per-customer fork)

Non c'è un singolo blueprint cloud "ufficiale": ogni fork sceglie il
suo deployment. Pattern raccomandati:

- **VPS singolo (1-5 mezzi)**: Aruba Cloud (IT-MI, IT-AR), Hetzner
  (DE, FI), 4-8 GB RAM, Docker Compose con backup volumi
  giornaliero su S3-compatibile.
- **Kubernetes piccolo (5-30 mezzi)**: 3 nodi minimi, MongoDB
  replicaset 3 nodi, Redis primary+replica, Caddy o nginx-ingress
  per TLS, CertManager per Let's Encrypt.
- **On-premise**: docker compose su un server cliente, VPN per
  l'accesso amministrativo del freelancer, backup su NAS cliente.

Lo zero-downtime deploy non è incluso nel kit "out of the box":
ogni fork lo aggiunge se necessario.

## 6. Modalità di fallimento e mitigazione

| Rischio | Impatto | Mitigazione |
| --- | --- | --- |
| MongoDB down | ingest interrotto | replicaset 3 nodi, election; backup snapshot giornaliero |
| Redis down | WS fan-out interrotto | client retry; SPA reconnect; replay storico via REST |
| OSRM down o non configurato | ETA meno preciso | straight-line fallback; one-shot WARN log |
| RENTRI sandbox/prod down | vidimazione bloccata | QueuedStub fallback (idempotency-keyed); coda manuale |
| Server VPS hard-down | tutto offline | restore da snapshot giornaliero; SLA freelancer 8×5 |

## 7. Service-level expectations

LogiTrack è un kit, non un SaaS. Non ci sono SLA contrattuali "di
prodotto". Il deployment per il singolo cliente può raggiungere:

| Metrica | Target tipico (best-effort) |
| --- | --- |
| Disponibilità API | 99% / mese su VPS singolo, ≥ 99.9% su K8s replicato |
| Latenza P95 REST | < 300 ms |
| Latenza WebSocket end-to-end | < 2 s |
| Ingest waypoint throughput | ≥ 500 eventi/s (dipende dalla CPU del VPS) |

Gli obiettivi reali sono materia di contratto retainer per cliente.

## 8. Module dependency sketch

```
modules/logistics       → (standalone)
modules/rifiuti         → (standalone)
modules/rifiuti/rentri  → (sub-package, no other modules)

repository → modules/logistics, modules/rifiuti
services   → modules/logistics, repository
handlers   → services, middleware, repository, modules/rifiuti, modules/rifiuti/rentri
middleware → config, audit, problem
audit      → repository
cmd/server → handlers, repository, services, demo, audit, config, modules/rifiuti/rentri
cmd/simulator → demo (only)
```

I moduli verticali NON si importano a vicenda. Aggiungere un nuovo
verticale = aggiungere un sotto-package sotto `modules/<name>/`,
seguire il contratto in [`DOMAIN-MODULES.md`](DOMAIN-MODULES.md), e
wire alla composition root in `cmd/server/main.go`.
