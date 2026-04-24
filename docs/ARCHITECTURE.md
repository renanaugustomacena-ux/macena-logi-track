# LogiTrack — Architettura

Documento di riferimento per l'architettura logica, fisica e di integrazione della piattaforma LogiTrack.

## 1. Vista d'insieme

LogiTrack è un sistema di visibilità supply-chain che combina ingestione
telematica ad alta frequenza, archiviazione documentale flessibile,
pub/sub in memoria per la fan-out WebSocket e integrazioni verticali
con Agenzia delle Dogane (AIDA), Rete Ferroviaria Italiana (RFI),
Telepass e i principali provider telematici (Viasat, Octo, Geotab).

```
   GPS / Telematics  ────────────┐
   Carrier webhooks ─────────────┼──> Ingest Gateway (Go)
   EDI partners    ──────────────┘         │
                                           │
                            ┌──────────────▼───────────────┐
                            │   Shipment Service (Go/Gin)  │
                            └──┬───────────┬────────────┬──┘
                               │           │            │
                         ┌─────▼───┐  ┌────▼───┐  ┌─────▼──────┐
                         │ MongoDB │  │ Redis  │  │   Kafka    │
                         │(shpmt,  │  │(cache, │  │(events,    │
                         │ routes) │  │  pub)  │  │ optional)  │
                         └─────────┘  └────────┘  └────────────┘
                               ▲
                               │ WebSocket / REST
                       ┌───────┴───────┐
                       │ Vue 3 SPA     │
                       │ Map Dashboard │
                       └───────────────┘
```

### 1.1 Componenti principali

- **Ingest Gateway** (parte del backend Go). Riceve webhook e chiamate
  REST, normalizza i payload dei provider telematici, arricchisce con
  metadati tenant/carrier e li pubblica come eventi `TrackingEvent`.
- **Shipment Service**. Aggregato di dominio che gestisce il ciclo di
  vita della spedizione (bozza → prenotata → in transito → consegnata),
  incluso il registro immutabile di catena di custodia.
- **Route Optimizer**. Proxy per un'istanza OSRM self-hosted; calcola
  rotte ottimali, ETA e geometrie poliline.
- **WebSocket Hub**. Fan-out in-process degli eventi pub/sub Redis
  verso i client del dashboard.
- **Frontend SPA**. Vue 3 con Composition API, Pinia, Vue Router 4 e
  Leaflet. Dashboard con mappa live, timeline eventi, filtri, export.

### 1.2 Principi architetturali applicati

1. **Twelve-Factor**: configurazione via env, processi stateless,
   logging strutturato su stdout, port binding esplicito, concurrency
   via processi (goroutine) e scaling orizzontale di default.
2. **Domain-Driven Design lite**: pacchetti `models`, `services`,
   `repository`, `handlers` con dipendenze a senso unico
   (handlers → services → repository → models).
3. **API-first**: lo schema REST/WebSocket è versionato in `/api/v1/`
   e documentato in `API.md` prima dell'implementazione.
4. **Security by design**: JWT HS256 con issuer pinnato, rate limiting
   per IP, CORS esplicito, Trivy in CI, distroless in produzione.
5. **Observability**: OpenTelemetry OTLP, log JSON zap, /metrics
   esposto in futuro via middleware Prometheus.

## 2. Modello dati (MongoDB)

Le collezioni primarie rispecchiano gli aggregati di dominio:

### 2.1 `shipments`

```json
{
  "_id": "5c3c...",
  "tenant_id": "ten_mozzecane_sml",
  "reference": "CMR-2026-04-0001",
  "carrier": "Autotrasporti Rossi Srl",
  "mode": "multimodal",
  "status": "in_transit",
  "consignor": {
    "name": "Officine Meccaniche Veronesi",
    "vat_number": "IT01234567890",
    "city": "Mozzecane",
    "province": "VR",
    "country": "IT"
  },
  "consignee": { "name": "Müller GmbH", "city": "München", "country": "DE" },
  "origin": { "type": "Point", "coordinates": [10.778, 45.333] },
  "destination": { "type": "Point", "coordinates": [11.5755, 48.1374] },
  "waypoints": [
    {
      "recorded_at": "2026-04-17T07:12:00Z",
      "position": { "type": "Point", "coordinates": [11.01, 45.93] },
      "speed_kph": 82.4,
      "source": "viasat"
    }
  ],
  "current_position": { "type": "Point", "coordinates": [11.01, 45.93] },
  "etd": "2026-04-17T05:00:00Z",
  "eta": "2026-04-17T13:30:00Z",
  "customs_status": { "declared": true, "document_type": "T1", "mrn": "26IT..." },
  "docs": [],
  "tags": ["quadrante-europa", "brennero"]
}
```

