# LogiTrack

**Traccia ogni spedizione in tempo reale, dal Quadrante Europa a tutta l'UE.**
**Track every shipment in real time, from Quadrante Europa to the whole EU.**

LogiTrack is a supply-chain visibility platform engineered for the Verona
intermodal corridor: carriers, shippers, customs brokers and freight
forwarders who live along the A22 Autostrada del Brennero and ship through
the Quadrante Europa freight terminal (8+ million tonnes/year, second-largest
intermodal hub in Europe).

The platform ingests telematics data (Viasat, Octo, Geotab), CMR electronic
consignment notes, AIDA customs declarations, and intermodal rail slots from
RFI, unifying them into a real-time map, a tamper-evident chain-of-custody
log, and SLA analytics.

---

## Stack

| Layer | Technology | Rationale |
| --- | --- | --- |
| Backend | Go 1.26 + Gin | Concurrent ingestion at low latency, cheap to scale horizontally |
| Frontend | Vue 3 + Vite + TailwindCSS + TypeScript | Lightweight SPA, great for mapping-heavy UIs |
| Primary DB | MongoDB 7 | Flexible document shape per carrier, geospatial indexes |
| Cache & Pub/Sub | Redis 7 | Hot position cache and WebSocket fan-out |
| Event Stream | Kafka (roadmap) | Phase 3 event-sourced history |
| Maps | Leaflet + OSRM (self-hosted in prod) | EU data residency by default |
| Observability | OpenTelemetry + Zap | OTLP traces, structured JSON logs |

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

The contract ports are: backend **8080**, frontend **5174**, MongoDB **27017**
(bound to 127.0.0.1), Redis **6380** (bound to 127.0.0.1; container-internal
Redis remains on 6379).

To run only the backend locally (with MongoDB/Redis in Docker):

```bash
cd backend
make tidy
make run
```

To run only the frontend:

```bash
cd frontend
npm install
npm run dev
```

---

## Endpoints

| Method | Path | Purpose |
| --- | --- | --- |
| GET | `/api/health` | Liveness + dependency status |
| POST | `/api/v1/shipments` | Create a shipment |
| GET | `/api/v1/shipments` | List shipments (filters: `status`, `carrier`, `from`, `to`) |
| GET | `/api/v1/shipments/{id}` | Retrieve a shipment |
| POST | `/api/v1/shipments/{id}/waypoints` | Ingest a telematics waypoint |
| GET | `/api/v1/shipments/{id}/trace` | Return the chain-of-custody log |
| GET | `/api/v1/shipments/{id}/position` | Latest cached position |
| POST | `/api/v1/routes/optimize` | OSRM-backed route optimisation |
| GET (WS) | `/api/v1/stream/tracking` | Live tracking-event stream |

Full protocol, request/response examples and WebSocket message schema in
[`docs/API.md`](docs/API.md).

---

## Documentation

- [`docs/MODUS_OPERANDI.md`](docs/MODUS_OPERANDI.md) — strategic, technical, operational playbook (13,000+ words).
- [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) — architecture, data model, sequence diagrams.
- [`docs/API.md`](docs/API.md) — REST + WebSocket API reference.
- [`docs/ITALIAN-COMPLIANCE.md`](docs/ITALIAN-COMPLIANCE.md) — regulatory map (Reg. UE 952/2013, AIDA, Reg. UE 165/2014, Albo Autotrasportatori, Telepass/AISCAT, ADR/ATP).
- [`docs/RUNBOOK.md`](docs/RUNBOOK.md) — on-call procedures, backup/restore, common incident response.
- [`docs/SLO.md`](docs/SLO.md) — latency, availability and error-budget objectives.
- [`docs/SECURITY.md`](docs/SECURITY.md) — threat model, secrets, vulnerability reporting.
- [`docs/A11Y.md`](docs/A11Y.md) — WCAG 2.1 AA checklist and keyboard-only walk-throughs.
- [`docs/PITCH.md`](docs/PITCH.md), [`docs/DEMO-SCRIPT.md`](docs/DEMO-SCRIPT.md), [`docs/PRICING.md`](docs/PRICING.md), [`docs/COMPLIANCE.md`](docs/COMPLIANCE.md), [`docs/SECURITY-SELF-ASSESSMENT.md`](docs/SECURITY-SELF-ASSESSMENT.md), [`docs/MIGRATION-FROM-LEGACY.md`](docs/MIGRATION-FROM-LEGACY.md), [`docs/OPERATIONS-CADENCE.md`](docs/OPERATIONS-CADENCE.md), [`docs/INTEGRATIONS.md`](docs/INTEGRATIONS.md), [`docs/DATA-RESIDENCY.md`](docs/DATA-RESIDENCY.md), [`docs/TECHNICAL-DEBT.md`](docs/TECHNICAL-DEBT.md), [`docs/RISK-ACCEPTANCES.md`](docs/RISK-ACCEPTANCES.md), [`docs/CHANGELOG.md`](docs/CHANGELOG.md), [`docs/CONTRIBUTING.md`](docs/CONTRIBUTING.md) — sales-enablement and operations artefacts.

