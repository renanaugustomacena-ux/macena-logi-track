<template>
  <ol class="relative border-l border-slate-200 pl-5 space-y-4">
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
    <li v-if="!events.length" class="text-sm text-slate-500">Nessun evento registrato.</li>
  </ol>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue';

interface TimelineEvent {
  id: string;
  type: string;
  description: string;
  occurredAt: string;
}

const props = defineProps<{ shipmentId: string }>();

const events = ref<TimelineEvent[]>([]);

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

function seed() {
  // Template data. Real implementation subscribes to
  // /api/v1/stream/tracking and appends incoming events.
  events.value = [
    { id: '1', type: 'departure', description: 'Carico presso il mittente', occurredAt: new Date(Date.now() - 7200_000).toISOString() },
    { id: '2', type: 'position_update', description: 'Transito A22 - casello Affi', occurredAt: new Date(Date.now() - 3600_000).toISOString() },
    { id: '3', type: 'geofence_enter', description: 'Ingresso Quadrante Europa', occurredAt: new Date(Date.now() - 1200_000).toISOString() },
  ];
}

onMounted(seed);
watch(() => props.shipmentId, seed);
</script>
