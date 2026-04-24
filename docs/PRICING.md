# LogiTrack — Pricing & Licensing

All prices in EUR, IVA esclusa, monthly billing, no multi-year lock-in
unless explicitly negotiated.

## Tier matrix

| Capability | Starter | Professionale | Enterprise |
| --- | --- | --- | --- |
| Monthly price | €199 | €599 | Preventivo |
| Shipments/day (fair use) | 10 | 100 | Unlimited |
| Operator users | 2 | 10 | Unlimited |
| Portale committenti | no | yes | yes |
| Telematica integrations | 1 | all | all |
| AIDA dogana | no | yes | yes |
| Slot RFI | no | roadmap | yes |
| SSO (SAML/OIDC) | no | no | yes |
| SLA | best-effort | 99.5% | 99.9% + penali |
| Regione dati | Milano (Aruba) | Milano | Milano or customer-private |
| Support | email 8×5 | email+chat 12×6 | dedicated CSM |

## Feature-gating logic
- The backend enforces tier limits at the `ShipmentService` layer
  (rate-limits, user-count, integration flags) driven by the
  `tenants.plan` field. A migration path from Starter → Professionale
  is a no-op for existing data.

## Self-hosted vs SaaS
- **SaaS** is the default: Aruba Cloud Italia, IT region.
- **Self-hosted** is available for Enterprise: docker compose or
  Kubernetes manifests, MongoDB Enterprise / Redis Enterprise
  supported. Quarterly image updates, patched by the customer via
  helm values or docker pull.
- **Hybrid** available: telematics ingestion on-premise, dashboards
  served from SaaS via reverse proxy — common for customers with
  strict data-residency requirements.

## Custom-enterprise terms
- Multi-year contract 2 / 3 years with 8% / 12% discount.
- SLA penali: 10% credit per 1% availability below target.
- On-site training at Verona / Mozzecane: €1.500 / giorno.
- Connettori custom (EDI tradizionale, mainframe IBM i, SAP ERP):
  preventivo a giornate €800/giorno T&M.

## Billing
- Monthly invoicing via FatturaPA (SDI).
- Payment terms: 30 gg fattura fine mese.
- Decadenza: tolerated one failed month + written notice, then
  suspension after 45 days.

## Trial
- 30-day free trial on Professionale tier, fully-featured. No credit
  card required; only a signed LOI for pilot scope.
