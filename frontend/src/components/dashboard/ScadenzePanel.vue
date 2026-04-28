<template>
  <article class="lt-card">
    <header class="lt-card-header flex items-center justify-between">
      <h2 class="text-lg font-semibold text-slate-800">Scadenze imminenti</h2>
      <span class="text-xs text-slate-500">{{ items.length }} elementi</span>
    </header>
    <ul v-if="items.length" class="divide-y divide-slate-100">
      <li
        v-for="s in items"
        :key="s.id"
        class="px-6 py-3 flex items-start gap-3 hover:bg-slate-50"
      >
        <span class="lt-pill shrink-0 mt-0.5" :class="pillClass(s.severity)">
          {{ pillLabel(s.severity, s.daysLeft) }}
        </span>
        <div class="min-w-0 flex-1">
          <p class="text-sm font-medium text-slate-900">{{ s.title }}</p>
          <p class="text-xs text-slate-600 truncate">{{ s.detail }}</p>
        </div>
        <div class="shrink-0 text-right text-xs">
          <p class="text-slate-700 font-medium">{{ formatDate(s.dueDate) }}</p>
          <p class="text-slate-500">{{ s.daysLeft }} giorni</p>
        </div>
      </li>
    </ul>
    <div v-else class="px-6 py-10 text-center text-sm text-slate-500">
      Nessuna scadenza imminente.
    </div>
  </article>
</template>

<script setup lang="ts">
interface Scadenza {
  id: string;
  kind: string;
  severity: 'ok' | 'warn' | 'bad';
  title: string;
  detail: string;
  dueDate: string;
  daysLeft: number;
}

defineProps<{ items: Scadenza[] }>();

function pillClass(s: Scadenza['severity']): string {
  switch (s) {
    case 'ok':
      return 'is-ok';
    case 'warn':
      return 'is-warn';
    case 'bad':
      return 'is-bad';
    default:
      return 'is-info';
  }
}

function pillLabel(severity: Scadenza['severity'], daysLeft: number): string {
  if (severity === 'bad') return `${daysLeft} gg`;
  if (severity === 'warn') return `${daysLeft} gg`;
  return `${daysLeft} gg`;
}

function formatDate(iso: string): string {
  try {
    return new Intl.DateTimeFormat('it-IT', { dateStyle: 'medium' }).format(new Date(iso));
  } catch {
    return iso;
  }
}
</script>
