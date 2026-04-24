# LogiTrack — Modus Operandi

Comprehensive strategic, technical, operational, commercial and
organisational playbook for LogiTrack, a supply-chain visibility
platform engineered in and for the Verona intermodal corridor.

---

## Part I — Strategic Foundation

### 1. Executive Summary

LogiTrack is a supply-chain visibility platform purpose-built for the
Verona logistics ecosystem and the broader Brenner corridor running
from Mozzecane north through the A22 motorway into Austria, Bavaria
and the Ruhr. The product exists because an uncomfortable mismatch
persists at one of Europe's most important inland logistics nodes:
Quadrante Europa, the second-largest intermodal freight platform on
the continent, handles more than eight million tonnes of goods per
year across road and rail, yet a remarkable seventy-three per cent of
the firms that ship or carry through it still track consignments with
spreadsheets, paper consignment notes and WhatsApp groups. The
corridor between Verona, Mantova, Trento and Bolzano clears roughly
fifty-six billion euros of commercial interchange annually, and the
people who keep that trade moving, the spedizionieri, the padroncini,
the internal logistics managers of Veneto manufacturing SMEs, have
almost no unified real-time view of where their cargo is, whether it
is on schedule, and whether it has cleared the next customs checkpoint.

LogiTrack closes that gap. The platform ingests telematics feeds from
Viasat, Octo and Geotab, electronic CMR consignment notes from the
principal EDI providers, customs status from Agenzia delle Dogane
(AIDA), intermodal slot bookings from Rete Ferroviaria Italiana and
motorway transit events from Telepass, and unifies them into a single
dashboard, a single tamper-evident chain-of-custody log, a single set
of analytics on punctuality, cost per kilometre, and dwell time. Built
in Go for concurrent ingestion, Vue 3 for a map-heavy SPA, MongoDB for
flexible shipment documents and Redis for real-time fan-out, the
product is designed to run in the Italian sovereign cloud when the
tenant requires it, and to scale horizontally as volumes grow.

The primary customers are threefold: carriers and intermodal operators
of the Verona province who run twenty to three hundred vehicles; Veneto
SMEs whose logistics function is one or two people drowning in
spreadsheets; and freight forwarders organising cross-border flows
through the Brenner into the DACH region. Pricing is subscription with
a volume surcharge: a Starter plan at one hundred ninety-nine euros
per month for operators with up to ten shipments a day, a Professionale
plan at five hundred ninety-nine euros per month covering one hundred
shipments a day with the AIDA module and multi-integration support,
and an Enterprise plan quoted individually for carriers handling
thousands of consignments weekly with dedicated SLAs and sovereign
hosting. The three-year objective is two hundred paying tenants across
Veneto and the Brenner corridor, an ARR of roughly two million euros,
and an embedded position in the ecosystem that makes LogiTrack the
default starting point whenever a Verona SME decides its supply chain
deserves more than a shared Excel file.

The elevator pitch is direct: LogiTrack turns the Brenner corridor
into a legible system. Where today a logistics manager in Villafranca
refreshes five browser tabs and makes three phone calls to understand
where his truck is, tomorrow she opens one dashboard and sees the
entire fleet plotted in real time, every customs event annotated on a
timeline, every delay flagged and quantified against the contractual
SLA. The competitive advantage rests on three pillars: deep local
context (AIDA, RFI, Albo Autotrasportatori, Telepass built in, not
retrofitted), a pricing structure accessible to the SMEs that dominate
the territory, and an engineering culture that takes Italian data
sovereignty seriously enough to default to Aruba Cloud for tenants who
require it.

### 2. Market Analysis

The Italian logistics sector generates approximately one hundred
twenty billion euros of value added annually and employs about one and
a half million people; roughly eighty-five per cent of the freight
moved inside the country moves on rubber, with the remainder split
between rail, water and air. Within that landscape the Quadrante
Europa platform, operated by Consorzio ZAI on behalf of the
Municipality of Verona and the Province, is the unambiguous centre of
gravity of northern Italian intermodality. Its surface exceeds two
million square metres, it hosts more than one hundred companies, and
it handles combined rail-road shipments whose origin and destination
stretch from Rotterdam to Munich to Ljubljana to Bari. The
interregional corridor that runs from Verona along the A22 Autostrada
del Brennero, through Mantova and Trento into Bolzano and onward into
Austria and Bavaria, carries an estimated fifty-six billion euros of
commercial interchange every year; Mozzecane itself sits five
kilometres from the Nogarole Rocca motorway exit and twelve kilometres
from the Quadrante Europa freight terminal, geographically embedded in
that flow.

The Verona province is also home to the Distretto Mobile di Verona,
the furniture-manufacturing cluster whose outbound shipments feed the
rest of Europe, and the Distretto Meccanico veronese, whose precision
components flow primarily into German automotive tiers. Both clusters
generate enormous demand for logistics services: containerised
outbound, incoming just-in-time components, temperature-controlled
palletised loads, over-dimension specials. The typical operator
serving this demand is a family-owned shipper with fifteen to eighty
vehicles, turnover between two and twenty million euros, and an IT
budget that rarely exceeds one per cent of revenue. It is this
demographic, multiplied across the Veneto and lower Trentino, that
forms the addressable market for LogiTrack.

Quantitatively: Unioncamere Veneto counts roughly three thousand five
hundred transport and logistics enterprises within a two-hundred-
kilometre radius of Verona — a mixture of road carriers on the Albo
Autotrasportatori, freight forwarders registered with the Camera di
Commercio, warehouse operators and intermodal specialists. To that
base must be added the estimated twelve thousand Veneto manufacturing
SMEs whose shipping volume justifies a light-touch visibility product
aimed at the shipper side rather than the carrier side. The total
addressable market (TAM) therefore encompasses roughly fifteen
thousand firms in the Veneto and adjacent provinces; the serviceable
addressable market (SAM), defined as firms with at least five
shipments per day and a digital-adoption posture compatible with a
SaaS product, is closer to four thousand; and the serviceable
obtainable market (SOM) over a three-year horizon, accounting for
competitive pressure and go-to-market realism, is two hundred to
three hundred tenants.

The competitive landscape is bifurcated. On the enterprise end sit
the European-scale platforms: Project44, FourKites, Transporeon,
Descartes and Shippeo, all of which sell into Fortune 500 shippers
and the top thirty European retailers. Their pricing starts at tens
of thousands of euros per year, their sales cycles are twelve to
eighteen months, and they treat Italian peculiarities (AIDA, Telepass,
Albo Autotrasportatori, PEC, partita IVA validation, CCNL driver-hour
rules) as edge cases to be handled by customer-success teams rather
than first-class product features. On the Italian end sit the legacy
ERP extensions: TeamSystem Logistic, Keolisa, Dogma, Sirio, and a
handful of small ISVs offering bolt-on modules for SAP and Zucchetti.
These systems excel at fleet management, cost accounting and payroll
integration but are either weak on real-time visibility (refresh every
fifteen minutes instead of every second) or tightly coupled to a
specific ERP backend. A gap remains in the middle: a cloud-native,
API-first product specifically tuned to the Verona-corridor context,
priced for the SME operator rather than the multinational shipper.
That is the slot LogiTrack occupies.

A SWOT summarises the posture. Strengths: local context baked in;
native Italian integrations (AIDA, SDI, Albo, Telepass); modern stack
with low runtime cost; founder network inside Consorzio ZAI, RFI
commerce, Confindustria Verona. Weaknesses: new entrant with no
reference clients; capital-light team limits sales capacity; telematics
integrations each demand a bespoke engineering effort. Opportunities:
PNRR Missione 3 invests nearly thirty billion euros in logistics
digitalisation and multimodal infrastructure, CSRD forces large
customers to ask suppliers for emissions data that LogiTrack can
produce, and Transizione 5.0 tax credits cover up to forty-five per
cent of qualifying software expenditure. Threats: Project44 could
launch a Verona data centre and deprive the sovereignty argument of
oxygen; a consolidator could acquire TeamSystem Logistic and graft on
a visibility module at scale; and an antitrust or data-flow ruling at
EU level could shift the compliance cost base.

Pricing benchmarks are revealing. A European-scale visibility
subscription for a mid-size 3PL runs fifty to ninety thousand euros
per year. A legacy Italian TMS module adds ten to twenty thousand euros
onto an existing Zucchetti or TeamSystem licence. Pure telematics
(Viasat box plus portal, excluding analytics or multi-vendor
consolidation) costs twenty to thirty euros per vehicle per month. By
pricing its Professionale plan at five hundred ninety-nine euros per
month inclusive, LogiTrack lands inside the budget an SME
controlling-office can approve without board-level review, while
offering a richer feature surface than the telematics-only providers.

### 3. Business Model & Revenue Strategy

Revenue is delivered through three intertwined streams: subscription
to the platform, per-shipment tracking fees on high-volume tenants,
and a narrow band of services revenue (integration accelerator
packages, training, one-off custom integrations). The subscription is
the anchor and the point of commercial simplicity: a monthly bill tied
to a plan level and a volume ceiling, with predictable upgrade paths
as the tenant's business grows. The per-shipment fee, applied above
plan ceilings and inside the Enterprise tier, transforms LogiTrack
revenue into a function of customer success: the more the tenant
ships, the more the platform earns.

The Starter plan at one hundred ninety-nine euros per month targets
small carriers with five to ten vehicles and up to ten shipments a
day, two concurrent users, one telematics integration, mapping,
timeline and basic SLA reporting. The Professionale plan at five
hundred ninety-nine euros per month targets operators with twenty to
eighty vehicles and up to one hundred shipments a day, ten users,
multiple telematics integrations, AIDA customs module, shipper portal
and webhook APIs. The Enterprise plan, quoted individually and
typically landing between one thousand five hundred and five thousand
euros per month, serves carriers above one hundred vehicles or
shippers above five hundred shipments a day, with dedicated SLAs,
single sign-on, sovereign-hosting options and a named customer-success
owner. Above plan ceilings, consumption is metered at a rate of
between fifty cents and one euro fifty per additional shipment,
decreasing as volumes rise. The freemium hook — a free tier capped at
fifty shipments per month, one user, no API — exists to lower the
activation barrier for shipper-side buyers who want to prove the
product on a single corridor before budgeting for a full subscription.

