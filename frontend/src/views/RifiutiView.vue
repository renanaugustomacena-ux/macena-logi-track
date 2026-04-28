<template>
  <section class="space-y-8">
    <header>
      <h1 class="text-2xl font-semibold tracking-tight">Rifiuti speciali — FIR</h1>
      <p class="text-sm text-slate-600 mt-1">
        Modulo RENTRI-ready. Stub di vidimazione attivo finché non è
        configurato il certificato digitale del cliente.
      </p>
    </header>

    <div v-if="banner" :class="bannerClass" role="status">
      {{ banner }}
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Form -->
      <article class="bg-white rounded-lg shadow p-6 space-y-4">
        <h2 class="text-lg font-semibold">Nuovo FIR (draft)</h2>

        <div>
          <label class="block text-sm font-medium text-slate-700">Produttore</label>
          <select v-model="form.produttoreId" class="mt-1 w-full border-slate-300 rounded">
            <option value="" disabled>— scegli —</option>
            <option v-for="p in produttori" :key="p.id" :value="p.id">
              {{ p.ragioneSociale }} ({{ p.comune }})
            </option>
          </select>
        </div>

        <div>
          <label class="block text-sm font-medium text-slate-700">Trasportatore</label>
          <select v-model="form.trasportatoreId" class="mt-1 w-full border-slate-300 rounded">
            <option value="" disabled>— scegli —</option>
            <option v-for="t in trasportatori" :key="t.id" :value="t.id">
              {{ t.ragioneSociale }} — Albo {{ t.alboCategoria }}/{{ t.alboClasse }}
            </option>
          </select>
        </div>

        <div>
          <label class="block text-sm font-medium text-slate-700">Destinatario</label>
          <select v-model="form.destinatarioId" class="mt-1 w-full border-slate-300 rounded">
            <option value="" disabled>— scegli —</option>
            <option v-for="d in destinatari" :key="d.id" :value="d.id">
              {{ d.ragioneSociale }} ({{ d.comune }} {{ d.provincia }})
            </option>
          </select>
        </div>

        <div>
          <label class="block text-sm font-medium text-slate-700">Codice CER (EER)</label>
          <input
            v-model="form.cer"
            placeholder="es. 150106 oppure 130205*"
            class="mt-1 w-full border-slate-300 rounded font-mono"
            @blur="checkCer"
          />
          <p v-if="cerInfo" class="mt-1 text-xs" :class="cerInfo.valid ? 'text-emerald-700' : 'text-rose-700'">
            <template v-if="cerInfo.valid">
              CER {{ cerInfo.normalised }} — capitolo {{ cerInfo.chapter }}
              <span v-if="cerInfo.pericoloso" class="font-semibold">— pericoloso (richiede Albo cat. 5 + ADR)</span>
              <span v-else>— non pericoloso</span>
            </template>
            <template v-else>{{ cerInfo.error }}</template>
          </p>
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium text-slate-700">Quantità (kg)</label>
            <input
              v-model.number="form.quantitaKg"
              type="number"
              min="0"
              step="0.1"
              class="mt-1 w-full border-slate-300 rounded"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-slate-700">Operazione destino</label>
            <select v-model="form.operazione" class="mt-1 w-full border-slate-300 rounded">
              <option value="R3">R3 — riciclo organici</option>
              <option value="R4">R4 — riciclo metalli</option>
              <option value="R5">R5 — riciclo altri inorganici</option>
              <option value="R13">R13 — messa in riserva</option>
              <option value="D1">D1 — discarica</option>
              <option value="D9">D9 — trattamento fisico-chimico</option>
              <option value="D15">D15 — deposito preliminare</option>
            </select>
          </div>
        </div>

        <div>
          <label class="block text-sm font-medium text-slate-700">Stato fisico</label>
          <select v-model="form.statoFisico" class="mt-1 w-full border-slate-300 rounded">
            <option value="solido_non_polverulento">Solido non polverulento</option>
            <option value="solido_polverulento">Solido polverulento</option>
            <option value="liquido">Liquido</option>
            <option value="fangoso_palabile">Fangoso palabile</option>
            <option value="aeriforme">Aeriforme</option>
            <option value="vischioso_sciropposo">Vischioso/sciropposo</option>
          </select>
        </div>

        <div>
          <label class="block text-sm font-medium text-slate-700">Descrizione rifiuto</label>
          <input
            v-model="form.descrizione"
            placeholder="es. assorbenti contaminati"
            class="mt-1 w-full border-slate-300 rounded"
          />
        </div>

        <div v-if="cerInfo?.pericoloso" class="border-t pt-3 space-y-3">
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-slate-700">Classe ADR</label>
              <select v-model="form.adrClass" class="mt-1 w-full border-slate-300 rounded">
                <option value="3">3 — liquidi infiammabili</option>
                <option value="6.1">6.1 — tossiche</option>
                <option value="8">8 — corrosive</option>
                <option value="9">9 — varie</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium text-slate-700">Numero ONU</label>
              <input
                v-model="form.numeroONU"
                placeholder="UN1202"
                class="mt-1 w-full border-slate-300 rounded font-mono"
              />
            </div>
          </div>
          <div>
            <label class="block text-sm font-medium text-slate-700">Caratteristiche HP</label>
            <div class="flex flex-wrap gap-2 mt-1">
              <label v-for="hp in HP_OPTIONS" :key="hp" class="inline-flex items-center gap-1 text-xs">
                <input type="checkbox" :value="hp" v-model="form.hpClasses" />
                {{ hp }}
              </label>
            </div>
          </div>
        </div>

        <button
          type="button"
          class="w-full bg-logitrack-blue hover:bg-blue-700 text-white py-2 px-4 rounded font-semibold disabled:opacity-50"
          :disabled="!canSubmit || submitting"
          @click="onCreate"
        >
          {{ submitting ? 'Creazione…' : 'Crea FIR (draft)' }}
        </button>
      </article>

      <!-- List -->
      <article class="bg-white rounded-lg shadow p-6 space-y-3">
        <div class="flex items-center justify-between">
          <h2 class="text-lg font-semibold">FIR esistenti</h2>
          <button
            type="button"
            class="text-xs text-slate-600 hover:text-slate-900 underline"
            @click="reloadAll"
          >Aggiorna</button>
        </div>

        <div v-if="firs.length === 0" class="text-sm text-slate-500 py-6 text-center">
          Nessun FIR ancora creato per questo tenant.
        </div>

        <ul class="space-y-2">
          <li
            v-for="f in firs"
            :key="f.id"
            class="border rounded p-3 hover:bg-slate-50 cursor-pointer"
            :class="{ 'ring-2 ring-logitrack-blue': selected?.id === f.id }"
            @click="select(f)"
          >
            <div class="flex items-center justify-between">
              <span class="font-mono text-xs text-slate-600">{{ f.id?.slice(0, 8) }}…</span>
              <span class="text-xs uppercase tracking-wide" :class="stateBadge(f.state)">
                {{ f.state }}
              </span>
            </div>
            <div class="text-sm mt-1">
              <span class="font-mono">{{ f.cer }}</span>
              · {{ (f.quantitaDichiarataGrammi / 1000).toFixed(1) }} kg
              · {{ f.operazioneDestino }}
            </div>
            <div v-if="f.numeroRentri" class="text-xs text-emerald-700 mt-1">
              RENTRI: <span class="font-mono">{{ f.numeroRentri }}</span>
            </div>
          </li>
        </ul>
      </article>
    </div>

    <!-- Detail + state machine -->
    <article v-if="selected" class="bg-white rounded-lg shadow p-6 space-y-4">
      <header class="flex items-center justify-between">
        <h2 class="text-lg font-semibold">FIR <span class="font-mono">{{ selected.id }}</span></h2>
        <span class="text-xs uppercase" :class="stateBadge(selected.state)">{{ selected.state }}</span>
      </header>

      <pre class="bg-slate-50 border rounded p-3 text-xs overflow-auto max-h-64">{{ JSON.stringify(selected, null, 2) }}</pre>

      <div class="flex flex-wrap gap-2">
        <button
          v-if="selected.state === 'draft'"
          class="bg-emerald-600 text-white text-sm px-3 py-1.5 rounded hover:bg-emerald-700"
          @click="onVidima(selected.id)"
        >Vidima (RENTRI stub)</button>

        <button
          v-for="t in nextTransitions(selected.state)"
          :key="t"
          class="bg-slate-700 text-white text-sm px-3 py-1.5 rounded hover:bg-slate-900"
          @click="onTransition(selected.id, t)"
        >→ {{ t }}</button>
      </div>
    </article>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import { api } from '@/api/client';

