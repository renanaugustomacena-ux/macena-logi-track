// Canned demo data for the static GitHub Pages demo. Activated when
// VITE_DEMO_MODE=true at build time. None of this code is reachable
// from a normal production build of the SPA.
//
// The data is deliberately fictional but plausible: real Italian
// place names, plausible CER codes, plausible Albo categorie. No
// real customer / driver / VAT number must ever appear here.

export const DEMO_TENANT_ID = 'demo-tenant';
export const DEMO_USER_ID = 'demo@logitrack.it';

// Shipments — three plausible Veneto / Brennero corridor scenarios.
// The first one is the "live" one whose marker animates along a
// polyline; the other two are static (delivered / booked).
export const DEMO_SHIPMENTS = [
  {
    id: 'demo-shp-001',
    tenantId: DEMO_TENANT_ID,
    reference: 'CMR-2026-04-001',
    carrier: 'Autotrasporti Veneti S.r.l.',
    mode: 'multimodal' as const,
    status: 'in_transit' as const,
    consignor: { name: 'Officine Meccaniche Veronesi', city: 'Mozzecane', country: 'IT' },
    consignee: { name: 'Müller GmbH', city: 'München', country: 'DE' },
    origin: { type: 'Point' as const, coordinates: [10.7950, 45.3567] as [number, number] },
    destination: { type: 'Point' as const, coordinates: [11.5755, 48.1374] as [number, number] },
    currentPosition: { type: 'Point' as const, coordinates: [11.0100, 45.9300] as [number, number] },
    routePolyline: '',
    etd: '2026-04-28T05:00:00Z',
    eta: '2026-04-28T13:30:00Z',
    updatedAt: '2026-04-28T07:12:00Z',
    adrClass: '',
    atpClass: '',
    vehiclePlate: 'GA512MX',
  },
  {
    id: 'demo-shp-002',
    tenantId: DEMO_TENANT_ID,
    reference: 'CMR-2026-04-002',
    carrier: 'Trasporti Garda S.n.c.',
    mode: 'road' as const,
    status: 'delivered' as const,
    consignor: { name: 'Cantina Sociale Bardolino', city: 'Bardolino', country: 'IT' },
    consignee: { name: 'Distributori Roma S.p.A.', city: 'Roma', country: 'IT' },
    origin: { type: 'Point' as const, coordinates: [10.7104, 45.5450] as [number, number] },
    destination: { type: 'Point' as const, coordinates: [12.4964, 41.9028] as [number, number] },
    currentPosition: { type: 'Point' as const, coordinates: [12.4964, 41.9028] as [number, number] },
    routePolyline: '',
    etd: '2026-04-26T04:00:00Z',
    eta: '2026-04-26T18:30:00Z',
    updatedAt: '2026-04-26T18:12:00Z',
    adrClass: '',
    atpClass: 'C',
    vehiclePlate: 'FA987BC',
  },
  {
    id: 'demo-shp-003',
    tenantId: DEMO_TENANT_ID,
    reference: 'CMR-2026-04-003',
    carrier: 'Spedizioniere Mantova S.r.l.',
    mode: 'road' as const,
    status: 'booked' as const,
    consignor: { name: 'Cooperativa Agricola Mantovana', city: 'Mantova', country: 'IT' },
    consignee: { name: 'Mercato Ortofrutticolo Milano', city: 'Milano', country: 'IT' },
    origin: { type: 'Point' as const, coordinates: [10.7914, 45.1564] as [number, number] },
    destination: { type: 'Point' as const, coordinates: [9.1900, 45.4642] as [number, number] },
    routePolyline: '',
    etd: '2026-04-29T03:30:00Z',
    eta: '2026-04-29T07:00:00Z',
    updatedAt: '2026-04-28T20:00:00Z',
    adrClass: '',
    atpClass: '',
    vehiclePlate: 'EA445DG',
  },
];

// Polyline waypoints (lon, lat) for the live shipment, Mozzecane →
// München via the Brennero. ~12 stops along the A22 + A12. The
// mock WebSocket walks through these on a timer to animate the marker.
export const DEMO_LIVE_POLYLINE: [number, number][] = [
  [10.7950, 45.3567], // Mozzecane (origin)
  [10.9870, 45.4380], // Verona Sud
  [11.0100, 45.9300], // Trento
  [11.1232, 46.0700], // Mezzocorona
  [11.2210, 46.4990], // Bolzano
  [11.4534, 46.7780], // Bressanone
  [11.5070, 47.0030], // Brennero
  [11.4022, 47.2680], // Innsbruck
  [11.3829, 47.6500], // Kufstein
  [11.4400, 47.9000], // Rosenheim
  [11.5300, 48.0500], // Holzkirchen
  [11.5755, 48.1374], // München (destination)
];