### 2.2 `tracking_events`

Eventi immutabili, uno per aggiornamento posizione o transizione di stato. Chiave (`shipment_id`, `sequence`).

### 2.3 `vehicles`

Anagrafica flotta (targa, categoria EURO, ADR, Telepass, provider telematico).

### 2.4 `drivers`

Anagrafica autisti (patente, CQC, iscrizione Albo Autotrasportatori, ore di guida secondo Reg. 561/2006).

### 2.5 `geofences`

Poligoni GeoJSON con tipo (warehouse, customs, port, depot, terminal, rest_area, loading_bay).

### 2.6 `chain_of_custody`

Append-only, ogni record firmato con hash del precedente (prev_hash → hash, SHA-256 del JSON canonico).

### 2.7 Indici principali

| Collezione | Indice | Scopo |
| --- | --- | --- |
| `shipments` | `{tenant_id:1, reference:1}` unique | deduplica reference per tenant |
| `shipments` | `{tenant_id:1, status:1, updated_at:-1}` | liste paginate dashboard |
| `shipments` | `{carrier:1, etd:1}` | reportistica per vettore |
| `shipments` | `{current_position: 2dsphere}` | query geospaziali (mappa live) |
| `geofences` | `{polygon: 2dsphere}` | $geoIntersects per dwell detection |
| `chain_of_custody` | `{shipment_id:1, sequence:1}` unique | append-only ordinato |

## 3. Diagrammi di sequenza

### 3.1 Ingestione webhook telematico

```
Carrier ──► POST /api/v1/shipments/{id}/waypoints ──► Gin handler
  Gin handler ──► JWT auth + rate limit
    ──► ShipmentService.RecordWaypoint
       ──► Mongo: $push waypoints, set current_position
       ──► Redis: SET logitrack:position:{id}
       ──► Mongo: insert tracking_events
       ──► Redis: PUBLISH logitrack:events:tracking
WebSocket hub ◄── Redis SUBSCRIBE
  hub ──► broadcast to matching subscribers
  subscriber ──► Vue SPA: update pin on map + append timeline
```

### 3.2 Subscription WebSocket

```
Browser ──► GET /api/v1/stream/tracking (Upgrade WS, Authorization: Bearer)
  Gin handler ──► JWT auth
    ──► upgrader.Upgrade
    ──► hub.Register(subscriber)
Browser ──► {op:"subscribe", shipmentId:"…"}
  handler ──► set subscriber.ShipmentID
Redis event arrives ──► hub.Broadcast(evt)
  matching subscribers ──► conn.Write {type:"event", event:{…}}
Browser periodic ──► {op:"ping"} ──► server {type:"pong"}
```

### 3.3 Append chain-of-custody

```
Operator ──► POST /api/v1/shipments/{id}/custody (handover)
  Gin handler ──► JWT auth, role=dispatcher|driver
    ──► ShipmentService.AppendCustody
       ──► Mongo: find latest sequence + hash
       ──► compute prev_hash = last.hash
       ──► compute hash = SHA256(canonical_json without hash field)
       ──► Mongo: insert chain_of_custody
    ──► return record
```

### 3.4 Dichiarazione dogana AIDA (outbound)

```
Operator ──► POST /api/v1/shipments/{id}/customs/declare
  handler ──► validate MRN / HS / country
    ──► CustomsService.Declare
       ──► AIDA client (SOAP/REST) with mTLS cert
       ──► persist MRN + document_type T1/T2
       ──► chain_of_custody append (action: "inspection")
  response: MRN, barcode, PDF URL
```

## 4. Integrazioni Italia

- **AIDA (Agenzia delle Dogane)** — dichiarazioni import/export, T1/T2, notifiche MRN.
- **RFI / FERTRAM** — prenotazioni slot ferroviari Quadrante Europa, interscambio con Verona QE.
- **Telepass / ViaCard** — dati di transito casello A22, fusione con waypoint telematici.
- **Albo Autotrasportatori (Ministero dei Trasporti)** — verifica iscrizione vettori.
- **Sistema di Interscambio (SDI)** — integrazione opzionale con FatturaPA per fatture di trasporto.

