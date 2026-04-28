<template>
  <section v-if="shipment" class="space-y-6">
    <!-- Page header — back link + title + status pill -->
    <header class="flex items-start justify-between gap-4">
      <div>
        <router-link to="/" class="text-xs text-logitrack-blue hover:underline">
          ← Torna all'elenco spedizioni
        </router-link>
        <h1 class="mt-1 text-2xl font-semibold tracking-tight text-slate-900">
          Spedizione <span class="font-mono">{{ shipment.reference }}</span>
        </h1>
        <p class="text-sm text-slate-600 mt-0.5">
          {{ shipment.consignor.name }} ({{ shipment.consignor.city }})
          <span class="mx-1" aria-hidden="true">→</span>
          {{ shipment.consignee.name }} ({{ shipment.consignee.city }})
        </p>
      </div>
      <div class="flex flex-col items-end gap-2">
        <span class="lt-pill" :class="statusPillClass">{{ statusLabel }}</span>
        <button
          type="button"
          class="text-sm px-3 py-1.5 rounded border border-slate-300 hover:bg-slate-100"
          @click="cmrPreviewOpen = true"
        >Anteprima CMR</button>
      </div>
    </header>

    <!-- KPI strip — shipment-specific -->
    <section class="grid grid-cols-2 lg:grid-cols-4 gap-4" aria-label="Indicatori spedizione">
      <KpiCard
        label="Distanza totale"
        :value="distanceKm"
        unit="km"
        sublabel="Sul percorso pianificato"
      />
      <KpiCard
        label="Velocità media"
        :value="averageKph"
        unit="km/h"
        sublabel="Smoothing 20 campioni"
      />
      <KpiCard
        label="ETA"
        :value="formattedEta"
        :severity="etaSeverity"
        :sublabel="etaSublabel"
      />
      <KpiCard
        label="Catena custodia"
        :value="custody.length"
        :severity="custody.length > 0 ? 'ok' : 'neutral'"
        sublabel="Eventi firmati SHA-256"
      />
    </section>

    <!-- 2-col map + timeline -->
    <section class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <article class="lt-card lg:col-span-2">
        <header class="lt-card-header flex items-center justify-between">
          <h2 class="text-base font-semibold text-slate-800">Mappa live</h2>
          <span class="text-xs text-slate-500">Aggiornamento WebSocket</span>
        </header>
        <div class="p-4">
          <ShipmentMap
            :shipment-id="shipment.id"
            :origin="shipment.origin"
            :destination="shipment.destination"
            :current="shipment.currentPosition ?? null"
            :polyline="shipment.routePolyline"
          />
        </div>
      </article>

      <article class="lt-card">
        <header class="lt-card-header flex items-center justify-between">
          <h2 class="text-base font-semibold text-slate-800">Eventi</h2>
        </header>
        <div class="p-4">
          <TrackingTimeline :shipment-id="id" />
        </div>
      </article>
    </section>

    <!-- Custody chain card -->
    <article class="lt-card">
      <header class="lt-card-header flex items-center justify-between">
        <h2 class="text-base font-semibold text-slate-800">Catena di custodia</h2>
        <span class="text-xs text-slate-500">{{ custody.length }} record append-only</span>
      </header>
      <ol v-if="custody.length" class="divide-y divide-slate-100">
        <li
          v-for="rec in custody"
          :key="rec.id"
          class="px-6 py-3 grid grid-cols-12 gap-3 items-start text-sm"
        >
          <span class="col-span-1 font-mono text-xs text-slate-500">#{{ rec.sequence }}</span>
          <span class="col-span-2 lt-pill is-info self-start">{{ actionLabel(rec.action) }}</span>
          <div class="col-span-4">
            <p class="font-medium text-slate-900">{{ rec.actor.name }}</p>
            <p class="text-xs text-slate-500">
              {{ rec.actor.role }} — {{ rec.actor.organisation }}
            </p>
          </div>
          <span class="col-span-2 text-xs text-slate-600 tabular">{{ formatDate(rec.occurredAt) }}</span>
          <span class="col-span-3 font-mono text-[10px] text-slate-500 truncate" :title="rec.hash">
            {{ rec.hash.slice(0, 16) }}…
          </span>
        </li>
      </ol>
      <div v-else class="px-6 py-8 text-sm text-slate-500 text-center">
        Catena di custodia non ancora popolata.
      </div>
    </article>

    <!-- Vehicle + driver block -->
    <section class="grid grid-cols-1 md:grid-cols-2 gap-6">
      <article class="lt-card">
        <header class="lt-card-header">
          <h2 class="text-base font-semibold text-slate-800">Mezzo di trasporto</h2>
        </header>
        <div class="lt-card-body grid grid-cols-2 gap-4 text-sm">
          <div>
            <p class="text-xs uppercase tracking-wide text-slate-500">Targa</p>
            <p class="mt-1 font-mono">{{ shipment.vehiclePlate ?? '—' }}</p>
          </div>
          <div>
            <p class="text-xs uppercase tracking-wide text-slate-500">Modalità</p>
            <p class="mt-1">{{ modeLabel(shipment.mode) }}</p>
          </div>
          <div v-if="shipment.adrClass">
            <p class="text-xs uppercase tracking-wide text-slate-500">ADR</p>
            <p class="mt-1">Classe {{ shipment.adrClass }}</p>
          </div>
          <div v-if="shipment.atpClass">
            <p class="text-xs uppercase tracking-wide text-slate-500">ATP</p>
            <p class="mt-1">Classe {{ shipment.atpClass }}</p>
          </div>
        </div>
      </article>
      <article class="lt-card">
        <header class="lt-card-header">
          <h2 class="text-base font-semibold text-slate-800">Pianificazione</h2>
        </header>
        <div class="lt-card-body grid grid-cols-2 gap-4 text-sm">
          <div>
            <p class="text-xs uppercase tracking-wide text-slate-500">ETD</p>
            <p class="mt-1 tabular">{{ formatDate(shipment.etd) }}</p>
          </div>
          <div>
            <p class="text-xs uppercase tracking-wide text-slate-500">ETA pianificato</p>
            <p class="mt-1 tabular">{{ formatDate(shipment.eta) }}</p>
          </div>
          <div>
            <p class="text-xs uppercase tracking-wide text-slate-500">Vettore</p>
            <p class="mt-1">{{ shipment.carrier }}</p>
          </div>
          <div>
            <p class="text-xs uppercase tracking-wide text-slate-500">Ultimo aggiornamento</p>
            <p class="mt-1 tabular">{{ formatDate(shipment.updatedAt) }}</p>
          </div>
        </div>
      </article>
    </section>

    <CmrPreview :open="cmrPreviewOpen" :shipment="shipment" :custody="custody" @close="cmrPreviewOpen = false" />
  </section>

  <section v-else-if="store.loading" class="text-slate-500 text-center py-16">
    Caricamento spedizione…
  </section>
  <section v-else class="text-slate-500 text-center py-16">
    Spedizione non trovata.
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import ShipmentMap from '@/components/ShipmentMap.vue';
import TrackingTimeline from '@/components/TrackingTimeline.vue';
import KpiCard from '@/components/dashboard/KpiCard.vue';
import CmrPreview from '@/components/CmrPreview.vue';
import { useShipmentStore } from '@/stores/shipment';
import { api } from '@/api/client';

