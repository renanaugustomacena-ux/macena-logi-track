# LogiTrack — Kit Playbook

The operating manual for the maintainer (Renan Augusto Macena). Read
this before starting a new client engagement, before merging a feature
into kit upstream, and before deciding "is this code generic enough to
live in the kit, or customer-specific enough to live in the overlay?"

## 1. What a kit is, and what it isn't

LogiTrack is **a kit**. That means:

- **One canonical upstream repository** (`main` branch on GitHub) that
  represents the freelancer's reusable baseline. The kit contains
  exactly the code that is generic across all current and reasonably-
  expected future customers.
- **Per-customer forks** that take the kit at a specific commit, layer
  customer-specific code on top in an `overlay/<customer>/` directory
  (or via env-driven configuration), and deploy. The customer owns the
  fork. The freelancer keeps a backup mirror only for retainer work.

The kit is **NOT**:

- A SaaS product. There is no shared cloud. There is no canone
  mensile per posto. There is no multi-tenant production deployment
  the freelancer operates.
- A platform competing with TeamSystem / Tesi / Modular / Passepartout.
  Those vendors have salesforces, customer-success teams, 24/7 NOCs.
  This is a freelancer kit; the operating model is fundamentally
  different.
- A consulting deliverable that is rewritten from scratch for each
  customer. The kit baseline is real, real customers run it, real
  patches propagate.

## 2. Kit invariants (rules that must not break)

These are the load-bearing claims about the kit. Every commit to
`main` must preserve them.

1. **Modules never import each other.** `modules/logistics` and
   `modules/rifiuti` are leaves. Adding a third vertical = a third
   leaf. Cross-vertical helpers go into `internal/<helper>` at the
   platform level, not into a module.

2. **Tenant scoping is a repository property, not a caller option.**
   Every Mongo query method must take `tenantID` as an explicit
   argument. There is no "global" lookup. There is no "list everything
   across tenants" admin endpoint in the kit core.

3. **Auth is access-token-only.** No refresh endpoint. No sliding
   session in the kit. Sliding sessions are the customer's IDP's
   responsibility (Keycloak, Azure AD, Okta) via the IdentityStore
   seam.

4. **Production refuses to boot with weak secrets.** JWT_SECRET <
   32 chars, known-weak placeholders, unauthenticated MONGO_URI /
   REDIS_URL, empty-or-weak demo identity password — all hard-fail at
   startup with a clear error.

5. **Outbound HTTP is allow-listed and redirect-blocked.** Any new
   HTTP client added to the kit must check the host against an
   allow-list before dialling AND set `CheckRedirect` to refuse
   redirects. SSRF is the threat we close at the architecture layer.

6. **Custody chain is append-only and tamper-evident.** Failing to
   compute a chain hash is a hard error, not a logged warning.
   Corrections are a new `CustodyException` entry, never an in-place
   edit.

7. **No claim ships ahead of code.** If `docs/INTEGRATIONS.md` says
   "the kit ships an X integration", the X integration must actually
   be in the codebase, exercised by tests, and reachable from a
   handler. The 2026-04-28 honesty pass purged 4 packages that violated
   this. Don't let it happen again.

8. **No fake regulatory citations.** Every Italian law / EUR-Lex /
   AgID reference in the docs must trace back to a primary source
   (normattiva.it, eur-lex.europa.eu, agid.gov.it, agenziaentrate.gov.it).
   When a citation changes (regulation amended, guideline superseded),
   update the doc; don't keep a stale reference.

## 3. Kit vs. overlay: where does this code go?

Decision tree for "I'm about to write a feature. Where does it live?"

```
Is the code useful for ≥ 2 (likely) customers?
├─ Yes → Is it specific to one vertical (logistics / rifiuti / future)?
│        ├─ Yes → modules/<vertical>/ (kit core)
│        └─ No  → internal/<helper>/  (platform layer of kit core)
└─ No  → overlay/<customer>/         (per-customer fork only)
```

Concrete examples:

| Code | Where | Why |
| --- | --- | --- |
| Plate validator (post-1994 Italian + storiche) | `modules/logistics/compliance.go` | Useful for every Italian customer in logistics |
| CER catalogue + chapter helper | `modules/rifiuti/compliance.go` | Useful for every rifiuti customer |
| RENTRI client interface + queued stub | `modules/rifiuti/rentri/` | Useful for every rifiuti customer; live HTTP adapter swappable |
| **FRO-specific custom Albo verification adapter** that calls FRO's bespoke verifier portal | `overlay/fro/` (in FRO's fork) | Specific to one customer. Keep out of kit core. |
| Customer-branded login page colours | `overlay/<customer>/frontend/` (per fork) | Branding is per-customer |
| Generic OSRM fallback to straight-line at 70 km/h | `services/route_optimizer.go` (kit core) | Generic, useful for every customer |

