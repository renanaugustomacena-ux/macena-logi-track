# LogiTrack — Client Fork Recipe

Mechanical steps from "the customer signed the preventivo" to "the
customer has a working deployed system". Run these in order. Each step
is independently reversible.

The recipe assumes:

- Kit upstream lives at `https://github.com/<your-account>/logitrack.git`
  (private repo on the freelancer's GitHub account).
- The customer is `acme-trasporti`. Replace this with the actual
  customer slug throughout.
- The customer wants `rifiuti` (vertical 2). Adjust for `logistics`
  (vertical 1) by replacing `rifiuti` with `logistics` where relevant.

## Day 0 — pre-engagement checklist

Before kicking off the fork, confirm:

- [ ] Preventivo signed (vedi
  [`FREELANCER-COMMERCIAL-MODEL.md`](FREELANCER-COMMERCIAL-MODEL.md)).
- [ ] Customer's hosting choice decided: Aruba IT VPS / Hetzner DE /
  Hetzner FI / on-premise / customer Kubernetes.
- [ ] Customer's domain name: `logistica.acme.it` or `gestionale.acme.it`.
  Customer must own DNS and TLS (Let's Encrypt or commercial cert).
- [ ] Customer's IDP (if any): Keycloak / Azure AD / Okta / "use the
  built-in memory store". Memory store = the freelancer seeds one
  demo user; customer adds users via `mongosh`.
- [ ] Customer's RENTRI status (rifiuti only): does the trasportatore
  already hold the certificato digitale? If yes, sandbox or prod?
- [ ] PEC + email contacts for the operations handover.

## Day 1 — fork setup

```bash
# 1. Create the per-customer repo on a private git host.
#    Recommended: customer's own GitHub Organization (so they own it
#    even before contract end). Failing that, the freelancer's
#    GitHub Org with a "transfer to customer" agreement.
gh repo create acme-trasporti/logitrack-acme --private --description "LogiTrack fork for ACME Trasporti"

# 2. Clone the new (empty) repo locally.
mkdir -p ~/clients/acme-trasporti
cd ~/clients/acme-trasporti
git clone git@github.com:acme-trasporti/logitrack-acme.git
cd logitrack-acme

# 3. Add the kit as a subtree. Pin to the latest tagged release.
KIT_REF=v0.3.0   # current released kit version
git subtree add --prefix=kit https://github.com/<your-account>/logitrack.git $KIT_REF --squash

# 4. Scaffolding for overlay + ops.
mkdir -p overlay/{backend,frontend,env,docs}
mkdir -p ops/{deploy,backup,monitoring}

# 5. Top-level docker-compose.yml that composes kit + overlay.
#    (Sample skeleton below; tune per customer)
cat > docker-compose.yml <<'EOF'
# ACME Trasporti deployment — composes kit + overlay
# Source of truth for kit version: ./kit/
# Source of truth for overlay: ./overlay/
services:
  logitrack-backend:
    extends:
      file: kit/docker-compose.yml
      service: logitrack-backend
    env_file:
      - overlay/env/backend.env
  logitrack-frontend:
    extends:
      file: kit/docker-compose.yml
      service: logitrack-frontend
  logitrack-mongodb:
    extends:
      file: kit/docker-compose.yml
      service: logitrack-mongodb
  logitrack-redis:
    extends:
      file: kit/docker-compose.yml
      service: logitrack-redis
EOF

# 6. Initial commit
git add .
git commit -m "chore: scaffold ACME fork from kit $KIT_REF"
git push -u origin main
```

## Day 2 — environment configuration

```bash
# 1. Copy the kit's .env.example as the starting point.
cp kit/.env.example overlay/env/backend.env

# 2. Generate strong secrets.
JWT_SECRET=$(openssl rand -hex 32)
MONGO_ROOT_PASSWORD=$(openssl rand -base64 24)
REDIS_PASSWORD=$(openssl rand -base64 24)
DEMO_PASSWORD=$(openssl rand -base64 18)  # >=12 chars, NIST 800-63B

# 3. Edit overlay/env/backend.env with the generated values + customer
#    specifics. At minimum:
#    - APP_ENV=production (for staging set =staging)
#    - JWT_SECRET=<generated>
#    - MONGO_ROOT_PASSWORD / REDIS_PASSWORD as above
#    - HTTP_CORS_ALLOWED=https://logistica.acme.it
#    - HTTP_WS_ORIGINS=https://logistica.acme.it
#    - LOGITRACK_IDENTITY_BACKEND=memory  (or "disabled" if customer IDP)
#    - LOGITRACK_IDENTITY_DEMO_USER=admin@acme.it
#    - LOGITRACK_IDENTITY_DEMO_PASSWORD=<generated>
#    - LOGITRACK_IDENTITY_DEMO_TENANT=acme-trasporti
#    - LOGITRACK_IDENTITY_DEMO_BREACH_ACK=true   (after verifying with HIBP)
#    - SEED_DEMO=false   (production: never seed demo data)

# 4. Encrypt the env file before committing. Use `sops` or `age`.
sops --encrypt --age <recipient-key> overlay/env/backend.env > overlay/env/backend.env.enc
git add overlay/env/backend.env.enc
echo "overlay/env/backend.env" >> .gitignore
git add .gitignore
git commit -m "chore: seal production env with sops/age"
```

## Day 3 — VPS provisioning (Aruba IT example)

```bash
# 1. Provision a VPS (8 GB RAM minimum for kit + Mongo + Redis + nginx).
#    Aruba Cloud Console → Crea Server → Cloud VPS Smart →
#    Plesk-free, SSD, Italy region (IT-MI or IT-AR), Ubuntu 24.04 LTS.

# 2. SSH in, harden basics.
adduser deploy
usermod -aG sudo deploy
sed -i 's/^#PasswordAuthentication.*/PasswordAuthentication no/' /etc/ssh/sshd_config
systemctl restart sshd

# 3. Firewall: 22, 80, 443 only from outside.
ufw allow 22/tcp
ufw allow 80/tcp
ufw allow 443/tcp
ufw --force enable

# 4. Docker engine + compose.
curl -fsSL https://get.docker.com | sh
usermod -aG docker deploy

# 5. Caddy as reverse proxy (auto Let's Encrypt).
apt install -y caddy
cat > /etc/caddy/Caddyfile <<'EOF'
logistica.acme.it {
    reverse_proxy /api/* localhost:8080
    reverse_proxy /metrics localhost:8080  # access-restrict at firewall layer
    reverse_proxy /* localhost:5174
    encode gzip
}
EOF
systemctl reload caddy

# 6. Backup script (run via cron daily at 03:00).
cat > /home/deploy/backup-mongo.sh <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
DATE=$(date +%F)
docker exec logitrack-mongodb mongodump \
  --authenticationDatabase admin \
  -u "$MONGO_ROOT_USERNAME" -p "$MONGO_ROOT_PASSWORD" \
  --db logitrack --out /data/backup/$DATE
# Off-host copy
rsync -az /data/backup/$DATE/ acme-backup@offsite.example:/backups/acme/$DATE/
# Retain 14 days locally, 90 days off-host
find /data/backup -type d -mtime +14 -exec rm -rf {} \;
EOF
chmod +x /home/deploy/backup-mongo.sh
echo "0 3 * * * deploy /home/deploy/backup-mongo.sh" | tee /etc/cron.d/logitrack-backup
```

## Day 4 — first deployment

```bash
# On the VPS, as user `deploy`:
git clone git@github.com:acme-trasporti/logitrack-acme.git
cd logitrack-acme

# Decrypt env
sops --decrypt overlay/env/backend.env.enc > overlay/env/backend.env

# Pull images / build
docker compose pull || true
docker compose build

# Boot
docker compose up -d

# Wait for health
sleep 20
curl -sf http://localhost:8080/api/health | jq

# Smoke test login
TOKEN=$(curl -sf http://localhost:8080/api/v1/auth/login \
  -H 'content-type: application/json' \
  -d "{\"username\":\"admin@acme.it\",\"password\":\"$DEMO_PASSWORD\"}" \
  | jq -r .accessToken)
echo "$TOKEN" | head -c 40 ; echo
```

## Day 5-15 — customisation period

Per-customer customisation goes in `overlay/`. Examples:

### Frontend branding

```bash
# Copy customer logo into the SPA assets.
cp ~/Downloads/acme-logo.svg overlay/frontend/public/logo.svg

# Override the SPA's brand component via Vite alias (set in
# overlay/frontend/vite.config.overlay.ts).
```

### Custom telematics adapter

```bash
mkdir -p overlay/backend/internal/integrations/acme-telematics
# Build the adapter that translates ACME's specific provider webhook
# into the kit's POST /api/v1/shipments/{id}/waypoints shape.
```

### Custom CER override

```bash
# If the customer handles a niche EER subset, document it but do not
# override the kit's CER catalogue. The kit catalogue is the
# authoritative D.U. 2014/955 source.
```

## Day 16-21 — pilot operations

- Real users log in.
- Real shipments / FIR run end-to-end.
- Issue list maintained in the customer's GitHub Issues.
- Daily check of `/api/health` from a monitoring service (UptimeRobot
  free tier is enough for a single VPS).

## Day 22-28 — handover

```bash
# 1. Transfer GitHub repo ownership to the customer.
#    GitHub UI → Settings → Transfer ownership → acme-trasporti.

# 2. Hand over secrets out-of-band:
#    - sops/age private key
#    - VPS SSH key
#    - Aruba Cloud Console credentials
#    - Caddy config + Let's Encrypt account info
#    Use 1Password / Bitwarden Vault export, encrypted USB stick, or
#    the customer's preferred secret-sharing method. NEVER email
#    plaintext.

# 3. Document the customer's runbook addendum (per-customer
#    quirks) in overlay/docs/RUNBOOK-ACME.md.

# 4. Final transfer email lists every credential, every URL, every
#    monthly cost (VPS, domain, backups).
```

## Day 29+ — retainer

Switch to monthly retainer (vedi
[`FREELANCER-COMMERCIAL-MODEL.md`](FREELANCER-COMMERCIAL-MODEL.md)).
Patch propagation cadence in
[`KIT-PLAYBOOK.md`](KIT-PLAYBOOK.md) §4.

## Checklist condensed

```
[ ] Preventivo signed
[ ] Hosting decided
[ ] Domain + TLS owned by customer
[ ] IDP decision (memory / external)
[ ] RENTRI status (rifiuti only)
[ ] Per-customer GitHub repo created
[ ] Kit pulled as subtree, pinned to a tagged release
[ ] Overlay scaffolding committed
[ ] Secrets generated + sealed with sops/age
[ ] VPS provisioned + hardened
[ ] Caddy + TLS auto-renewing
[ ] Backup cron + off-host copy
[ ] First deployment green health check
[ ] Customisation work merged
[ ] Pilot users on real data
[ ] Handover: repo, secrets, docs, monthly costs
[ ] Retainer scheduled
```