Unit economics are designed to hold through the early growth phase.
Gross margin on the subscription is approximately eighty-two per cent
after cloud infrastructure, OSRM self-hosting and observability
spend. Customer acquisition cost averages one thousand two hundred
euros for Professionale-tier tenants, dominated by outbound sales
touches (trade shows, partnership introductions, content marketing)
and onboarding engineering time. Average contract value lands at
seven thousand two hundred euros annually for Professionale tenants
and approximately twenty-eight thousand euros for Enterprise, giving a
blended lifetime value — assuming a three-year mean tenancy and gross
margin constant — of roughly seventeen thousand euros. The resulting
LTV to CAC ratio is comfortably above ten to one; payback on
Professionale acquisitions is around five months, which allows
reinvestment of the first annual recurring revenue into either sales
or product without equity dilution.

Churn is the single largest economic variable. The Italian logistics
SME is a long-relationship buyer once convinced, but the platform
must earn the convincing. Churn mitigation is therefore built into
the onboarding itself: every Professionale tenant gets a six-week
structured rollout, a named success owner, a corridor-specific
integration template (Verona → DACH, Verona → Lombardia, Verona →
Emilia), and a Monday-morning dashboard review for the first quarter.
Beyond onboarding, the churn levers are three: integration depth,
where every additional connected carrier, telematics provider or
client portal increases switching cost; analytic habituation, where
recurring SLA reports inside the platform create executive muscle
memory; and contract timing, where LogiTrack moves tenants onto
quarterly invoicing with a five per cent loyalty discount after the
first renewal.

The pricing psychology is deliberate. Five hundred ninety-nine euros
per month for the Professionale plan sits one cent below six hundred,
the traditional commercial ceiling below which approvals flow through
middle management. Twenty-three euros per day works out to roughly
one-tenth the fuel bill of a mid-sized truck for a single day, a
framing LogiTrack uses explicitly in its pricing page copy. The
Starter plan at one hundred ninety-nine euros per month is framed as
"less than a daily lunch for a small team" during sales calls. Most
importantly, the Professionale plan includes everything needed for
AIDA cross-border operations out of the box — no add-ons, no modules,
no separate contract — because cross-border is the single use case
that converts interest into signature inside the Verona corridor.

The two strategic pricing trade-offs deserve explicit articulation.
The first is whether to charge the carrier or the shipper. LogiTrack
charges both, but differently. The carrier pays the subscription,
because the carrier is the source of telematics truth and the one who
wins new contracts by presenting visibility as a differentiator. The
shipper pays a portal-seat fee, typically fifty euros per user per
month, which is billed to the carrier and passed through at cost —
the carrier's commercial sales team can therefore offer free portal
access to prospective customers as a sweetener without dipping into
its own margin. The second trade-off is how to price the per-shipment
tier. Metered pricing inside a small number of strategic Enterprise
accounts lets LogiTrack capture upside when volumes explode (a common
pattern during the autumn peak, when e-commerce and agri-food coincide)
without forcing smaller tenants into a pricing model they perceive as
unpredictable. Starter and Professionale tenants therefore face a hard
ceiling with an automatic upgrade prompt, while Enterprise tenants
face a metered overage curve that decreases as volumes rise.

Revenue recognition follows straightforward SaaS norms: monthly
subscription recognised ratably, per-shipment overages recognised as
earned, professional services recognised on delivery. The revenue
model is modelled to hit a break-even run rate at month thirty with
roughly one hundred twenty paying tenants and an average revenue per
user of six hundred fifty euros per month; the sensitivity analysis
shows the plan still funds itself at month forty-two if the average
ARPU falls to five hundred euros, and breaks even at month twenty-two
if enterprise attach rates exceed fifteen per cent instead of the
plan's ten per cent. The capital envelope assumed to reach those
milestones is approximately two million euros of seed plus Series A
pre-seed bridging, enough to fund two-and-a-half years of plan
execution.

A more granular slice of the pricing strategy is worth articulating
here because it clarifies the commercial architecture to any reader
who has not yet seen the price list. The Starter plan is explicitly
loss-leading in customer-acquisition terms: its margin contribution
is small once the costs of onboarding, telematics integration setup
and first-quarter customer success are amortised. Its strategic
purpose is not to be a profit centre but to lower the activation
barrier for the fifteen-vehicle family-owned carrier that cannot
justify six hundred euros of monthly software spend to its board of
directors but will readily authorise two hundred. Those Starter
tenants are the breeding ground for Professionale upgrades, which
occur statistically around month nine of tenancy when the tenant's
vehicle count or shipment volume crosses the plan ceiling. The
Professionale plan is where the unit economics work: its margin
after onboarding costs is approximately five hundred twenty euros
per month, and the modal tenant renews for three to four years. The
Enterprise plan is the economic engine: its gross margin exceeds
ninety per cent, and its strategic role is to anchor the company's
annual recurring revenue trajectory while giving the Professionale
customers a visible upgrade target that keeps them engaged with the
roadmap.

The sales-cycle economics follow naturally from this segmentation.
Starter deals close in four to six weeks from first contact, largely
via inbound and self-service channels; Professionale deals close in
twelve to sixteen weeks with a one-pilot-before-signing pattern;
Enterprise deals close in six to nine months and typically involve
a named procurement officer, a competing evaluation against one or
two European platforms, and an Information Security Questionnaire
that the security engineer hired in month eighteen is dedicated to
answering. Contract structures mirror this: Starter is monthly
rolling with thirty-day cancellation; Professionale defaults to
twelve-month terms with a five per cent loyalty discount on
renewal; Enterprise lands at twenty-four to thirty-six-month
committed terms with explicit service credits, named contacts on
both sides and an annual joint roadmap planning session.

---

## Part II — Technical Deep Dive

### 4. Technical Architecture

LogiTrack is designed as a service-oriented backend, a reactive
single-page frontend, and an integration layer that is deliberately
thin. The Go backend is written in Gin, a lightweight HTTP framework
that lets the team keep a single process per replica while handling
thousands of concurrent connections; Go's goroutine model is a
natural fit for telematics ingestion, where each incoming webhook
spawns a lightweight worker that normalises, persists and publishes
without touching a thread pool. MongoDB was selected as the primary
store because the document shape of a shipment varies substantially
between carriers (one provider attaches ADR compliance data, another
tracks reefer temperature, a third records only bare waypoints), and
forcing that variability through a rigid relational schema would
either produce dozens of JOINs per read or degenerate into a generic
JSONB column in a relational database, which is the worst of both
worlds. Redis serves two complementary roles: a hot cache for the
latest position of each in-flight shipment, read by the dashboard on
every page load, and a publish/subscribe channel that fans tracking
events out to the WebSocket hub without coupling producers to
consumers.

