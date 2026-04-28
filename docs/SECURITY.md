# LogiTrack — Security Posture

## 1. Threat model

**Assets**
- Shipment metadata, geolocation traces.
- Chain-of-custody log (tamper-evidence claim).
- Tenant credentials, JWT signing secret.
- Operator PII (driver names, phone numbers), customer PII (consignor /
  consignee anagrafiche).
- For rifiuti: produttore / trasportatore / destinatario anagrafiche,
  FIR, registro carico/scarico, RENTRI numero.

**Attackers**
- External unauthenticated: HTTP probers, bot scanners, attempted DoS.
- External authenticated: a competing tenant attempting horizontal
  privilege escalation.
- Insider: an operator with a valid JWT attempting to enumerate other
  tenants' shipments.
- Supply-chain: a compromised upstream Go or JS dependency.
- Compromised upstream service: a hijacked OSRM redirecting outbound
  HTTP to attacker-controlled hosts (SSRF pivot).

**Primary threats**
- Tenant boundary crossing via id-guessing or JWT tampering.
- Chain-of-custody tampering (rewrite history of a consignment).
- OSRM SSRF pivot.
- WebSocket hijack from a foreign origin or via stolen JWT.
- Rate-limit bypass via IP rotation (memory-exhaustion DoS on the
  per-IP limiter map).
- XML injection at the RENTRI submission boundary.
- Secret leakage via logs.

## 2. Controls

### 2.1 Authentication
- HS256 JWT with `JWT_SECRET` from env, minimum 32 chars in
  production. Production refuses to boot with known-weak placeholders
  (`change-me`, `dev`, etc.) or short secrets.
- Algorithm pinned: only `HS256` accepted by the underlying
  `github.com/golang-jwt/jwt/v5` library; `alg:none` and HMAC-family
  hopping are rejected.
- Issuer check: `jwt.WithIssuer(cfg.JWT.Issuer)` applied in REST
  middleware and WebSocket handshake.
- Access tokens only — there is no refresh endpoint. Customers needing
  sliding sessions integrate their corporate IDP at the IdentityStore
  seam in `backend/internal/handlers/auth.go`.
- Login is rate-limited (5/min effective per IP via the auth-tier
  override) and account-locked after 5 consecutive failures (NIST SP
  800-63B-style 15-minute lockout window).
- The in-memory identity store in production with `memory` backend
  requires `LOGITRACK_IDENTITY_DEMO_BREACH_ACK=true` so the operator
  explicitly acknowledges the seed password was checked against HIBP.

### 2.2 Authorisation
- Tenant scoping at the repository layer: every Mongo query filters
  by `tenant_id`. No endpoint returns data without the filter.
- Role gating via `middleware.RequireRole("admin")` is available; not
  yet applied to every mutating endpoint. Tracked in
  [`TECHNICAL-DEBT.md`](TECHNICAL-DEBT.md).
- The `tenantId` claim is required on every JWT; missing claim → 401
  `invalid_claims`.

### 2.3 Input validation
- `Shipment.Validate()` enforces plate format, coordinate cardinality
  and required fields before persistence.
- `services.OptimiseRoute` rejects requests with < 2 waypoints.
- Rifiuti `FIR.Validate()` enforces CER + ADR + HP coherence; Albo
  categoria + autorizzazione impianto enforced at handler boundary
  via `Trasportatore.CanCarry` + `Destinatario.CanReceive`.
- Gin `ShouldBindJSON` validates JSON shape and catches malformed
  payloads.

### 2.4 WebSocket hardening
- Origin allow-list against `HTTP_WS_ORIGINS`.
- JWT accepted in three places (Authorization header, query
  `?access_token`, `Sec-WebSocket-Protocol` subprotocol).
- HS256-only with issuer check (same as REST middleware).
- Per-IP handshake rate-limit with **eviction** (idle buckets
  removed every 5 min after 30 min idle) — the limiter cannot be
  drained to OOM via IP rotation.
- Per-connection inbound rate-limit (token bucket).
- 5-minute idle disconnect via read deadline.
- 64 KiB max read frame; compression disabled.

### 2.5 OSRM SSRF mitigation
- Host allow-list (`OSRM_ALLOWED_HOSTS`); the BaseURL host is checked
  before any DNS or socket.
- IP literals compared by string match; metadata addresses
  (169.254.169.254, loopback) are refused unless explicitly listed.
- The HTTP client refuses to follow redirects
  (`http.ErrUseLastResponse`) so a compromised OSRM cannot redirect
  to internal hosts and bypass the allow-list.

### 2.6 HTTP hardening
- `middleware.SecurityHeaders` sets X-Content-Type-Options,
  X-Frame-Options DENY, Referrer-Policy, Permissions-Policy, HSTS
  (with `preload`), and a deliberately minimal CSP for the JSON API
  surface (`default-src 'none'; frame-ancestors 'none'`).
- CORS restricted to `HTTP_CORS_ALLOWED`.
- Token-bucket rate limiter per client IP, with **eviction** after
  30 min idle (cap memory under IP-rotation attack).
- Per-route override tier for auth + mutating routes.

