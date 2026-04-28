// Demo-mode fetch interceptor.
//
// In production builds this file is unreachable (the import in main.ts
// is gated on import.meta.env.VITE_DEMO_MODE === 'true').
//
// In demo builds, calling installDemoFetchMock() replaces window.fetch
// with a wrapper that:
//   - intercepts every /api/v1/* and /api/health request and returns a
//     canned JSON response built from src/demo/data.ts
//   - delegates everything else (asset URLs, third-party tile servers,
//     etc.) to the original fetch
//
// The mock is intentionally permissive: it never returns 401, never
// rate-limits, and never validates the body. It is a static demo,
// not a backend simulator.

import {
  DEMO_DESTINATARI,
  DEMO_FIR,
  DEMO_PRODUTTORI,
  DEMO_SHIPMENTS,
  DEMO_TRASPORTATORI,
} from './data';

type AnyJSON = Record<string, unknown> | unknown[];

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

function handle(method: string, path: string): Response | null {
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
    return jsonResponse({
      shipmentId: id,
      records: [
        {
          id: `${id}-cust-1`,
          sequence: 1,
          action: 'created',
          occurredAt: '2026-04-28T05:00:00Z',
          actor: { name: 'system', role: 'creator', organisation: 'Autotrasporti Veneti S.r.l.' },
          prevHash: '',
          hash: '4f7b9c2e8a1f3b5d7e9c2a4f6b8d1e3f5a7c9b1d3e5f7a9c1e3b5d7f9a1c3e5',
        },
        {
          id: `${id}-cust-2`,
          sequence: 2,
          action: 'loaded',
          occurredAt: '2026-04-28T05:30:00Z',
          actor: { name: 'Marco R.', role: 'driver', organisation: 'Autotrasporti Veneti S.r.l.' },
          prevHash: '4f7b9c2e8a1f3b5d7e9c2a4f6b8d1e3f5a7c9b1d3e5f7a9c1e3b5d7f9a1c3e5',
          hash: '8a1f3b5d7e9c2a4f6b8d1e3f5a7c9b1d3e5f7a9c1e3b5d7f9a1c3e54f7b9c2e',
        },
        {
          id: `${id}-cust-3`,
          sequence: 3,
          action: 'sealed',
          occurredAt: '2026-04-28T05:45:00Z',
          actor: { name: 'Marco R.', role: 'driver', organisation: 'Autotrasporti Veneti S.r.l.' },
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

  // Rifiuti
  if (path === '/api/v1/rifiuti/produttori' && method === 'GET') {
    return jsonResponse({ items: DEMO_PRODUTTORI, limit: 100, offset: 0 });
  }
  if (path === '/api/v1/rifiuti/trasportatori' && method === 'GET') {
    return jsonResponse({ items: DEMO_TRASPORTATORI, limit: 100, offset: 0 });
  }
  if (path === '/api/v1/rifiuti/destinatari' && method === 'GET') {
    return jsonResponse({ items: DEMO_DESTINATARI, limit: 100, offset: 0 });
  }
  if (path === '/api/v1/rifiuti/fir' && method === 'GET') {
    return jsonResponse({ items: DEMO_FIR, limit: 100, offset: 0 });
  }
  const firMatch = /^\/api\/v1\/rifiuti\/fir\/([^/]+)$/.exec(path);
  if (firMatch && method === 'GET') {
    const f = DEMO_FIR.find((x) => x.id === firMatch[1]);
    return f ? jsonResponse(f as unknown as AnyJSON) : notFound();
  }
  const cerMatch = /^\/api\/v1\/rifiuti\/cer\/([^/]+)$/.exec(path);
  if (cerMatch && method === 'GET') {
    const code = cerMatch[1];
    const valid = /^\d{6}$/.test(code);
    if (!valid) {
      return jsonResponse({
        input: code,
        valid: false,
        error: 'CER code must be 6 digits',
        normalised: code,
      });
    }
    const pericoloso = code.endsWith('*') || ['170504', '160601'].includes(code);
    return jsonResponse({
      input: code,
      valid: true,
      normalised: code,
      pericoloso,
      chapter: code.substring(0, 2),
    });
  }

  // Generic POST fallback for create/transition/vidima — return 200
  // with a fabricated id. Demo never persists.
  if (method === 'POST') {
    return jsonResponse({
      id: 'demo-' + Math.random().toString(36).slice(2, 10),
      demo: true,
      message: 'Demo mode: nessuna modifica salvata.',
    }, 201);
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
      const handled = handle(method, path);
      if (handled) {
        // Tiny artificial latency so the UI shows loading skeletons.
        await new Promise((r) => setTimeout(r, 80));
        return handled;
      }
    }
    return originalFetch!(input, init);
  };
}
