# LogiTrack — Competitor Landscape (research 2026-04-29)

Source-attributed snapshot of the Italian gestionale rifiuti and TMS
competitive landscape as of April 2026. Used to set the bar for our
own UI polish, feature scope, and pricing transparency. NOT for
publication: this document is an internal sales tool.

## TL;DR

- **The Italian competitor bar is much lower than expected.** Public
  product pages of TeamSystem Waste 360, Modular SVFOR02,
  SistemaRentri, Tesisquare TMS all share the same pattern: vague
  feature lists, no real product screenshots, no public pricing,
  dated 2000s-aesthetic for the smaller vendors.
- **No competitor publishes a working demo URL.** Every CTA is
  "request a free consultation" or "contattaci". Our public demo
  on GitHub Pages is already a differentiator.
- **No competitor publishes pricing transparently except Modular**,
  whose pricing turns out to be far lower than rumoured (€245-695
  setup + €95-345/semester, NOT the €1.311 figure circulating).
- **What needs polishing in LogiTrack to compete**: KPI cards on
  the dashboard, scadenze panel (Albo / autorizzazioni / copia
  produttore 90gg), recent activity feed, denser demo data, a
  consistent visual rhythm. None of these competitors does these
  better than Tailwind defaults — getting a clean 7/10 closes the
  visual gap.
- **What to skip**: AI-agent claims (project44 territory),
  enterprise scorecards, multi-municipality consortium reporting.
  Out of ICP.

---

## 1. Italian rifiuti software competitors

### 1.1 TeamSystem Waste / Waste 360

URL: <https://www.teamsystem.com/aziende/software-gestione-rifiuti/>

| Aspect | Finding |
|---|---|
| Position | "Soluzioni affidabili e conformi alla normativa per la gestione dei rifiuti e RENTRI" |
| Two products | **TeamSystem Waste** (operational) and **TeamSystem Waste 360** (advanced/integrated, mid-large enterprise) |
| Features | Digital register + FIR management, RENTRI compliance, integration with TeamSystem Enterprise ecosystem, modular architecture, "centralised information management" |
| Screenshots | Two generic product images, no captions, no real interface views |
| Pricing | **Not public.** "Varies based on functionality"; CTA = "consulenza gratuita senza impegno" |
| Target | Mid-to-large enterprise, ERP-bundle preferred |
| Style | Conservative Italian corporate, FAQ-driven |
| Polish | Medium |

**LogiTrack delta**: We target 5-30 mezzi PMI which is below their
sweet spot. They want to bundle their ERP — we don't compete on
ERP at all. Our pricing is public and project-based (€18-48k
one-off + €350-1500/mese retainer), against their sales-led "varies".

### 1.2 Modular SVFOR02 — RIFIUTI: FORMULARIO E CARICO SCARICO

URL: <https://www.modularsoftware.it/main.php?pagina=info2&cod_prog=SVFOR02>

| Aspect | Finding |
|---|---|
| Position | "Specifico per la tua attività di rifiuti, rapido e intuitivo" |
| Pricing (public!) | **Online software**: €245 setup + €95/semestre (first free) — i.e. **~€190/year ongoing**. **Online + website**: €295 + €155/sem. **Online + website + app**: €695 + €345/sem |
| Features | FIR generation (VI.VI.FIR digital cert), registro carico/scarico, multi-warehouse (4 default), tracking by lot/expiration, vehicle/transporter, CER tables, barcode |
| Reporting | FIR printing, registro printing, subject listings, waste schedules |
| Screenshots | None of actual interface — only stock office photography |
| RENTRI | **Not mentioned.** Likely lagging |
| Target | Small operators, single-warehouse SMEs |
| Style | Mid-2000s aesthetic, dense text, character encoding issues (€ shown as `�`) |
| Polish | **5/10** by their own visual standards |

**LogiTrack delta**: We are explicitly RENTRI-ready (they aren't on
the public page). We're more expensive (€18k+ vs €245+€95/sem) but
deliver a customer-owned, customer-installed application with source
code intestato al cliente. They sell a hosted multi-tenant subscription
with no source-code option. Different product, different segment.

### 1.3 SistemaRentri — gestionale rifiuti con dashboard