The principal data flow begins with ingestion. A carrier's telematics
provider pushes a webhook into LogiTrack at `POST
/api/v1/shipments/{id}/waypoints`, or a direct Kafka consumer pulls
events from a provider stream when the provider supports it.
Authentication is JWT-based today and mTLS-pinned for enterprise
integrations; rate limiting runs per-tenant per-provider, with
circuit breakers that open after three consecutive errors and close
when a probe succeeds. Once past the gate, the handler deserialises
the payload, validates it (coordinates inside plausible bounds,
timestamp not in the future, shipment exists and belongs to the
carrier's tenant), and invokes `ShipmentService.RecordWaypoint`. The
service atomically appends the waypoint to the Mongo document, sets
`currentPosition`, writes a `tracking_events` record with a sequence
number derived from the UTC nanosecond, caches the position in Redis
and publishes the event on the tracking channel. The WebSocket hub,
running in a separate goroutine per replica, subscribes to the Redis
channel at boot and broadcasts each incoming event to matching
connected subscribers.

The Shipment aggregate is intentionally denormalised. When a
consignment is booked, copies of the consignor, consignee and
carrier at that instant are captured into the Mongo document, so
later edits to master data do not retroactively rewrite historical
shipments. The chain-of-custody log is stored separately in its own
collection, indexed on `(shipment_id, sequence)` uniquely. Every
custody record embeds the SHA-256 of the previous record's canonical
JSON, forming a hash chain that an external auditor can replay
without access to the internal database: given the genesis entry and
the sequence of actions, the hash at any position can be recomputed
and compared. Corrections are never done in place; they append a
compensating record with action `exception` and explanatory text.
This design choice is driven less by fear of internal tampering and
more by the courtroom reality that Italian commercial litigation
occasionally demands a demonstrable immutability argument when a
consignment is claimed to have been altered in transit.

The WebSocket protocol is deliberately simple. The client opens the
connection with a bearer token (header for non-browser clients, query
string for browsers because Fetch API cannot set Authorization on an
upgrade request), sends an initial `subscribe` message with an
optional shipment id and an optional carrier filter, and from that
point receives JSON-encoded tracking events. Pings at thirty-second
intervals keep the connection alive through corporate proxies, and a
slow-consumer policy drops messages rather than buffering, because a
dashboard that lags ten seconds behind reality is worse than one
that resyncs on the next REST refresh. In production, the WebSocket
layer sits behind an Application Load Balancer with idle timeouts set
explicitly to sixty seconds to match the ping cadence.

Scalability is engineered horizontally. The Go backend is stateless
with respect to request routing; any replica can serve any request
because all state lives in MongoDB or Redis. At low volumes a single
replica handles thousands of ingested events per second, and Go's
memory footprint (forty to eighty megabytes per replica in steady
state) keeps infrastructure costs modest. As volumes rise the
MongoDB tier grows through sharding: the shard key is a hash of
`tenant_id` for OLTP reads and, for a secondary analytical collection,
a compound of `(tenant_id, etd_day)` for time-range queries. Redis is
replaced by Redis Cluster or ElastiCache Redis with failover enabled
for uptime guarantees; the pub/sub channel is replaced by Kafka when
cross-service event consumers are introduced (for example, a
batch-analytics service consuming from the tracking topic to compute
SLA statistics overnight). The Kafka migration is pre-engineered: the
service publishes events to both Redis and a Kafka placeholder
topic as soon as the brokers are configured, giving the team time to
validate the migration with real traffic before flipping consumers.

Consistency is treated pragmatically. Writes to the shipment document
and the tracking-events collection are not wrapped in a multi-document
transaction by default; MongoDB supports them, but the cost is non-
trivial at ingestion rates north of five hundred events per second.
Instead, the service design tolerates eventual consistency for the
write-then-read sequence and idempotency for every ingestion path, so
a duplicate webhook never double-books a waypoint (deduplication on
`rawEventId` is enforced at the repository boundary). The chain-of-
custody append is the single operation that does demand strict
ordering and durability: the sequence computation and insert run
inside a short Mongo transaction, with a retryable pattern for the
rare `WriteConflict` under contention.

Security architecture follows the platform-wide baseline (Section 12
of the parent specification) and layers LogiTrack-specific controls on
top. Tokens are issued by a separate identity service (not shown in
the reference implementation, a placeholder for SSO integration) and
signed with a 256-bit secret rotated quarterly. Per-tenant isolation
is enforced at the repository layer — every query includes the tenant
id from the authenticated claims — and confirmed by a contract test
that asserts cross-tenant reads return 404. Geolocation data is
treated as sensitive personal data under GDPR; retention defaults to
thirteen months for granular waypoints and five years for aggregated
route summaries, and tenants can configure tighter retention via an
admin API. Field-level encryption is applied to driver personal data
(name, contact, licence number) when the tenant enables CCNL payroll
integration.

Failure modes and their mitigations deserve a concrete run-through.
If the MongoDB primary fails, the replica set elects a new primary in
under ten seconds; during that window the backend replicas buffer
ingestion events into the Redis stream `logitrack:buffer:ingest` with
a five-minute TTL, replaying them once the primary returns. If Redis
itself fails, the WebSocket fan-out is interrupted but ingestion
continues: the frontend falls back to a two-second REST polling loop
automatically, driven by a feature flag that the backend reports
through `/api/health`. If OSRM is unreachable, ETAs are computed with
a fallback that uses the great-circle distance and a historical mean
speed for the corridor; users see a banner marking the ETA as
degraded. If AIDA is unreachable, customs declarations are queued in a
retry loop for up to twenty-four hours and escalated to human
operators if the queue clears without success. None of these failures
take the dashboard fully offline; each mode degrades gracefully with
explicit user-visible signals.

A note on the specific geospatial-index choice is warranted because
it is the performance-critical hinge of the product. MongoDB 2dsphere
indexes are built on the GeoHash tree, supporting the `$geoWithin`,
`$geoIntersects` and `$near` operators at logarithmic complexity. The
LogiTrack geofence check — "does this position fall inside any
geofence relevant to this shipment?" — runs on every ingested
waypoint, and at two thousand waypoints per second a naive design
would hammer the database. The actual implementation combines three
layers. First, a client-side bloom filter at the ingestion gateway
rules out waypoints that are clearly outside any known geofence by
comparing the H3 cell index (resolution 9, approximately 174 metres
per edge) against a precomputed set of cells that intersect at least
one geofence. Second, a Redis-cached per-shipment set of relevant
geofences — populated at shipment creation time from the Mongo
collection — narrows the candidate list to typically five or fewer
polygons for a given shipment. Third, the MongoDB 2dsphere index
resolves the final `$geoIntersects` against those candidates. The
measured latency for this three-layer check is sub-millisecond at
the ninety-fifth percentile, letting the ingestion service sustain
peak rates without running into index saturation.

The ETA calculation deserves a similar architectural note. A naive
approach would call OSRM on every waypoint ingestion, but OSRM round
trips cost tens of milliseconds and dominate the latency budget. The
implementation calls OSRM once per shipment at creation (to compute
the planned route and the reference ETA), caches the Polyline6
geometry in the shipment document, and thereafter snaps every
incoming waypoint to the nearest point on that geometry using a
fast spatial algorithm implemented directly in Go. The "distance
remaining along route" is derived by subtracting the pre-computed
distance-along-route value from the total route length; the "time
remaining" is estimated by applying the provider's historical
average speed for the corridor's time-of-day and day-of-week slot.
OSRM is re-called only when the snap distance exceeds a threshold —
typically five hundred metres — indicating the vehicle has deviated
materially from the planned route and the ETA needs to be
recomputed. This architecture keeps the OSRM call rate at roughly
one call per shipment-hour, a comfortable regime for a
self-hosted instance.

Multi-tenancy is implemented as a soft-isolation model at the
application layer, not a hard-isolation model at the infrastructure
layer. Every document in every collection carries a `tenant_id`
field, every query in the repository layer is enforced to include
that field, and a contract test running in CI asserts cross-tenant
reads return 404 across the full API surface. For Enterprise
tenants with sovereignty requirements, a per-tenant database option
is available: the tenant is provisioned a dedicated MongoDB cluster,
a dedicated Redis instance and a dedicated namespace on the Kubernetes
cluster, all routed by a tenant-aware ingress that inspects the JWT
to route traffic to the correct backend deployment. The application
code is unchanged between shared and dedicated modes; only the
infrastructure provisioning differs, which keeps the engineering
overhead of supporting both models modest.

The service-oriented internal decomposition, while appearing as a
single Go binary in the reference implementation, is deliberately
organised so a future split into independent services is a minimally
invasive refactoring. The `internal/handlers`, `internal/services`,
`internal/repository` packages are coupled only by interface, and
each service struct is constructed in the main composition root;
extracting, for example, the custody service into a dedicated
microservice is a matter of exposing its interface over gRPC and
replacing the in-process call with a client stub. The planning
assumption is that the first extraction happens in Phase 3, when the
AIDA customs workflow becomes sufficiently complex to merit its own
deployment cadence. Until then the monolith is preserved because its
operational simplicity — one image, one log stream, one deployment
pipeline — saves engineering time that can be spent on features
instead of on distributed-systems plumbing.

Database migration strategy follows a similar disciplined approach.
Schema evolution in MongoDB is nominally free because the engine is
document-oriented, but the service code's expectations around
document shape constitute an implicit schema that must evolve
carefully. Every change to the document structure is paired with a
migration script stored in `backend/migrations/`, a version number
recorded on each document, and a dual-read code path that tolerates
both the old and new shapes for a defined overlap period (usually
two release cycles). Migrations that require backfilling rows run
as Kubernetes jobs scheduled off-peak, with per-batch progress
reporting and resumability on failure. The strategy has proven
sufficient for the migrations executed so far — the addition of the
`tenant_id` field during the multi-tenancy retrofit, the
introduction of the hash-chain custody fields, the reshaping of the
waypoints array into a capped subdocument when volumes grew.

### 5. Development Roadmap

The roadmap is carved into four phases spanning approximately
twenty-four months from MVP to mature multi-modal platform. Each
phase has a clearly bounded scope, a shipping milestone and a
MoSCoW-classified backlog distilled from the customer-interview
programme.

Phase 1, months one through five, is the MVP: core tracking, manual
entries and single-integration ingestion. The Must items are the
shipment CRUD, the waypoint ingestion, the dashboard map, the
chain-of-custody log, the basic SLA report and the AIDA declaration
stub; the Should items include the WebSocket stream and the Leaflet
integration; the Could items are the CSV import wizard and the
shipper portal; the Won't items, scoped deliberately out, include
multimodal rail and customer-managed custom dashboards. The sprint
cadence is two-week iterations, each ending with a demo session for
two design-partner carriers and a shipper prospect. The MVP is
declared complete when ten paying tenants run production workloads
for a full calendar month without critical incidents.

Phase 2, months six through eleven, integrates the principal
telematics providers and hardens the ingestion surface. Must items:
certified Viasat, Octo and Geotab connectors with webhook + pull
fallback, a replay harness for carrier-reported anomalies, rate
limiting per provider, and a reconciliation job that cross-checks
Telepass transits against recorded waypoints. Should items: the
geofence workflow (warehouse, customs, port, depot typing, dwell-time
alarm), a carrier-side reporting module exposing on-time-in-full and
cost per kilometre, and customer-success tooling (success-plan
templates, health-score dashboards). Could items: a mobile companion
for drivers to submit proof-of-delivery photos; Won't: automated
insurance-claim initiation. Sprint cadence shifts to three-week
iterations to reduce the tax of release engineering now that multiple
integrations are in flight.

Phase 3, months twelve through eighteen, addresses customs via AIDA
and hardens Italian-sovereignty options. Must items: a production-
grade AIDA client supporting T1 and T2 declarations, MRN lifecycle,
inspection workflow and the electronic seal integration; Aruba Cloud
sovereign deployment option with a documented migration runbook; a
DPIA template and Garante-ready audit trail for geolocation data
processing. Should items: Albo Autotrasportatori carrier verification
hooks, automatic due-diligence checks for new carrier onboarding;
Could items: GDPR data-subject request handling via a self-service
portal; Won't: consumer-facing tracking pages.

Phase 4, months nineteen through twenty-four, reaches multimodal and
strategic scale. Must items: RFI intermodal-slot integration, a
combined timeline view that joins road waypoints with rail events,
multi-tenant reseller management, and an SDK for partners to build
vertical extensions. Should items: Kafka event-sourcing migration,
replay tooling for historical reconstruction, a data-export module
aligned to CSRD reporting. Could items: machine-learning delay
prediction based on two years of accumulated data; Won't: autonomous
routing replacing dispatcher decisions.

Between the phases sit well-defined transition gates with explicit
exit criteria. The gate from Phase 1 to Phase 2 requires: ten paying
tenants in production, a first-response support SLA achieved at
ninety-five per cent, a backend P95 latency below three hundred
milliseconds on the list-view query, an uptime of ninety-nine point
five per cent measured over a rolling thirty-day window, and at least
one tenant who has formally signed an AIDA-adjacent pilot commitment.
The gate from Phase 2 to Phase 3 requires: fifty paying tenants,
three production telematics integrations each handling more than a
thousand events per day, carrier-reported false-positive rate on
delay detection below two per cent, and at least one paying tenant
running a DACH-corridor shipment volume above fifty consignments per
week. The gate from Phase 3 to Phase 4 requires: one hundred paying
tenants, the Aruba Cloud sovereign deployment live with at least one
anchor tenant, ISO 27001 certification in progress, an AIDA
declaration success rate above ninety-nine per cent, and a net
revenue retention measurement above one hundred ten per cent.

These gates exist to resist the natural tendency to chase the next
capability before the current capability is genuinely production-
mature. Engineering teams without explicit gates drift into
simultaneous construction of Phase N and Phase N+1; the result is
two half-finished phases rather than one solid one. The planning
process institutionalises the discipline of finishing. Each gate is
reviewed monthly at the board-level metrics meeting, and the
go-ahead is a formal decision with recorded assent; the alternative
outcomes — extension, course correction, or resource reallocation —
are equally formal and equally welcomed.

The technical-debt budget is not only fixed at twenty per cent of
engineering capacity; it is also earmarked for specific categories
on a rolling schedule. The categories, in descending priority order,
are: security and compliance (patching, dependency upgrades, access
review), observability and reliability (tracing coverage, runbook
drills, chaos engineering), developer velocity (build times, test
runtime, local environment reliability), and finally cosmetic
cleanup (naming, dead code, documentation gaps). Each cycle
allocates the twenty per cent across categories based on the
observed state of the debt register; an unhealthy observability
state, for example, may consume the full cycle's debt budget for
two consecutive cycles until the baseline recovers.

Sprint planning follows a light-weight Shape Up rhythm inside each
phase: six-week cycles with a dedicated two-week cool-down. Each
cycle ships two or three "bets" sized to a pair or a single senior
engineer, with a cool-down reserved for refactoring, technical-debt
reduction and exploration work. The technical-debt budget is fixed at
twenty per cent of engineering capacity across the life of the
product; the team tracks a single debt register in the engineering
wiki, and any item that stays on the register for two cycles
automatically escalates to the next planning meeting. Feature
prioritisation uses MoSCoW as above, but every Must candidate must
demonstrate either a named customer commitment or a concrete revenue
impact before entering the cycle shape-up; this rule is the single
most important defence against the subtle drift toward speculative
features that kills small teams.

---

## Part III — Scaling & Operations

### 6. Scaling Strategy

Scaling LogiTrack is an exercise in anticipating three distinct
pressures: read concurrency on the dashboards, write throughput on
ingestion and storage growth on historical data. Each pressure
demands a different mitigation, and the roadmap sequences them so
that each investment pays off before the next becomes necessary.

Read concurrency arrives first. A single Professionale tenant with a
hundred shipments a day and ten dashboard users refreshes the list
view roughly forty times per hour and opens the detail view on
specific shipments another hundred times per hour; across the first
hundred tenants that is a few thousand requests per minute. At those
volumes a single backend replica serves comfortably, but tenant
growth past two hundred, or the activation of the shipper-portal
seats that multiply dashboard users by a factor of five, pushes the
single replica toward its CPU headroom. The mitigation is a standard
horizontal scale-out behind an ALB: the backend is stateless, so
adding replicas increases capacity linearly up to the MongoDB query
capacity, which is the next bottleneck. MongoDB read replicas serve
the list queries with a secondary-preferred read preference, while
the write-back remains on the primary; at the scale of a thousand
tenants a three-node replica set is saturated, and the cluster moves
to a sharded configuration with three shards and a dedicated config
server replica set. The list-view query, by far the hottest, is
served primarily from a covered index on `{tenant_id, status,
updated_at}` so even large tenants see sub-hundred-millisecond
response times.

Write throughput is the second pressure. At the MVP scale, the
telematics inflow is in the low hundreds of events per second; at
Phase 2 scale, with Viasat, Octo and Geotab streaming for hundreds of
carriers, it peaks around two thousand events per second during the
autumn corridor rush. MongoDB handles that rate on appropriately
sized hardware — the bottleneck is not insert throughput but the
index maintenance on the secondary indexes we use for dashboards. The
mitigation is twofold: first, we split the tracking-events collection
into a hot tier (last thirty days, fully indexed) and a cold tier
(thirty days to three years, indexed only on `shipment_id`), using
Atlas Online Archive or a custom ETL to move documents across tiers
on a nightly schedule. Second, we migrate the publish side to Kafka,
which decouples the ingestion rate from the query side entirely: the
Go service writes to Kafka, consumers materialise to MongoDB at their
own pace, and the WebSocket hub subscribes to a compacted topic. At
extreme throughput we move the hot tier to TimescaleDB for time-series
compression, while keeping the aggregate shipment document in MongoDB.

Storage growth is the third pressure. Each waypoint occupies roughly
four hundred bytes on disk including indexes; a hundred thousand
shipments a year each generating a hundred waypoints means forty
gigabytes of raw data per year per hundred thousand shipments, before
compression. Mongo compresses well — roughly three-to-one on our
document shape — so the actual growth is about thirteen gigabytes per
year per hundred thousand shipments. That is trivially manageable
even at ten-million-shipment annual volume, provided the indexing
strategy is respected. Where storage growth does bite is on
observability: the structured log stream at ten thousand requests per
minute generates roughly one gigabyte of compressed logs per day, and
a thirty-day retention inside Loki or CloudWatch adds up to a
thousand euros per month at scale. The mitigation is a tiered logging
policy: full request logs retained seven days, audit-grade logs
retained thirteen months in object storage, everything above that
shipped to S3 Glacier Deep Archive and indexed via Athena when an
auditor asks.

Caching hierarchy is three layers deep. The CDN layer (CloudFront in
the reference deployment) serves the frontend bundle and the landing
page with a one-hour cache and a stale-while-revalidate strategy,
cutting the origin read count to essentially zero. The application
cache (Redis) holds the last-known position, pre-computed SLA
rollups, refresh-token JTIs, rate-limit counters and short-lived
authorisation decisions; TTLs are tight, on the order of minutes for
most entries, because stale logistics data is actively misleading.
The database cache (MongoDB's WiredTiger cache plus Atlas query
shapes) is sized to hold the working set of hot shipments in RAM —
empirically, ninety-five per cent of queries hit shipments updated in
the last seven days, which comfortably fits in a sixty-four-gigabyte
primary.

Multi-region strategy is sequenced by customer need rather than
engineering ambition. In Phase 1 the platform runs entirely in AWS
eu-south-1 (Milano), with backups replicated to eu-west-1 (Dublin)
for disaster recovery. In Phase 3, when the first customer explicitly
requests Italian-sovereign hosting for a government-adjacent workload,
an Aruba Cloud Arezzo deployment is spun up as a second production
region, with tenant-level routing such that a sovereign tenant's
traffic never leaves Italian soil. In Phase 4, a DACH co-location is
added either at a Frankfurt colocation or on Hetzner's Nuremberg
facility to reduce WebSocket latency for customers hub-and-spoking
through Munich. Cross-region data replication is intentionally
avoided where not strictly required; the operational burden exceeds
the latency savings until a single region genuinely cannot carry the
load.

Geographic expansion follows a similar cadence. Year one is Verona
province with a concentrated spoke into Milan. Year two expands into
the Veneto and Lombardia more broadly, with a dedicated go-to-market
push around the Interporto Bologna ecosystem. Year three reaches the
pan-Italian level and begins systematic DACH expansion, leveraging the
fact that every successful Verona-corridor customer already runs
cross-border traffic north and therefore serves as a natural bridge.
Year four and beyond, the platform enters the French corridor through
Turin and the southern Iberian corridor via Genoa, with localised
integrations for the respective customs authorities and local
languages.

Performance testing is continuous rather than episodic. A k6 suite
exercises the principal API surfaces every night against a
staging environment that mirrors production topology at one-tenth
scale. The suite simulates three load profiles: the steady-state
daytime profile with approximately two hundred concurrent dashboard
users and eight hundred ingestion events per second; the rush-hour
profile peaking at one thousand concurrent users and two thousand
events per second; and a spike profile that quadruples the rush-hour
load for a three-minute window to verify the system's ability to
absorb burst traffic without cascading failures. Each nightly run
emits a structured report into a Grafana dashboard; regressions
exceeding the ten per cent tolerance threshold against the previous
seven-day rolling median open an automatic Jira ticket assigned to
the on-call engineer and are triaged at the next morning stand-up.
The discipline has caught a dozen subtle regressions over the first
eighteen months — a GC tuning issue after a Go toolchain upgrade, a
Redis key sprawl from an inadvertently chatty microservice, a
MongoDB query plan regression after a documentation-motivated index
rename — each time before the regression reached paying customers.

The capacity headroom target is deliberately aggressive: production
runs at no more than fifty per cent of the tested safe load at any
moment, providing a 2x margin for unanticipated spikes and enough
buffer to survive a full regional failover without performance
degradation. This headroom is expensive in infrastructure terms
(roughly twenty per cent more compute than strict
demand-driven sizing would imply) but it is consciously chosen as
insurance against the reputational cost of an outage. Logistics
customers, more than many SaaS buyers, equate availability with
trust; an outage of an hour during a customs peak costs the customer
real money and the platform real goodwill, and the cost of the
incident is an order of magnitude larger than the cost of the
insurance headroom.

### 7. Security & Compliance

LogiTrack processes personal data — driver names, contact details,
real-time location traces — and regulated commercial data — commercial
invoices, customs declarations, transit route optimisation output.
The security and compliance stance reflects that dual nature: it is
GDPR-compliant by construction, respectful of the Italian D.Lgs.
196/2003 transposition, and aligned with the sector-specific
obligations imposed by Agenzia delle Dogane on its API consumers.

On the GDPR side, the platform appoints a Data Protection Officer as
part of every Enterprise agreement and provides a template DPIA that
customers can reuse as the starting point for their own
accountability documentation. Lawful basis for processing is
contractual performance for the tenant's employees and legitimate
interest — carefully balanced — for third-party drivers tracked
through telematics integrations, with a documented legitimate-interest
assessment that customers can append to their internal register.
Data-subject requests are handled through a dedicated administrative
API: a tenant administrator with the `dpo` role can export all
personal data relating to a named data subject in a structured JSON
format, or initiate erasure (with the hard-guaranteed exception of
chain-of-custody entries that the customer is legally required to
retain). Retention is configurable per tenant with sensible defaults:
thirteen months for granular positional data, five years for
aggregated SLA reports, ten years for custody and customs records in
alignment with Italian tax-law retention.

The Italian D.Lgs. 196/2003, as amended by D.Lgs. 101/2018, layers
specific obligations around employee monitoring. Because tracking
company vehicles amounts to indirect tracking of the drivers driving
them, Italian law requires either an agreement with the trade-union
representatives (accordo sindacale) or, failing that, an
authorisation from the Ispettorato Nazionale del Lavoro. LogiTrack
does not replace the customer's obligation to secure this agreement,
but it documents the obligation explicitly in the onboarding
checklist and provides a template union-agreement clause that large
customers' human-resources teams can take to negotiation.

Encryption is applied at rest (AES-256 on MongoDB Atlas, AWS-managed
KMS for S3 buckets hosting backups and documents) and in transit (TLS
1.3 only, with HSTS one-year max-age and preloaded at the domain
level). Field-level encryption is applied to driver names and contact
details when the payroll-integration flag is enabled; the encryption
keys live in AWS Secrets Manager and rotate quarterly. Client-side
encryption for document attachments is an opt-in for Enterprise
tenants with particularly sensitive goods (pharma cold chain,
defence-sector exports under D.Lgs. 105/2003).

Access control is role-based and tenant-scoped. Seven roles are
standard: `admin`, `dispatcher`, `driver`, `shipper_read`,
`shipper_write`, `dpo`, `auditor`. Every role has an explicit
permission matrix, maintained in `docs/PERMISSIONS.md` (referenced
from the product's administration portal). Audit logging captures
every state-changing operation in an append-only `audit_log`
collection with fields `who, what, when, from, to, correlation_id`;
entries are immutable in the sense that the repository layer refuses
updates or deletes, and the collection is backed up to a separate
write-once S3 bucket with object-lock enabled.

Incident response follows a tiered SLA: severity-one incidents
(data breach, prolonged outage, customs-declaration failure) are
triaged within fifteen minutes and escalated to the CEO within an
hour; severity-two incidents (functional regression, integration
outage limited to a single provider) are triaged within an hour and
resolved within a calendar day; severity-three incidents (minor UI
defects, cosmetic issues) are resolved at the next sprint boundary.
For GDPR breaches, the seventy-two-hour notification window to the
Garante is tracked explicitly through a dedicated incident runbook,
and the customer is informed in parallel with the same SLA.
Certifications are sequenced: ISO 27001 Stage 1 audit in month
fifteen, full certification in month twenty; SOC 2 Type I in month
twenty-four for the DACH expansion; TISAX remains a future
consideration driven by automotive-sector customer demand.

### 8. Infrastructure & DevOps

The production infrastructure is deliberately minimal in moving
parts. Primary region is AWS eu-south-1 (Milano), chosen because it
keeps data inside Italy for GDPR simplicity while giving access to
managed services at the price point a seed-stage company can afford.
Compute runs on Amazon EKS with three node groups: a system node
group for ingress, observability and cert-manager; a stateful node
group for MongoDB sidecars when self-hosted is chosen over Atlas; and
a general node group for the backend, frontend and OSRM replicas.
MongoDB is MongoDB Atlas M30 in the reference deployment, upgraded to
M60 at the Phase 3 threshold; Redis is ElastiCache with a failover
replica.

For sovereignty-sensitive tenants, the same stack is replicated on
Aruba Cloud Arezzo using Aruba's Kubernetes service for compute, a
self-hosted MongoDB replica set and a self-hosted Redis sentinel. The
runbook for switching a tenant between AWS and Aruba is explicit and
tested quarterly; it involves draining WebSocket connections on the
source cluster, freezing writes for a fifteen-minute window,
replicating the most recent Mongo snapshot, and reopening writes on
the target cluster. The process takes approximately ninety minutes
end-to-end and is accepted by sovereignty-aware customers as the
price of compliance.

Infrastructure-as-code is Terraform for cloud resources and
Helm-plus-Kustomize for Kubernetes manifests. The Terraform state
lives in an encrypted S3 bucket with DynamoDB locking; every change
runs through a pull request that requires a plan output to be
attached in the description. The CI/CD pipeline is GitHub Actions for
both CI and CD; production deployments use a blue-green strategy for
the backend (a new ReplicaSet stands up, health checks pass, traffic
cuts over, the old ReplicaSet is drained) and a canary strategy for
the frontend (five per cent of traffic sees the new bundle for thirty
minutes before full cutover). Disaster recovery targets an RPO of
five minutes — the Atlas continuous backup guarantees this, plus
Redis AOF replay — and an RTO of one hour, achieved by having the DR
region's Kubernetes manifests pre-materialised and the Atlas cluster
pre-provisioned in passive mode.

Observability is OpenTelemetry end-to-end. Traces flow to a managed
OTLP collector and into Grafana Tempo; logs are Loki; metrics are
Prometheus scraped from the `/metrics` endpoint. Alerting baselines
mirror the platform-wide standard — error rate above one per cent,
P95 latency above three hundred milliseconds, pod saturation above
eighty per cent — plus LogiTrack-specific alerts for ingestion lag
(events queued for more than thirty seconds) and customs declaration
failures (more than three consecutive AIDA errors). On-call rotation
is one engineer per week, with a warm secondary; the service has run
for the first six months without a severity-one incident, a track
record the team intends to protect fiercely as volumes grow.

Cost optimisation becomes a dedicated discipline from month twelve
onwards. At pre-product-market-fit scale, engineering resources are
better spent on features than on cost tuning; once revenue stabilises
and the infrastructure bill becomes meaningful, a quarterly
cost-review meeting aligns engineering priorities against the
cost-per-tenant metric. Specific measures taken during the first
cost-review iteration include: moving the frontend bundle to
CloudFront's HTTP/3 configuration for better cache behaviour, moving
OSRM to a pair of spot-instance replicas with a fallback on-demand
baseline, consolidating Loki log retention from thirty to fourteen
days for non-audit categories, and right-sizing the MongoDB Atlas
tier based on observed working-set memory. Each iteration targets a
five-to-ten per cent reduction in cost-per-tenant without regressing
availability or latency SLOs; the cumulative effect compounds into a
meaningful gross-margin improvement by Year Three.

Compliance-driven audit preparation is folded into the standard
engineering workflow. The Security engineer hired in month eighteen
owns an evidence-collection pipeline that automatically harvests
access logs, code-review records, deployment approvals and security
scan outputs into a structured archive aligned to ISO 27001 Annex A
controls. When the Stage 1 audit arrives, the evidence is ready;
when Stage 2 arrives, the auditor finds a mature process rather than
a scramble. The same pipeline feeds the SOC 2 Type I preparation in
month twenty-four, reusing the ISO investment rather than duplicating
it. This approach — treating compliance as a byproduct of
well-instrumented operations rather than a parallel project — is the
single most important operational decision the company makes in its
first two years.

---

## Part IV — Business Operations

### 9. Team Structure & Hiring Plan

The founding team is a pragmatic four-person unit sized to ship the
MVP and acquire the first twenty paying tenants without requiring
institutional capital. The founder, responsible for product, sales
and investor relations, holds roughly seventy per cent of the cap
table at inception. The CTO, a Go-and-MongoDB specialist brought in
as a co-founder with ten per cent equity and a competitive Verona
salary in the sixty-to-seventy-thousand-euro range, owns the backend
architecture and the integration engineering roadmap. A senior
frontend engineer with Vue and TypeScript seniority at eight per cent
equity and a salary in the fifty-to-sixty-thousand range owns the
SPA, the dashboard visualisations and the WebSocket client. A
logistics-domain-expert hire, drawn from the Verona spedizioniere
community, sits on two per cent equity and a salary between
forty-five and fifty-five thousand euros plus a commission tied to
closed-won deals; his role is the bridge between the product and the
reality of day-to-day operations at Quadrante Europa.

Month six expands the team by two: a carrier-integrations engineer
specialising in EDI, webhook normalisation and ad-hoc connector
engineering, salary fifty to sixty thousand euros, and a SRE focused
on observability, deployment pipelines and the Kubernetes operations,
salary fifty-five to sixty-five thousand euros. Month twelve adds
three: a second frontend engineer, a dedicated customer-success lead
based in Verona and a DACH-market BDR. Month eighteen brings on a
data engineer for the analytics product, a security engineer to shepherd
ISO 27001 certification and a second account executive to cover the
Lombardia territory. The headcount at month twenty-four is
approximately fifteen, with gross personnel cost roughly one million
euros per year — carefully matched against the ARR trajectory so
personnel burn never exceeds sixty per cent of revenue at the run
rate of the current month.

Remote-hybrid is the default work model. The Verona office, located
in a coworking space inside the Polo Tecnologico or rented near the
A22 Nogarole Rocca exit, hosts a Tuesday-Thursday in-office rhythm
for local staff; the DACH BDR works remotely from Munich with monthly
visits; engineering hires from outside Verona province are
fully-remote with quarterly in-person gatherings. Compensation is
aligned to Ichnos-Market Italian benchmarks: below the Milan premium,
above the Bologna norm, with a meaningful equity slice for early
hires. Benefits include Italian TFR (treatment de fin de rapporto),
welfare vouchers (Ticket Restaurant plus Edenred Flex), private
health insurance via an UNISalute plan, and pension-fund matching up
to four per cent for employees on open-ended contracts. The company
explicitly commits to the metal-mechanical (CCNL Metalmeccanico
Industria) contract when that becomes economically sensible, since
most Veneto engineering candidates already understand it.

Recruiting channels are LinkedIn for senior roles, the Universities
of Verona and Trento computer-science departments for internships
and junior hires, and the logistics-industry network (Confindustria
Verona, Albo Autotrasportatori) for the domain-expert hires. A formal
graduate programme targeting the University of Verona's Data Science
Master's curriculum opens in Year Two, pipelining a handful of
interns per year into junior engineering or customer-success seats.

### 10. Marketing & Sales Strategy

Marketing and sales are designed as a single integrated machine
because, at this stage of the company, the two functions share
people, materials and success criteria. The go-to-market motion is
founder-led and outbound-heavy in months one through twelve, pivoting
to inbound-augmented in months thirteen through twenty-four, and
reaching a balanced mix thereafter.

The ideal-customer profile is explicit. On the carrier side: a
family-owned road-freight operator with twenty to one hundred
vehicles, headquartered in the Verona province or along the A22
between Trento and Modena, generating two to twenty million euros of
annual revenue, already using at least one telematics provider for
regulatory reasons, and with a designated logistics manager who
personally feels the pain of spreadsheet-tracking. On the shipper
side: a Veneto manufacturing SME with fifty to five hundred employees,
exporting at least thirty per cent of revenue into the DACH region,
running a controlling office that already speaks the language of OTIF
and cost-per-kilogramme.

Outbound sequencing is tight. Week one: research and pre-qualification
using Cerved and the Confindustria Verona member directory. Week two:
a LinkedIn connection to the logistics manager with a personalised
opener referring to a concrete operational reality (an A22 closure,
an upcoming customs rule change, a Vinitaly shipping window for
agri-food customers). Week three: a phone call from the
logistics-domain-expert co-founder. Week four: an in-person
thirty-minute meeting at the prospect's office or at the LogiTrack
Verona space; the discovery session ends either with a demo scheduled
or an explicit no-go. Week six: a pilot proposal for a specific
corridor with a thirty-day no-commitment window.

Trade shows are the second pillar. The priority list is SPS Italia in
Parma for the automation crossover, LogiMAT Stuttgart for the DACH
market presence, and SITL Paris for the French-adjacent corridor. The
founder and the customer-success lead cover all three annually; a
booth at SPS Italia begins in year two and scales in year three.
Verona-specific events — Fieragricola for the wine and agri-food
crossover, FERTRAM Verona in the logistics-oriented editions, the
ZAI-hosted "Quadrante Europa Innovation Day" — are attended from year
one. Specific named partnerships anchor the strategy: a joint
outreach with Consorzio ZAI (offering co-promotion to the tenants of
the Quadrante Europa industrial park), a referral arrangement with
Confindustria Verona's logistics commission, and a reseller
agreement with five to eight Verona-based IT systems integrators who
serve SMEs as their long-term IT partner.

Digital marketing is organic-heavy. The marketing blog publishes one
long-form post per fortnight on operationally-grounded topics: "Come
calcolare l'OTIF sul corridoio del Brennero", "Cosa cambia con AIDA
2026 per le PMI di Mozzecane", "Il vero costo di un ritardo di
tre ore sul trasporto Verona-Monaco". The target channels are
Logistica Management magazine (which syndicates selected posts),
Uomini e Trasporti (the Italian road-freight trade press), and a
LinkedIn newsletter that currently reaches approximately three
thousand logistics and supply-chain readers across the Veneto and
Lombardia. SEO focuses on Italian long-tail keywords: "tracking
spedizioni tempo reale Verona", "integrazione AIDA dogane",
"visibilità intermodale Quadrante Europa". Paid acquisition is
reserved for specific campaigns — a LinkedIn retargeting campaign
after SPS Italia, a Google Ads burst around Vinitaly — with a
carefully measured cost per qualified meeting target of roughly two
hundred euros.

Content marketing also includes white papers co-authored with
Confindustria Verona and the University of Verona's Department of
Business Administration, typically one per year, positioning
LogiTrack as a credible voice in the territorial conversation about
logistics digitalisation. Customer case studies are published as
video testimonials (one per quarter once the first references are
live) and distributed through the trade-press channels.

Sales enablement is a growing repository of assets: pitch decks
customised by persona (dispatcher, CFO, controlling-office manager,
family owner), an ROI calculator that takes five inputs (number of
shipments, average corridor distance, current time spent tracking per
shipment, fuel cost per kilometre, average delay cost) and outputs a
twelve-month ROI chart, and a trial-enablement kit with pre-built
dashboard templates for Verona-Monaco, Verona-Rotterdam and
Verona-Bologna corridors. Pricing exceptions above five per cent
require CEO approval; the company resists the temptation to
list-price-discount as a matter of strategic discipline.

The measurement framework tracks meetings booked per week, pilots
started per month, pilots converted to paid per quarter and ARR per
month. The target at month twelve is twenty paying tenants generating
approximately eighty thousand euros in monthly recurring revenue;
month twenty-four targets one hundred tenants and three hundred
thousand euros MRR. Those numbers are public to every employee via
a live dashboard inside the company Notion, reinforcing shared
accountability.

Partner channel economics deserve a dedicated paragraph because they
are a meaningful contributor to growth from year two onwards. The
reseller programme for IT systems integrators pays a fifteen per cent
commission on the first-year contract value and an ongoing five per
cent recurring commission for the life of the tenant, in exchange for
the reseller handling the majority of the sales cycle, the initial
onboarding and the tier-one support. Typical resellers in the Verona
ecosystem — mid-sized ICT houses with ten to fifty employees serving
a portfolio of SME clients — generate between three and eight
LogiTrack deals per year once the relationship is mature. The
commission structure is deliberately at the high end of the SaaS
industry norm because the trust architecture in the Italian SME
segment is reseller-centric: the commercialista, the legal advisor
and the ICT integrator form the three-party advisory triangle that
most SMEs consult before making any material technology decision.
Paying the reseller meaningfully recognises their role as the
custodian of that trust.

Events strategy is complemented by direct-outbound motion designed
for asynchronous efficiency. The founder and the logistics-domain
hire maintain a curated prospect list of approximately six hundred
Verona-corridor operators, scored on revenue band, vehicle-count band
and digital-maturity signal (LinkedIn profile richness, website
modernity, presence on industry trade press, observable investment
in telematics). The list is worked in cohorts of fifty per month,
with a sequenced touchpoint programme that blends LinkedIn outreach,
personalised email, warm introductions sourced from Confindustria
Verona and the three founding advisors, and an in-person meeting
whenever the prospect is within an hour's drive of Mozzecane. The
cohort conversion rate has been calibrated against the first two
months of execution: ten per cent of a cohort accepts an initial
meeting, thirty per cent of those progress to a pilot proposal,
and forty per cent of pilots close as paid Professionale contracts.
Those conversion rates imply a cohort yield of approximately six new
customers per month of fifty prospects worked; the planning model
builds in a seasonality adjustment for the summer period (mid-July
through late August) when Verona businesses largely close and
outreach yields plunge to near zero.

### 11. Financial Projections

The financial model assumes a seed round of one point two million
euros raised in month three, sufficient to fund operations through
month twenty-four at a run rate below burn. Year one revenue is
projected at approximately one hundred twenty thousand euros
(predominantly Professionale subscriptions, a handful of Starter and
one Enterprise), with operating expenses of approximately eight
hundred thousand euros, giving a net loss of six hundred eighty
thousand euros and cash reserves at year-end of approximately five
hundred thousand euros. Year two revenue climbs to eight hundred
fifty thousand euros as the Phase 2 feature rollouts and trade-show
visibility materialise into new customers, with operating expenses of
roughly one million three hundred thousand euros, producing a smaller
net loss and a Series A fundraise targeted at month twenty-two of
four to six million euros at a pre-money valuation of twelve to
sixteen million. Year three revenue crosses two point one million
euros on the back of DACH expansion and the Enterprise deal cadence,
operating expenses of two million euros, producing a near break-even
year and positioning the company for growth-stage milestones.

Break-even on a cash basis is projected for month thirty; break-even
on a GAAP basis, including the Series A capital runway, is month
thirty-six. Burn rate across the first twenty-four months averages
seventy-two thousand euros per month, rising from thirty thousand
euros in the early engineering-heavy months to one hundred twenty
thousand euros during the Phase 3 integration push. The fundraising
plan is sequenced as seed at twelve months of runway, Series A at
eighteen months of runway, Series B only if the DACH expansion
demands capital beyond organic cash flow. The company deliberately
retains optionality on whether to raise a Series B at all; the
product economics, if customer acquisition cost stays under one
thousand five hundred euros and churn stays under eight per cent
annually, support profitable growth from Series A onwards.

---

## Part V — Excellence & Growth

### 12. Operations Manual

Day-to-day operations run on a small number of explicit rituals. A
fifteen-minute stand-up every morning for engineering, alternated
with a twenty-minute go-to-market stand-up for sales and customer
success; a thirty-minute product review on Wednesday afternoons
where the founder, CTO and customer-success lead align on the next
two sprints; a ninety-minute board-level metrics review on the first
Monday of every month. All rituals are documented in the engineering
wiki with the template inputs, expected outputs and escalation paths,
so new hires absorb the rhythm through observation rather than
anecdote.

Service level agreements are tiered by plan. Starter tenants get
best-effort support from Monday to Friday nine to eighteen, with a
published email address and a four-hour first-response SLA during
business hours. Professionale tenants get a shared-queue Slack
Connect channel with the customer-success team plus a two-hour
first-response SLA during business hours and a four-hour SLA outside
them. Enterprise tenants get a named customer-success owner, a
24x7 on-call escalation path for severity-one issues, and a
contractually binding ninety-nine point nine per cent uptime SLA
with service credits structured as a percentage of the monthly fee.

Runbooks live in the operations wiki and are tested quarterly via
game-day exercises. The catalogue includes: MongoDB primary failover,
Redis cluster failure, AWS region failure with Aruba Cloud cutover,
OSRM capacity spike, AIDA outage with manual declaration fallback,
telematics provider quota exhaustion, JWT secret rotation, data-
subject erasure request, GDPR breach notification to the Garante, and
customs inspection escalation to the tenant's broker. Every runbook
carries an explicit owner, a last-tested date and a post-exercise
retrospective log.

### 13. Risk Management

The principal risks are enumerated in a matrix with likelihood,
impact and mitigation ownership. Operational risks dominate: a
telematics provider API change (likelihood medium, impact high,
mitigated by a connector-abstraction layer and contractual
change-notice windows); an AIDA outage during customs peak season
(likelihood low, impact critical, mitigated by retry queues and a
manual-fallback workflow trained on every customer-success hire); a
key-person dependency on the CTO (likelihood low, impact catastrophic
in the first year, mitigated by documented architecture, pair
programming on critical modules, and a second senior backend hire at
month nine).

Commercial risks follow: a large European competitor launches a
Verona region (likelihood medium over five years, impact high,
mitigated by the sovereignty and local-context moat plus an aggressive
land-grab on the Italian SME segment); commoditisation pressure from
open-source alternatives (likelihood low in the short term, impact
medium, mitigated by pivoting to vertical modules and integration
depth); a major customer churns at renewal (likelihood medium for
any single year, impact measurable but recoverable through the
portfolio-effect of a well-diversified book).

Regulatory risks are treated with a risk officer reviewing the
quarterly legislative landscape and flagging changes that affect the
product. The AIDA rule evolution toward the "Single Window"
architecture, for example, is already on the radar for Phase 4.

Contingency plans cover the two worst-case scenarios explicitly.
First, a twelve-month runway shortfall before Series A: the plan is
a forty per cent headcount cut concentrated in late-stage sales
hires, a pivot to a services-heavy revenue mix to lengthen runway,
and an aggressive bridge round from existing investors or strategic
customers. Second, a major data breach: the plan triggers the GDPR
seventy-two-hour notification workflow, engages external counsel
from Studio Legale Tonucci, activates a public-relations response
template pre-drafted with a Verona-based PR agency, and commits to a
post-incident product remediation timeline published within thirty
days of the incident.

A third contingency, less frequently discussed in SaaS playbooks but
materially relevant to the logistics context, is a prolonged regional
infrastructure disruption — for example, a multi-day closure of the
A22 due to a landslide or a systemic freight-rail strike affecting
Quadrante Europa. Such disruptions do not directly threaten the
platform's operation, but they dramatically reshape the tenant's
workload: traffic drops, support load shifts, delay reporting
becomes the single hottest dashboard, and the success team's capacity
is stretched thin. The contingency plan anticipates this pattern and
pre-stages two responses: a narrative template for communicating with
tenants ("we see the A22 closure; here is what your dashboard will
show and how we'll help you explain delays to your customers"), and
a temporary feature toggle for a "disruption-annotated" timeline
view that flags every affected shipment with a regional-event tag
traceable from the tenant's report back to the infrastructure event.
These responses have proved useful in practice during the first year,
when an unplanned A22 closure at the Affi exit disrupted cross-border
traffic for three days.

A fourth contingency plan covers the scenario of a strategic
acquisition rumour or offer at an inopportune moment. Early-stage
companies that find themselves in informal acquisition conversations
often lose focus; the contingency plan is to route any inbound
acquisition contact to a single named founder, to require all
conversations past the initial introduction to involve external
advisors (a boutique M&A firm in Milan is pre-retained on a modest
annual fee), and to commit publicly to the team that no offer will
be pursued below a clearly defined valuation threshold. The policy's
purpose is not to preclude a sale but to protect the company against
the distraction and the internal uncertainty that premature
acquisition discussions inflict on a small team.

### 14. Quality Assurance

Quality is enforced through a layered testing pyramid documented in
`tests/` across the codebase. Unit tests cover pure functions,
service methods and repository methods with fifty per cent minimum
coverage at the package level and eighty per cent on the
shipment-service and custody-service modules. Integration tests
exercise real MongoDB and Redis via docker-compose fixtures inside
GitHub Actions; every REST endpoint has at least one happy-path and
one failure-path test. Contract tests verify that the OpenAPI schema
exposed by the backend is honoured by the frontend client. End-to-end
tests run Playwright against a full docker-compose spin-up and cover
the ten most-critical user journeys: tenant onboarding, first
shipment creation, telematics ingestion, geofence dwell alert, AIDA
declaration, chain-of-custody append, shipper-portal view, CSV
export, WebSocket reconnect and multi-tenant isolation.

Code review is mandatory and two-person: every pull request needs an
approval from a reviewer who did not author any of the changes, and
the reviewer checklist covers correctness, tests, observability,
security, documentation and backward compatibility. Static analysis
runs in CI on every pull request: go vet, golangci-lint with the
house configuration, Semgrep with the OWASP ruleset, Gitleaks for
secret detection, Trivy for container image vulnerability scanning,
and npm audit for the frontend dependency tree.

Continuous integration gates are: linters pass, unit tests pass,
integration tests pass, coverage holds or improves, Semgrep finds no
new critical issues, Trivy finds no new high-severity vulnerabilities.
Continuous deployment is triggered only on merges to `main`; a
deploy to staging runs automatically, a deploy to production
requires manual approval from a named on-call engineer. Post-deploy,
the health endpoint is probed for fifteen minutes; a failing probe
triggers automatic rollback.

### 15. Customer Success

Customer success is treated as a product feature rather than a cost
centre. The onboarding programme is six weeks for Professionale
tenants and ten weeks for Enterprise: week one is kickoff with the
success owner, week two is data migration and master-data import,
week three is the first telematics integration, week four is the
dashboard configuration and the user-training session, week five is
the shadow-operation period where the tenant runs LogiTrack alongside
their existing process, and week six is the cutover and the first
SLA review. A tenant is considered successfully onboarded when
seventy per cent of its weekly shipments flow through the platform
for two consecutive weeks without manual reconciliation.

Ongoing success follows a quarterly-business-review cadence for
Professionale and Enterprise tenants and a monthly email newsletter
plus an annual check-in for Starter tenants. The QBR is a structured
sixty-minute session with the customer-success owner and the tenant's
logistics manager or equivalent executive: a review of the last
ninety days' SLA performance, a preview of the next quarter's
product roadmap relevant to the tenant, an action-plan for any
outstanding support tickets, and a two-question satisfaction survey
(NPS plus a single-question CSAT). Retention is measured monthly;
annual net retention target is one hundred fifteen per cent,
accounting for plan upgrades and per-shipment overages offsetting
the natural five per cent annual churn.

The customer-health score is an internal composite metric updated
weekly from seven inputs: login frequency of the primary user,
percentage of shipments flowing through the platform versus through
legacy processes, integration error rate, open support ticket count
and age, invoice payment punctuality, the result of the most recent
quarterly business review, and the pipeline of upsell conversations
logged in the CRM. Tenants scoring below a threshold trigger a
proactive intervention: a call from the customer-success owner,
sometimes a founder touch, occasionally a dedicated engineering
effort to resolve a specific integration pain point. The intervention
programme has driven the early cohort's annualised churn to below
three per cent, well inside the planning assumption and a meaningful
competitive differentiator when recruiting anchor tenants for the
DACH expansion.

Support tooling is deliberately overbuilt relative to company size
because the Verona SME logistics buyer is exceptionally
support-intensive in the first six months of tenancy. The platform
ships with an in-product chat widget routed to the customer-success
inbox during business hours, a self-service knowledge base with
Italian-language articles covering the top fifty support scenarios,
a weekly office-hours webinar where any tenant user can drop in and
ask questions, and an escalation path that routes any ticket
unresolved after four business hours to a named executive. The
investment is substantial for a company of this size — roughly
fifteen per cent of total cost base in the early years — but the
economics of a one-hundred-fifteen-per-cent net retention number
justify it many times over.

### 16. Partnerships & Ecosystem

Partnerships are sequenced by strategic leverage. Consorzio ZAI sits
at the top as the governance body of Quadrante Europa; a formal
memorandum of understanding signed in year one anchors LogiTrack as
a recommended digital tool for ZAI tenants. Rete Ferroviaria Italiana
follows as the intermodal infrastructure partner; a commercial
arrangement with RFI's Terminali Italia unit opens up direct booking
APIs into the Verona QE slots. Telepass supplies the transit data
that materially improves ETA accuracy on the A22; the partnership is
framed as a data-exchange agreement rather than a reseller deal.
Agenzia delle Dogane, while not a commercial partner in the
traditional sense, is the regulatory counterparty whose AIDA API
access is secured through the standard Autorizzazione Tecnica process.
Ecosystem technology partners include Viasat, Octo, Geotab, Movyon,
Zucchetti and TeamSystem (integrations rather than resellers). Albo
Autotrasportatori integration provides carrier-verification hooks.
Confindustria Verona, Fondazione Edulife and the University of Verona
round out the institutional relationships.

Each partnership has a dedicated owner, a quarterly review cadence
and a documented exit criterion so that relationships that are not
producing business value are wound down rather than allowed to
consume attention indefinitely. The Consorzio ZAI relationship, for
example, is reviewed annually against the count of ZAI tenants
actively using LogiTrack; the arrangement continues as long as
LogiTrack delivers visibility value inside the ZAI ecosystem and
ZAI provides co-promotion access to its tenant base. The RFI
agreement hinges on maintaining a functioning API to the intermodal
booking system and on LogiTrack's compliance with RFI's data-handling
requirements for operationally sensitive information.

The ecosystem extension beyond the initial partners is pursued on a
rolling basis. Year two targets a partnership with Autostrada del
Brennero S.p.A. to integrate live A22 incident feeds; year three
targets a partnership with Autorità di Regolazione dei Trasporti
(ART) for regulatory-grade reporting; year four explores a federated
relationship with equivalent visibility platforms in other European
corridors (for example, a Rhine-corridor platform in the Netherlands
or a Channel-tunnel-corridor platform based in the UK) to offer
customers end-to-end visibility across platform boundaries, similar
in spirit to the old GDS interoperability that once united
independent airline reservation systems.

### 17. Exit Strategy

The exit strategy is articulated without compulsion: the founders
prefer an eight-to-ten-year build toward market leadership and will
resist pressure to sell early. Realistic acquisition targets,
plausible at Series B or later, are the European arm of Project44
(most natural strategic fit, consolidating visibility with a strong
Italian footprint), Transporeon or its parent Trimble (complementary
freight-forwarder focus), TeamSystem Logistic (domestic ERP
consolidator that would benefit from a visibility product), and
Descartes Systems Group (global platform that regularly acquires
regional specialists). A trade sale valuation model at five to seven
times run-rate ARR implies a range of fifty to two hundred million
euros at the Series B milestone. An IPO path, while not the primary
scenario, is compatible with the plan if ARR reaches fifty million
euros and the company achieves pan-European coverage; Euronext
Growth Milan is the most plausible listing venue given the Italian
anchor and the PIR-compatible investor appetite for mid-cap tech
stories.

The succession considerations that accompany any exit discussion are
planned ahead of time. Key-person insurance is obtained on the
founder and the CTO from the moment the seed round closes; the cap
table is kept clean with a single class of common stock and a
standard four-year vesting with one-year cliff on all employee
grants; the employment agreements contain reasonable non-compete
clauses enforceable under Italian law (valid for twelve months
post-termination, geographically limited to the Veneto, compensated
at fifty per cent of base salary). The shareholders' agreement
provides for a drag-along right above a seventy-five per cent
threshold and a tag-along right on any secondary sale above ten per
cent of the cap table, balancing the founders' control over strategic
direction against minority shareholders' liquidity rights. These
governance details are not glamorous but they materially increase the
valuation multiple at exit because they remove the friction and the
legal-due-diligence concerns that cause acquirers to discount
offers.

A realistic probability-weighted view of exit paths assigns the
following rough odds across a ten-year horizon: fifty per cent for a
strategic acquisition at Series B or Series C stage by one of the
European visibility platforms or by a domestic ERP consolidator,
twenty per cent for a private-equity buyout once revenue crosses
thirty million euros with demonstrated profitability, fifteen per cent
for a continued independent growth trajectory without exit, ten per
cent for an IPO on Euronext Growth Milan or on a pan-European venue,
and five per cent for a failure scenario captured by the contingency
plans in the risk-management section. The expected value calculation
favours strategic acquisition but the planning posture is to build
the company as if independent growth were the outcome, because that
discipline maximises the enterprise value regardless of which path
eventually wins.

---

## Appendix A — Glossary (non-exhaustive)

- **AIDA** — Automazione Integrata Dogane e Accise, the customs
  information system of Agenzia delle Dogane e dei Monopoli.
- **Albo Autotrasportatori** — the Italian national register of road
  freight carriers maintained by the Ministry of Transport.
- **CMR** — Convention on the Contract for the International Carriage
  of Goods by Road, shorthand for the consignment note.
- **Consorzio ZAI** — the public consortium that manages the Verona
  industrial zone and the Quadrante Europa intermodal platform.
- **ETA** — Estimated Time of Arrival.
- **Geofence** — a virtual polygon on the map that triggers events
  when a vehicle enters or exits it.
- **OSRM** — Open Source Routing Machine, a routing engine used for
  ETA and path geometry.
- **OTIF** — On-Time, In-Full, a standard supply-chain KPI.
- **PEC** — Posta Elettronica Certificata, the Italian certified email
  system.
- **Quadrante Europa** — the intermodal freight terminal of Verona
  operated by Consorzio ZAI.
- **RFI** — Rete Ferroviaria Italiana, the Italian rail infrastructure
  operator.
- **SDI** — Sistema di Interscambio, the Italian electronic-invoicing
  clearing system.
- **Telepass** — the electronic toll payment and transit data
  platform on Italian motorways.
- **T1/T2** — the transit document types used in the European Union
  customs framework; T1 applies to non-union goods under transit, T2
  to union goods moving through a third-country transit.
- **MRN** — Master Reference Number assigned by customs to each
  transit declaration.
- **CCNL** — Contratto Collettivo Nazionale di Lavoro, the national
  collective bargaining agreement; the `Autotrasporto Merci e
  Logistica` CCNL is the reference contract for road-freight drivers.
- **Partita IVA** — the Italian VAT number uniquely identifying every
  commercial entity.
- **DPIA** — Data Protection Impact Assessment, required under GDPR
  for high-risk processing activities.
- **RPO / RTO** — Recovery Point Objective and Recovery Time
  Objective, the two canonical disaster-recovery targets.

## Appendix B — Verona corridor data points

- Quadrante Europa annual throughput: 8+ million tonnes across road
  and rail.
- Verona-Mantova-Trento-Bolzano annual commercial interchange:
  approximately €56 billion.
- Italian logistics sector annual value added: approximately
  €120 billion.
- Share of Italian freight moved by road: approximately 85%.
- Mozzecane distance to A22 Nogarole Rocca exit: 5 km; to Quadrante
  Europa freight terminal: 12 km.
- Estimated number of road-freight operators within 200 km of Verona:
  approximately 3,500.
- Share of Italian SMEs still tracking supply chain via spreadsheets:
  approximately 73%.
- Quadrante Europa surface area: over 2,000,000 m²; tenant companies
  hosted: over 100.
- Principal Brenner corridor motorway: A22 Autostrada del Brennero,
  administered by Autostrada del Brennero S.p.A.
- Neighbouring comuni of Mozzecane relevant to the corridor
  catchment: Nogarole Rocca, Povegliano Veronese, Villafranca di
  Verona, Vigasio, Valeggio sul Mincio.
- Primary Veneto industrial districts intersecting the LogiTrack
  customer base: Distretto Mobile di Verona, Distretto Meccanico
  Veronese, Distretto del Vino (Valpolicella, Bardolino, Custoza,
  Soave), and the agri-food cluster (meat, dairy, preserved
  vegetables) of the Verona plain.
- Reference figure for ERP adoption among Veneto SMEs of ten to two
  hundred fifty employees: approximately thirty-eight per cent.
- Reference figure for cloud service adoption among Veneto SMEs:
  approximately forty-two per cent.
- Reference figure for Piano Transizione 4.0 uptake in the Veneto:
  approximately eleven point seven per cent of eligible SMEs.

## Appendix C — Strategic anchor facts

The following ten facts form the quantitative spine of every sales
pitch, investor update and customer-success review LogiTrack
produces. They are memorised verbatim by every commercial employee
during onboarding.

1. Quadrante Europa is Europe's second-largest intermodal freight
   platform, handling over eight million tonnes a year.
2. The Verona-Mantova-Trento-Bolzano corridor processes approximately
   fifty-six billion euros of commercial interchange annually.
3. Seventy-three per cent of Italian supply-chain operators still
   track consignments with spreadsheets or paper workflows.
4. Mozzecane sits five kilometres from the A22 Nogarole Rocca exit
   and twelve kilometres from the Quadrante Europa freight terminal.
5. The Italian logistics sector contributes roughly one hundred
   twenty billion euros of value added per year.
6. Eighty-five per cent of Italian freight moves by road.
7. Consorzio ZAI governs Quadrante Europa on behalf of the
   Municipality of Verona and the Province.
8. The A22 Autostrada del Brennero is administered by Autostrada del
   Brennero S.p.A., a joint-stock company with public shareholders.
9. The Albo Autotrasportatori is the national register maintained by
   the Ministry of Transport for road-freight carriers.
10. AIDA (Automazione Integrata Dogane e Accise) is the customs
    information system of Agenzia delle Dogane e dei Monopoli and
    the central integration target for cross-border operations on
    the Brenner corridor.

---

## Part XVIII — Mission II Consolidation Record (v0.2.0, 2026-04-17)

### Scope of the consolidation

The second mission focused on closing the five gaps identified in
the Mission I audit: a realistic telematics simulator, an OSRM-backed
route optimiser with a local fallback, an ETA computation service
with moving-average smoothing, a hardened WebSocket handler, and a
live-map Vue component that consumes it. The five gaps are closed
with the following modules:

- `backend/cmd/simulator/main.go` — the simulator authenticates with
  a short-lived HS256 JWT minted against the same `JWT_SECRET` the
  server runs with, then posts waypoints at 1 Hz to the three demo
  shipments. The interpolation routine converts the pre-computed
  polylines into smooth per-second movement without introducing a
  spline library. When run in looping mode it is a long-lived
  companion container for demos; in the E2E test it runs for a
  bounded 30–60 seconds and exits cleanly.

- `backend/internal/services/route_optimizer.go` — the route
  optimiser acquires three defences against common failure modes:
  a 1000-entry LRU cache keyed by the canonical coordinate string,
  a host allow-list against SSRF, and a straight-line fallback
  estimator computed from great-circle distance at 70 km/h. The
  allow-list is deliberately strict — IP literals are accepted only
  when explicitly listed, so even an operator with write access to
  `OSRM_BASE_URL` cannot pivot to the cloud metadata service on
  169.254.169.254.

- `backend/internal/services/eta_service.go` — the ETA service
  maintains a per-shipment speed filter. Each waypoint contributes
  either its explicit `speedKph` telemetry or a derived speed from
  the interval between the previous and current position. The filter
  is a bounded moving average with 20 samples, clipped to a sanity
  band of 5–120 km/h to blunt both "parked at customs" and "GPS
  jitter" extremes.

- `backend/internal/handlers/stream.go` — the WebSocket handler is
  rewritten to expose a JWT-over-subprotocol path suitable for
  browsers (`Sec-WebSocket-Protocol: logitrack.jwt.v1,<token>`),
  plus the legacy `?access_token=` and the header-based path for
  CLI clients. Origin enforcement is a closure over the configured
  allow-list; a per-IP handshake rate limiter and a per-connection
  inbound rate limiter collaborate to blunt DoS. Idle timeout is
  lifted to five minutes, matching the Part IX runbook expectation.

- `frontend/src/components/ShipmentMap.vue` — the map is no longer
  a placeholder. It lazy-loads Leaflet and its CSS, decodes the
  OSRM Polyline6 geometry, plots the planned route, and repositions
  a pulsing amber circle marker on every incoming WebSocket event.
  The ETA label is `aria-live="polite"` so screen readers announce
  updates without trapping navigation focus.

### Security posture

`govulncheck` returns zero findings after the consolidation. Trivy
FS reports the project clean apart from an informational lag in its
vulnerability-database refresh, noted in `docs/RISK-ACCEPTANCES.md`.
Semgrep's WebSocket origin-check rule produces one finding that the
`nosemgrep` annotation explains: the check IS present but hidden
inside a closure. No HIGH or CRITICAL dependency vulnerabilities
remain after bumping `golang-jwt/jwt/v5`, `go-redis/v9`,
`golang.org/x/net`, `google.golang.org/grpc` and
`go.opentelemetry.io/otel`.

### Italian compliance additions

Three changes strengthen the regulatory envelope:

1. Plate validation now accepts both post-1994 and historical Italian
   plate formats, matching the realities of carriers that still
   operate vintage Fiat 242 vans registered before May 1994.
2. Shipments carry explicit `adrClass` (ADR 2023 classes 1–9 with
   sub-classes) and `atpClass` (ATP 1970 categories IR/RNA/RRB/FRC/IN)
   so dispatchers can assign loads to compliant vehicles.
3. A `TelepassTollCode` value object captures the AISCAT gazetted
   code plus a human label per tolled stretch; the quarterly refresh
   procedure is documented in the runbook.

### Release status

LogiTrack 0.2.0 is tagged as PASS WITH NOTES: the functional
acceptance criteria are met, govulncheck and Trivy are clean, the
three-shipment simulator reaches the browser in under two seconds
for the first waypoint, and the sales-enablement collateral is
shipped. Outstanding items are logged in `docs/TECHNICAL-DEBT.md`
and `docs/RISK-ACCEPTANCES.md` and do not block commercial pilot
engagement on the Starter / Professionale tiers.

