# LogiTrack

**Kit modulare per software di logistica, su misura per la PMI italiana.**
**A modular logistics-software kit, tailored per Italian SME engagement.**

LogiTrack is **not a SaaS product**. It is a freelancer-grade software
kit: a Go + MongoDB + Redis platform plus per-vertical domain modules,
designed to be **forked once per customer**, customised, deployed on
that customer's infrastructure, and maintained under a project +
retainer contract.

There is no shared multi-tenant cloud, no canone mensile, no
subscription. Each customer owns their fork, their data, and their
deployment timeline. The freelancer (Renan Augusto Macena) ships the
kit, customises it during a 4–8 week engagement, hands over the keys,
and stays on retainer for compliance updates and feature work.

> Read [`docs/KIT-PLAYBOOK.md`](docs/KIT-PLAYBOOK.md) before forking.
> Read [`docs/CLIENT-FORK-RECIPE.md`](docs/CLIENT-FORK-RECIPE.md) on
> the day of a new engagement. Read
> [`docs/FREELANCER-COMMERCIAL-MODEL.md`](docs/FREELANCER-COMMERCIAL-MODEL.md)
> before you quote.

## Verticals shipped today

- **`logistics`** — supply-chain visibility for carriers, shippers and
  freight forwarders along the A22 Brennero corridor and through the
  Verona Quadrante Europa intermodal terminal. Live tracking from
  telematics webhooks, SHA-256-chained chain-of-custody log, OSRM-backed
  route optimisation with fallback. Pitch in
  [`docs/PITCH.md`](docs/PITCH.md).
- **`rifiuti`** — RENTRI-ready waste-transport vertical for
  trasportatori di rifiuti speciali iscritti Albo cat. 4/5/8. Owns the
  FIR (Formulario Identificazione Rifiuti) state machine, the registro
  cronologico carico/scarico, the EER catalogue and the RENTRI client
  adapter (queued-stub default; live HTTP one constructor away). Pitch
  in [`docs/PITCH-RIFIUTI.md`](docs/PITCH-RIFIUTI.md), regulatory
  anchors in [`docs/MODULE-RIFIUTI.md`](docs/MODULE-RIFIUTI.md).

The kit doctrine is in
[`docs/DOMAIN-MODULES.md`](docs/DOMAIN-MODULES.md): shared platform +
leaf modules + per-customer overlay; modules never import each other.

---

## Stack

| Layer | Technology | Why |
| --- | --- | --- |
| Backend | Go 1.26 + Gin | Concurrent ingestion at low latency, fast cold-start, single binary in distroless |
| Frontend | Vue 3 + Vite + Tailwind + TypeScript | Lightweight, mapping-friendly, easy to fork per customer |
| Primary DB | MongoDB 7 | Flexible document shape per carrier/cliente, geospatial indexes |
| Cache & Pub/Sub | Redis 7 | Hot position cache, WebSocket fan-out |
| Maps | Leaflet + OSRM (optional, opt-in self-hosted) | EU data residency by default; falls back to straight-line if OSRM not configured |
| Observability | OpenTelemetry + zap | OTLP tracing, structured JSON logs, Prometheus exposition |

---

## Quick start

```bash
# 1. Clone and configure
cp .env.example .env
# edit .env — at minimum:
#   - JWT_SECRET                          (openssl rand -hex 32)
#   - MONGO_ROOT_PASSWORD                 (rotate from the dev placeholder)
#   - REDIS_PASSWORD                      (rotate from the dev placeholder)
#   - LOGITRACK_IDENTITY_DEMO_USER        (e.g. demo@logitrack.it)
#   - LOGITRACK_IDENTITY_DEMO_PASSWORD    (>= 12 chars, NIST SP 800-63B)

# 2. Boot the full stack
docker compose up --build

# 3. Verify health
curl -s http://localhost:8080/api/health | jq

# 4. Open the UI
open http://localhost:5174
```

Contract ports: backend **8080**, frontend **5174**, MongoDB **27017**
(bound to 127.0.0.1), Redis **6380** (bound to 127.0.0.1; container-
internal Redis still on 6379).

Backend-only run (Mongo/Redis still in Docker):

```bash
cd backend
make tidy
make run
```

Frontend-only run:

```bash
cd frontend
npm install
npm run dev
```

---

## Endpoints (high level)

