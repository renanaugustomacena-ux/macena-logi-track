# LogiTrack — Operations Runbook

## Daily
- Automated CI on every commit (`.github/workflows/ci.yml`).
- Trivy rescan of production image nightly via bot.
- Automated Mongo + oplog snapshot 02:00 Europe/Rome.
- Alertmanager daily summary to the on-call channel.

## Smoke-test (manual, ~3 min)
```bash
curl -sf http://localhost:8080/api/health | jq  # expect status:ok, deps populated
curl -sf http://localhost:8080/metrics | head -5 # expect logitrack_build_info
docker compose --profile demo up -d --wait
sleep 30
curl -sf http://localhost:8080/api/v1/shipments  -H "Authorization: Bearer $DEMO_JWT" | jq '.items | length'  # expect 3
```

## Backup & restore

### MongoDB
- **Backup:** `docker exec logitrack-mongodb mongodump --db logitrack --out /data/backup/$(date +%F)` (daily full).
- **Oplog archiving:** `--oplog` flag on mongodump.
- **Restore:** `mongorestore --drop /data/backup/YYYY-MM-DD/logitrack`.
- **Test:** monthly restore drill (see OPERATIONS-CADENCE).

### Redis
- Treated as rebuildable; the cache repopulates from Mongo on miss.

## Alert routing
- `error_rate > 1%` 5 min → warning → on-call channel.
- `error_rate > 5%` 5 min → critical → page on-call.
- `p95_latency > slo * 1.5` 10 min → warning.
- `unauthenticated_401_rate > 10%` 5 min → warning (brute-force probe).
- `ws_connections_active` flapping > 50% 15 min → warning.
- `database_pool_exhaustion` → critical.

## Common incidents

### "DB pool exhausted"
1. Check `logitrack_http_requests_total` trend — sudden spike?
2. `docker exec logitrack-mongodb mongosh --eval 'db.currentOp()'` — long-running queries?
3. Raise `MONGO_MAX_POOL`, redeploy. Root cause usually a missing
   index; confirm indexes via `EnsureIndexes` log line at boot.

### "WebSocket clients disconnect every 5 minutes"
- Expected: idle disconnect at `WS_IDLE_TIMEOUT` = 5 min.
- Clients must send `{"op":"ping"}` or respond to server pings.

### "OSRM 502"
- `OSRM_BASE_URL` unreachable. Fall back to in-memory estimate is
  automatic; `source="fallback"` in the response announces this.
- If the public demo is rate-limited, switch to the self-hosted OSRM
  (production deployment) by updating `OSRM_BASE_URL` and adding the
  host to `OSRM_ALLOWED_HOSTS`.

### "Chain-of-custody verification fails"
- Run `services.VerifyChain` audit CLI (planned `cmd/verifychain`).
- Identify first-broken sequence; inspect `chain_of_custody` collection
  for manual edits. Appendicate a `CustodyException` entry with the
  actor explaining the drift. NEVER patch the historical record.

## Emergency shutdown
```bash
docker compose down
# or, leave Mongo/Redis running:
docker compose stop logitrack-backend logitrack-frontend logitrack-simulator
```

## Emergency rollback
```bash
docker compose pull logitrack-backend:<previous-tag>
docker compose up -d --no-deps logitrack-backend
```

## Telepass / AISCAT code refresh
- Cadence: quarterly.
- Source: https://www.aiscat.it → gazzettino.
- Manual import into `telepass_codes` collection via a Mongo script
  (see `scripts/refresh-telepass.js`, planned).

## MODUS_OPERANDI word count
- G12 requires ≥ 13,000 words. Verify with
  `wc -w docs/MODUS_OPERANDI.md` before tagging a release.
