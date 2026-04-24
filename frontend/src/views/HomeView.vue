<template>
  <section class="space-y-6">
    <Dashboard :count="store.count" :in-transit="store.inTransit.length" :delayed="store.delayed.length" />

    <ShipmentFilters
      :filters="store.filters"
      @update="onFilterUpdate"
      @refresh="store.fetchAll"
    />

    <article class="bg-white rounded-lg shadow-sm border border-slate-200">
      <header class="px-6 py-4 border-b border-slate-100 flex items-center justify-between">
        <h2 class="text-lg font-semibold text-slate-800">Spedizioni recenti</h2>
        <span class="text-sm text-slate-500">{{ store.count }} risultati</span>
      </header>

      <div v-if="store.loading" class="px-6 py-10 text-center text-slate-500" role="status">
        Caricamento in corso...
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
            {{ s.consignor.city }} <span aria-hidden="true">-></span> {{ s.consignee.city }}
          </span>
          <span class="col-span-2 text-sm">
            <StatusBadge :status="s.status" />
          </span>
          <span class="col-span-2 text-xs text-slate-500 text-right">
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
import { onMounted, h, defineComponent } from 'vue';
import Dashboard from '@/components/Dashboard.vue';
import ShipmentFilters from '@/components/ShipmentFilters.vue';
import { useShipmentStore, type ShipmentFilters as Filters } from '@/stores/shipment';

const store = useShipmentStore();

onMounted(() => {
  store.fetchAll();
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
    return () =>
      h(
        'span',
        { class: `inline-block px-2 py-0.5 rounded text-xs font-medium ${palette[props.status] ?? 'bg-slate-100 text-slate-700'}` },
        props.status,
      );
  },
});

function onFilterUpdate<K extends keyof Filters>(payload: { key: K; value: Filters[K] }) {
  store.setFilter(payload.key, payload.value);
  store.fetchAll();
}

function formatDate(iso: string): string {
  if (!iso) return '-';
  try {
    return new Intl.DateTimeFormat('it-IT', { dateStyle: 'short', timeStyle: 'short' }).format(new Date(iso));
  } catch {
    return iso;
  }
}
</script>
