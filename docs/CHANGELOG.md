# Changelog

LogiTrack follows semantic versioning (`MAJOR.MINOR.PATCH`). Each
release pairs a git tag with a line in this file.

## 0.2.0 — 2026-04-17 — Consolidation, Mission II

### Added
- `cmd/simulator/main.go` — GPS waypoint publisher for the three
  demo shipments (Verona→Milano via A4, Verona→Napoli via A1,
  Verona→München via A22/Brennero) at 1 Hz.
- `internal/services/eta_service.go` — ETA computation with
  moving-average smoothing (20-sample window, 5–120 km/h bounds).
- `internal/services/route_optimizer.go` — in-memory LRU cache
  (1000 entries), straight-line fallback at 70 km/h, OSRM
  domain allow-list (SSRF mitigation).
- `internal/handlers/stream.go` — WebSocket hardening: origin
  check, JWT via Authorization / query / subprotocol, per-IP
  handshake rate limit, per-connection inbound rate limit,
  5-minute idle disconnect.
- `internal/handlers/metrics.go` + `internal/obs/` — Prometheus
  exposition at `/metrics` with build info and HTTP counters.
- `internal/demo/` — three deterministic demo shipments +
  seed-on-boot helper (`SEED_DEMO=true`).
- `internal/models/compliance.go` — Italian plate validator
  (post-1994 + historical), ADR/ATP enums, Telepass toll-code
  shape.
- `frontend/src/components/ShipmentMap.vue` — live Leaflet map
  with WebSocket subscription, planned polyline decoding, moving
  marker, accessible ETA announcement.
- Documentation: `ITALIAN-COMPLIANCE.md`, `SECURITY.md`, `RUNBOOK.md`,
  `SLO.md`, `A11Y.md`, `CHANGELOG.md`, `CONTRIBUTING.md`, `PITCH.md`,
  `DEMO-SCRIPT.md`, `PRICING.md`, `COMPLIANCE.md`,
  `SECURITY-SELF-ASSESSMENT.md`, `MIGRATION-FROM-LEGACY.md`,
  `OPERATIONS-CADENCE.md`, `INTEGRATIONS.md`, `DATA-RESIDENCY.md`,
  `TECHNICAL-DEBT.md`, `RISK-ACCEPTANCES.md`.

### Changed
- `docker-compose.yml` — frontend on :5174, Redis host-mapped 6380.
- `frontend/vite.config.ts`, `frontend/nginx.conf`,
  `frontend/Dockerfile` — port 5174, CSP header, non-root runtime.
- `backend/go.mod` — bumped `golang-jwt/jwt/v5` → v5.2.2,
  `go-redis/v9` → v9.6.3, `golang.org/x/net` → v0.52.0,
  `google.golang.org/grpc` → v1.80.0,
  `go.opentelemetry.io/otel/sdk` → v1.43.0. govulncheck clean.
- `Shipment` — new fields `adrClass`, `atpClass`, `telepassCodes`,
  `routePolyline`.
- `ShipmentService` — integrates RouteOptimizer (pre-computed
  polyline on create) and ETAService (waypoint → speed filter).

### Fixed
- Plate validator no longer falsely accepts 6-character historical
  inputs like `AB12CD`; post-1994 plates normalised to compact form.
- WebSocket handler now authenticates via Sec-WebSocket-Protocol
  (browser-friendly path), not just Authorization header.
- Trusted-proxy misconfiguration warning turned into a log line
  rather than a startup abort, so containerised deployments with a
  non-default gateway do not fail health checks.

### Security
- Host allow-list on outbound OSRM traffic.
- Content-Security-Policy emitted by the API (`default-src 'none'`).
- CSP on the nginx frontend now constrains `connect-src` to same-origin
  plus ws/wss; tile provider whitelisted.

## 0.1.0 — 2026-04-16 — Initial scaffold, Mission I
- Go/Gin backend with Mongo + Redis wiring.
- Vue 3 / Vite / Tailwind dashboard skeleton.
- Italian landing page.
- Initial MODUS_OPERANDI.md, ARCHITECTURE.md, API.md.
