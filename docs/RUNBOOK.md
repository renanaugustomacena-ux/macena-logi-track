# LogiTrack — Per-Fork Operations Runbook

This runbook covers a single deployed fork on a single VPS / single
Kubernetes namespace. Multi-customer aggregation is not a kit
concern; each customer fork has its own runbook copy.

## Boot smoke-test (manual, ~3 min)

```bash
curl -sf http://localhost:8080/api/health | jq
# expect: { status:"ok", dependencies:{ mongodb:"ok", redis:"ok" } }

curl -sf http://localhost:8080/api/ready | jq
# expect: { status:"ready" } once seed (if SEED_DEMO=true) finishes.

curl -sf http://localhost:8080/metrics | head -5
# expect: logitrack_build_info gauge, http counters

# Login + smoke shipment fetch
TOKEN=$(curl -sf http://localhost:8080/api/v1/auth/login \
  -H 'content-type: application/json' \
  -d '{"username":"<demo>","password":"<demo>"}' | jq -r .accessToken)
curl -sf http://localhost:8080/api/v1/shipments \
  -H "Authorization: Bearer $TOKEN" | jq '.items | length'
```

## Backup & restore

### MongoDB

- **Daily full**:
  ```bash
  docker exec logitrack-mongodb mongodump \
    --authenticationDatabase admin \
    -u "$MONGO_ROOT_USERNAME" -p "$MONGO_ROOT_PASSWORD" \
    --db logitrack --out /data/backup/$(date +%F)
  ```
- **Restore**:
  ```bash
  docker exec logitrack-mongodb mongorestore \
    --authenticationDatabase admin \
    -u "$MONGO_ROOT_USERNAME" -p "$MONGO_ROOT_PASSWORD" \
    --drop /data/backup/YYYY-MM-DD/logitrack
  ```
- **Off-host copy**: rsync the `/data/backup/` directory to S3 / B2 /
  customer NAS via cron.
- **Test**: monthly restore drill on a fresh container; verify
  `services.VerifyChain` passes on a sample tenant.

### Redis

- Treated as rebuildable: position cache repopulates from Mongo on
  miss; pub/sub channel is fire-and-forget. No backup needed unless
  the customer adds Redis-persisted data later.

## Common incidents

### "Backend boot fails with config error"

Most production-guard failures are about secrets:

```
config: JWT_SECRET is a known-weak placeholder
config: JWT_SECRET must be >= 32 characters in production, got <N>
config: MONGO_URI lacks credentials in production (must be mongodb://user:pass@host)
config: REDIS_URL lacks credentials in production
config: LOGITRACK_IDENTITY_BACKEND=memory but ... DEMO_PASSWORD is empty or weak
```

Fix the offending env var, redeploy. Do not lower `APP_ENV` away from
`production` to bypass — the guard exists for a reason.

### "DB pool exhausted"

1. Check `logitrack_http_requests_total` — sudden spike?
2. Check long-running queries:
   ```bash
   docker exec logitrack-mongodb mongosh -u "$MONGO_ROOT_USERNAME" \
     -p "$MONGO_ROOT_PASSWORD" --authenticationDatabase admin \
     --eval 'db.currentOp()'
   ```
3. Raise `MONGO_MAX_POOL`, redeploy. Root cause is usually a missing
   index; confirm the `EnsureIndexes` log line at boot.

### "WebSocket clients disconnect every 5 minutes"

Expected: idle disconnect at `WS_IDLE_TIMEOUT` (default 5 min).
Clients must send `{"op":"ping"}` or respond to server pings.

### "OSRM unreachable / 502"

- OSRM is opt-in. With `OSRM_BASE_URL` empty (kit default) the
  optimiser silently falls back to a great-circle estimate after a
  one-shot WARN.
- If OSRM is configured and unreachable, calls fall back to
  straight-line on a per-call basis (`source: "fallback"` in the
  response).
- The HTTP client refuses redirects, so OSRM serving 30x will be
  treated as an upstream error → fallback. By design.

### "Chain-of-custody verification fails"

Run the verifier:

```go
ok, brokenSeq, err := services.VerifyChain(records)
```

If `ok==false`, the first broken sequence is `brokenSeq`. Inspect
the `chain_of_custody` collection for manual edits. **Never patch
the historical record.** Append a `CustodyException` entry with the
actor explaining the drift.

### "RENTRI vidimazione blocked"

- Default adapter is `rentri.QueuedStub`; if you see "rentri_failed"
  errors, the stub itself is buggy or the FIR shape is invalid.
- Live HTTP adapter (when wired) errors map to the RENTRI sandbox /
  production endpoint behaviour. Check the log for the upstream
  status code.
- Idempotency keys are deterministic per `(tenantID, firID)` — the
  same FIR retried yields the same numero.

## Emergency procedures

### Emergency stop
```bash
docker compose stop logitrack-backend logitrack-frontend logitrack-simulator
# Mongo + Redis stay up; no data loss
```

### Emergency rollback
```bash
docker compose pull logitrack-backend:<previous-tag>
docker compose up -d --no-deps logitrack-backend
```

### Full teardown (data preserved)
```bash
docker compose down
# volumes logitrack-mongo-data and logitrack-mongo-config are kept
```

### Full teardown (data WIPED — destructive)
```bash
docker compose down -v
# Volumes deleted. ONLY in test / staging.
```

## Maintenance cadence (per-fork suggestion)

| Cadence | Task |
| --- | --- |
| Daily | Mongo dump + off-host copy. Monitor `/api/health`. |
| Weekly | Review `audit_log` dropped-record counter. Review error logs. |
| Monthly | Restore drill on fresh container. `services.VerifyChain` audit on a sample. Review `govulncheck` output. |
| Quarterly | Rotate `JWT_SECRET`. Rotate Mongo + Redis passwords. Review `LOGITRACK_IDENTITY_DEMO_PASSWORD` (or move to IDP). |
| On regulation change | Update `ITALIAN-COMPLIANCE.md`, push patch to all forks via the kit changelog. |
