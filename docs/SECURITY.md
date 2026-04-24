# LogiTrack — Security Posture

## Threat model (summary)

- **Assets:** shipment identities, geolocation traces, chain-of-custody
  log, tenant credentials, operator PII (driver names, contacts).
- **Attackers:**
  - External unauthenticated: HTTP probers, bot scanners, DDoS.
  - External authenticated: a competing tenant attempting horizontal
    privilege escalation.
  - Insider: operator with valid JWT attempting to enumerate other
    tenants' shipments.
  - Supply-chain: compromised upstream Go or JS dependency.
- **Primary threats:**
  - Tenant boundary crossing via ID-guessing or JWT tampering.
  - Chain-of-custody tampering (rewrite history of a consignment).
  - OSRM SSRF pivot (outbound HTTP with user-controlled URL).
  - WebSocket hijack from a foreign origin.
  - Secret leakage via logs.

## Controls

### Authentication
- HS256 JWT with `JWT_SECRET` from environment, minimum 256 bits
  (enforced by documentation; CI check planned). `.env.example`
  points at `openssl rand -hex 32`.
- Algorithm pinned: only `HS256` is accepted; the `alg:none` bypass is
  rejected by the underlying `github.com/golang-jwt/jwt/v5` library.
- Issuer check: `jwt.WithIssuer(cfg.JWT.Issuer)` applied in both the
  REST middleware and the WebSocket handshake.

### Authorisation
- Tenant scoping at the repository layer: every Mongo query filters by
  `tenant_id`. No endpoint returns data without the filter.
- Role gating via `middleware.RequireRole("admin")` etc., available
  but not yet applied to every mutating endpoint. Planned for a
  future iteration; tracked in [`docs/TECHNICAL-DEBT.md`](TECHNICAL-DEBT.md).

### Input validation
- `Shipment.Validate()` enforces plate format, coordinate cardinality
  and required fields before persistence.
- `services.OptimiseRoute` rejects requests with < 2 waypoints.
- Gin bindings validate JSON shapes.

### WebSocket hardening (`handlers/stream.go`)
- Origin check against `HTTP_WS_ORIGINS`.
- JWT accepted via Authorization header, `?access_token=`, or
  `Sec-WebSocket-Protocol` subprotocol (`logitrack.jwt.v1,<token>`).
- Per-IP handshake rate-limit + per-connection inbound rate-limit.
- 5-minute idle disconnect via read deadline.
- 64 KiB max read frame; compression disabled.

### OSRM SSRF mitigation (`services/route_optimizer.go`)
- Host allow-list: `OSRM_ALLOWED_HOSTS` (comma-separated).
- Parsed BaseURL's host is checked BEFORE dialling.
- IP literals compared by string match; metadata addresses
  (169.254.169.254, loopback, link-local) are refused by default.

### HTTP hardening
- `middleware.SecurityHeaders` sets X-Content-Type-Options,
  X-Frame-Options, Referrer-Policy, Permissions-Policy, HSTS, CSP.
- CORS restricted to `HTTP_CORS_ALLOWED`.
- Token-bucket rate limiter per client IP.

### Supply-chain
- `go mod verify` target in Makefile.
- govulncheck run at consolidation (clean as of 2026-04-17).
- Dependencies pinned at module-cache level; rebuilds use
  `go mod download` off the public proxy + sum.golang.org checksum DB.

### Secret management
- `.env.example` documents every variable.
- `.gitignore` excludes `.env`, `*.pem`, `*.key`, `*.crt`.
- Gitleaks/detect-secrets run at CI time (planned; currently manual).

### Chain-of-custody integrity
- `services.computeHash` + `services.VerifyChain` cover append + audit.
- Records are append-only; corrections are a new `CustodyException`
  entry referencing the erroneous sequence.

## Accepted risks / false positives

| Finding | Rationale |
| --- | --- |
| Semgrep `websocket-missing-origin-check` on `handlers/stream.go:L91` | CheckOrigin is an allow-list closure; the generic rule cannot see through it. Annotated with `nosemgrep`. |
| Demo JWT_SECRET default in compose | The value contains the literal string `change-me`; startup refuses production mode without an override. Documented in RUNBOOK. |

## Reporting

Please report suspected vulnerabilities to `security@logitrack.it`
(PGP key on request). We follow a 90-day coordinated disclosure
window with CVSS-based severity triage.
