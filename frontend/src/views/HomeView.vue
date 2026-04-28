<template>
  <section class="space-y-6">
    <!-- KPI strip -->
    <section class="grid grid-cols-2 lg:grid-cols-4 gap-4" aria-label="Indicatori chiave">
      <KpiCard
        label="Spedizioni totali"
        :value="store.count"
        sublabel="Ultime 30 giornate operative"
      />
      <KpiCard
        label="In transito ora"
        :value="store.inTransit.length"
        :severity="store.inTransit.length > 0 ? 'ok' : 'neutral'"
        sublabel="Aggiornato live"
      />
      <KpiCard
        label="In ritardo"
        :value="store.delayed.length"
        :severity="store.delayed.length > 0 ? 'warn' : 'neutral'"
        sublabel="ETA superato di oltre 30 min"
      />
      <KpiCard
        label="Scadenze imminenti"
        :value="scadenzeCount"
        :severity="scadenzeSeverity"
        :sublabel="scadenzeSummary"
      />
    </section>

    <!-- Two-column secondary row: scadenze + activity feed -->
    <section class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <ScadenzePanel :items="scadenze" />
      <RecentActivity :items="activity.slice(0, 8)" />
    </section>

    <!-- Filters + shipment list -->
    <ShipmentFilters
      :filters="store.filters"
      @update="onFilterUpdate"
      @refresh="store.fetchAll"
    />

    <article class="lt-card">
      <header class="lt-card-header flex items-center justify-between">
        <h2 class="text-lg font-semibold text-slate-800">Spedizioni recenti</h2>
        <span class="text-sm text-slate-500">{{ store.count }} risultati</span>
      </header>

      <div v-if="store.loading" class="px-6 py-10 text-center text-slate-500" role="status">
        Caricamento in corso…
      </div>
      <div v-else-if="store.error" class="px-6 py-10 text-center text-red-600">
        {{ store.error }}
      </div>
      <ul v-else-if="store.items.length" class="divide-y divide-slate-100">
        <li
          v-for="s in store.items"
          :key="s.id"
          class="px-6 py-4 grid grid-cols-12 gap-4 items-center hover:bg-slate-50"
        >
          <router-link
            :to="{ name: 'shipment-detail', params: { id: s.id } }"
            class="col-span-3 font-medium text-logitrack-blue hover:underline"
          >
            {{ s.reference }}
          </router-link>
          <span class="col-span-2 text-sm text-slate-600">{{ s.carrier }}</span>
          <span class="col-span-3 text-sm text-slate-600 truncate">
            {{ s.consignor.city }} <span aria-hidden="true">→</span> {{ s.consignee.city }}
          </span>
          <span class="col-span-2 text-sm">
            <StatusBadge :status="s.status" />
          </span>
          <span class="col-span-2 text-xs text-slate-500 text-right tabular">
            {{ formatDate(s.eta) }}
          </span>
        </li>
      </ul>
      <div v-else class="px-6 py-10 text-center text-slate-500">
        Nessuna spedizione trovata con i filtri applicati.
      </div>
    </article>
  </section>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onMounted, ref } from 'vue';
import KpiCard from '@/components/dashboard/KpiCard.vue';
import ScadenzePanel from '@/components/dashboard/ScadenzePanel.vue';
import RecentActivity from '@/components/dashboard/RecentActivity.vue';
import ShipmentFilters from '@/components/ShipmentFilters.vue';
import { useShipmentStore, type ShipmentFilters as Filters } from '@/stores/shipment';
import { api } from '@/api/client';

interface Scadenza {
  id: string;
  kind: string;
  severity: 'ok' | 'warn' | 'bad';
  title: string;
  detail: string;
  dueDate: string;
  daysLeft: number;
}
interface ActivityEvent {
  id: string;
  type: string;
  actor: string;
  target: string;
  description: string;
  occurredAt: string;
}
interface ListResponse<T> {
  items: T[];
}

const store = useShipmentStore();
const scadenze = ref<Scadenza[]>([]);
const activity = ref<ActivityEvent[]>([]);

onMounted(async () => {
  store.fetchAll();
  try {
    const [s, a] = await Promise.all([
      api.get<ListResponse<Scadenza>>('/dashboard/scadenze').catch(() => ({ items: [] })),
      api.get<ListResponse<ActivityEvent>>('/dashboard/activity').catch(() => ({ items: [] })),
    ]);
    scadenze.value = s.items;
    activity.value = a.items;
  } catch {
    /* fall back to empty panels */
  }
});

const scadenzeCount = computed(() => scadenze.value.length);

const scadenzeSeverity = computed<'ok' | 'warn' | 'bad' | 'neutral'>(() => {
  if (scadenze.value.some((s) => s.severity === 'bad')) return 'bad';
  if (scadenze.value.some((s) => s.severity === 'warn')) return 'warn';
  if (scadenze.value.length > 0) return 'ok';
  return 'neutral';
});

const scadenzeSummary = computed(() => {
  const bad = scadenze.value.filter((s) => s.severity === 'bad').length;
  const warn = scadenze.value.filter((s) => s.severity === 'warn').length;
  if (bad > 0) return `${bad} critiche, ${warn} in attenzione`;
  if (warn > 0) return `${warn} in attenzione`;
  return 'Tutto in regola';
});

const StatusBadge = defineComponent({
  props: { status: { type: String, required: true } },
  setup(props) {
    const palette: Record<string, string> = {
      in_transit: 'bg-emerald-100 text-emerald-800',
      delayed: 'bg-amber-100 text-amber-800',
      at_customs: 'bg-indigo-100 text-indigo-800',
      delivered: 'bg-slate-100 text-slate-700',
      cancelled: 'bg-red-100 text-red-800',
      booked: 'bg-blue-100 text-blue-800',
      picked_up: 'bg-sky-100 text-sky-800',
      draft: 'bg-slate-100 text-slate-600',
    };
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
    return () =>
      h(
        'span',
        {
          class: `inline-block px-2 py-0.5 rounded text-xs font-medium ${palette[props.status] ?? 'bg-slate-100 text-slate-700'}`,
        },
        labels[props.status] ?? props.status,
      );
  },
});

function onFilterUpdate<K extends keyof Filters>(payload: { key: K; value: Filters[K] }) {
  store.setFilter(payload.key, payload.value);
  store.fetchAll();
}

function formatDate(iso: string): string {
  if (!iso) return '—';
  try {
    return new Intl.DateTimeFormat('it-IT', { dateStyle: 'short', timeStyle: 'short' }).format(new Date(iso));
  } catch {
    return iso;
  }
}
</script>
