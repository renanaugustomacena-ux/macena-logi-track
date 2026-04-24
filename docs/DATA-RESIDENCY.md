# LogiTrack — Data Residency

## Default topology (SaaS)

- Compute: Aruba Cloud, region `Milano (Italy)`.
- MongoDB: replica set in the same region, encrypted at rest (LUKS
  under the VM disks; WiredTiger encrypted storage engine on top).
- Redis: managed cluster in the same region.
- Object store (documents attached to shipments): Aruba Cloud Object
  Storage, region `Milano`.
- Backups: cold-stored in a second Aruba region (`Bergamo`).
- CDN (static assets, landing): Aruba CDN; DNS via Aruba.
- Metrics/logs: Grafana Cloud `eu-west` region, with PII scrubbing
  applied at export time (see SECURITY.md).

All personal data, including telematics traces and operator PII,
remains within the **EEA**. No sub-processor outside the EEA is used
by default.

## Sub-processors registry

| Sub-processor | Purpose | Region | DPA status |
| --- | --- | --- | --- |
| Aruba S.p.A. | IaaS + object storage | Italy | signed |
| OpenStreetMap Foundation | tile serving (dev only) | UK | best-effort CC BY-SA |
| Project OSRM (demo) | route optimisation (dev only) | Heidelberg, DE | best-effort public demo |
| Grafana Labs EMEA SARL | metrics/log SaaS | FR | signed SCC |
| Cloudflare | WAF/CDN (optional) | EU PoPs | signed SCC |

In production, the OSRM public demo is replaced by a customer-
hosted OSRM instance (`osrm.logitrack.local`) and the
`OSRM_ALLOWED_HOSTS` allow-list is updated accordingly.

## Self-hosted option

Customers with strict internal data-residency policies (PA, health
sector, aerospace) can deploy LogiTrack in their own infrastructure:
- Kubernetes manifests / docker-compose published under the same
  GitHub repo.
- MongoDB Enterprise / Redis Enterprise supported.
- No outbound connection required by default — OSRM runs on-premise,
  Grafana points at the customer's Prometheus, logs stay local.
- Image pulls from LogiTrack's Harbor registry; can be mirrored to
  Nexus / JFrog in the customer's zone.

## Retention

- Shipment documents: 10 years (art. 2220 c.c. business records) —
  configurable down to 5 years per customer's internal policy.
- Chain-of-custody records: retained for as long as the parent
  shipment exists (never auto-deleted).
- Telematics waypoints: by default 24 months; beyond that, aggregated
  into daily summaries kept for 10 years.
- Audit log: 24 months.
- Backups: daily incrementals retained for 30 days, weekly fulls for
  12 months, monthly fulls for 36 months.

## Right-to-erasure workflow

- GDPR Art. 17 request → admin endpoint `DELETE /api/v1/operators/{id}`
  (planned) → hard-delete of personal fields; analytic aggregates
  retained with the operator identifier hashed.
- Chain-of-custody `Actor.Name` replaced with "anonimizzato",
  `Actor.VATNumber` retained (it is the business identifier of the
  carrier, not a personal datum).