// Demo timeline events that get pushed into the TrackingTimeline as
// the live shipment progresses. Index aligns with DEMO_LIVE_POLYLINE.
export const DEMO_TIMELINE_TYPES: string[] = [
  'departure',
  'position_update',
  'position_update',
  'position_update',
  'geofence_enter',
  'position_update',
  'position_update',
  'position_update',
  'position_update',
  'position_update',
  'position_update',
  'arrival',
];

// Rifiuti anagrafiche — minimal but plausible.
export const DEMO_PRODUTTORI = [
  {
    id: 'demo-prod-001',
    tenantId: DEMO_TENANT_ID,
    ragioneSociale: 'Officina Meccanica Demo S.r.l.',
    codiceFiscale: '00000000001',
    sede: { via: 'Via Industria 12', comune: 'Mozzecane', provincia: 'VR', cap: '37060' },
    createdAt: '2026-04-01T08:00:00Z',
    updatedAt: '2026-04-01T08:00:00Z',
  },
  {
    id: 'demo-prod-002',
    tenantId: DEMO_TENANT_ID,
    ragioneSociale: 'Carrozzeria Demo & Figli S.n.c.',
    codiceFiscale: '00000000002',
    sede: { via: 'Via del Lavoro 3', comune: 'Villafranca', provincia: 'VR', cap: '37069' },
    createdAt: '2026-04-02T08:00:00Z',
    updatedAt: '2026-04-02T08:00:00Z',
  },
];

export const DEMO_TRASPORTATORI = [
  {
    id: 'demo-trasp-001',
    tenantId: DEMO_TENANT_ID,
    ragioneSociale: 'FRO Trasporti Demo S.r.l.',
    codiceFiscale: '00000000010',
    alboCategoria: '4',
    alboClasse: 'C',
    alboNumero: 'VR/00000/C/2025',
    alboScadenza: '2030-12-31T00:00:00Z',
    createdAt: '2026-03-01T08:00:00Z',
    updatedAt: '2026-03-01T08:00:00Z',
  },
];

export const DEMO_DESTINATARI = [
  {
    id: 'demo-dest-001',
    tenantId: DEMO_TENANT_ID,
    ragioneSociale: 'Impianto di Recupero Demo S.p.A.',
    codiceFiscale: '00000000020',
    autorizzazioni: [
      { cer: '200201', operazione: 'R3', scadenza: '2030-06-30T00:00:00Z' },
      { cer: '170504', operazione: 'D9', scadenza: '2030-06-30T00:00:00Z' },
    ],
    createdAt: '2026-03-10T08:00:00Z',
    updatedAt: '2026-03-10T08:00:00Z',
  },
];

export const DEMO_FIR = [
  {
    id: 'demo-fir-001',
    tenantId: DEMO_TENANT_ID,
    state: 'vidimato',
    cer: '200201',
    operazioneDestino: 'R3',
    pesoKg: 2400,
    produttoreId: 'demo-prod-001',
    trasportatoreId: 'demo-trasp-001',
    destinatarioId: 'demo-dest-001',
    numeroRENTRI: 'FIR-DEMO-A1B2C3D4',
    vidimatoAt: '2026-04-28T08:30:00Z',
    firmaProduttoreAt: '2026-04-28T08:30:00Z',
    createdAt: '2026-04-28T08:00:00Z',
    updatedAt: '2026-04-28T08:30:00Z',
  },
  {
    id: 'demo-fir-002',
    tenantId: DEMO_TENANT_ID,
    state: 'draft',
    cer: '170504',
    operazioneDestino: 'D9',
    pesoKg: 8200,
    produttoreId: 'demo-prod-002',
    trasportatoreId: 'demo-trasp-001',
    destinatarioId: 'demo-dest-001',
    createdAt: '2026-04-28T09:00:00Z',
    updatedAt: '2026-04-28T09:00:00Z',
  },
];

export const DEMO_DURATION_MS = 12 * 30_000; // 6 minutes total animation
export const DEMO_STEP_MS = 30_000;          // 30 s per waypoint
