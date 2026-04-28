<template>
  <article class="lt-card">
    <header class="lt-card-header flex items-center justify-between">
      <h2 class="text-lg font-semibold text-slate-800">Attività recente</h2>
      <span class="text-xs text-slate-500">ultime {{ items.length }}</span>
    </header>
    <ol v-if="items.length" class="divide-y divide-slate-100">
      <li v-for="evt in items" :key="evt.id" class="px-6 py-3 flex items-start gap-3 hover:bg-slate-50">
        <span
          class="shrink-0 mt-1 h-2 w-2 rounded-full"
          :class="dotClass(evt.type)"
          aria-hidden="true"
        ></span>
        <div class="min-w-0 flex-1">
          <p class="text-sm text-slate-900 truncate">{{ evt.description }}</p>
          <p class="text-xs text-slate-500">
            <span class="font-mono">{{ evt.actor }}</span>
            · {{ relativeTime(evt.occurredAt) }}
          </p>
        </div>
        <span class="shrink-0 text-[10px] uppercase tracking-wide text-slate-500">
          {{ evt.type.replace(/_/g, '.') }}
        </span>
      </li>
    </ol>
    <div v-else class="px-6 py-10 text-center text-sm text-slate-500">
      Nessuna attività recente.
    </div>
  </article>
</template>

<script setup lang="ts">
interface ActivityEvent {
  id: string;
  type: string;
  actor: string;
  target: string;
  description: string;
  occurredAt: string;
}

defineProps<{ items: ActivityEvent[] }>();

function dotClass(type: string): string {
  if (type.startsWith('fir.vidima')) return 'bg-emerald-500';
  if (type.startsWith('fir.respinto')) return 'bg-red-500';
  if (type.startsWith('fir.chiusura')) return 'bg-slate-500';
  if (type.startsWith('fir.transition')) return 'bg-blue-500';
  if (type.startsWith('fir.create')) return 'bg-indigo-500';
  if (type.startsWith('shipment.delayed')) return 'bg-amber-500';
  if (type.startsWith('shipment.create')) return 'bg-sky-500';
  if (type.startsWith('auth')) return 'bg-slate-400';
  return 'bg-slate-300';
}

function relativeTime(iso: string): string {
  const now = Date.now();
  const then = new Date(iso).getTime();
  const diffSec = Math.floor((now - then) / 1000);
  if (diffSec < 60) return 'ora';
  const diffMin = Math.floor(diffSec / 60);
  if (diffMin < 60) return `${diffMin} min fa`;
  const diffHour = Math.floor(diffMin / 60);
  if (diffHour < 24) return `${diffHour} h fa`;
  const diffDay = Math.floor(diffHour / 24);
  if (diffDay < 7) return `${diffDay} g fa`;
  try {
    return new Intl.DateTimeFormat('it-IT', { dateStyle: 'short' }).format(new Date(iso));
  } catch {
    return iso;
  }
}
</script>