### 2.7 Audit log
- Audit middleware enqueues a `Record` on every state-changing 2xx/4xx
  request. 5xx are excluded (they go through error logs, not the
  audit trail).
- The writer is **lossy under backpressure**: under extreme load, it
  drops records and increments a counter logged at WARN. For
  guarantee-grade compliance audit, increase buffer or switch to
  synchronous insert. Documented honestly here so customers know the
  posture before the DPA.

### 2.8 RENTRI submission boundary
- `VidimaFIR` builds the xFIR payload via `encoding/xml` so all
  user-controlled fields (CER, etc.) are escaped. The placeholder
  payload is shipped pending the full XSD encoder; the encoding/xml
  pattern is the same one the XSD encoder will use.

### 2.9 Supply-chain
- `go mod verify` target in Makefile.
- `govulncheck` run at consolidation (clean as of last audit pass).
- Dependencies pinned via go.sum + module cache; rebuilds use
  sum.golang.org checksum DB.

### 2.10 Secret management
- `.env.example` documents every variable.
- `.gitignore` excludes `.env`, `*.pem`, `*.key`, `*.crt`.
- Production guard refuses to boot with:
  - `JWT_SECRET` < 32 chars or matching a known-weak placeholder.
  - `MONGO_URI` without `@` (no credentials).
  - `REDIS_URL` without `@` (no credentials).
  - `LOGITRACK_IDENTITY_BACKEND=memory` with empty / weak demo
    password.
- Gitleaks / detect-secrets scanning is the customer's responsibility
  to wire into their CI.

### 2.11 Chain-of-custody integrity
- `services.computeHash` + `services.VerifyChain` cover append +
  audit. The hash is over the canonical JSON of the record (sans
  Hash field, with timestamps UTC-truncated to ms for BSON round-trip
  stability). Failing to compute the genesis hash now returns an
  error instead of being silently swallowed.
- Records are append-only; corrections are a new `CustodyException`
  entry referencing the erroneous sequence.

## 3. Self-assessment vs. baseline frameworks

Snapshot date: 2026-04-28. Scores 0 (not implemented) — 3 (measured
and audited). Single-customer-fork posture; multi-customer aggregate
posture is the customer's hosting choice, not the kit's.

### CIS Controls v8 (subset relevant to a single-server kit)

| Control | Score | Evidence |
| --- | --- | --- |
| 1. Inventory of assets | 2 | `internal/modules/logistics/{vehicle,driver,geofence}.go` |
| 2. Inventory of software | 2 | `go.mod` pinned + go.sum verified |
| 3. Data protection | 2 | tenant-scoped repo + audit log |
| 4. Secure configuration | 2 | distroless image, non-root, HSTS preload, CSP |
| 5. Account management | 2 | NIST 800-63B-style lockout, breach-ack opt-in |
| 6. Access control | 2 | repository tenant-filter mandatory; RequireRole available |
| 7. Continuous vulnerability management | 1 | `govulncheck` manual; CI integration is per-fork |
| 8. Audit log management | 2 | `audit_log` collection (lossy under backpressure — documented) |
| 11. Data recovery | 1 | runbook documents Mongo dump cadence; backup wiring per fork |
| 16. Application software security | 2 | this document + threat model + 2026-04-28 audit pass |

### AgID Misure Minime ICT (Circ. 2/2017) — Standard level

| Cluster | Score |
| --- | --- |
| ABSC 1 — Inventario asset | 2 |
| ABSC 2 — Inventario software | 2 |
| ABSC 3 — Protezione configurazioni | 2 |
| ABSC 4 — Valutazione vulnerabilità | 1 |
| ABSC 5 — Privilegi amministrativi | 2 |
| ABSC 8 — Difesa contro malware | 2 |
| ABSC 10 — Backup | 1 |
| ABSC 13 — Protezione dei dati | 2 |

The "Alto" level requires independent penetration testing and a
formal incident-response plan — both customer-deployment-specific.

## 4. Accepted residual findings (last audit pass: 2026-04-28)

| Finding | Severity | Rationale |
| --- | --- | --- |
| Audit log is lossy under backpressure | M | Documented behaviour. Customer needing guaranteed audit can increase buffer or switch to synchronous insert in their fork. |
| Refresh endpoint not provided | (intentional) | Kit ships access tokens only. Customer's IDP owns sliding sessions. |
| No audience claim validation on JWT | L | Single-service kit; only relevant if multiple services share a secret. |
| No CSRF protection on POST | L | SPA uses Authorization header, never cookies. CSRF surface = 0 with the current auth shape. |
| `error.Error()` echoed to client on 500 paths | L | Some `problem.Internal(c, …, err.Error())` paths surface DB driver text. Customer-facing paths (4xx) use stable error codes; 500s log full context server-side. |

## 5. Reporting

This is a freelancer-grade kit. Report suspected issues directly to
the maintainer:

- **Renan Augusto Macena** — renanaugustomacena@gmail.com
- Acknowledgement within 72 h. Coordinated disclosure window: 60 days.
- A customer with a deployed fork should also notify their own DPO
  and, if appropriate, the Garante per la Protezione dei Dati
  Personali.
