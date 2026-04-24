<template>
  <section class="grid grid-cols-1 md:grid-cols-3 gap-4" aria-label="Indicatori chiave">
    <div class="bg-white rounded-lg shadow-sm border border-slate-200 p-5">
      <p class="text-xs uppercase tracking-wide text-slate-500">Spedizioni totali</p>
      <p class="mt-2 text-3xl font-semibold text-logitrack-teal">{{ count }}</p>
      <p class="text-xs text-slate-400 mt-1">Nelle ultime 30 giornate operative</p>
    </div>
    <div class="bg-white rounded-lg shadow-sm border border-slate-200 p-5">
      <p class="text-xs uppercase tracking-wide text-slate-500">In transito ora</p>
      <p class="mt-2 text-3xl font-semibold text-logitrack-blue">{{ inTransit }}</p>
      <p class="text-xs text-slate-400 mt-1">Aggiornato in tempo reale via WebSocket</p>
    </div>
    <div class="bg-white rounded-lg shadow-sm border border-slate-200 p-5">
      <p class="text-xs uppercase tracking-wide text-slate-500">In ritardo</p>
      <p class="mt-2 text-3xl font-semibold" :class="delayed > 0 ? 'text-amber-600' : 'text-slate-400'">
        {{ delayed }}
      </p>
      <p class="text-xs text-slate-400 mt-1">Superata la finestra ETA di 30 minuti</p>
    </div>
  </section>
</template>

<script setup lang="ts">
defineProps<{ count: number; inTransit: number; delayed: number }>();
</script>
