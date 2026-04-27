# LogiTrack — Service Level Objectives

Customer-facing declaration:

## Availability
- Monthly availability target: **99.5%** (Starter/Professionale),
  **99.9%** (Enterprise with SLA contract).
- Measured via an external synthetic probe hitting `/api/health`
  every 60 seconds.

## Latency (P95, under 10 RPS)
- `GET /api/v1/shipments` (list): **< 400 ms**.
- `GET /api/v1/shipments/:id`: **< 200 ms**.
- `POST /api/v1/shipments/:id/waypoints`: **< 500 ms** (includes Mongo
  write + Redis publish).
- `GET /api/v1/shipments/:id/eta`: **< 400 ms** (includes OSRM round
  trip; cache hit should be < 50 ms).
- WebSocket event end-to-end (ingestion → broadcast): **P95 < 150 ms**.

## Error budget
- `5xx` error rate: **< 0.5%** monthly.
- Breach ⇒ PIR + root-cause publication within 7 days.

## Data freshness
- Live map marker update: within **2 seconds** of an ingested waypoint
  (tested by the integration suite).
- Chain-of-custody append-visibility: within **500 ms**.

## Backups (RPO/RTO)
- **RPO:** 1 hour (oplog archiving).
- **RTO:** 4 hours (MongoDB restore + smoke test).

## Planned maintenance
- Announced ≥ 72 h in advance on status.logitrack.it.
- Never during Italian business hours (08:00–20:00 Europe/Rome
  Lun–Ven) unless emergency.

## Italian holidays

LogiTrack's on-call team observes the full list of Italian national
holidays (including patron-saint day for Verona: 12 aprile San Zeno).
SLA response-time clocks pause on these dates; SLA uptime clocks do
not.