## 5. Architettura di deploy

### 5.1 Sviluppo (docker-compose)

| Servizio | Immagine | Porta | Volume |
| --- | --- | --- | --- |
| logitrack-backend | build ./backend | 8080 | — |
| logitrack-frontend | build ./frontend | 5173 | — |
| logitrack-mongodb | mongo:7 | 27017 | logitrack-mongo-data |
| logitrack-redis | redis:7-alpine | 6379 | — |

### 5.2 Produzione

- **Region**: AWS eu-south-1 (Milano) come primario, Aruba Cloud Arezzo
  come fallback per tenant a requisito sovrano PA.
- **Compute**: EKS cluster, 3 node group (system, stateful, general).
- **Storage**: MongoDB Atlas M30+ con encryption at rest, retention
  backup 7 giorni; snapshot giornaliero su S3 eu-south-1.
- **Networking**: VPC 10.0.0.0/16, 3 subnet AZ, NAT gateway per egress,
  ALB + AWS WAF managed rules + CloudFront davanti al frontend.
- **Secrets**: AWS Secrets Manager, rotate JWT_SECRET ogni 90 giorni.

## 6. Modalità di fallimento e mitigazione

| Rischio | Impatto | Mitigazione |
| --- | --- | --- |
| MongoDB primary down | ingestione interrotta | Replica set 3 nodi, election automatica, buffer Redis fino a 5 min |
| Redis down | WebSocket fan-out interrotto | Circuit breaker su publish, riconnessione client, replay da Mongo |
| OSRM self-host down | ETA non aggiornati | fallback su haversine + storico medie A22 |
| AIDA irraggiungibile | dichiarazioni bloccate | retry esponenziale 24h + coda manuale + alert |
| Kafka (futuro) lag | eventi in ritardo ma non persi | alert su consumer lag > 30s |

## 7. SLO

| Metrica | Target | Misurazione |
| --- | --- | --- |
| Disponibilità API | 99.9% mensile | probes `/api/health` ogni 30s |
| Latenza P95 REST | < 300 ms | OpenTelemetry tracing |
| Latenza WebSocket | < 2 s end-to-end | synthetic agent |
| Ingest waypoint throughput | ≥ 500 eventi/s | load test k6 |

## 8. Mission II additions (v0.2.0)

- **`cmd/simulator`** — telematics emitter. Publishes waypoints at
  1 Hz for the 3 demo shipments (Verona→Milano A4, Verona→Napoli A1,
  Verona→München A22/Brennero). See `docs/DEMO-SCRIPT.md`.
- **`internal/services/eta_service.go`** — moving-average smoother over
  the latest 20 speed samples (bounded 5–120 km/h). Exposed via
  `GET /api/v1/shipments/:id/eta`.
- **`internal/services/route_optimizer.go`** — LRU cache (1000
  entries), domain allow-list on outbound HTTP, straight-line fallback.
- **`internal/obs/` + `internal/middleware/prom_mw.go` +
  `internal/handlers/metrics.go`** — hand-rolled Prometheus exposition
  broken out of the handlers package to avoid a middleware ↔ handlers
  import cycle.
- **`internal/demo/`** — deterministic routes and seeding. Idempotent
  seed on boot when `SEED_DEMO=true`.
- **`internal/models/compliance.go`** — Italian plate validator
  (post-1994 + historical), ADR/ATP class enums, Telepass toll-code
  record.
- **`frontend/src/components/ShipmentMap.vue`** — live Leaflet map
  consuming the WS stream, decoding Polyline6 and placing a moving
  marker.

### Data-flow: golden path

```
  simulator → POST /api/v1/shipments/{id}/waypoints
           → ShipmentService.RecordWaypoint
              ├─ Mongo.AppendWaypoint
              ├─ Redis.CacheLatestPosition
              ├─ ETAService.UpdateSpeed (moving avg)
              └─ Redis.PublishTrackingEvent
                    → WebSocketHub.Run
                         → StreamHandler (gorilla/websocket)
                             → browser ShipmentMap.vue
                                 ↳ marker updates
                                 ↳ GET /api/v1/shipments/{id}/eta
                                        → OSRM (cached) or fallback
```

### Module dependency sketch

```
models → (standalone)
obs    → (standalone)
repository → models
services → models, repository
handlers → services, middleware, config
middleware → config, obs
cmd/server → handlers, repository, services, demo, config
cmd/simulator → demo  (only; no repository/handler dep)
```
