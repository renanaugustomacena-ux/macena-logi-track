<template>
  <section class="grid grid-cols-1 lg:grid-cols-3 gap-6">
    <article class="lg:col-span-2 bg-white rounded-lg shadow-sm border border-slate-200">
      <header class="px-6 py-4 border-b border-slate-100 flex items-start justify-between">
        <div>
          <h2 class="text-xl font-semibold text-slate-800">Spedizione {{ shipment?.reference ?? '—' }}</h2>
          <p class="text-sm text-slate-500 mt-1">
            <template v-if="shipment">
              {{ shipment.consignor.city }}, {{ shipment.consignor.country }}
              <span aria-hidden="true">-></span>
              {{ shipment.consignee.city }}, {{ shipment.consignee.country }}
            </template>
          </p>
        </div>
        <router-link to="/" class="text-sm text-logitrack-blue hover:underline">Torna all'elenco</router-link>
      </header>
      <div class="p-6">
        <ShipmentMap
          v-if="shipment"
          :shipment-id="shipment.id"
          :origin="shipment.origin"
          :destination="shipment.destination"
          :current="shipment.currentPosition ?? null"
          :polyline="shipment.routePolyline"
        />
      </div>
    </article>

    <aside class="bg-white rounded-lg shadow-sm border border-slate-200">
      <header class="px-6 py-4 border-b border-slate-100">
        <h3 class="text-lg font-semibold text-slate-800">Timeline tracciamento</h3>
      </header>
      <div class="p-6">
        <TrackingTimeline :shipment-id="id" />
      </div>
    </aside>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue';
import ShipmentMap from '@/components/ShipmentMap.vue';
import TrackingTimeline from '@/components/TrackingTimeline.vue';
import { useShipmentStore } from '@/stores/shipment';

const props = defineProps<{ id: string }>();

const store = useShipmentStore();
const shipment = computed(() => store.selected);

onMounted(() => {
  store.fetchOne(props.id);
});
</script>
