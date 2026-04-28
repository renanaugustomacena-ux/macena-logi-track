<template>
  <div class="lt-kpi" :class="severityClass">
    <p class="text-xs uppercase tracking-wider text-slate-500">{{ label }}</p>
    <p class="mt-1 flex items-baseline gap-2">
      <span class="text-3xl font-semibold tabular text-slate-900">{{ formattedValue }}</span>
      <span v-if="unit" class="text-sm text-slate-500">{{ unit }}</span>
    </p>
    <p v-if="sublabel" class="mt-1 text-xs text-slate-500">{{ sublabel }}</p>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';

type Severity = 'ok' | 'warn' | 'bad' | 'neutral';

const props = defineProps<{
  label: string;
  value: number | string;
  unit?: string;
  sublabel?: string;
  severity?: Severity;
}>();

const severityClass = computed(() => {
  switch (props.severity) {
    case 'ok':
      return 'is-ok';
    case 'warn':
      return 'is-warn';
    case 'bad':
      return 'is-bad';
    default:
      return '';
  }
});

const formattedValue = computed(() => {
  if (typeof props.value === 'number') {
    return props.value.toLocaleString('it-IT');
  }
  return props.value;
});
</script>
