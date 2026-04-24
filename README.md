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
# edit .env — set JWT_SECRET at minimum

# 2. Boot the full stack
docker compose up --build

# 3. Verify health
curl -s http://localhost:8080/api/health | jq

# 4. Open the UI
open http://localhost:5174
```

The contract ports are: backend **8080**, frontend **5174**, MongoDB **27017**,
Redis **6380** (host-mapped; container-internal Redis remains on 6379).

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

```
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

## License

MIT. See [`LICENSE`](LICENSE).
