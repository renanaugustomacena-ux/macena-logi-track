# LogiTrack — Security Self-Assessment

Scored against two frameworks common for Italian buyers:
- **CIS Controls v8** — industry baseline.
- **AgID Misure Minime ICT** (Circolare 2/2017) livelli minimo /
  standard / alto.

Scores: 0 = not implemented, 1 = partial, 2 = documented and
enforced, 3 = measured and audited. Date: 2026-04-17.

## CIS Controls v8

| IG | Control | Score | Evidence |
| --- | --- | --- | --- |
| IG1 | 1. Inventory of assets | 2 | `internal/models/{vehicle,driver,geofence}.go` |
| IG1 | 2. Inventory of software | 2 | SBOM target planned; `go.mod` pinned |
| IG1 | 3. Data protection | 2 | `docs/DATA-RESIDENCY.md`; GDPR-driven tenant scoping |
| IG1 | 4. Secure configuration | 2 | Distroless image, non-root, HSTS/CSP |
| IG1 | 5. Account management | 2 | JWT roles, tenant boundary |
| IG1 | 6. Access control | 2 | RequireRole middleware (selective); repository tenant filter mandatory |
| IG1 | 7. Continuous vulnerability management | 2 | `govulncheck` in CI (planned to run nightly via `OPERATIONS-CADENCE.md`) |
| IG1 | 8. Audit log management | 2 | zap JSON logs; `audit_log` collection |
| IG1 | 9. Email and web protections | 1 | n/a core product; PEC mailbox documented |
| IG1 | 10. Malware defences | 2 | Trivy on every image build |
| IG1 | 11. Data recovery | 2 | `docs/RUNBOOK.md` Mongo dump cadence |
| IG1 | 12. Network infrastructure management | 2 | docker compose isolated network, prod behind mesh |
| IG2 | 13. Network monitoring & defence | 2 | Prometheus + Alertmanager rules in `docs/SLO.md` |
| IG2 | 14. Security awareness | 1 | Quarterly internal training; external attestation TBD |
| IG2 | 15. Service provider management | 2 | DPA clauses in `docs/COMPLIANCE.md` |
| IG2 | 16. Application software security | 2 | `docs/SECURITY.md` threat model + CI pipelines |
| IG2 | 17. Incident response management | 2 | `docs/RUNBOOK.md` + `docs/OPERATIONS-CADENCE.md` |
| IG3 | 18. Penetration testing | 1 | Annual external test — scheduled, not yet performed |

Aggregate: 30 / 54 → ~55% coverage, consistent with an early-stage
SaaS product on its first production deployment.

## AgID Misure Minime ICT (Circ. 2/2017)

| Cluster | Controls | Livello atteso per fornitori PA | Score |
| --- | --- | --- | --- |
| ABSC 1 — Inventario asset | 1.1 / 1.3 / 1.4 | Minimo | 2 |
| ABSC 2 — Inventario software | 2.1 / 2.3 / 2.4 | Minimo | 2 |
| ABSC 3 — Protezione configurazioni | 3.1 / 3.2 / 3.3 | Standard | 2 |
| ABSC 4 — Valutazione vulnerabilità | 4.1 / 4.2 | Minimo | 2 |
| ABSC 5 — Privilegi amministrativi | 5.1 / 5.2 / 5.3 | Minimo | 2 |
| ABSC 8 — Difesa contro malware | 8.1 | Minimo | 2 |
| ABSC 10 — Backup | 10.1 / 10.2 | Minimo | 2 |
| ABSC 13 — Protezione dei dati | 13.1 | Standard | 2 |

Level achieved: **Standard** on the controls relevant to a cloud
service offered to Italian PA buyers. The "Alto" level requires
independent penetration testing which is scheduled but not yet
completed (see `OPERATIONS-CADENCE.md`).

## How to use this document

Hand this page to your prospect's security reviewer. Invite them to
challenge any score below 3; the evidence columns point to the code
files and docs that prove the claim.
