<template>
  <form
    class="bg-white rounded-lg shadow-sm border border-slate-200 p-4 flex flex-wrap gap-4 items-end"
    @submit.prevent="$emit('refresh')"
  >
    <div>
      <label for="f-status" class="block text-xs text-slate-500 mb-1">Stato</label>
      <select
        id="f-status"
        :value="filters.status"
        class="border border-slate-200 rounded px-2 py-1 text-sm"
        @change="onChange('status', ($event.target as HTMLSelectElement).value)"
      >
        <option value="">Tutti</option>
        <option value="booked">Prenotata</option>
        <option value="picked_up">Ritirata</option>
        <option value="in_transit">In transito</option>
        <option value="delayed">In ritardo</option>
        <option value="at_customs">In dogana</option>
        <option value="delivered">Consegnata</option>
      </select>
    </div>
    <div>
      <label for="f-carrier" class="block text-xs text-slate-500 mb-1">Vettore</label>
      <input
        id="f-carrier"
        type="text"
        :value="filters.carrier"
        placeholder="Tutti"
        class="border border-slate-200 rounded px-2 py-1 text-sm"
        @input="onChange('carrier', ($event.target as HTMLInputElement).value)"
      />
    </div>
    <div>
      <label for="f-from" class="block text-xs text-slate-500 mb-1">Dal</label>
      <input
        id="f-from"
        type="date"
        :value="filters.from"
        class="border border-slate-200 rounded px-2 py-1 text-sm"
        @change="onChange('from', ($event.target as HTMLInputElement).value)"
      />
    </div>
    <div>
      <label for="f-to" class="block text-xs text-slate-500 mb-1">Al</label>
      <input
        id="f-to"
        type="date"
        :value="filters.to"
        class="border border-slate-200 rounded px-2 py-1 text-sm"
        @change="onChange('to', ($event.target as HTMLInputElement).value)"
      />
    </div>
    <button type="submit" class="bg-logitrack-blue text-white text-sm font-medium rounded px-4 py-1.5 hover:bg-blue-700">
      Aggiorna
    </button>
  </form>
</template>

<script setup lang="ts">
import type { ShipmentFilters } from '@/stores/shipment';

defineProps<{ filters: ShipmentFilters }>();
const emit = defineEmits<{
  (e: 'update', payload: { key: keyof ShipmentFilters; value: string }): void;
  (e: 'refresh'): void;
}>();

function onChange(key: keyof ShipmentFilters, value: string) {
  emit('update', { key, value });
}
</script>