interface Produttore {
  id: string;
  ragioneSociale: string;
  comune: string;
}
interface Trasportatore {
  id: string;
  ragioneSociale: string;
  alboCategoria: string;
  alboClasse: string;
}
interface Destinatario {
  id: string;
  ragioneSociale: string;
  comune: string;
  provincia: string;
}

interface FIR {
  id?: string;
  tenantId?: string;
  produttoreId: string;
  trasportatoreId: string;
  destinatarioId: string;
  cer: string;
  state: string;
  quantitaDichiarataGrammi: number;
  operazioneDestino: string;
  statoFisico: string;
  descrizioneRifiuto: string;
  adrClass?: string;
  numeroOnu?: string;
  caratteristiche?: string[];
  numeroRentri?: string;
  vidimatoAt?: string;
}

interface CERInfo {
  valid: boolean;
  normalised: string;
  pericoloso?: boolean;
  chapter?: string;
  error?: string;
}

interface ListResponse<T> {
  items: T[];
  limit: number;
  offset: number;
}

const HP_OPTIONS = [
  'HP1', 'HP2', 'HP3', 'HP4', 'HP5', 'HP6', 'HP7', 'HP8',
  'HP9', 'HP10', 'HP11', 'HP12', 'HP13', 'HP14', 'HP15',
] as const;

