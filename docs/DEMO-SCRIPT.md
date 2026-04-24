# LogiTrack — Demo Script (10 minutes)

Audience: prospect operations director, 8–50 vehicles, cross-border
flows. Environment: a browser and a terminal.

## 0. Pre-flight (90 s)
```bash
cd LogiTrack
cp .env.example .env
# optional: edit JWT_SECRET
docker compose --profile demo up --build --wait
```
Expected:
- Backend healthy (`curl http://localhost:8080/api/health | jq`).
- Simulator container logs "simulator starting".

## 1. Landing page (1 min)
- Browse `http://localhost:5174/` (Italian landing by default at
  `landing-page/index.html` in production).
- Highlight: hero copy, Quadrante Europa mention, Italian SaaS tiers.

## 2. Dashboard overview (2 min)
- Open `http://localhost:5174/` (SPA).
- Show the three KPI tiles (totali / in transito / in ritardo).
- Show the shipment list — the three demo shipments are visible with
  Verona→Milano, Verona→Napoli, Verona→München.

## 3. Live map (3 min)
- Click into `CMR-VRMU-003` (Verona → München).
- The map renders the planned polyline in blue and the live marker
  (amber circle). Say: "questo marker si muove di secondo in secondo
  grazie al simulatore — in produzione arriva dalla telematica Viasat".
- After 20 seconds show the ETA strip update below the map.

## 4. Chain-of-custody (1 min)
- `curl http://localhost:8080/api/v1/shipments/demo-shipment-verona-munchen/trace -H "Authorization: Bearer $DEMO_JWT" | jq`
- Point at the `hash` / `prevHash` linkage. "Ogni passaggio firmato
  SHA-256: modificare la storia diventa matematicamente evidente."

## 5. Prometheus / observability (1 min)
- `curl http://localhost:8080/metrics | head -20`
- Show `logitrack_ws_connections_active`, `logitrack_http_requests_total`.
- Frame: "in produzione il dato va a Grafana sul vostro mesh o sul
  nostro cluster Aruba Cloud Italia".

## 6. Security posture (1 min)
- `curl -i http://localhost:8080/api/v1/shipments`
  → 401 missing_authorization. "La piattaforma è cieca senza token
  valido — scope di tenant applicato al repository."
- Mention: OSRM allow-list, WS origin check, Italian data residency.

## 7. Q&A + next steps (1 min)
- Offer 30-day pilot su 3 spedizioni reali.
- Leave Pitch + Compliance one-pager with the buyer.

## Teardown
```bash
docker compose down
```
