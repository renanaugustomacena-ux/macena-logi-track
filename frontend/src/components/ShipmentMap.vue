<template>
  <div
    ref="mapRoot"
    class="h-96 w-full rounded border border-slate-200 bg-slate-100"
    role="application"
    aria-label="Mappa spedizione con posizione in tempo reale"
    tabindex="0"
  >
    <div v-if="!mapReady" class="h-full flex items-center justify-center text-slate-500 text-sm">
      Inizializzazione mappa...
    </div>
  </div>
  <p v-if="etaLabel" class="mt-2 text-sm text-slate-600" aria-live="polite">
    ETA stimato: <strong>{{ etaLabel }}</strong> —
    distanza residua {{ remainingKm }} km — velocità media {{ smoothedKph }} km/h.
  </p>
</template>

<script setup lang="ts">
/**
 * ShipmentMap — live Leaflet map for a single shipment.
 *
 * Draws:
 *   1. Origin marker (logitrack-teal).
 *   2. Destination marker (logitrack-blue).
 *   3. The planned polyline (decoded from OSRM Polyline6 if present,
 *      else a great-circle fallback between origin and destination).
 *   4. The live vehicle marker.
 *
 * Consumes the backend WebSocket at /api/v1/stream/tracking with
 * the Sec-WebSocket-Protocol handshake "logitrack.jwt.v1,<JWT>",
 * or an ?access_token=<JWT> fallback for servers that do not honour
 * subprotocol proxying.
 *
 * ETA is fetched from /api/v1/shipments/:id/eta on each live update.
 *
 * The component is accessible: the container is focusable, reports
 * `role="application"` and an Italian aria-label; ETA updates are
 * announced via `aria-live="polite"`.
 */

import { onMounted, onBeforeUnmount, ref, watch, computed } from 'vue';
import type { GeoPoint } from '@/stores/shipment';
import { getAccessToken } from '@/lib/tokenStore';
import { createTrackingSocket, type WSLike } from '@/lib/createWS';

interface ShipmentEvent {
  shipmentId: string;
  type: string;
  position?: GeoPoint;
  occurredAt?: string;
}

interface ETAResult {
  shipmentId: string;
  remainingMeters: number;
  smoothedSpeedKph: number;
  eta: string;
  source?: string;
  geometry?: string;
}

const props = defineProps<{
  shipmentId: string;
  origin: GeoPoint;
  destination: GeoPoint;
  current: GeoPoint | null;
  polyline?: string;
  tileUrl?: string;
  wsPath?: string;
}>();

const mapRoot = ref<HTMLDivElement | null>(null);
const mapReady = ref(false);
const currentPos = ref<GeoPoint | null>(props.current ?? null);
const etaResult = ref<ETAResult | null>(null);

let mapInstance: unknown = null;
let liveMarker: unknown = null;
let plannedLayer: unknown = null;
let L: typeof import('leaflet') | null = null;
let socket: WSLike | null = null;
let etaTimer: ReturnType<typeof setInterval> | null = null;

const etaLabel = computed(() => {
  if (!etaResult.value?.eta) return '';
  try {
    return new Intl.DateTimeFormat('it-IT', { dateStyle: 'short', timeStyle: 'short' }).format(new Date(etaResult.value.eta));
  } catch {
    return etaResult.value.eta;
  }
});
const remainingKm = computed(() =>
  etaResult.value ? Math.round(etaResult.value.remainingMeters / 100) / 10 : 0,
);
const smoothedKph = computed(() =>
  etaResult.value ? Math.round(etaResult.value.smoothedSpeedKph) : 0,
);

// Token sourcing goes through tokenStore exclusively; no direct
// localStorage / sessionStorage access here. See lib/tokenStore.ts
// for the storage contract.
function token(): string | null {
  return getAccessToken();
}

async function initMap() {
  if (!mapRoot.value) return;
  L = await import('leaflet');
  await import('leaflet/dist/leaflet.css');
  const url = props.tileUrl ?? 'https://tile.openstreetmap.org/{z}/{x}/{y}.png';
  const map = L.map(mapRoot.value);
  L.tileLayer(url, {
    attribution: '&copy; OpenStreetMap contributors',
    maxZoom: 19,
  }).addTo(map);
  const originLL: [number, number] = [props.origin.coordinates[1], props.origin.coordinates[0]];
  const destLL: [number, number] = [props.destination.coordinates[1], props.destination.coordinates[0]];
  L.marker(originLL, { title: 'Origine' }).addTo(map).bindPopup('Origine');
  L.marker(destLL, { title: 'Destinazione' }).addTo(map).bindPopup('Destinazione');

  drawPlanned(map, originLL, destLL, props.polyline);
  if (currentPos.value) {
    placeLiveMarker(map, currentPos.value);
  }
  map.fitBounds([originLL, destLL], { padding: [30, 30] });
  mapInstance = map;
  mapReady.value = true;
}

function drawPlanned(
  map: import('leaflet').Map,
  origin: [number, number],
  dest: [number, number],
  polyline?: string,
) {
  if (!L) return;
  const points = polyline ? decodePolyline6(polyline) : [origin, dest];
  if (plannedLayer) {
    (plannedLayer as import('leaflet').Polyline).remove();
  }
  plannedLayer = L.polyline(points, {
    color: '#1d4ed8',
    weight: 4,
    opacity: 0.6,
    dashArray: polyline ? undefined : '6 6',
  }).addTo(map);
}

