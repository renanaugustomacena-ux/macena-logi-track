# LogiTrack — Integrations

Honesty: the kit ships only the integrations actually wired in the
codebase today. The "Italian regulatory" integrations (AIDA, FERTRAM,
Telepass, Albo Autotrasportatori) are real things real customers ask
for, but they are deliberately NOT in the kit core: they are
per-customer adapters added during the engagement when the customer
has the credentials, the contract, and the budget for them.

If you are forking the kit, expect to write the adapter for the
provider your customer actually uses, not to plug in a pre-built
generic one.

## 1. Telematics ingestion (built-in)

Single generic ingest endpoint, provider-agnostic:

```
POST /api/v1/shipments/{id}/waypoints
Authorization: Bearer <jwt>
Content-Type: application/json

{
  "recordedAt": "2026-04-17T07:12:00Z",
  "position": { "type": "Point", "coordinates": [11.01, 45.93] },
  "speedKph": 82.4,
  "headingDeg": 40.2,
  "source": "viasat",
  "rawEventId": "v-98f1…"
}
```

- Authentication: per-tenant JWT (HS256, issued by the kit's `/auth/login`
  or by the customer's IDP at the IdentityStore seam).
- Idempotency: providers optionally set `rawEventId`; duplicates
  deduplicated downstream by the customer adapter when needed.
- Side effects: appends waypoint, updates `current_position`, caches
  in Redis, inserts a `position_update` tracking event, publishes on
  the Redis tracking channel for WebSocket fan-out.

Provider-specific adapters (Viasat, Octo, Geotab, etc.) translate the
provider's webhook into this shape. Each adapter is small (a few
hundred lines of Go) and lives outside the kit core, in a per-customer
overlay package.

## 2. RENTRI (built-in: stub default, live adapter pending)

The rifiuti module ships a `rentri.Client` interface and a
`rentri.QueuedStub` default adapter. The stub:

- Accepts `VidimazioneRequest` calls, returns a deterministic
  `NumeroRENTRI` keyed on the idempotency key.
- Persists submissions in an in-memory FIFO queue, race-tested under
  16-worker contention.
- Is the right answer for development, customer demos, and the
  engagement period before the customer's RENTRI certificato digitale
  arrives.

The live HTTP adapter (sandbox `https://demoapi.rentri.gov.it`,
production `https://api.rentri.gov.it`) materialises in
[`backend/internal/modules/rifiuti/rentri/`](../backend/internal/modules/rifiuti/rentri/)
when the customer provides the SPID/CIE/CNS-bound Entratel delegation
that unlocks the API certificate. Cutover is a one-file change at the
composition root in `cmd/server/main.go`:

```go
deps.Rifiuto = handlers.NewRifiutoHandler(mongoRepo, rentri.NewQueuedStub())
//                                                  ^ swap with live HTTP adapter
```

Sandbox base URL + production base URL are constants in
`rentri/endpoints.go`. xFIR payload encoding is currently a
`encoding/xml`-escaped placeholder; the full RENTRI v1.0 XSD encoder
lands together with the live adapter.

## 3. OSRM route optimisation (opt-in)

Set `OSRM_BASE_URL` and `OSRM_ALLOWED_HOSTS` to point at a
self-hosted OSRM instance with a custom truck profile (`OSRM_TRUCK_PROFILE`).
When unconfigured (kit default) the optimiser silently falls back to
a great-circle + 70 km/h estimate after a one-shot WARN.

SSRF guard: the BaseURL host is checked against the allow-list before
any DNS or socket. The HTTP client refuses to follow redirects, so a
compromised or MITM'd OSRM cannot escape the allow-list.

Never point at the public `https://router.project-osrm.org` in
production: US-hosted, no SLA, no truck profile, EU customer route
data leaving the EU (GDPR Art. 44).

## 4. Frontend (built-in)

The Vue 3 SPA in `frontend/` consumes only the kit's own REST + WebSocket
surface. There is no third-party widget, embed or analytics tag in the
default kit. Customers who want analytics (Plausible, Matomo) add it
during the engagement.

## 5. Prometheus / Grafana (opt-in)

`/metrics` exposes Prometheus exposition v0.0.4 unconditionally. In a
service-mesh deployment a sidecar fronts the endpoint and restricts
access to the monitoring namespace. In a non-mesh deployment you must
restrict the endpoint at the reverse proxy layer (e.g. `deny` outside
the customer's Prometheus scrape source IP).

A starter Grafana dashboard JSON is not currently in the repo; one
will be added when the first customer asks for it.

## 6. What is NOT in the kit and why

| Provider | Why not in the kit |
| --- | --- |
| AIDA (Agenzia delle Dogane) | Per-customer accreditation (PKCS#12). Customer-specific. Adapter built during engagement. |
| FERTRAM / RFI | mTLS certificate per operator. Customer-specific. Adapter built during engagement. |
| Telepass / ViaCard | Read-only monthly CSV; customer's own contract; trivial CSV importer per fork. |
| Albo Autotrasportatori | Public registry, no machine-readable API as of 2026-04-28. Manual verification per fork. |
| FatturaPA / SDI | Out of scope for LogiTrack. Customer keeps their existing fiscal gestionale; LogiTrack writes shipment metadata to a CSV export it consumes. |
| Conservazione AgID | Customer chooses an accreditato (https://www.agid.gov.it/it/piattaforme/conservazione). Per-deployment, not per-kit. |

A previous version of this kit shipped half-finished AIDA / FERTRAM /
Telepass / Albo client packages with `_ = deps.AidaClient`-style dead
references. They have been removed from the codebase as part of the
2026-04-28 honesty pass: the kit no longer claims to integrate
something it does not actually integrate.
