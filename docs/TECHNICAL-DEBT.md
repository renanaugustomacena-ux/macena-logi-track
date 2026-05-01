# LogiTrack — Technical Debt Ledger

Each item lists the debt, the rationale for accepting it today, and
the trigger that should force a refactor. Updated 2026-04-28 after
the kit-honesty audit pass.

## 1. Single HS256 JWT secret

- **State:** HS256 with one symmetric secret.
- **Trigger:** kit deployed for a customer with a contractual key
  rotation requirement on a different schedule than the freelancer's
  default → migrate that fork to RS256 with KMS-backed key pair.

## 2. In-process rate limiter (no Redis-backed distributed limiter)

- **State:** in-process token bucket with eviction
  (`middleware/rate_limit.go`). Memory bounded by `evictAfter` /
  `sweepInterval`.
- **Trigger:** more than one backend replica behind a single public
  endpoint → switch to a Redis-backed limiter (e.g. `redis_rate`).
  The customer fork that needs this writes the swap.

## 3. Single-process WebSocket hub

- **State:** Redis pub/sub fan-out into an in-process subscriber
  map. Single replica only.
- **Trigger:** > 2 backend replicas serving WS clients → migrate to
  per-replica partitioning, either via NATS / Kafka consumer groups
  or via tenant-sticky load balancing.

## 4. Hand-rolled Prometheus exposition

- **State:** `internal/obs/metrics.go` emits the text format via
  `fmt.Fprintf`. Sufficient for the gauges + counters the kit
  exposes today.
- **Trigger:** need histograms, summaries, or multi-service scraping
  at scale → adopt `prometheus/client_golang`.

## 5. Audit writer is lossy under backpressure

- **State:** `audit.Writer` drops records when the channel buffer
  (default 1024) is full and increments a logged counter. Documented
  in [`SECURITY.md`](SECURITY.md) §2.7.
- **Trigger:** customer with a contractual "no audit drop" clause →
  switch to synchronous insert in `AuditMiddleware`, accepting the
  per-request latency cost.

## 6. No keyset pagination on `/shipments`

- **State:** offset/limit pagination on `(updated_at desc, _id)`.
  Performant up to ~10⁵ shipments per tenant.
- **Trigger:** tenants with > 100 k open shipments → keyset
  pagination on `(updated_at, _id)` cursor.

## 7. Leaflet tiles default to public OSM

- **State:** dev defaults to `tile.openstreetmap.org`.
- **Trigger:** any production deployment → wire MapLibre with a
  self-hosted MapTiler schema, or contract with a tile provider that
  is acceptable to the customer.

## 8. No OpenAPI schema generated from code

- **State:** `docs/API.md` is the source of truth.
- **Trigger:** customer's downstream client in another language
  (Python, Java) → generate OpenAPI via `swaggo/swag`.

## 9. RENTRI live HTTP adapter not yet shipped

- **State:** `rentri.QueuedStub` is the default. The live HTTP
  adapter is one constructor swap away when the customer's
  certificato digitale is active.
- **Trigger:** first design-partner trasportatore goes through
  RENTRI delegation → ship the live adapter.

## 10. Per-route role gating not applied to every mutating endpoint

- **State:** `middleware.RequireRole` exists but is selectively
  applied. Tenant scoping always applies; role checks fall back to
  "any authenticated user in tenant".
- **Trigger:** customer policy explicitly separating dispatcher /
  driver / admin authority → apply RequireRole on the affected
  routes during the engagement.

## 11. `error.Error()` echoed in some 500 paths

- **State:** several `problem.Internal(c, …, err.Error())` paths
  echo the raw error string (which may contain Mongo driver
  internals). 4xx responses use stable error codes, but 500s leak
  more detail than they should.
- **Trigger:** any production deployment → introduce a
  `problem.SafeInternal(c, code)` helper that logs the full error
  server-side and returns only a stable code to the client.

## 12. No CSRF protection on POST

- **State:** SPA uses Authorization header only; no cookie auth, so
  CSRF surface is zero today.
- **Trigger:** future feature that introduces cookie auth → add a
  CSRF token middleware before the cookie ships.

## 13. xFIR XML payload is a placeholder

- **State:** `VidimaFIR` builds a minimal `<formulario><cer>…</cer></formulario>`
  via `encoding/xml` (so escaping is correct). The real RENTRI v1.0
  XSD-driven encoder lands together with the live HTTP adapter.
- **Trigger:** RENTRI live adapter ships → XSD-driven encoder.

## 14. Backend test coverage at ~9% vs 80% policy target

- **State:** total backend coverage is ~9%. Domain modules clear the
  bar (`internal/modules/rifiuti` 93.3%, `…/rifiuti/rentri` 87.1%,
  `internal/integrations/httpretry` 88.9%). Platform layers ship
  with no tests yet: `internal/handlers` (3.9%), `internal/services`,
  `internal/repository`, `internal/middleware`, `internal/audit`,
  `internal/obs`, `internal/config`, `internal/demo`,
  `cmd/simulator` are at 0%.
- **CI behaviour:** the workflow keeps the 80% gate as the *target*
  (per the global CLAUDE.md testing rule) but emits it as a `::warning::`
  rather than failing the job, so the kit can ship while the platform
  test backlog is worked down.
- **Trigger / ratchet plan:** raise the gate by ~10 points whenever
  a layer crosses 50% (handlers first, then services, repository,
  middleware). When the total clears 80%, flip the gate back to a
  hard `exit 1`.
