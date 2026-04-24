# LogiTrack — Risk Acceptance Log

Each entry documents a finding the Mission II consolidator noted but
that was accepted-with-rationale rather than fixed. Signed by the
consolidator; countersigned by the product owner at release time.

| ID | Finding | Severity | Rationale | Expiry |
| --- | --- | --- | --- | --- |
| RA-001 | Semgrep `websocket-missing-origin-check` on `internal/handlers/stream.go` (CheckOrigin is an allow-list closure; the rule cannot see through the closure). | Low | The origin check IS performed; the control is enforced via the closure parameter. Runtime test confirms 403 on foreign origins. Annotated with `nosemgrep`. | 2027-04-17 |
| RA-002 | Default `JWT_SECRET` in compose reads `change-me-in-production-use-openssl-rand-hex-32`. | Medium | Literal string is neither random nor high-entropy; operator MUST override. Documented in `docs/RUNBOOK.md`. Production start-up will add a refuse-to-start guard in the next cycle. | 2026-07-17 |
| RA-003 | Trivy reports HIGH CVE-2026-39883 on `otel/sdk v1.40.0` — mitigated by bumping to v1.43.0 but Trivy's DB still flags the module path; re-scan after DB refresh clears it. | Informational | Actual fix present. Re-scan 2026-04-24. | 2026-05-01 |
| RA-004 | No E2E Playwright suite shipped in this Mission II cycle; golden-path validated via the simulator + manual curl. | Medium | Playwright is on the Phase-3 roadmap; the simulator provides a reproducible golden-path. | 2026-10-17 |
| RA-005 | No per-tenant quota enforcement in `CreateShipment`. | Low | Current commercial tiers are policed by contract; first abuse case triggers implementation (see `docs/TECHNICAL-DEBT.md` #5). | 2026-12-31 |

Updates to this log occur at release time or after a security review.