interface CustodyRecord {
  id: string;
  sequence: number;
  action: string;
  occurredAt: string;
  actor: { name: string; role: string; organisation: string };
  prevHash: string;
  hash: string;
  sealNumber?: string;
}
interface TraceResponse { shipmentId: string; records: CustodyRecord[] }
interface EtaResponse { eta?: string; remainingMeters?: number; smoothedSpeedKph?: number }

const props = defineProps<{ id: string }>();

const store = useShipmentStore();
const shipment = computed(() => store.selected);

const custody = ref<CustodyRecord[]>([]);
const eta = ref<EtaResponse | null>(null);
const cmrPreviewOpen = ref(false);

onMounted(async () => {
  await store.fetchOne(props.id);
  // Fetch custody + ETA in parallel; tolerate failures so the rest of
  // the page still renders.
  const [c, e] = await Promise.all([
    api.get<TraceResponse>(`/shipments/${props.id}/trace`).catch(() => ({ shipmentId: props.id, records: [] as CustodyRecord[] })),
    api.get<EtaResponse>(`/shipments/${props.id}/eta`).catch(() => null),
  ]);
  custody.value = c.records;
  eta.value = e;
});

const distanceKm = computed(() => {
  if (eta.value?.remainingMeters) {
    return Math.round(eta.value.remainingMeters / 100) / 10;
  }
  if (!shipment.value) return 0;
  // fallback: great-circle from origin to destination
  const a = shipment.value.origin.coordinates;
  const b = shipment.value.destination.coordinates;
  return Math.round(haversineKm(a[1], a[0], b[1], b[0]) * 10) / 10;
});

