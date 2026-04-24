# LogiTrack — Operations Cadence

## Daily
- CI on every commit.
- `trivy image logitrack/backend:<tag>` nightly (cron).
- Mongo `mongodump --oplog` at 02:00 Europe/Rome, pushed to the
  customer's S3-compatible bucket (or Aruba Cloud Object Storage for
  SaaS).
- Alertmanager summary 07:45 Europe/Rome to the on-call channel.

## Weekly
- Dependency advisory DB refresh (govulncheck, npm audit, gitleaks).
  PRs auto-opened for patch bumps.
- SLO report generated and shared with the steering committee.
- On-call rotation hand-off Monday 09:00 — written retrospective of
  the week's incidents.

## Monthly
- Full restore drill: restore the latest backup to a throwaway
  container, run `tests/integration/golden_path_test.go`, tear down.
- Penetration-test lite: ZAP baseline + Nuclei against the staging
  instance.
- Dependency major-version review: can we ship the largest-risk
  upgrade this cycle?
- Status-page post: SLA vs. actual for the month.

## Quarterly
- Base-image refresh (distroless, nginx, mongo, redis) and rebuild
  of all images.
- Threat model refresh (`docs/SECURITY.md`).
- Compliance re-verification: confirm every link in
  `docs/ITALIAN-COMPLIANCE.md` still resolves.
- Telepass / AISCAT toll-code dictionary refresh.
- Customer-voice review: are the open issues reflecting real pain?

## Annually
- External penetration test by a Clusit-certified firm.
- Disaster-recovery test: full-stack failover to a secondary region.
- Licence audit on every third-party dependency.
- Pricing review against CRM data.
- Internal GDPR DPIA refresh.