URL: <https://www.sistemarentri.it/funzioni-software-rifiuti/dashboard/>

| Aspect | Finding |
|---|---|
| Position | "Analisi statistico-grafica della movimentazione dei rifiuti aziendali" |
| Dashboard | Tabella riassuntiva per CER + grafici dei dati tabulati |
| Filters | Impianto destino, trasportatore, classificazione R/D, flag pericolosi |
| Example shown | Filtraggio "rifiuti pericolosi movimentati verso impianto Eurorecuperi Modena" |
| KPI cards / alerts | **None.** No real-time alerts, no threshold metrics, no widget cards |
| Reporting | Filter-driven, manual export |

**LogiTrack delta**: Their dashboard is "filter the data and see a
chart". Ours can do better with: scadenze panel (Albo, autorizzazioni,
copia produttore 90gg), recent activity feed, KPI cards (FIR
pendenti, mezzi attivi, FIR vidimati questo mese, peso totale).

### 1.4 Other Italian rifiuti vendors mentioned

- **VAR4Team** — TeamSystem reseller, focus on cloud + RENTRI.
- **Esaedro** — TeamSystem reseller, "ideale per digitalizzare".
- **DataConsult** — TeamSystem Enterprise reseller.
- **Catamacro** — TeamSystem reseller, RENTRI compliance.
- **Winwaste / Ambiente.it / Rifiutoo / Ecofacile** — couldn't find
  product pages with screenshots. Same opacity pattern.
- **EcoWasteManagement (ecowastemanagement.it)** — Italian startup
  focused on **municipalities** (Comuni), not trasportatori. ARERA
  + TARIP billing. **OUT of our ICP.**

## 2. Italian TMS competitors

### 2.1 Tesisquare TESI TMS

URL: <https://www.tesisquare.com/it/tesisquare-platform/tesi-tms>

| Aspect | Finding |
|---|---|
| Position | "Riduce i costi, ottimizza i servizi, condivide dati" |
| Modules | Planning (route optimisation), Collaboration (carrier engagement), Execution (event tracking), Costing (real-time freight rates) |
| Screenshots | None — high-level messaging only |
| Pricing | Not public |
| Target | 3PL, forwarders, carriers, shippers — mid-to-enterprise |
| Polish | Professional Italian copy, no proof-points (no testimonials, no metrics, no UI mockups) |

**LogiTrack delta**: Tesisquare is enterprise. We're PMI 5-30 mezzi.
Different segment.

### 2.2 Other TMS

- **CTSI Global TMS** — End-to-end shipper/3PL TMS. Enterprise.
- **Stesi Consulting** — Odoo-based TMS. Custom dev studio.
- **Tesi Group, GTS, AS-Software, Mertel** — no public product pages
  found. Sales-led entirely.

## 3. Enterprise visibility platforms (reference, OUT of league)

### 3.1 project44

URL: <https://www.project44.com>

| Aspect | Finding |
|---|---|
| Position | "Decision Intelligence Platform" |
| Scale claims | 3.7T data points/year, 700M+ events/day, 259K carriers, 1M+ facilities |
| Customer outcomes | $16M cost avoidance (Baltimore Bridge rerouting), $12M annual savings, 99% reduction detention/demurrage, 86% reduction gate wait |
| Dashboard | KPI cards (shipment SN/weight/dimensions), agent-monitoring panels, performance scorecards, "AI agents" doing autonomous decisions |
| Style | Modern, data-heavy, AI-forward |

