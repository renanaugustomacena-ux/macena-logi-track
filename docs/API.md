# LogiTrack — API Reference

Base URL: per-deployment (e.g. `https://api.<customer>.example/api/v1`)
Development: `http://localhost:8080/api/v1`

All endpoints (except `/api/health`, `/api/ready`, `/metrics`,
`/api/v1/auth/login`) require a Bearer JWT issued by the LogiTrack
auth endpoint or by the customer's IDP. Tokens are HS256-signed,
carry the tenant identifier under `tenantId`, and expire after the
`JWT_ACCESS_TTL` configured (default 15 minutes). The kit issues
**access tokens only** — there is no refresh endpoint.

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6Ikp...
```

Content type defaults to `application/json; charset=utf-8`. Dates are
ISO-8601 with a trailing `Z` for UTC. Coordinates follow GeoJSON
convention — `[longitude, latitude]`.

Error envelopes are uniform:

```json
{ "error": "not_found", "detail": "optional human-friendly message" }
```

## Table of contents

1. [Health + Ready + Metrics](#1-health)
2. [Auth](#2-auth)
3. [Shipments](#3-shipments)
4. [Tracking](#4-tracking)
5. [Routes](#5-routes)
6. [WebSocket stream](#6-websocket-stream)
7. [Fleet (vehicles, drivers, geofences)](#7-fleet)
8. [Rifiuti (anagrafiche + FIR + RENTRI)](#8-rifiuti)
9. [Observability](#9-observability)
10. [Error codes](#10-error-codes)

---

## 1. Health

### `GET /api/health`

Public. Returns liveness plus dependency status. Used by Kubernetes
liveness probes and by the landing page.

**Response 200**

```json
{
  "status": "ok",
  "service": "logitrack",
  "version": "0.1.0",
  "uptime_seconds": 321,
  "time": "2026-04-17T10:15:02.123456Z",
  "dependencies": { "mongodb": "ok", "redis": "ok" }
}
```

`status` is `ok` if every dependency reports `ok`, `degraded` otherwise.

### `GET /api/ready`

Public. Readiness probe (seed-aware). Returns 200 with
`{"status":"ready"}` once dependencies are connected and demo seed
(if enabled) has finished. Used by orchestrator readiness probes.

---

## 2. Auth

### `POST /api/v1/auth/login`

Public. Issue an access token.

**Request**
```json
{ "username": "operator@customer.example", "password": "<password>" }
```

**Response 200**
```json
{
  "tokenType": "Bearer",
  "accessToken": "eyJhbGciOiJIUzI1NiIs…",
  "expiresIn": 900,
  "issuedAt": 1745835600
}
```

NIST SP 800-63B-style lockout: 5 failed attempts → 15-minute window
where the username returns `invalid_credentials` regardless of the
password supplied. Successful login resets the counter.

The kit issues **access tokens only**. There is no `/refresh`
endpoint. On expiry the client re-authenticates. Customers needing
sliding sessions integrate their corporate IDP at the IdentityStore
seam (`backend/internal/handlers/auth.go`).

When `LOGITRACK_IDENTITY_BACKEND=disabled` (or the seed user/password
are empty), this endpoint returns 503 `identity_backend_disabled`.

---

## 3. Shipments

### `POST /api/v1/shipments`

Create a new shipment. The caller's tenant id is injected from the JWT;
any `tenantId` sent in the body is ignored.

**Request**

```json
{
  "reference": "CMR-2026-04-0001",
  "carrier": "Autotrasporti Rossi Srl",
  "mode": "multimodal",
  "consignor": {
    "name": "Officine Meccaniche Veronesi",
    "vatNumber": "IT01234567890",
    "city": "Mozzecane",
    "province": "VR",
    "country": "IT"
  },
  "consignee": { "name": "Müller GmbH", "city": "München", "country": "DE" },
  "origin": { "type": "Point", "coordinates": [10.778, 45.333] },
  "destination": { "type": "Point", "coordinates": [11.5755, 48.1374] },
  "etd": "2026-04-17T05:00:00Z",
  "eta": "2026-04-17T13:30:00Z",
  "tags": ["quadrante-europa", "brennero"]
}
```

**Response 201** — the persisted shipment with `id`, `createdAt`, `updatedAt` populated.

### `GET /api/v1/shipments`

List shipments belonging to the caller tenant. Supported query
parameters:

| Param | Type | Notes |
| --- | --- | --- |
| `status` | string | One of `draft`, `booked`, `picked_up`, `in_transit`, `delayed`, `at_customs`, `delivered`, `cancelled` |
| `carrier` | string | Exact match |
| `from` | RFC3339 date-time | Lower bound on `etd` |
| `to` | RFC3339 date-time | Upper bound on `etd` |
| `limit` | int | Page size, max 500, default 100 |
| `offset` | int | Offset into the sorted list (most recent first) |

**Response 200**

```json
{
  "items": [ /* Shipment objects */ ],
  "limit": 100,
  "offset": 0
}
```

### `GET /api/v1/shipments/{id}`

Fetch a single shipment. Returns 404 when the id does not exist within
the caller tenant (no cross-tenant leak).

### `POST /api/v1/shipments/{id}/waypoints`

Append a telematics waypoint. The request body is a single
`Waypoint`:

```json
{
  "recordedAt": "2026-04-17T07:12:00Z",
  "position": { "type": "Point", "coordinates": [11.01, 45.93] },
  "speedKph": 82.4,
  "headingDeg": 40.2,
  "source": "viasat",
  "rawEventId": "v-98f1..."
}
```

**Response 202** — `{ "accepted": true }`

Side effects: appends the waypoint, updates `currentPosition`, caches
the latest position in Redis, inserts a `position_update` tracking event
and publishes on the Redis tracking channel.

### `GET /api/v1/shipments/{id}/trace`

Returns the full chain-of-custody log. The array is already ordered
by `sequence` ascending.

**Response 200**

```json
{
  "shipmentId": "5c3c...",
  "records": [
    {
      "id": "…",
      "sequence": 1,
      "action": "created",
      "occurredAt": "2026-04-17T05:02:00Z",
      "actor": { "name": "system", "role": "creator", "organisation": "Autotrasporti Rossi Srl" },
      "prevHash": "",
      "hash": "4f7b..."
    }
  ]
}
```

### `GET /api/v1/shipments/{id}/position`

Returns the latest cached waypoint from Redis, falling back to MongoDB
when the cache has expired. Returns 404 when there is no position yet.

### `GET /api/v1/shipments/{id}/eta`

Returns the predicted ETA for the shipment. Combines the current
position, the remaining planned polyline (from the OSRM cache or the
straight-line fallback) and a moving-average smoother over the most
recent 20 speed samples.

**Response 200**

```json
{
  "shipmentId": "5c3c...",
  "remainingMeters": 184200,
  "smoothedSpeedKph": 72.3,
  "eta": "2026-04-17T13:34:12Z",
  "source": "osrm",
  "geometry": "q~~yF}jiZ..."
}
```

- `source` is one of `osrm`, `cache`, `fallback`, `stub`.
- When no waypoint has been ingested yet, the endpoint returns **202
  Accepted** with `{"status":"awaiting_first_waypoint"}`.

### ETA algorithm

Let `d_m` be the remaining distance in metres (from the route optimiser)
and `v_kph` the smoothed speed in km/h:

```
v_kph = clip(mean(last 20 valid samples), 5 kph, 120 kph)
eta   = now + (d_m / 1000) / v_kph  hours
```

A sample is "valid" if it is either the explicit `speedKph` of the
telematics payload or a derived speed from `Δdistance / Δtime` when
the interval > 0. See `internal/services/eta_service.go`.

---

## 4. Tracking

`TrackingEvent` is emitted on every significant state transition. The
full taxonomy is documented inline in the Go type
`models.TrackingEventType` and kept stable across versions.

| Type | Emitted when |
| --- | --- |
| `position_update` | a waypoint is recorded |
| `departure` | shipment enters `in_transit` |
| `arrival` | vehicle reaches destination geofence |
| `geofence_enter` / `geofence_exit` | polygon traversal detected |
| `delay_detected` | ETA threshold exceeded |
| `customs_hold` / `customs_cleared` | AIDA workflow events |
| `seal_broken`, `temperature_alarm` | device-level alarms |
| `handover`, `delivery` | custody transitions |

---

## 5. Routes

### `POST /api/v1/routes/optimize`

Thin proxy in front of an OSRM instance (self-hosted in production).
The optimiser caches responses in a 1000-entry LRU keyed by the
(vehicle, tolls, coords) tuple; subsequent identical calls return
the cached result with `source: "cache"`.

**SSRF envelope.** The base URL must resolve to a host on
`OSRM_ALLOWED_HOSTS`. Unknown hosts produce `503 optimise_failed` with
`detail: "host not in allow-list"` — no outbound request is made.

**Fallback envelope.** If OSRM returns non-200 or times out, the
optimiser replies with a straight-line great-circle estimate at
`70 km/h` average speed and `source: "fallback"`. The geometry is
empty so the frontend knows to draw a dashed line between waypoints.

**Request**

```json
{
  "waypoints": [
    { "type": "Point", "coordinates": [10.778, 45.333] },
    { "type": "Point", "coordinates": [11.5755, 48.1374] }
  ],
  "vehicle": "truck",
  "avoidTolls": false
}
```

**Response 200**

```json
{
  "distanceMeters": 412987,
  "durationSeconds": 24300,
  "geometry": "q~~yF}jiZ…",
  "legs": [
    { "distanceMeters": 412987, "durationSeconds": 24300 }
  ]
}
```

Errors surface as 502 `optimise_failed` when the upstream refuses the
request or returns a non-OK code.

---

## 6. WebSocket stream

### `GET /api/v1/stream/tracking` (upgrade)

The WebSocket endpoint delivers live tracking events to authenticated
clients. The JWT can be supplied in three places (in this priority
order):

1. `Authorization: Bearer <jwt>` header (servers and CLIs).
2. `?access_token=<jwt>` query parameter (legacy browsers).
3. `Sec-WebSocket-Protocol: logitrack.jwt.v1,<jwt>` (modern browsers).
   The server echoes back `logitrack.jwt.v1` to complete the
   handshake.

Handshake controls:
- **Origin check**: the `Origin` header must match
  `HTTP_WS_ORIGINS` (configurable) when present. Missing `Origin`
  headers (non-browser clients) are permitted because the JWT still
  binds the caller to a tenant.
- **Per-IP handshake rate limit**: default 20 rps burst 40.
- **JWT**: HS256 only; alg:none, HMAC-family hopping and other
  bypasses are rejected.

Runtime controls:
- **Per-connection inbound rate limit**: 20 rps burst 40.
- **Read-size cap**: 64 KiB.
- **Idle disconnect**: 5 minutes (configurable via `WS_IDLE_TIMEOUT`).
- **Keepalive**: the server sends a WS ping every 30 s; clients may
  also issue `{"op":"ping"}` to keep the connection alive.

### 5.1 Messages · client → server

```json
{ "op": "subscribe", "shipmentId": "5c3c...", "carriers": ["Rossi Srl"] }
{ "op": "ping" }
```

- `shipmentId` is optional. When empty the client receives every event
  within its tenant.
- `carriers` is an optional allow-list applied server-side.

### 5.2 Messages · server → client

```json
{ "type": "event", "event": { /* TrackingEvent */ } }
{ "type": "pong" }
{ "type": "error", "code": "unauthorized" }
```

### 5.3 Lifecycle

- Idle timeout: 5 minutes. The server sends WS ping frames every 30 s;
  clients' pong handlers reset the server-side read deadline.
- Slow-consumer policy: if a client's outbound queue is full, the
  server drops the message and logs `ws slow consumer drop`. The
  client is expected to reconnect and, if needed, replay recent events
  via `GET /api/v1/shipments/{id}`.
- Close codes follow RFC 6455; `4401` means authentication invalid,
  `4429` means rate limited.

### 5.4 Prometheus metrics

- `logitrack_ws_connections_active` — gauge of active subscribers.
- `logitrack_ws_broadcasts_total` — counter of messages fanned out.

## 7. Fleet

Master-data endpoints for vehicles, drivers and geofences. All
authenticated, all tenant-scoped.

### Vehicles

| Method | Path | Notes |
| --- | --- | --- |
| POST | `/api/v1/vehicles` | Create. Body: `Vehicle` (plate, EURO class, ADR/ATP markers, telematics provider). |
| GET | `/api/v1/vehicles` | List for tenant. Query: `limit`, `offset`. |
| GET | `/api/v1/vehicles/{id}` | Retrieve. |

### Drivers

| Method | Path | Notes |
| --- | --- | --- |
| POST | `/api/v1/drivers` | Create. Body: `Driver` (name, patente, CQC, Albo flag). |
| GET | `/api/v1/drivers` | List for tenant. |
| GET | `/api/v1/drivers/{id}` | Retrieve. |

### Geofences

| Method | Path | Notes |
| --- | --- | --- |
| POST | `/api/v1/geofences` | Create. Body: `Geofence` with GeoJSON `polygon` and `type` (warehouse, customs, port, depot, terminal, rest_area, loading_bay). |
| GET | `/api/v1/geofences` | List for tenant. |
| GET | `/api/v1/geofences/{id}` | Retrieve. |

---

## 8. Rifiuti

Module 2: trasporto rifiuti speciali, orientato al flusso RENTRI
(coda di trasmissione oggi, adapter live per ingaggio). Regulatory
anchors in [`MODULE-RIFIUTI.md`](MODULE-RIFIUTI.md).

### Anagrafiche

| Method | Path | Notes |
| --- | --- | --- |
| POST | `/api/v1/rifiuti/produttori` | Create produttore. `ragione_sociale` and `codice_fiscale` required. |
| GET | `/api/v1/rifiuti/produttori` | List for tenant. |
| POST | `/api/v1/rifiuti/trasportatori` | Create trasportatore. Validates `albo_categoria` (1, 2-bis, 4, 5, 6, 8, 9, 10). |
| GET | `/api/v1/rifiuti/trasportatori` | List for tenant. |
| POST | `/api/v1/rifiuti/destinatari` | Create destinatario (impianto). |
| GET | `/api/v1/rifiuti/destinatari` | List for tenant. |

### FIR

| Method | Path | Notes |
| --- | --- | --- |
| POST | `/api/v1/rifiuti/fir` | Create. Validates FIR shape, runs `Trasportatore.CanCarry(cer, now)` and `Destinatario.CanReceive(cer, op, now)` cross-checks. Persists in `FIRDraft` state. |
| GET | `/api/v1/rifiuti/fir` | List for tenant. Query: `state` to filter by FIR state. |
| GET | `/api/v1/rifiuti/fir/{id}` | Retrieve. |
| POST | `/api/v1/rifiuti/fir/{id}/transition` | Body `{"to":"<FIRState>"}`. Guards every legal edge in `firTransitions`; stamps signature timestamps for the relevant role. |
| POST | `/api/v1/rifiuti/fir/{id}/vidima` | Submit FIR to RENTRI. Idempotency key derived from `(tenantID, firID)`. State must be `FIRDraft` on entry; advances to `FIRVidimato` on success. Returns `numero_rentri`, `vidimato_at`, `qr_code_payload`. |

### CER inline validator

| Method | Path | Notes |
| --- | --- | --- |
| GET | `/api/v1/rifiuti/cer/{code}` | Returns `{ valid, normalised, pericoloso, chapter }` or `{ valid:false, error }`. Useful for the operator UI to validate as the user types. |

---

## 9. Observability

### `GET /metrics`

Public on the container network; in production the endpoint is
accessible only from the service-mesh monitoring namespace. Returns
Prometheus text exposition version 0.0.4.

Minimum exposed series:
- `logitrack_build_info{service,version,go_version}` — gauge = 1.
- `logitrack_http_requests_total{method,path,status}` — counter.
- `logitrack_ws_connections_active` — gauge.
- `logitrack_ws_broadcasts_total` — counter.
- `process_start_time_seconds` — gauge.
- `go_goroutines` — gauge.

---

## 10. Error codes

| Code | Meaning |
| --- | --- |
| `missing_authorization` | Authorization header absent |
| `invalid_authorization_scheme` | Header present but not `Bearer …` |
| `invalid_token` | JWT signature / issuer / expiry invalid |
| `invalid_claims` | Token parsed but required claim missing |
| `insufficient_role` | Authenticated but lacking required role |
| `invalid_credentials` | Login: wrong username/password or locked out |
| `identity_backend_disabled` | Login: backend disabled (set `LOGITRACK_IDENTITY_BACKEND=memory` + seed user/password) |
| `invalid_body` | JSON decoding failed |
| `missing_fields` | Required field absent in body |
| `invalid_albo_categoria` | Rifiuti: trasportatore Albo categoria not in {1, 2-bis, 4, 5, 6, 8, 9, 10} |
| `fir_invalid` | Rifiuti: FIR validation failed |
| `fir_not_draft` | Rifiuti: vidimazione attempted on a FIR not in `FIRDraft` state |
| `invalid_transition` | Rifiuti: FIR state transition rejected by state machine |
| `trasportatore_not_found` / `destinatario_not_found` / `produttore_not_found` | Rifiuti: referenced anagrafica missing |
| `trasportatore_cannot_carry` | Rifiuti: Albo categoria/expiry rejects this CER |
| `destinatario_cannot_receive` | Rifiuti: impianto authorisation rejects this CER + operazione |
| `rentri_failed` | Rifiuti: RENTRI client returned an error |
| `xfir_marshal_failed` | Rifiuti: XML encoding of placeholder xFIR failed |
| `cannot_create` | Domain validation rejected (shipment / fleet) |
| `cannot_record` | Waypoint rejected (bad geo-point, unknown shipment) |
| `not_found` | Resource not found in tenant |
| `list_failed`, `lookup_failed`, `trace_failed`, `update_failed`, `create_failed` | Internal DB error |
| `optimise_failed` | OSRM upstream error |
| `rate_limited` | Per-IP bucket exceeded |
| `handshake_rate_limited` | WS handshake bucket exceeded |
| `unauthorized` (WS) | Token rejected during upgrade |