function placeLiveMarker(map: import('leaflet').Map, pos: GeoPoint) {
  if (!L) return;
  const ll: [number, number] = [pos.coordinates[1], pos.coordinates[0]];
  if (liveMarker) {
    (liveMarker as import('leaflet').CircleMarker).setLatLng(ll);
    return;
  }
  liveMarker = L.circleMarker(ll, {
    radius: 9,
    color: '#f59e0b',
    fillColor: '#f59e0b',
    fillOpacity: 1,
  }).addTo(map);
  (liveMarker as import('leaflet').CircleMarker).bindPopup('Posizione attuale');
}

// Reconnect state — bounded exponential backoff with jitter so a
// backend outage does not produce a thundering herd. Cleared on
// onBeforeUnmount via reconnectTimer.
let reconnectAttempts = 0;
const reconnectMaxMs = 30_000;
let reconnectTimer: ReturnType<typeof setTimeout> | null = null;

function scheduleReconnect() {
  // 1s, 2s, 4s, 8s, 16s, 30s (capped). Jitter ±20% to break the herd.
  const base = Math.min(reconnectMaxMs, 1000 * 2 ** reconnectAttempts);
  const jitter = base * 0.2 * (Math.random() * 2 - 1);
  const delay = Math.max(500, base + jitter);
  reconnectAttempts++;
  reconnectTimer = setTimeout(connectWS, delay);
}

function connectWS() {
  const tok = token();
  if (!tok) {
    // Without a token we cannot subscribe. The router auth guard
    // normally prevents this path; if we arrive here it means the
    // token expired between map mount and WS open. Bail; the user
    // will be redirected to /login on the next API call.
    return;
  }
  const wsPath = props.wsPath ?? '/api/v1/stream/tracking';
  const scheme = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const url = `${scheme}//${window.location.host}${wsPath}`;
  // Token is passed via Sec-WebSocket-Protocol exclusively. The
  // legacy ?access_token= query-string fallback was dropped during
  // the 2026-04-27 audit follow-up — query strings hit access logs,
  // browser history and reverse-proxy caches.
  try {
    socket = createTrackingSocket(url, ['logitrack.jwt.v1', tok]);
  } catch {
    // SecurityError from some browsers blocking mixed schemes — fail
    // soft and try again with backoff.
    scheduleReconnect();
    return;
  }
  socket.onopen = () => {
    reconnectAttempts = 0;
    socket?.send(JSON.stringify({ op: 'subscribe', shipmentId: props.shipmentId }));
  };
  socket.onmessage = (ev: { data: string }) => {
    try {
      const msg = JSON.parse(ev.data) as ShipmentEvent | { type: string };
      const m = msg as ShipmentEvent;
      if (m.shipmentId && m.shipmentId !== props.shipmentId) return;
      if (m.position) {
        currentPos.value = m.position;
        if (mapInstance) placeLiveMarker(mapInstance as import('leaflet').Map, m.position);
        void refreshETA();
      }
    } catch {
      /* ignore malformed payload */
    }
  };
  socket.onclose = (ev: { code: number }) => {
    // 4xxx close codes mean auth/protocol failure — do not retry.
    if (ev.code >= 4000 && ev.code < 5000) return;
    scheduleReconnect();
  };
}

async function refreshETA() {
  const tok = token();
  try {
    const res = await fetch(`/api/v1/shipments/${props.shipmentId}/eta`, {
      headers: tok ? { Authorization: `Bearer ${tok}` } : {},
    });
    if (!res.ok) return;
    etaResult.value = (await res.json()) as ETAResult;
  } catch {
    /* network error — leave previous value */
  }
}

onMounted(() => {
  initMap()
    .then(() => {
      connectWS();
      void refreshETA();
      etaTimer = setInterval(refreshETA, 30_000);
    })
    .catch(() => {
      /* fail-soft: keep the loading placeholder */
    });
});

onBeforeUnmount(() => {
  if (etaTimer) clearInterval(etaTimer);
  if (reconnectTimer) clearTimeout(reconnectTimer);
  socket?.close();
  (mapInstance as { remove?: () => void } | null)?.remove?.();
});

watch(
  () => props.current,
  (p) => {
    if (p && mapInstance) placeLiveMarker(mapInstance as import('leaflet').Map, p);
  },
);

/**
 * Decode an OSRM Polyline6 string. The algorithm is the Google
 * Encoded Polyline Algorithm with a precision of 1e-6 (instead of
 * the default 1e-5).
 *
 * Source: https://developers.google.com/maps/documentation/utilities/polylinealgorithm
 */
function decodePolyline6(str: string): [number, number][] {
  const precision = 1e-6;
  let index = 0;
  let lat = 0;
  let lng = 0;
  const coordinates: [number, number][] = [];
  while (index < str.length) {
    let b: number;
    let shift = 0;
    let result = 0;
    do {
      b = str.charCodeAt(index++) - 63;
      result |= (b & 0x1f) << shift;
      shift += 5;
    } while (b >= 0x20);
    const dlat = (result & 1) !== 0 ? ~(result >> 1) : result >> 1;
    lat += dlat;

    shift = 0;
    result = 0;
    do {
      b = str.charCodeAt(index++) - 63;
      result |= (b & 0x1f) << shift;
      shift += 5;
    } while (b >= 0x20);
    const dlng = (result & 1) !== 0 ? ~(result >> 1) : result >> 1;
    lng += dlng;

    coordinates.push([lat * precision, lng * precision]);
  }
  return coordinates;
}
</script>