const averageKph = computed(() => {
  if (eta.value?.smoothedSpeedKph) return Math.round(eta.value.smoothedSpeedKph);
  return 70; // fallback used by the route optimiser too
});

const formattedEta = computed(() => {
  const isoEta = eta.value?.eta ?? shipment.value?.eta;
  if (!isoEta) return '—';
  try {
    return new Intl.DateTimeFormat('it-IT', { dateStyle: 'short', timeStyle: 'short' }).format(new Date(isoEta));
  } catch {
    return isoEta;
  }
});

const etaSeverity = computed<'ok' | 'warn' | 'bad' | 'neutral'>(() => {
  if (!shipment.value) return 'neutral';
  if (shipment.value.status === 'delayed') return 'bad';
  if (shipment.value.status === 'in_transit') return 'ok';
  if (shipment.value.status === 'delivered') return 'ok';
  return 'neutral';
});

const etaSublabel = computed(() => {
  if (!shipment.value) return '';
  if (shipment.value.status === 'delayed') return 'ETA superato';
  if (shipment.value.status === 'delivered') return 'Consegna completata';
  return 'Aggiornato live';
});

const statusPillClass = computed(() => {
  if (!shipment.value) return 'is-info';
  switch (shipment.value.status) {
    case 'delayed':
      return 'is-bad';
    case 'in_transit':
    case 'picked_up':
      return 'is-ok';
    case 'delivered':
    case 'cancelled':
      return 'is-info';
    case 'at_customs':
      return 'is-warn';
    default:
      return 'is-info';
  }
});

const statusLabel = computed(() => {
  const labels: Record<string, string> = {
    in_transit: 'in transito',
    delayed: 'in ritardo',
    at_customs: 'in dogana',
    delivered: 'consegnata',
    cancelled: 'annullata',
    booked: 'prenotata',
    picked_up: 'caricata',
    draft: 'bozza',
  };
  return labels[shipment.value?.status ?? ''] ?? shipment.value?.status ?? '—';
});

function actionLabel(action: string): string {
  const map: Record<string, string> = {
    created: 'Creazione',
    loaded: 'Caricamento',
    sealed: 'Sigillo',
    picked_up: 'Ritiro',
    handover: 'Passaggio',
    unsealed: 'Apertura sigillo',
    unloaded: 'Scarico',
    delivered: 'Consegna',
    inspection: 'Ispezione',
    exception: 'Eccezione',
  };
  return map[action] ?? action;
}

function modeLabel(mode: string): string {
  const map: Record<string, string> = {
    road: 'Strada',
    rail: 'Ferrovia',
    sea: 'Mare',
    air: 'Aria',
    multimodal: 'Multimodale',
  };
  return map[mode] ?? mode;
}

function formatDate(iso?: string): string {
  if (!iso) return '—';
  try {
    return new Intl.DateTimeFormat('it-IT', { dateStyle: 'short', timeStyle: 'short' }).format(new Date(iso));
  } catch {
    return iso;
  }
}

function haversineKm(lat1: number, lon1: number, lat2: number, lon2: number): number {
  const R = 6371;
  const dLat = ((lat2 - lat1) * Math.PI) / 180;
  const dLon = ((lon2 - lon1) * Math.PI) / 180;
  const a =
    Math.sin(dLat / 2) ** 2 +
    Math.cos((lat1 * Math.PI) / 180) * Math.cos((lat2 * Math.PI) / 180) * Math.sin(dLon / 2) ** 2;
  return R * 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
}
</script>
