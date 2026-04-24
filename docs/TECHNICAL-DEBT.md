# LogiTrack — Technical Debt Ledger

Each item lists the debt, rationale for accepting it today, and the
trigger that should force a refactor.

## 1. JWT secret is a single HS256 key
- **State:** HS256 with a single symmetric secret.
- **Trigger:** second tenant requiring key rotation on their own
  schedule → migrate to RS256 with per-tenant key pairs.

## 2. No Redis-backed distributed rate limiter
- **State:** in-process token bucket (`middleware/rate_limit.go`).
- **Trigger:** more than 1 backend replica behind a single public
  endpoint → switch to `redis_rate` package.

## 3. WebSocket hub is single-process
- **State:** Redis pub/sub fan-out into an in-process map.
- **Trigger:** > 2 replicas serving WS clients → migrate to a Kafka
  consumer group with per-replica partitioning (already stubbed in
  `KafkaConfig`).

## 4. Prometheus exposition is hand-rolled
- **State:** `internal/obs/metrics.go` uses `fmt.Fprintf` to emit
  the text format.
- **Trigger:** need for histograms, summaries, or multi-service
  scraping at scale → adopt `prometheus/client_golang`.

## 5. No per-tenant quota enforcement at the edge
- **State:** tier limits documented but not enforced in code for
  "shipments per day".
- **Trigger:** first abusive tenant → add a daily-counter check in
  `ShipmentService.CreateShipment`.

## 6. AIDA client is a stub
- **State:** envelope documented, client not shipped.
- **Trigger:** first customer production deployment with AIDA
  accreditation → ship the accredited client under an opt-in feature
  flag.

## 7. Digital tachograph ingestion not implemented
- **State:** LogiTrack relies on the operator's tachograph vendor to
  produce daily summaries and POSTs them via webhook.
- **Trigger:** regulatory change forcing carriers to submit raw DDD
  to a centralised authority → build the parser against the
  smart-tachograph v2+ spec.

## 8. No horizontal pagination for shipments
- **State:** offset/limit pagination.
- **Trigger:** tenants with > 100 k open shipments → keyset
  pagination on `(updated_at, _id)`.

## 9. Leaflet tiles from public OSM in dev
- **State:** dev defaults to `tile.openstreetmap.org`.
- **Trigger:** first paying customer in SaaS → ship MapLibre with a
  self-hosted MapTiler schema, EU-resident.

## 10. No OpenAPI schema generated from code
- **State:** `docs/API.md` is the source of truth.
- **Trigger:** second client language (Python, Java) → generate
  OpenAPI via `swaggo/swag`.