const produttori = ref<Produttore[]>([]);
const trasportatori = ref<Trasportatore[]>([]);
const destinatari = ref<Destinatario[]>([]);
const firs = ref<FIR[]>([]);
const selected = ref<FIR | null>(null);
const cerInfo = ref<CERInfo | null>(null);
const banner = ref<string>('');
const bannerKind = ref<'info' | 'error' | 'success'>('info');
const submitting = ref<boolean>(false);

const bannerClass = computed(() => {
  const base = 'rounded p-3 text-sm';
  if (bannerKind.value === 'error') return `${base} bg-rose-50 text-rose-800 border border-rose-200`;
  if (bannerKind.value === 'success') return `${base} bg-emerald-50 text-emerald-800 border border-emerald-200`;
  return `${base} bg-slate-50 text-slate-700 border border-slate-200`;
});

const form = reactive({
  produttoreId: '',
  trasportatoreId: '',
  destinatarioId: '',
  cer: '',
  quantitaKg: 100,
  operazione: 'R3',
  statoFisico: 'solido_non_polverulento',
  descrizione: '',
  adrClass: '3',
  numeroONU: '',
  hpClasses: [] as string[],
});

const canSubmit = computed(() => {
  if (!form.produttoreId || !form.trasportatoreId || !form.destinatarioId) return false;
  if (!form.cer || !cerInfo.value?.valid) return false;
  if (!form.quantitaKg || form.quantitaKg <= 0) return false;
  if (cerInfo.value.pericoloso) {
    if (!form.adrClass || !form.numeroONU || form.hpClasses.length === 0) return false;
  }
  return true;
});

function stateBadge(state: string): string {
  const map: Record<string, string> = {
    draft: 'text-slate-600',
    vidimato: 'text-emerald-700 font-semibold',
    consegnato_trasportatore: 'text-blue-700',
    in_transito: 'text-amber-700',
    consegnato_destinatario: 'text-purple-700',
    chiuso: 'text-slate-500',
    respinto: 'text-rose-700',
    annullato: 'text-rose-500 line-through',
  };
  return map[state] ?? 'text-slate-600';
}

const TRANSITIONS: Record<string, string[]> = {
  draft: ['annullato'],
  vidimato: ['consegnato_trasportatore', 'annullato'],
  consegnato_trasportatore: ['in_transito'],
  in_transito: ['consegnato_destinatario', 'respinto'],
  consegnato_destinatario: ['chiuso'],
  respinto: ['chiuso'],
};

function nextTransitions(state: string): string[] {
  return TRANSITIONS[state] ?? [];
}