**LogiTrack delta**: NOT a competitor — different universe. But the
**KPI card density and "actionable signals" framing** is worth
borrowing for our dashboard. Strip the AI-agent claims (we don't have
them and don't pretend).

### 3.2 Shippeo / FourKites

Couldn't fetch direct product pages. Same enterprise-visibility
segment as project44.

## 4. What this means for LogiTrack

### 4.1 The visual polish bar to clear

Specifically what LogiTrack's dashboard needs to have at parity with
even the lower-end Italian competition (= "looks professional in a
sales meeting"):

1. **KPI cards** at the top of the home view. Suggested:
   - Spedizioni attive (in_transit + delayed)
   - Spedizioni in ritardo
   - FIR pendenti (vidimato + in_transito)
   - Scadenze imminenti (Albo / autorizzazioni / copia produttore)
2. **Scadenze panel** — a list of upcoming deadlines, each as a
   clickable card with severity colour. None of the Italian
   competitors does this well.
3. **Recent activity feed** — last 10 audit-log events, human-friendly
   labels, time-ago format.
4. **One small chart** on the home — sparkline of FIR per giorno
   ultimi 30, OR donut FIR per stato, OR map heatmap of where the
   mezzi have been.
5. **Consistent typography + spacing** — current SPA uses Tailwind
   defaults; a 30-line `tokens.css` with 4 type sizes + 6 spacings
   + 3 radii + 4 shadows would close 60% of the polish gap.
6. **Denser demo data** — 12-15 spedizioni, 8-10 FIR (a mix of
   states), 4-5 trasportatori, 6-8 produttori, 3-4 destinatari.
   Real-feel quantity for the demo.

### 4.2 What we deliberately skip

- **AI / ML claims of any kind.** We don't ship them. project44 can.
- **"Real-time agent orchestration".** project44 territory.
- **Carrier scorecards / performance benchmarking.** Enterprise.
- **Multi-municipality consortium dashboards.** EWM territory.
- **ARERA / TARIP municipal billing.** Wrong segment.
- **Stock business stock-photography.** Modular and Tesisquare lean
  on this; it dates them. Use real Verona / Brennero photography
  or none at all.
- **"Conformità" as the only narrative.** TeamSystem leads with this
  because they have nothing else. We have RENTRI + tracking + custody
  chain + IP transparency.

### 4.3 Technical-debt items worth touching during enhancement

From `TECHNICAL-DEBT.md`, the items that matter for a sales-quality
demo:

| # | Item | Effort | Sales impact if fixed |
|---|---|---|---|
| 11 | `error.Error()` echoed in 500 paths | 1 day | High — demo accidentally reveals Mongo internals on bad input |
| 10 | RequireRole not applied on every mutating endpoint | 0.5 day | Low — demo never exercises non-admin roles |
| 5 | Audit log lossy under backpressure | 1 day | Medium — talking point in compliance review |
| 13 | xFIR XSD placeholder | 2 days | Defer until first RENTRI live customer |
| 6 | Keyset pagination | 1 day | Low — only relevant >100k spedizioni |

### 4.4 Pricing positioning

Modular's actual pricing (€245 setup + €95/sem ≈ €435/year ongoing)
is NOT comparable to ours (€18-48k one-off + €350-1500/mese retainer).
Different products, different segments. We are a **bespoke
project + ongoing maintenance**, they are a **subscription to a
hosted product**.

Honest pitch to a prospect comparing both:
- "Modular costa molto meno il primo anno. Dopo 5 anni hanno
  incassato circa €2.500. Tu hai pagato 5 anni di abbonamento e non
  possiedi niente.
- Con LogiTrack paghi più all'inizio (€18-25k) ma dopo 5 anni hai
  spesso speso meno (€18-25k + €750×60 = circa €60k vs €2.500), MA
  hai un'applicazione tua, sul tuo server, brandizzata, con il
  codice sorgente intestato a te.
- È un trade-off sincero: se ti basta un FIR digitale generico,
  Modular va bene. Se vuoi un software che sia tuo, brandizzato per
  la tua azienda, integrato con la tua telematica e i tuoi flussi,
  serve un progetto."

## 5. Sources

- TeamSystem Waste / Waste 360 — <https://www.teamsystem.com/aziende/software-gestione-rifiuti/>
- Modular SVFOR02 — <https://www.modularsoftware.it/main.php?pagina=info2&cod_prog=SVFOR02>
- VAR4Team case study — <https://www.var4team.it/casi-di-successo/il-software-cloud-per-la-gestione-dei-rifiuti-rentri/>
- SistemaRentri dashboard — <https://www.sistemarentri.it/funzioni-software-rifiuti/dashboard/>
- EcoWasteManagement — <https://www.ecowastemanagement.it>
- Tesisquare TESI TMS — <https://www.tesisquare.com/it/tesisquare-platform/tesi-tms>
- project44 — <https://www.project44.com>

Research conducted via DuckDuckGo (per global rules), 2026-04-29.