| Method | Path | Purpose |
| --- | --- | --- |
| GET | `/api/health` | Liveness + dependency status |
| GET | `/api/ready` | Readiness (seed-done aware) |
| GET | `/metrics` | Prometheus exposition |
| POST | `/api/v1/auth/login` | Issue access token (15-min TTL) |
| POST | `/api/v1/shipments` | Create shipment |
| GET | `/api/v1/shipments` | List shipments |
| GET | `/api/v1/shipments/{id}` | Retrieve shipment |
| POST | `/api/v1/shipments/{id}/waypoints` | Ingest telematics waypoint |
| GET | `/api/v1/shipments/{id}/trace` | Chain-of-custody log |
| GET | `/api/v1/shipments/{id}/position` | Latest cached position |
| GET | `/api/v1/shipments/{id}/eta` | Predicted ETA |
| POST | `/api/v1/routes/optimize` | OSRM-backed (or fallback) route optimisation |
| GET (WS) | `/api/v1/stream/tracking` | Live tracking event stream |
| POST/GET | `/api/v1/vehicles\|drivers\|geofences` | Fleet master data |
| POST/GET | `/api/v1/rifiuti/produttori\|trasportatori\|destinatari\|fir` | Rifiuti master data + FIR lifecycle |
| GET | `/api/v1/rifiuti/cer/{code}` | Inline CER validator |

Full protocol, request/response examples and WebSocket message schema in
[`docs/API.md`](docs/API.md).

---

## Documentation

**Kit playbook (read these before a new client engagement):**

- [`docs/KIT-PLAYBOOK.md`](docs/KIT-PLAYBOOK.md) — kit invariants,
  upstream-vs-overlay rule, patch-propagation discipline.
- [`docs/CLIENT-FORK-RECIPE.md`](docs/CLIENT-FORK-RECIPE.md) — mechanical
  steps to spin up a per-client fork.
- [`docs/MOZZECANE-PITCH.md`](docs/MOZZECANE-PITCH.md) — door-to-door
  + cold-call discovery script for Verona-area SMEs.
- [`docs/FREELANCER-COMMERCIAL-MODEL.md`](docs/FREELANCER-COMMERCIAL-MODEL.md)
  — pricing, contratto d'opera, FatturaPA, IVA/INPS reality.

**Vertical-specific:**

- [`docs/PITCH.md`](docs/PITCH.md) — logistics one-pager.
- [`docs/PITCH-RIFIUTI.md`](docs/PITCH-RIFIUTI.md) — rifiuti one-pager.
- [`docs/MODULE-RIFIUTI.md`](docs/MODULE-RIFIUTI.md) — rifiuti
  regulatory anchors + state machine.

**Engineering:**

- [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) — architecture, data
  model, sequence diagrams.
- [`docs/DOMAIN-MODULES.md`](docs/DOMAIN-MODULES.md) — module contract
  and roadmap.
- [`docs/API.md`](docs/API.md) — REST + WebSocket reference.
- [`docs/SECURITY.md`](docs/SECURITY.md) — threat model, controls,
  self-assessment.
- [`docs/RUNBOOK.md`](docs/RUNBOOK.md) — per-deployment operations.
- [`docs/INTEGRATIONS.md`](docs/INTEGRATIONS.md) — telematics + RENTRI
  envelopes (the only integrations actually wired in the kit today).
- [`docs/ITALIAN-COMPLIANCE.md`](docs/ITALIAN-COMPLIANCE.md) —
  regulatory map (Reg. UE 952/2013, Reg. UE 165/2014, Albo, ADR/ATP,
  D.Lgs. 152/2006, RENTRI).
- [`docs/COMPLIANCE.md`](docs/COMPLIANCE.md) — one-pager for the
  prospect's legal counsel.
- [`docs/A11Y.md`](docs/A11Y.md) — WCAG 2.1 AA target.
- [`docs/TECHNICAL-DEBT.md`](docs/TECHNICAL-DEBT.md) — known gaps,
  honestly listed.
- [`docs/CHANGELOG.md`](docs/CHANGELOG.md), [`docs/CONTRIBUTING.md`](docs/CONTRIBUTING.md).

---

## Project structure

```text
LogiTrack/
├── backend/           # Go/Gin API
├── frontend/          # Vue 3 + Vite SPA
├── landing-page/      # Italian-language one-page portfolio piece (pure HTML/CSS)
├── docs/              # Engineering + commercial playbook
├── .github/workflows/ # CI/CD
├── docker-compose.yml
└── .env.example
```

