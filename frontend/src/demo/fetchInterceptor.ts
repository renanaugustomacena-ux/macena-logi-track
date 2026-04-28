// Demo-mode fetch interceptor.
//
// In production builds this file is unreachable (the import in main.ts
// is gated on import.meta.env.VITE_DEMO_MODE === 'true').
//
// In demo builds, calling installDemoFetchMock() replaces window.fetch
// with a wrapper that:
//   - intercepts every /api/v1/* and /api/health request and returns a
//     canned JSON response built from src/demo/data.ts
//   - persists newly created / mutated FIR documents in an in-memory
//     mutable copy of DEMO_FIR for the lifetime of the page session
//     so the operator can create-then-vidima-then-transition like a
//     real backend
//   - delegates everything else (asset URLs, third-party tile servers,
//     etc.) to the original fetch
//
// The mock is intentionally permissive: it never returns 401, never
// rate-limits, and never validates the body beyond the bare minimum
// needed to keep the demo coherent. It is a static demo, not a
// backend simulator.

import {
  DEMO_ACTIVITY,
  DEMO_DESTINATARI,
  DEMO_FIR,
  DEMO_PRODUTTORI,
  DEMO_SCADENZE,
  DEMO_SHIPMENTS,
  DEMO_TRASPORTATORI,
} from './data';

type AnyJSON = Record<string, unknown> | unknown[];
type FirRecord = (typeof DEMO_FIR)[number] & {
  pericoloso?: boolean;
  classiPericolo?: string[];
  adrClasse?: string;
  adrNumeroONU?: string;
  adrGruppoImballaggio?: string;
  pesoLordoKg?: number;
  pesoNettoKg?: number;
  numeroColli?: number;
  tipoImballaggio?: string;
  vehicle?: { targa: string; modello: string; adr?: boolean; atp?: boolean };
  driver?: { nome: string; cognome: string; patente: string };
  percorso?: { da: string; a: string; kmStimati: number };
  numeroProgressivo?: string;
  dataEmissione?: string;
  numeroRENTRI?: string;
  qrCodePayload?: string;
  vidimatoAt?: string;
  firmaProduttoreAt?: string | null;
  firmaTrasportatoreAt?: string | null;
  firmaDestinatarioAt?: string | null;
  motivoAnnullamento?: string;
  motivoRespingimento?: string;
  chiusuraAt?: string;
};

// Mutable session-scoped copy of the FIR list. Spread copy so the
// imported array is never mutated (which would break HMR + the demo's
// data fixture intent).
const sessionFIR: FirRecord[] = DEMO_FIR.map((f) => ({ ...f } as FirRecord));

// Activity feed similarly mutable so creations / vidima show up.
const sessionActivity = DEMO_ACTIVITY.map((a) => ({ ...a }));

