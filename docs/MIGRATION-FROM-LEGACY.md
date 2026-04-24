# LogiTrack — Migration From Legacy

Short plans for the three commonest legacies we meet at Italian
spedizioniere. Each plan is ~1 page and is meant to anchor the
commercial conversation; a full migration engagement is quoted
separately.

---

## 1. From Zucchetti Gedriver / Trasporti

Zucchetti's road-freight bundle stores shipments in a SQL Server
database with bespoke stored procedures for dispatching, and a
Windows client for operators.

Data mapping:
| Zucchetti column | LogiTrack field | Note |
| --- | --- | --- |
| `TDOC.NUMERO` | `Shipment.Reference` | Consignment note number |
| `TDOC.CLI_COD` | `Shipment.Consignor.VATNumber` | VAT cross-reference |
| `TDOC.DEST_COD` | `Shipment.Consignee.VATNumber` |  |
| `TDOC.DATAPART` | `Shipment.ETD` | Departure date |
| `TDOC.DATAARR` | `Shipment.ETA` | Arrival date |
| `AUTM.TARGA` | `Shipment.VehiclePlate` | Validated with regex |
| `AUTI.COD_CF` | `Driver.ID` | CF not stored unless opt-in |
| `STORIA.MRN` | `Shipment.Customs.MRN` | AIDA MRN |

Plan (30 days):
1. **Week 1** — read-only SQL extract to CSV, ETL notebook to
   transform into LogiTrack JSONL, dry-run import to a staging tenant.
2. **Week 2** — dual-run: Zucchetti remains authoritative, LogiTrack
   receives telematica feeds; operators observe the new dashboard in
   read mode.
3. **Week 3** — cut dispatch to LogiTrack for the first 10%
   shipments; Zucchetti remains for legacy. Observability is KPI'd
   against parity.
4. **Week 4** — full cut-over; Zucchetti archived read-only for 24
   months per art. 2220 c.c.

---

## 2. From TeamSystem TS Enterprise Trasporti

TeamSystem's product is tightly integrated with the ERP
TS Enterprise; extraction is via the open REST API introduced in
TS release 2024.

Plan (45 days):
1. **Week 1** — API credentials + scope negotiation; test the GraphQL
   `/consignments` endpoint.
2. **Week 2–3** — sync job: pull open consignments every 15 min,
   upsert into LogiTrack via `POST /api/v1/shipments`. Duplicates
   prevented by the `(tenant_id, reference)` unique index.
3. **Week 4–5** — bidirectional sync: LogiTrack webhooks
   `shipment.updated` POST back into TS Enterprise's events API.
4. **Week 6** — training for dispatcher + warehouse ops; cut-over.

---

## 3. From Excel + WhatsApp (the most common starting point)

Yes, this is a real migration. Here is what works.

Plan (14 days):
1. **Day 1–3** — anagrafiche: mittenti, destinatari, mezzi, autisti
   into CSV. Ship a pre-built CSV template with Italian headings.
2. **Day 4–5** — upload via `POST /api/v1/imports/anagrafiche` (TBD
   in roadmap; for Mission-II era, use an admin Python script).
3. **Day 6–10** — operator training (2 sessions, 2 hours each).
   Provide a printed quick-reference card.
4. **Day 11–12** — first 3 live spedizioni end-to-end in LogiTrack.
5. **Day 13–14** — parallel operation ends; WhatsApp group becomes
   "amministrativo" only.

### Anti-patterns we have seen
- Importing 5 years of historical Excel rows — not worth the effort.
  Start with open shipments, archive legacy as files.
- Trying to match every local Excel column to a LogiTrack field —
  accept that 20% of columns were never used.
- Skipping operator training — the platform's true value emerges only
  when the person on the phone at 14:30 knows where to click.
