# LogiTrack — Domain Modules

> Architecture note. This document defines what a "domain module"
> is, what its contract with the platform looks like, and the
> roadmap of modules planned for the kit.

## Why modules

LogiTrack is a kit, not a product. The platform layers (auth,
audit, chain-of-custody, observability, repositories, middleware,
HTTP handlers for cross-vertical concerns) are reused unchanged
across every customer engagement. The vertical-specific layer —
entities, state machines, validators, integrations with regulators
that only matter for one vertical — lives in a domain module.

Each customer engagement = platform + 1 module + per-customer
customization. The module is the unit of vertical specialization.
This is what lets a single freelancer ship a customer-fit operations
platform in 6-10 weeks instead of 6 months: 70% of the codebase is
already production-grade.

## Module contract

A domain module under `backend/internal/modules/<name>/` MUST:

1. **Own its entities.** All persistent types specific to the
   vertical (Cantina, WineLot, AcciseDocument for wine; Frantoio,
   OliveBatch, OilLot for oil; etc.) live in the module package.
2. **Be a leaf in the dependency graph.** A module imports the
   platform (config, audit, problem, middleware, repository
   primitives) but NEVER imports another module. Vertical
   modules do not know about each other.
3. **Expose domain validators and state-machine guards** as pure
   Go (no I/O). Persistence and HTTP wiring live in the platform.
4. **Document its Italian regulatory anchors** in a sibling
   `docs/MODULE-<NAME>.md` so a customer-facing audit can name
   the laws and articles the module implements.
5. **Ship with its own integration sub-packages** for regulators
   or providers that only matter for that vertical (e.g. EMCS for
   wine, SIAN Vitivinicolo for wine, frantoio-specific cooperative
   gestionali for oil).

A domain module SHOULD:

- **Compose, not duplicate.** A wine shipment IS a logistics
  shipment plus wine-specific dispatch metadata. Reuse the
  logistics types where the abstraction holds; specialize only
  the parts that differ.
- **Keep its public surface narrow.** A module exports what the
  platform's HTTP/CLI layers need, not internal helpers.

## Layer map

```text
backend/internal/
├── audit/                       (platform — append-only log)
├── config/                      (platform — env loader)
├── handlers/                    (platform — HTTP layer; mounts module routes)
├── integrations/                (platform — shared external clients)
│   ├── aida/                    (Italian customs — used by logistics + rifiuti when transfrontaliero)
│   ├── albo/                    (Italian carrier register — used by logistics; rifiuti has its own albo gestori check)
│   ├── httpretry/               (shared retry+backoff)
│   ├── rfi/                     (Italian rail — used by logistics)
│   └── telepass/                (Italian tolls — used by logistics + rifiuti)
├── middleware/                  (platform — Gin middlewares)
├── modules/
│   ├── logistics/               (MODULE 1 — freight visibility)
│   │   ├── doc.go
│   │   ├── shipment.go
│   │   ├── tracking_event.go
│   │   ├── chain_of_custody.go
│   │   ├── compliance.go        (plate, ADR, ATP, Telepass codes)
│   │   ├── driver.go
│   │   ├── geofence.go
│   │   └── vehicle.go
│   └── rifiuti/                 (MODULE 2 — Italian SME waste-transport, RENTRI-ready)
│       ├── doc.go
│       ├── compliance.go        (CER/EER + Albo categoria + impianto operazione)
│       ├── party.go             (Produttore, Trasportatore, Destinatario + Albo guards)
│       ├── fir.go               (Formulario Identificazione Rifiuti + state machine)
│       ├── registro.go          (registro cronologico carico/scarico, append-only)
│       └── rentri/              (sub-package: RENTRI client adapter + queued stub)
│           ├── doc.go
│           ├── endpoints.go     (sandbox + production base URLs)
│           ├── client.go        (Client interface + xFIR types)
│           └── queued_stub.go   (default adapter until SPID/CNS cert lands)
├── obs/                         (platform — Prometheus + counters)
├── problem/                     (platform — RFC 7807 errors)
├── repository/                  (platform — Mongo + Redis primitives;
│                                 collection-specific accessors will move
│                                 into the owning module in a future commit)
└── services/                    (platform — current location of logistics
                                  services; will move into modules/logistics
                                  in a follow-up commit so the boundary is
                                  clean)
```

The current commit moves only the entity layer to honour the
module contract. Services, repository accessors and handlers
follow in subsequent commits because their move is more
invasive — separating mechanical refactor from boundary
clarification keeps each diff reviewable.

## Module roadmap

| # | Module | Status | First-customer profile |
|---|---|---|---|
| 1 | `logistics` | shipped | (not customer-led — historical baseline) |
| 2 | `rifiuti` | shipped (entities + RENTRI client stub + tests) | Trasportatore rifiuti speciali SME, 5-30 mezzi, Albo Cat 4/5/8, Verona Sud / Mantova / Brescia corridor — first design-partner target: FRO S.r.l., Mozzecane (VR) |
| 3 | `oil` | future | frantoio in Garda or Veneto, DOP/IGP olive oil |
| 4 | `cheese` | future | caseificio in Asiago / Grana zone |
| 5 | `aquaculture` | future | mitilicoltura cooperativa, costa veneta or ligure |
| 6 | `funeral` | future | impresa funebre famigliare, network regionale |

The wine vertical was originally listed as module 2 but has been
permanently delegated to the sibling project TraceVino (Python /
FastAPI), which already implements SIAN, MVV-E, e-label, GS1+NFC,
HACCP, the lab adapters, and the eleven Verona DOC/DOCG
disciplinari. LogiTrack does not duplicate that work; the kit
boundary keeps the two products independently shippable to
non-overlapping customer sets.

Each row is a separate go-to-market wedge. They share the
platform; they do not share entities, state machines or
regulatory integrations.

## Adding a new module

Mechanical recipe (will be tightened into a generator later):

1. Create `backend/internal/modules/<name>/doc.go` with the
   `package <name>` declaration and the placeholder entities.
2. Create `docs/MODULE-<NAME>.md` documenting the regulatory
   anchors and the first-customer profile.
3. Add the module-specific Italian compliance entries to
   `docs/ITALIAN-COMPLIANCE.md` under a new section.
4. Add the module-specific integrations under
   `internal/integrations/<provider>/` (only when that provider
   is genuinely shared across modules) or under
   `internal/modules/<name>/<provider>/` (when it is module-
   specific, like RENTRI for rifiuti).
5. Wire the module into the composition root in
   `cmd/server/main.go` and `internal/handlers/routes.go`.

Each step is reviewable on its own. No big-bang refactor.

## Why this matters

The kit architecture lets a single engagement land in 6-10 weeks
because most of the work is already done. It also lets a customer
who wants something LogiTrack does not yet support get a clear
answer: "we will add a `<name>` module — that is a 4-6 week build
at €X". The module boundary is the unit you quote, the unit you
deliver, and the unit you reuse on the next customer in the same
vertical.

The platform is the moat. The modules are the wedges.