function jsonResponse(body: AnyJSON, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

function notFound(): Response {
  return jsonResponse({ code: 'not_found', detail: 'demo: resource not found' }, 404);
}

function pathOf(input: RequestInfo | URL): string {
  if (typeof input === 'string') return new URL(input, window.location.origin).pathname;
  if (input instanceof URL) return input.pathname;
  if (input instanceof Request) return new URL(input.url).pathname;
  return '';
}

function methodOf(input: RequestInfo | URL, init?: RequestInit): string {
  if (init?.method) return init.method.toUpperCase();
  if (input instanceof Request) return input.method.toUpperCase();
  return 'GET';
}

async function bodyOf(input: RequestInfo | URL, init?: RequestInit): Promise<unknown> {
  if (init?.body !== undefined && init.body !== null) {
    if (typeof init.body === 'string') {
      try {
        return JSON.parse(init.body);
      } catch {
        return null;
      }
    }
  }
  if (input instanceof Request) {
    try {
      return await input.clone().json();
    } catch {
      return null;
    }
  }
  return null;
}

function nextProgressivo(): string {
  const year = new Date().getFullYear();
  const numbers = sessionFIR
    .map((f) => f.numeroProgressivo ?? '')
    .map((n) => Number.parseInt(n.split('/')[1] ?? '0', 10))
    .filter((n) => Number.isFinite(n));
  const next = (numbers.length ? Math.max(...numbers) : 0) + 1;
  return `${year}/${String(next).padStart(5, '0')}`;
}

function deterministicNumeroRentri(tenantId: string, firId: string): string {
  // Deterministic 16-hex from a SHA-1-ish FNV; collision-acceptable
  // for the demo. The real backend uses sha256(tenantID:firID)[:8].
  let h = 0xdeadbeef;
  const seed = `${tenantId}:${firId}`;
  for (let i = 0; i < seed.length; i++) {
    h = Math.imul(h ^ seed.charCodeAt(i), 0x01000193);
  }
  const hex = (h >>> 0).toString(16).padStart(8, '0').toUpperCase();
  return `FIR-${hex}${hex}`.slice(0, 22);
}

function pushActivity(type: string, target: string, description: string): void {
  sessionActivity.unshift({
    id: `act-live-${Date.now()}-${Math.random().toString(36).slice(2, 6)}`,
    type,
    actor: 'demo@logitrack.it',
    target,
    description,
    occurredAt: new Date().toISOString(),
  });
  if (sessionActivity.length > 30) sessionActivity.length = 30;
}

interface FirCreateBody {
  produttoreId: string;
  trasportatoreId: string;
  destinatarioId: string;
  cer: string;
  quantitaDichiarataGrammi?: number;
  pesoNettoKg?: number;
  operazioneDestino: string;
  statoFisico?: string;
  descrizioneRifiuto?: string;
  adrClass?: string;
  numeroOnu?: string;
  caratteristiche?: string[];
}

function createFir(body: FirCreateBody): FirRecord {
  const id = 'demo-fir-' + Math.random().toString(36).slice(2, 10);
  const trasp = DEMO_TRASPORTATORI.find((t) => t.id === body.trasportatoreId);
  const prod = DEMO_PRODUTTORI.find((p) => p.id === body.produttoreId);
  const dest = DEMO_DESTINATARI.find((d) => d.id === body.destinatarioId);
  const pesoNetto =
    body.pesoNettoKg ??
    (body.quantitaDichiarataGrammi ? Math.round(body.quantitaDichiarataGrammi / 1000) : 0);
  const pericoloso = body.cer.endsWith('*');
  const fir: FirRecord = {
    id,
    tenantId: 'demo-tenant',
    state: 'draft',
    numeroProgressivo: nextProgressivo(),
    dataEmissione: new Date().toISOString(),
    produttoreId: body.produttoreId,
    trasportatoreId: body.trasportatoreId,
    destinatarioId: body.destinatarioId,
    cer: body.cer,
    descrizioneRifiuto: body.descrizioneRifiuto || 'Rifiuto demo',
    statoFisico: body.statoFisico || 'solido_non_polverulento',
    pericoloso,
    classiPericolo: body.caratteristiche,
    adrClasse: pericoloso ? body.adrClass : undefined,
    adrNumeroONU: pericoloso ? body.numeroOnu : undefined,
    adrGruppoImballaggio: pericoloso ? 'III' : undefined,
    pesoLordoKg: pesoNetto + 20,
    pesoNettoKg: pesoNetto,
    numeroColli: Math.max(1, Math.round(pesoNetto / 200)),
    tipoImballaggio: pericoloso ? 'fusti_metallici_200l' : 'big_bag',
    operazioneDestino: body.operazioneDestino,
    percorso: {
      da: `${prod?.sede.comune ?? 'Origine'} (${prod?.sede.provincia ?? '—'})`,
      a: `${dest?.sede.comune ?? 'Destinazione'} (${dest?.sede.provincia ?? '—'})`,
      kmStimati: 50,
    },
    vehicle: { targa: 'GA512MX', modello: 'Iveco Stralis AS440', adr: pericoloso },
    driver: { nome: 'Marco', cognome: 'Rossi', patente: 'CE' },
    annotazioni: '',
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  } as FirRecord;
  void trasp;
  sessionFIR.unshift(fir);
  pushActivity(
    'fir.create',
    id,
    `Nuovo FIR ${fir.numeroProgressivo} creato in stato bozza (CER ${fir.cer}, peso ${pesoNetto} kg)`,
  );
  return fir;
}

function vidimaFir(id: string): FirRecord | null {
  const fir = sessionFIR.find((f) => f.id === id);
  if (!fir) return null;
  if (fir.state !== 'draft') return fir;
  fir.state = 'vidimato';
  fir.numeroRENTRI = fir.numeroRENTRI ?? deterministicNumeroRentri('demo-tenant', id);
  fir.qrCodePayload = `rentri:${fir.numeroRENTRI}:checksum:${fir.numeroRENTRI.slice(-6)}`;
  fir.vidimatoAt = new Date().toISOString();
  fir.firmaProduttoreAt = fir.vidimatoAt;
  fir.updatedAt = fir.vidimatoAt;
  pushActivity(
    'fir.vidima',
    id,
    `FIR ${fir.numeroProgressivo} vidimato (${fir.cer}${fir.pericoloso ? ' pericoloso' : ''})`,
  );
  return fir;
}

function transitionFir(id: string, to: string): FirRecord | null {
  const fir = sessionFIR.find((f) => f.id === id);
  if (!fir) return null;
  fir.state = to;
  fir.updatedAt = new Date().toISOString();
  if (to === 'consegnato_trasportatore' && !fir.firmaTrasportatoreAt) {
    fir.firmaTrasportatoreAt = fir.updatedAt;
  }
  if (to === 'consegnato_destinatario' && !fir.firmaDestinatarioAt) {
    fir.firmaDestinatarioAt = fir.updatedAt;
  }
  if (to === 'chiuso' && !fir.chiusuraAt) {
    fir.chiusuraAt = fir.updatedAt;
  }
  pushActivity('fir.transition', id, `FIR ${fir.numeroProgressivo} → ${to}`);
  return fir;
}

async function handle(method: string, path: string, input: RequestInfo | URL, init?: RequestInit): Promise<Response | null> {
  // Health
  if (path === '/api/health') {
    return jsonResponse({
      status: 'ok',
      service: 'logitrack',
      version: 'demo',
      uptime_seconds: 0,
      time: new Date().toISOString(),
      dependencies: { mongodb: 'demo', redis: 'demo' },
    });
  }

  // Auth
  if (path === '/api/v1/auth/login' && method === 'POST') {
    return jsonResponse({
      tokenType: 'Bearer',
      accessToken: 'demo-token',
      expiresIn: 900,
      issuedAt: Math.floor(Date.now() / 1000),
    });
  }

  // Shipments
  if (path === '/api/v1/shipments' && method === 'GET') {
    return jsonResponse({ items: DEMO_SHIPMENTS, limit: 100, offset: 0 });
  }
  const shipMatch = /^\/api\/v1\/shipments\/([^/]+)$/.exec(path);
  if (shipMatch && method === 'GET') {
    const s = DEMO_SHIPMENTS.find((x) => x.id === shipMatch[1]);
    return s ? jsonResponse(s as unknown as AnyJSON) : notFound();
  }
  const traceMatch = /^\/api\/v1\/shipments\/([^/]+)\/trace$/.exec(path);
  if (traceMatch && method === 'GET') {
    const id = traceMatch[1];
    const s = DEMO_SHIPMENTS.find((x) => x.id === id);
    const orgName = s?.carrier ?? 'Trasportatore demo';
    return jsonResponse({
      shipmentId: id,
      records: [
        {
          id: `${id}-cust-1`,
          sequence: 1,
          action: 'created',
          occurredAt: s?.etd ?? '2026-04-28T05:00:00Z',
          actor: { name: 'system', role: 'creator', organisation: orgName },
          prevHash: '',
          hash: '4f7b9c2e8a1f3b5d7e9c2a4f6b8d1e3f5a7c9b1d3e5f7a9c1e3b5d7f9a1c3e5',
        },
        {
          id: `${id}-cust-2`,
          sequence: 2,
          action: 'loaded',
          occurredAt: s?.etd ? new Date(new Date(s.etd).getTime() + 30 * 60_000).toISOString() : '2026-04-28T05:30:00Z',
          actor: { name: 'Marco R.', role: 'driver', organisation: orgName },
          prevHash: '4f7b9c2e8a1f3b5d7e9c2a4f6b8d1e3f5a7c9b1d3e5f7a9c1e3b5d7f9a1c3e5',
          hash: '8a1f3b5d7e9c2a4f6b8d1e3f5a7c9b1d3e5f7a9c1e3b5d7f9a1c3e54f7b9c2e',
        },
        {
          id: `${id}-cust-3`,
          sequence: 3,
          action: 'sealed',
          occurredAt: s?.etd ? new Date(new Date(s.etd).getTime() + 45 * 60_000).toISOString() : '2026-04-28T05:45:00Z',
          actor: { name: 'Marco R.', role: 'driver', organisation: orgName },
          prevHash: '8a1f3b5d7e9c2a4f6b8d1e3f5a7c9b1d3e5f7a9c1e3b5d7f9a1c3e54f7b9c2e',
          hash: 'b8d1e3f5a7c9b1d3e5f7a9c1e3b5d7f9a1c3e54f7b9c2e8a1f3b5d7e9c2a4f6',
          sealNumber: 'SIG-2026-04-001',
        },
      ],
    });
  }
  const positionMatch = /^\/api\/v1\/shipments\/([^/]+)\/position$/.exec(path);
  if (positionMatch && method === 'GET') {
    const s = DEMO_SHIPMENTS.find((x) => x.id === positionMatch[1]);
    if (!s?.currentPosition) return notFound();
    return jsonResponse({
      shipmentId: s.id,
      position: s.currentPosition,
      recordedAt: s.updatedAt,
    });
  }
  const etaMatch = /^\/api\/v1\/shipments\/([^/]+)\/eta$/.exec(path);
  if (etaMatch && method === 'GET') {
    const s = DEMO_SHIPMENTS.find((x) => x.id === etaMatch[1]);
    if (!s) return notFound();
    return jsonResponse({
      shipmentId: s.id,
      remainingMeters: 184_200,
      smoothedSpeedKph: 72.3,
      eta: s.eta,
      source: 'demo',
      geometry: '',
    });
  }

  // Routes
  if (path === '/api/v1/routes/optimize' && method === 'POST') {
    return jsonResponse({
      distanceMeters: 412_987,
      durationSeconds: 24_300,
      geometry: '',
      legs: [{ distanceMeters: 412_987, durationSeconds: 24_300 }],
      source: 'demo',
    });
  }

  // Rifiuti — anagrafiche (read-only in demo)
  if (path === '/api/v1/rifiuti/produttori' && method === 'GET') {
    return jsonResponse({ items: DEMO_PRODUTTORI, limit: 100, offset: 0 });
  }
  if (path === '/api/v1/rifiuti/trasportatori' && method === 'GET') {
    return jsonResponse({ items: DEMO_TRASPORTATORI, limit: 100, offset: 0 });
  }
  if (path === '/api/v1/rifiuti/destinatari' && method === 'GET') {
    return jsonResponse({ items: DEMO_DESTINATARI, limit: 100, offset: 0 });
  }

  // Rifiuti — FIR list / detail / create / vidima / transition
  if (path === '/api/v1/rifiuti/fir' && method === 'GET') {
    return jsonResponse({ items: sessionFIR, limit: 100, offset: 0 });
  }
  if (path === '/api/v1/rifiuti/fir' && method === 'POST') {
    const body = (await bodyOf(input, init)) as FirCreateBody | null;
    if (!body) {
      return jsonResponse({ code: 'invalid_body', detail: 'demo: body required' }, 400);
    }
    const fir = createFir(body);
    return jsonResponse(fir as unknown as AnyJSON, 201);
  }
  const firDetailMatch = /^\/api\/v1\/rifiuti\/fir\/([^/]+)$/.exec(path);
  if (firDetailMatch && method === 'GET') {
    const f = sessionFIR.find((x) => x.id === firDetailMatch[1]);
    return f ? jsonResponse(f as unknown as AnyJSON) : notFound();
  }
  const firVidimaMatch = /^\/api\/v1\/rifiuti\/fir\/([^/]+)\/vidima$/.exec(path);
  if (firVidimaMatch && method === 'POST') {
    const fir = vidimaFir(firVidimaMatch[1]);
    if (!fir) return notFound();
    return jsonResponse({
      fir,
      numero_rentri: fir.numeroRENTRI,
      vidimato_at: fir.vidimatoAt,
      qr_code_payload: fir.qrCodePayload,
    });
  }
  const firTransitionMatch = /^\/api\/v1\/rifiuti\/fir\/([^/]+)\/transition$/.exec(path);
  if (firTransitionMatch && method === 'POST') {
    const body = (await bodyOf(input, init)) as { to?: string } | null;
    const to = body?.to;
    if (!to) return jsonResponse({ code: 'invalid_body', detail: 'demo: to required' }, 400);
    const fir = transitionFir(firTransitionMatch[1], to);
    if (!fir) return notFound();
    return jsonResponse(fir as unknown as AnyJSON);
  }

  const cerMatch = /^\/api\/v1\/rifiuti\/cer\/([^/]+)$/.exec(path);
  if (cerMatch && method === 'GET') {
    const raw = decodeURIComponent(cerMatch[1]).trim();
    const normalised = raw.replace(/\s+/g, '');
    const isPericoloso = normalised.endsWith('*');
    const digits = normalised.replace(/\*$/, '');
    const valid = /^\d{6}$/.test(digits);
    if (!valid) {
      return jsonResponse({
        input: raw,
        valid: false,
        error: 'Codice CER non valido (atteso: 6 cifre, opzionale * finale per pericolosi)',
        normalised,
      });
    }
    return jsonResponse({
      input: raw,
      valid: true,
      normalised,
      pericoloso: isPericoloso,
      chapter: digits.substring(0, 2),
    });
  }

  // Dashboard summary endpoints (demo-only)
  if (path === '/api/v1/dashboard/scadenze' && method === 'GET') {
    return jsonResponse({ items: DEMO_SCADENZE });
  }
  if (path === '/api/v1/dashboard/activity' && method === 'GET') {
    return jsonResponse({ items: sessionActivity });
  }

  // Generic POST fallback for any other create — return 200 with a
  // fabricated id. Demo never persists for unmatched POST paths.
  if (method === 'POST') {
    return jsonResponse(
      {
        id: 'demo-' + Math.random().toString(36).slice(2, 10),
        demo: true,
        message: 'Demo mode: nessuna modifica salvata.',
      },
      201,
    );
  }

  return null;
}

let installed = false;
let originalFetch: typeof fetch | null = null;

export function installDemoFetchMock(): void {
  if (installed) return;
  installed = true;
  originalFetch = window.fetch.bind(window);
  window.fetch = async (input: RequestInfo | URL, init?: RequestInit): Promise<Response> => {
    const path = pathOf(input);
    if (path.startsWith('/api/')) {
      const method = methodOf(input, init);
      const handled = await handle(method, path, input, init);
      if (handled) {
        // Tiny artificial latency so the UI shows loading skeletons.
        await new Promise((r) => setTimeout(r, 80));
        return handled;
      }
    }
    return originalFetch!(input, init);
  };
}
