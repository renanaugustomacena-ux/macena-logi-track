# Changelog

LogiTrack follows semantic versioning (`MAJOR.MINOR.PATCH`). Each
release pairs a git tag with a line in this file.

## 0.3.0 — 2026-04-28 — Kit honesty pass

The kit is reframed: no longer a SaaS product roadmap, now an
honestly-positioned freelancer-grade kit (fork-per-customer model,
contratto d'opera ex art. 2222 c.c., no canone mensile per posto).

### Removed (because the code did not back the claim)

- `internal/integrations/aida` — never wired beyond `_ = deps.AidaClient`.
- `internal/integrations/albo` — same.
- `internal/integrations/rfi` — same.
- `internal/integrations/telepass` — same.
- `POST /api/v1/auth/refresh` — refresh tokens were issued but the SPA
  never stored them. Now access tokens only; sliding sessions belong
  to the customer's IDP via the IdentityStore seam.
- `Redis.StoreRefreshToken` / `RevokeRefreshToken` — orphaned with refresh.
- `KafkaConfig`, `SessionConfig`, `AIDAConfig`, `RFIConfig`,
  `TelepassConfig`, `AlboConfig` — placeholders never read.
- `breachCheckStub` silent no-op — replaced with explicit
  `LOGITRACK_IDENTITY_DEMO_BREACH_ACK=true` opt-in for production +
  memory backend.
- Docs: `MODUS_OPERANDI.md`, `PRICING.md`, `PRICING-RIFIUTI.md`,
  `OPERATIONS-CADENCE.md`, `SLO.md`, `DATA-RESIDENCY.md`,
  `MIGRATION-FROM-LEGACY.md`, `DEMO-SCRIPT.md`, `RISK-ACCEPTANCES.md`,
  `SECURITY-SELF-ASSESSMENT.md` (merged into `SECURITY.md`).

### Security (CTF-style audit fixes)

- C-1 `middleware/rate_limit.go`: added eviction to the per-IP bucket
  map (30-min idle → swept every 5 min). Closes the IP-rotation
  memory-exhaustion DoS vector.
- C-2 `services/route_optimizer.go`: the OSRM HTTP client now refuses
  redirects (`http.ErrUseLastResponse`). Closes the SSRF bypass via a
  compromised OSRM serving 30x to internal hosts.
- C-3 `handlers/rifiuti.go`: xFIR placeholder switched from string
  concatenation to `encoding/xml` so user-controlled CER values are
  escaped.
- H-1 OSRM default → empty. Public US-hosted demo no longer the
  default; customer routes stay in the EU unless deliberately wired
  to a self-hosted OSRM.
- H-4 `config.guardProductionSecrets` extended: rejects
  unauthenticated `MONGO_URI` / `REDIS_URL` and weak/empty demo
  identity password in production.
- M-2 `services/shipment_service.go`: genesis custody hash failure
  now surfaces to caller instead of being silently swallowed.
- L-1 `middleware/security_headers.go`: HSTS now includes `preload`.

### Added

- New playbook docs: `KIT-PLAYBOOK.md`, `CLIENT-FORK-RECIPE.md`,
  `MOZZECANE-PITCH.md`, `FREELANCER-COMMERCIAL-MODEL.md`.

### Changed

- `README.md`, `ARCHITECTURE.md`, `INTEGRATIONS.md`, `COMPLIANCE.md`,
  `RUNBOOK.md`, `TECHNICAL-DEBT.md`, `MODULE-RIFIUTI.md`,
  `landing-page/index.html`, `PITCH.md`, `PITCH-RIFIUTI.md`: SaaS
  framing stripped, kit framing applied, dead integration references
  removed.
- `API.md`: added the 19 routes that existed in code but were not
  documented (auth login, fleet vehicles/drivers/geofences, rifiuti
  anagrafiche/FIR/CER/vidima/transition).

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