---

## Production hardening checklist

The compose stack ships with safe-by-default-but-not-production
credentials. Before any non-localhost deployment, walk this checklist —
every item is a real audit finding with a concrete fix in the repo.

**Secrets and credentials**

- [ ] Rotate `JWT_SECRET` to a fresh 32-byte hex value
      (`openssl rand -hex 32`). The backend refuses to boot in
      production with a known-weak placeholder, < 32 chars, or with an
      unauthenticated `MONGO_URI` / `REDIS_URL`.
- [ ] Rotate `MONGO_ROOT_PASSWORD` and `REDIS_PASSWORD`. Defaults end
      in `devonly-rotate-before-deploy` so a leaked `.env` is loud.
- [ ] Rotate `LOGITRACK_IDENTITY_DEMO_PASSWORD` (or set
      `LOGITRACK_IDENTITY_BACKEND=disabled` and integrate the
      customer's IDP via the `IdentityStore` interface in
      [`backend/internal/handlers/auth.go`](backend/internal/handlers/auth.go)).
      In production with the memory backend you must also set
      `LOGITRACK_IDENTITY_DEMO_BREACH_ACK=true` to acknowledge that the
      seed password has been verified externally against HIBP.
- [ ] Inject every secret via the orchestrator's secret manager
      (Kubernetes Secret, Docker Compose `secrets:`, AWS Secrets
      Manager, Vault). Plain env vars are visible to every process in
      the container and to anyone with `docker inspect`.

**Network exposure**

- [ ] MongoDB port `27017` and Redis port `6380` are bound to
      `127.0.0.1` in the compose file. Confirm production deployment
      has no equivalent host-network exposure (only the backend talks
      to Mongo and Redis on the internal network).
- [ ] Both Mongo and Redis require authentication. Confirm
      `MONGO_URI` / `REDIS_URL` carry credentials.

**Identity backend**

- [ ] In production set `LOGITRACK_IDENTITY_BACKEND=disabled` unless
      the customer wants the in-memory demo store. Disabled mode
      returns 503 with a clear pointer to the IDP integration path —
      so a misrouted request is loud, not silent.

**Route optimisation**

- [ ] If you need OSRM, self-host it with a custom truck profile, set
      `OSRM_BASE_URL` to it, and add the host to `OSRM_ALLOWED_HOSTS`.
      The kit defaults `OSRM_BASE_URL` empty: the optimiser falls
      back to a straight-line + 70 km/h estimate. Do **not** point at
      the public `router.project-osrm.org` demo (US-hosted, no SLA, no
      truck profile, sends EU customer route data outside the EU).
- [ ] Set `OSRM_TRUCK_PROFILE` to the profile name your OSRM server
      exposes (e.g. `truck` or `hgv`). Without it, every truck
      request falls back to the driving profile and ignores HGV
      restrictions, weight limits and ZTL.

**TLS and reverse proxy**

- [ ] Front the SPA with a reverse proxy that terminates TLS (Caddy,
      Cloudflare, AWS ALB, nginx-ingress). HSTS in
      [`backend/internal/middleware/security_headers.go`](backend/internal/middleware/security_headers.go)
      is honoured only over HTTPS.
- [ ] Tighten `Content-Security-Policy: connect-src` from
      `'self' ws: wss:` to the specific WebSocket host once the
      deployment URL is known.

**Observability**

- [ ] Wire `OTEL_EXPORTER_OTLP_ENDPOINT` to the customer's collector.
      Without it, tracing is a no-op (the runtime is OpenTelemetry-
      ready but inert).
- [ ] Scrape `GET /metrics` into Prometheus. The endpoint is
      unauth-ed because in production a service-mesh sidecar fronts
      it; if there is no service mesh, restrict the metrics endpoint
      at the reverse proxy layer.

**Backups and operations**

- [ ] Configure Mongo dumps. The compose volume `logitrack-mongo-data`
      is the source of truth for shipments, custody chain, audit log
      and fleet master data — all of which are append-only or
      tamper-evident, so a backup gap directly degrades regulatory
      claims.
- [ ] Read [`docs/RUNBOOK.md`](docs/RUNBOOK.md) before going on-call.

---

## License

MIT. See [`LICENSE`](LICENSE).