function showBanner(text: string, kind: 'info' | 'error' | 'success') {
  banner.value = text;
  bannerKind.value = kind;
  if (text) {
    window.setTimeout(() => {
      if (banner.value === text) banner.value = '';
    }, 5_000);
  }
}

async function reloadAll(): Promise<void> {
  try {
    const [pp, tt, dd, ff] = await Promise.all([
      api.get<ListResponse<Produttore>>('/rifiuti/produttori'),
      api.get<ListResponse<Trasportatore>>('/rifiuti/trasportatori'),
      api.get<ListResponse<Destinatario>>('/rifiuti/destinatari'),
      api.get<ListResponse<FIR>>('/rifiuti/fir'),
    ]);
    produttori.value = pp.items;
    trasportatori.value = tt.items;
    destinatari.value = dd.items;
    firs.value = ff.items;
    if (produttori.value.length && !form.produttoreId) form.produttoreId = produttori.value[0].id;
    if (trasportatori.value.length && !form.trasportatoreId) form.trasportatoreId = trasportatori.value[0].id;
    if (destinatari.value.length && !form.destinatarioId) form.destinatarioId = destinatari.value[0].id;
  } catch (err) {
    showBanner(`Errore di caricamento: ${getErrorMessage(err)}`, 'error');
  }
}

async function checkCer(): Promise<void> {
  const code = form.cer.trim();
  if (!code) {
    cerInfo.value = null;
    return;
  }
  try {
    cerInfo.value = await api.get<CERInfo>(`/rifiuti/cer/${encodeURIComponent(code)}`);
  } catch (err) {
    cerInfo.value = { valid: false, normalised: code, error: getErrorMessage(err) };
  }
}

function select(f: FIR): void {
  selected.value = f;
}

async function onCreate(): Promise<void> {
  if (!canSubmit.value || submitting.value) return;
  submitting.value = true;
  const payload: FIR = {
    produttoreId: form.produttoreId,
    trasportatoreId: form.trasportatoreId,
    destinatarioId: form.destinatarioId,
    cer: form.cer.trim(),
    state: 'draft',
    quantitaDichiarataGrammi: Math.round(form.quantitaKg * 1000),
    operazioneDestino: form.operazione,
    statoFisico: form.statoFisico,
    descrizioneRifiuto: form.descrizione || 'Rifiuto demo',
  };
  if (cerInfo.value?.pericoloso) {
    payload.adrClass = form.adrClass;
    payload.numeroOnu = form.numeroONU;
    payload.caratteristiche = [...form.hpClasses];
  }
  try {
    const created = await api.post<FIR, FIR>('/rifiuti/fir', payload);
    showBanner(`FIR creato: ${created.id?.slice(0, 8)}…`, 'success');
    selected.value = created;
    await reloadAll();
  } catch (err) {
    showBanner(`Creazione fallita: ${getErrorMessage(err)}`, 'error');
  } finally {
    submitting.value = false;
  }
}

interface VidimaResponse { fir: FIR; numero_rentri: string; vidimato_at: string; qr_code_payload: string }

async function onVidima(id?: string): Promise<void> {
  if (!id) return;
  try {
    const resp = await api.post<VidimaResponse, Record<string, never>>(`/rifiuti/fir/${id}/vidima`, {});
    showBanner(`Vidimato: ${resp.numero_rentri}`, 'success');
    selected.value = resp.fir;
    await reloadAll();
  } catch (err) {
    showBanner(`Vidima fallita: ${getErrorMessage(err)}`, 'error');
  }
}

async function onTransition(id: string | undefined, to: string): Promise<void> {
  if (!id) return;
  try {
    const f = await api.post<FIR, { to: string }>(`/rifiuti/fir/${id}/transition`, { to });
    showBanner(`FIR → ${to}`, 'success');
    selected.value = f;
    await reloadAll();
  } catch (err) {
    showBanner(`Transizione fallita: ${getErrorMessage(err)}`, 'error');
  }
}

function getErrorMessage(err: unknown): string {
  if (err instanceof Error) return err.message;
  if (typeof err === 'object' && err !== null && 'detail' in err) {
    return String((err as { detail?: string }).detail ?? 'errore');
  }
  return 'errore';
}

onMounted(() => {
  void reloadAll();
});
</script>