---

## Project structure

```text
LogiTrack/
├── backend/           # Go/Gin API
├── frontend/          # Vue 3 + Vite SPA
├── landing-page/      # Italian-language marketing page (pure HTML/CSS)
├── docs/              # MODUS, ARCHITECTURE, API
├── .github/workflows/ # CI/CD
├── docker-compose.yml
└── .env.example
```

---

## Production hardening checklist

The compose stack ships with safe-by-default-but-not-production credentials.
Before any non-localhost deployment, walk this checklist. Every item is a
real audit finding from the 2026-04-27 audit and has a concrete fix in this
repository.

**Secrets and credentials**

- [ ] Rotate `JWT_SECRET` to a fresh 32-byte hex value (`openssl rand -hex 32`).
      The backend refuses to boot with the documented placeholder when
      `APP_ENV=production`.
- [ ] Rotate `MONGO_ROOT_PASSWORD` and `REDIS_PASSWORD`. The defaults end in
      `devonly-rotate-before-deploy` so a leaked `.env` is at least loud
      about its provenance.
- [ ] Rotate `LOGITRACK_IDENTITY_DEMO_PASSWORD` (or move to
      `LOGITRACK_IDENTITY_BACKEND=disabled` and integrate the corporate IDP
      via the `IdentityStore` interface in
      [`backend/internal/handlers/auth.go`](backend/internal/handlers/auth.go)).
- [ ] Inject every secret via the orchestrator's secret manager (Kubernetes
      Secret, Docker Compose `secrets:`, AWS Secrets Manager, Vault). Plain
      env vars are visible to every process in the container and to anyone
      with `docker inspect`.

**Network exposure**

- [ ] MongoDB port `27017` and Redis port `6380` are bound to `127.0.0.1`
      in the compose file. Confirm your production deployment has no
      equivalent host-network exposure (i.e. only the backend talks to
      Mongo and Redis, both reachable by service name on the internal
      network).
- [ ] Both Mongo and Redis now require authentication. Confirm your
      `MONGO_URI` / `REDIS_URL` carry credentials (`mongodb://user:pass@…`
      and `redis://default:pass@…`).

**Identity backend**

- [ ] In production set `LOGITRACK_IDENTITY_BACKEND=disabled` unless you
      want every login attempt to validate against the in-memory demo
      store. Disabled mode returns 503 with a clear pointer to the IDP
      integration path, so a misrouted request is loud, not silent.

**Italian rail integration (FERTRAM / RFI)**

- [ ] If you use `LOGITRACK_RFI_*`, supply the mTLS certificate pair as
      both `LOGITRACK_RFI_MTLS_CERT_FILE` and
      `LOGITRACK_RFI_MTLS_KEY_FILE`. The backend refuses to boot if only
      one of the pair is set. Optional `LOGITRACK_RFI_MTLS_CA_FILE` for
      private CAs.

**Route optimisation**

- [ ] Self-host an OSRM server with a custom truck profile and set
      `OSRM_BASE_URL` to it; add the host to `OSRM_ALLOWED_HOSTS`. The
      public `router.project-osrm.org` has no SLA, no truck profile, and
      sends EU customer route data to a US-hosted demo service.
- [ ] Set `OSRM_TRUCK_PROFILE` to the profile name your OSRM server
      exposes (e.g. `truck` or `hgv`). Without it, every truck
      request falls back to the driving profile and ignores HGV
      restrictions, weight limits and ZTL.

**TLS and reverse proxy**

- [ ] Front the SPA with a reverse proxy that terminates TLS (Caddy,
      Cloudflare, AWS ALB, nginx-ingress). The HSTS header set in
      [`frontend/nginx.conf`](frontend/nginx.conf) is honoured only over
      HTTPS.
- [ ] Tighten `Content-Security-Policy: connect-src` from `'self' ws: wss:`
      to the specific WebSocket host once the deployment URL is known.
      Tracked in [`docs/TECHNICAL-DEBT.md`](docs/TECHNICAL-DEBT.md).

**Observability**

- [ ] Wire `OTEL_EXPORTER_OTLP_ENDPOINT` to your collector. Without it,
      tracing is a no-op (the runtime is OpenTelemetry-ready but inert).
- [ ] Scrape `GET /metrics` into Prometheus. The endpoint is unauth-ed
      because in production a service-mesh sidecar fronts it; if you do
      not run a service mesh, restrict the metrics endpoint at the
      reverse proxy layer.

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
