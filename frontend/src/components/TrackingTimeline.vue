<template>
  <ol v-if="events.length" class="relative border-l border-slate-200 pl-5 space-y-4">
    <li v-for="evt in events" :key="evt.id" class="ml-2">
      <span
        class="absolute -left-[7px] mt-1 h-3 w-3 rounded-full"
        :class="dotClass(evt.type)"
        aria-hidden="true"
      />
      <p class="text-xs uppercase tracking-wide text-slate-500">{{ formatType(evt.type) }}</p>
      <p class="text-sm text-slate-800 font-medium">{{ evt.description }}</p>
      <p class="text-xs text-slate-500 mt-0.5">{{ formatDate(evt.occurredAt) }}</p>
    </li>
  </ol>
  <p v-else class="text-sm text-slate-500" role="status" aria-live="polite">
    {{ wsConnected ? 'In ascolto degli eventi in tempo reale...' : 'Connessione al feed eventi non disponibile.' }}
  </p>
</template>

<script setup lang="ts">
// Live tracking-event timeline.
//
// Subscribes to the same /api/v1/stream/tracking WebSocket the map
// uses; every TrackingEvent for the visible shipmentId is appended
// to the timeline. There is no initial backfill yet — the events
// list starts empty when the user opens the detail view and grows
// as the simulator (or real telematics) emits events. A future
// commit can backfill the last N events from a new
// /api/v1/shipments/:id/events endpoint when that lands.
//
// Hardening applied during the 2026-04-27 audit follow-up:
//  - The previous seed() with three hardcoded fake events was
//    removed. Showing fictional events to a real operator is a
//    correctness defect, regardless of how clean the screenshot
//    looks.
//  - Token sourcing routes through tokenStore (not localStorage).
//  - Token is passed via Sec-WebSocket-Protocol only — no
//    ?access_token= query string.
//  - Reconnect uses bounded exponential backoff with jitter,
//    matching the map widget's policy.

import { onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { getAccessToken } from '@/lib/tokenStore';

interface TimelineEvent {
  id: string;
  type: string;
  description: string;
  occurredAt: string;
}

interface IncomingEvent {
  id?: string;
  shipmentId?: string;
  type?: string;
  source?: string;
  occurredAt?: string;
  recordedAt?: string;
  position?: { coordinates?: number[] };
}

const props = defineProps<{ shipmentId: string }>();

const events = ref<TimelineEvent[]>([]);
const wsConnected = ref(false);

let socket: WebSocket | null = null;
let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
let reconnectAttempts = 0;
const reconnectMaxMs = 30_000;

function dotClass(type: string): string {
  switch (type) {
    case 'delivery':
    case 'arrival':
      return 'bg-emerald-500';
    case 'delay_detected':
      return 'bg-amber-500';
    case 'customs_hold':
    case 'customs_cleared':
      return 'bg-indigo-500';
    case 'seal_broken':
    case 'temperature_alarm':
      return 'bg-red-500';
    default:
      return 'bg-logitrack-blue';
  }
}

function formatType(type: string): string {
  const map: Record<string, string> = {
    position_update: 'Aggiornamento posizione',
    departure: 'Partenza',
    arrival: 'Arrivo',
    geofence_enter: 'Ingresso area',
    geofence_exit: 'Uscita area',
    delay_detected: 'Ritardo rilevato',
    customs_hold: 'Fermo dogana',
    customs_cleared: 'Svincolo dogana',
    document_attached: 'Documento allegato',
    driver_assigned: 'Autista assegnato',
    vehicle_assigned: 'Mezzo assegnato',
    seal_broken: 'Sigillo violato',
    temperature_alarm: 'Allarme temperatura',
    handover: 'Passaggio consegne',
    delivery: 'Consegna',
  };
  return map[type] ?? type;
}

function formatDate(iso: string): string {
  if (!iso) return '-';
  try {
    return new Intl.DateTimeFormat('it-IT', { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(iso));
  } catch {
    return iso;
  }
}

function describe(evt: IncomingEvent): string {
  // Best-effort human description from the available fields. The
  // backend does not currently emit a `description` so we synthesise
  // one from `type` plus position when present.
  if (evt.type === 'position_update' && evt.position?.coordinates?.length === 2) {
    const [lon, lat] = evt.position.coordinates;
    return `Posizione: ${lat.toFixed(4)}, ${lon.toFixed(4)}`;
  }
  return formatType(evt.type ?? '');
}

function scheduleReconnect() {
  const base = Math.min(reconnectMaxMs, 1000 * 2 ** reconnectAttempts);
  const jitter = base * 0.2 * (Math.random() * 2 - 1);
  const delay = Math.max(500, base + jitter);
  reconnectAttempts++;
  reconnectTimer = setTimeout(connect, delay);
}

function connect() {
  const tok = getAccessToken();
  if (!tok) return;
  const scheme = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const url = `${scheme}//${window.location.host}/api/v1/stream/tracking`;
  try {
    socket = new WebSocket(url, ['logitrack.jwt.v1', tok]);
  } catch {
    scheduleReconnect();
    return;
  }
  socket.onopen = () => {
    reconnectAttempts = 0;
    wsConnected.value = true;
    socket?.send(JSON.stringify({ op: 'subscribe', shipmentId: props.shipmentId }));
  };
  socket.onmessage = (ev) => {
    try {
      const msg = JSON.parse(ev.data as string) as IncomingEvent | { type: string };
      const m = msg as IncomingEvent;
      if (!m.type || !m.shipmentId || m.shipmentId !== props.shipmentId) return;
      const occurredAt = m.occurredAt ?? m.recordedAt ?? new Date().toISOString();
      events.value = [
        {
          id: m.id ?? `${m.type}-${occurredAt}`,
          type: m.type,
          description: describe(m),
          occurredAt,
        },
        ...events.value,
      ].slice(0, 50);
    } catch {
      /* ignore malformed payload */
    }
  };
  socket.onclose = (ev) => {
    wsConnected.value = false;
    if (ev.code >= 4000 && ev.code < 5000) return;
    scheduleReconnect();
  };
  socket.onerror = () => {
    wsConnected.value = false;
  };
}

function reset() {
  events.value = [];
  reconnectAttempts = 0;
  socket?.close();
  socket = null;
  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }
}

onMounted(connect);
onBeforeUnmount(() => {
  reset();
});
watch(
  () => props.shipmentId,
  () => {
    reset();
    connect();
  },
);
</script>