The temptation that breaks the kit is "I'll put this customer-specific
hack in the kit core because it's faster." Resist. Customer-specific
code in kit core means the next customer inherits it, gets confused,
and either (a) carries dead code or (b) you have to add a feature flag
to disable it. Both are kit decay.

## 4. Patch propagation discipline

A bug found in the kit must reach every customer fork. Without this
discipline, the kit dies.

### Recommended fork structure

```
customer-fork/
├── kit/                  ← git subtree of kit upstream main
├── overlay/              ← customer-specific code
│   ├── frontend/
│   ├── backend/
│   └── env/
└── docker-compose.yml    ← composes kit + overlay
```

`git subtree` (or `git submodule`, but subtree is friendlier for
distribution) lets each fork pull kit upstream patches in a single
command:

```bash
# In a customer fork
git subtree pull --prefix=kit https://github.com/renanaugustomacena/logitrack.git main --squash
```

### Cadence

- **CRITICAL / security**: push to all forks within 72 h of the kit
  upstream patch. Maintain a contact list per fork.
- **HIGH (regulatory change, e.g. RENTRI deadline shift)**: within
  2 weeks. Coordinate with the customer's retainer schedule.
- **MEDIUM / LOW (quality, performance)**: opportunistic, batch in the
  next quarterly retainer touchpoint.

### Compliance changelog

Every regulatory-related kit change ships with a `KIT-CHANGELOG-IT.md`
entry naming the law / decree that triggered it and the per-fork
upgrade steps. This is what justifies the retainer to the customer.

## 5. Promoting overlay code to kit

Sometimes a customer-specific feature turns out to be useful for ≥ 2
customers. Promotion path:

1. **Generalise the interface.** Strip every customer-specific name,
   constant, hardcoded URL.
2. **Open-source-style PR review on the kit `main`.** Even if you are
   the only reviewer, write the PR description as if a future colleague
   were reading it.
3. **Add tests.** Promotion without tests is debt acceptance.
4. **Update `INTEGRATIONS.md` / `MODULE-<X>.md` / `API.md`** to reflect
   the new kit feature.
5. **Patch every other fork.** Even forks that don't use the new
   feature should pull the kit update so they stay current.

## 6. Demoting kit code to overlay

The opposite path: a kit feature that turns out to only matter for one
customer. The 2026-04-28 honesty pass was a mass demote: AIDA, RFI,
Telepass and Albo clients were demoted from "kit core" to "future
per-customer overlay" because no customer was actually using them in
the kit core.

Do this when:

- The feature has been in kit core for > 6 months and no customer has
  asked for it.
- The feature has accreted customer-specific assumptions in code or
  config.
- Maintaining the feature in kit core is more work than rebuilding it
  per-customer when needed.

## 7. Kit deprecation policy

Removing a feature from the kit:

1. Mark it deprecated in the changelog of the next minor release.
2. Wait one minor release.
3. Remove it.

If a customer fork still uses the feature, the customer's overlay can
copy the code in. The maintainer commits to keeping forks compatible
through one minor release of deprecation, not forever.

## 8. Anti-patterns the kit must reject

- **"While I'm here" refactors.** Every change must be traceable to a
  customer ask, a regulatory change, or a security finding. Random
  cleanup goes in its own commit and PR, not bundled with feature
  work.
- **Speculative abstractions.** No "let's build a plugin system in
  case a future customer wants to ship custom modules." Build the
  third module first, generalise after.
- **Feature flags as a way to half-ship.** A feature is either in or
  out. If it's behind a flag, the flag has a removal date in the
  changelog.
- **TODO / FIXME comments in kit core.** Every known issue lives in
  `TECHNICAL-DEBT.md` with a trigger, not as a TODO that rots in the
  source.
- **Mock data shipping in production.** Demo seed is opt-in
  (`SEED_DEMO=true`), idempotent, and namespaced to a `demo-tenant`.
  Never seed real customer tenants.

## 9. Doctrine in one paragraph

The kit is the freelancer's leverage. Every customer the kit lets you
serve in 4-8 weeks instead of 6 months is paid for by the discipline
that keeps the kit honest, lean and current. When that discipline
breaks — when the kit accretes claims it doesn't back, customer-
specific code it can't promote, integrations nobody uses, regulatory
citations that have rotted — the leverage evaporates. Treat every
commit to `main` as if it has to survive the next ten customers,
because it does.
