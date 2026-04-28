<template>
  <section class="space-y-6">
    <header>
      <h1 class="text-2xl font-semibold tracking-tight text-slate-900">Rifiuti speciali — FIR</h1>
      <p class="text-sm text-slate-600 mt-1">
        Modulo RENTRI-ready. Stub di vidimazione attivo finché non è
        configurato il certificato digitale del cliente.
      </p>
    </header>

    <!-- KPI strip -->
    <section class="grid grid-cols-2 lg:grid-cols-4 gap-4" aria-label="Indicatori rifiuti">
      <KpiCard
        label="FIR totali"
        :value="firs.length"
        sublabel="Tutti gli stati nel registro"
      />
      <KpiCard
        label="Vidimati"
        :value="counts.vidimati"
        :severity="counts.vidimati > 0 ? 'ok' : 'neutral'"
        sublabel="Pronti alla consegna"
      />
      <KpiCard
        label="In transito"
        :value="counts.transito"
        :severity="counts.transito > 0 ? 'ok' : 'neutral'"
        sublabel="Rifiuti in viaggio verso destinatario"
      />
      <KpiCard
        label="Da chiudere"
        :value="counts.daChiudere"
        :severity="counts.daChiudere > 2 ? 'warn' : 'neutral'"
        sublabel="Consegnati, in attesa di chiusura"
      />
    </section>

    <div v-if="banner" :class="bannerClass" role="status">
      {{ banner }}
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- FORM (1/3 col) -->
      <article class="lt-card lg:col-span-1">
        <header class="lt-card-header">
          <h2 class="text-base font-semibold text-slate-800">Nuovo FIR (bozza)</h2>
          <p class="text-xs text-slate-500 mt-0.5">Validazione CER + Albo + autorizzazione impianto</p>
        </header>
        <div class="lt-card-body space-y-3">
          <div>
            <label class="block text-xs font-medium text-slate-700">Produttore</label>
            <select v-model="form.produttoreId" class="mt-1 w-full text-sm border-slate-300 rounded">
              <option value="" disabled>— scegli —</option>
              <option v-for="p in produttori" :key="p.id" :value="p.id">
                {{ p.ragioneSociale }} ({{ comuneOf(p) }})
              </option>
            </select>
          </div>

          <div>
            <label class="block text-xs font-medium text-slate-700">Trasportatore</label>
            <select v-model="form.trasportatoreId" class="mt-1 w-full text-sm border-slate-300 rounded">
              <option value="" disabled>— scegli —</option>
              <option v-for="t in trasportatori" :key="t.id" :value="t.id">
                {{ t.ragioneSociale }} — Albo {{ t.alboCategoria }}/{{ t.alboClasse }}
              </option>
            </select>
          </div>

          <div>
            <label class="block text-xs font-medium text-slate-700">Destinatario</label>
            <select v-model="form.destinatarioId" class="mt-1 w-full text-sm border-slate-300 rounded">
              <option value="" disabled>— scegli —</option>
              <option v-for="d in destinatari" :key="d.id" :value="d.id">
                {{ d.ragioneSociale }} ({{ comuneOf(d) }})
              </option>
            </select>
          </div>

          <div>
            <label class="block text-xs font-medium text-slate-700">Codice CER (EER)</label>
            <input
              v-model="form.cer"
              placeholder="es. 150106 oppure 130205*"
              class="mt-1 w-full text-sm border-slate-300 rounded font-mono"
              @blur="checkCer"
            />
            <p v-if="cerInfo" class="mt-1 text-xs" :class="cerInfo.valid ? 'text-emerald-700' : 'text-rose-700'">
              <template v-if="cerInfo.valid">
                CER {{ cerInfo.normalised }} — capitolo {{ cerInfo.chapter }}
                <span v-if="cerInfo.pericoloso" class="font-semibold">— pericoloso (Albo cat. 5 + ADR)</span>
                <span v-else>— non pericoloso</span>
              </template>
              <template v-else>{{ cerInfo.error }}</template>
            </p>
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-xs font-medium text-slate-700">Quantità (kg)</label>
              <input
                v-model.number="form.quantitaKg"
                type="number"
                min="0"
                step="0.1"
                class="mt-1 w-full text-sm border-slate-300 rounded tabular"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-slate-700">Operazione</label>
              <select v-model="form.operazione" class="mt-1 w-full text-sm border-slate-300 rounded">
                <option value="R3">R3 — riciclo organici</option>
                <option value="R4">R4 — riciclo metalli</option>
                <option value="R5">R5 — riciclo altri inorganici</option>
                <option value="R13">R13 — messa in riserva</option>
                <option value="D1">D1 — discarica</option>
                <option value="D9">D9 — fisico-chimico</option>
                <option value="D10">D10 — incenerimento</option>
                <option value="D15">D15 — deposito preliminare</option>
              </select>
            </div>
          </div>

          <div>
            <label class="block text-xs font-medium text-slate-700">Stato fisico</label>
            <select v-model="form.statoFisico" class="mt-1 w-full text-sm border-slate-300 rounded">
              <option value="solido_non_polverulento">Solido non polverulento</option>
              <option value="solido_polverulento">Solido polverulento</option>
              <option value="liquido">Liquido</option>
              <option value="fangoso_palabile">Fangoso palabile</option>
              <option value="aeriforme">Aeriforme</option>
              <option value="vischioso_sciropposo">Vischioso/sciropposo</option>
            </select>
          </div>

          <div>
            <label class="block text-xs font-medium text-slate-700">Descrizione rifiuto</label>
            <input
              v-model="form.descrizione"
              placeholder="es. assorbenti contaminati"
              class="mt-1 w-full text-sm border-slate-300 rounded"
            />
          </div>

          <div v-if="cerInfo?.pericoloso" class="border-t border-slate-200 pt-3 space-y-2">
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="block text-xs font-medium text-slate-700">Classe ADR</label>
                <select v-model="form.adrClass" class="mt-1 w-full text-sm border-slate-300 rounded">
                  <option value="3">3 — liquidi infiammabili</option>
                  <option value="6.1">6.1 — tossiche</option>
                  <option value="8">8 — corrosive</option>
                  <option value="9">9 — varie</option>
                </select>
              </div>
              <div>
                <label class="block text-xs font-medium text-slate-700">Numero ONU</label>
                <input
                  v-model="form.numeroONU"
                  placeholder="UN1202"
                  class="mt-1 w-full text-sm border-slate-300 rounded font-mono"
                />
              </div>
            </div>
            <div>
              <label class="block text-xs font-medium text-slate-700">Caratteristiche HP</label>
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
            class="w-full bg-logitrack-blue hover:bg-blue-700 text-white py-2 px-4 rounded font-semibold disabled:opacity-50 text-sm"
            :disabled="!canSubmit || submitting"
            @click="onCreate"
          >
            {{ submitting ? 'Creazione…' : 'Crea FIR (bozza)' }}
          </button>
        </div>
      </article>

      <!-- LIST (2/3 col) -->
      <article class="lt-card lg:col-span-2">
        <header class="lt-card-header flex items-center justify-between">
          <h2 class="text-base font-semibold text-slate-800">Registro FIR</h2>
          <button
            type="button"
            class="text-xs text-slate-600 hover:text-slate-900 underline"
            @click="reloadAll"
          >Aggiorna</button>
        </header>

        <div v-if="firs.length === 0" class="px-6 py-10 text-center text-sm text-slate-500">
          Nessun FIR ancora creato.
        </div>

        <ul v-else class="divide-y divide-slate-100">
          <li
            v-for="f in firs"
            :key="f.id"
            class="px-6 py-3 hover:bg-slate-50 cursor-pointer"
            :class="{ 'bg-blue-50 ring-1 ring-blue-200': selected?.id === f.id }"
            @click="select(f)"
          >
            <div class="flex items-center justify-between gap-3">
              <div class="min-w-0 flex-1">
                <div class="flex items-center gap-2">
                  <span class="font-mono text-sm font-semibold text-slate-900">{{ f.numeroProgressivo ?? f.id?.slice(0, 8) }}</span>
                  <span class="lt-pill" :class="stateBadgeClass(f.state)">{{ stateLabel(f.state) }}</span>
                  <span v-if="f.pericoloso" class="lt-pill is-bad">pericoloso</span>
                </div>
                <p class="text-xs text-slate-600 mt-0.5 truncate">
                  <span class="font-mono">{{ f.cer }}</span> ·
                  {{ f.descrizioneRifiuto }}
                </p>
                <p class="text-xs text-slate-500 mt-0.5">
                  {{ produttoreLabel(f.produttoreId) }} →
                  {{ destinatarioLabel(f.destinatarioId) }}
                  <span class="ml-1">· {{ (f.pesoNettoKg || 0).toLocaleString('it-IT') }} kg</span>
                  <span class="ml-1">· {{ f.operazioneDestino }}</span>
                </p>
              </div>
              <div class="shrink-0 flex flex-col items-end gap-1">
                <button
                  type="button"
                  class="text-xs px-2 py-1 rounded border border-slate-300 hover:bg-slate-100"
                  @click.stop="openPreview(f)"
                >Anteprima FIR</button>
                <span v-if="f.numeroRENTRI" class="text-[10px] font-mono text-emerald-700">
                  {{ f.numeroRENTRI }}
                </span>
              </div>
            </div>
          </li>
        </ul>
      </article>
    </div>

    <!-- Detail card (selected FIR actions) -->
    <article v-if="selected" class="lt-card">
      <header class="lt-card-header flex items-center justify-between">
        <div>
          <h2 class="text-base font-semibold text-slate-800">
            FIR <span class="font-mono">{{ selected.numeroProgressivo ?? selected.id }}</span>
          </h2>
          <p class="text-xs text-slate-500 mt-0.5">
            Stato corrente: <span class="font-semibold">{{ stateLabel(selected.state) }}</span>
            <span v-if="selected.numeroRENTRI" class="ml-2">
              · RENTRI: <span class="font-mono">{{ selected.numeroRENTRI }}</span>
            </span>
          </p>
        </div>
        <button
          type="button"
          class="text-sm px-3 py-1.5 rounded border border-slate-300 hover:bg-slate-100"
          @click="openPreview(selected)"
        >Apri anteprima</button>
      </header>

      <div class="lt-card-body grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
        <div>
          <p class="text-xs uppercase tracking-wide text-slate-500">Caratteristiche</p>
          <p class="mt-1"><span class="font-mono">{{ selected.cer }}</span> · {{ selected.descrizioneRifiuto }}</p>
          <p class="text-slate-600 mt-1">
            {{ statoFisicoLabel(selected.statoFisico) }}
            <span v-if="selected.pericoloso"> · pericoloso ({{ (selected.classiPericolo ?? []).join(', ') }})</span>
          </p>
        </div>
        <div>
          <p class="text-xs uppercase tracking-wide text-slate-500">Quantità</p>
          <p class="mt-1 tabular">
            {{ (selected.pesoNettoKg || 0).toLocaleString('it-IT') }} kg netti ·
            {{ selected.numeroColli }} colli
          </p>
          <p class="text-slate-600 mt-1">{{ selected.percorso?.da }} → {{ selected.percorso?.a }} ({{ selected.percorso?.kmStimati }} km)</p>
        </div>
        <div>
          <p class="text-xs uppercase tracking-wide text-slate-500">Mezzo / autista</p>
          <p class="mt-1 tabular">{{ selected.vehicle?.targa ?? '—' }} · {{ selected.vehicle?.modello ?? '' }}</p>
          <p class="text-slate-600 mt-1">{{ selected.driver?.nome }} {{ selected.driver?.cognome }}</p>
        </div>
      </div>

      <div class="px-6 pb-6 flex flex-wrap gap-2">
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
        >→ {{ stateLabel(t) }}</button>
      </div>
    </article>

    <FirPreview :open="previewOpen" :fir="previewFir" @close="previewOpen = false" />
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import { api } from '@/api/client';
import KpiCard from '@/components/dashboard/KpiCard.vue';
import FirPreview from '@/components/FirPreview.vue';

interface SedeShape {
  via: string;
  comune: string;
  provincia: string;
  cap: string;
  paese?: string;
}
interface Produttore {
  id: string;
  ragioneSociale: string;
  sede: SedeShape;
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
  sede: SedeShape;
}
interface PercorsoShape {
  da: string;
  a: string;
  kmStimati: number;
}
interface Vehicle {
  targa: string;
  modello: string;
}
interface Driver {
  nome: string;
  cognome: string;
  patente: string;
}
interface FIR {
  id?: string;
  numeroProgressivo?: string;
  state: string;
  cer: string;
  descrizioneRifiuto: string;
  statoFisico: string;
  pericoloso: boolean;
  classiPericolo?: string[];
  produttoreId: string;
  trasportatoreId: string;
  destinatarioId: string;
  pesoLordoKg?: number;
  pesoNettoKg?: number;
  numeroColli?: number;
  operazioneDestino: string;
  percorso?: PercorsoShape;
  vehicle?: Vehicle;
  driver?: Driver;
  numeroRENTRI?: string;
  vidimatoAt?: string;
  dataEmissione?: string;
}

interface CERInfo {
  valid: boolean;
  normalised: string;
  pericoloso?: boolean;
  chapter?: string;
  error?: string;
}
interface ListResponse<T> { items: T[]; limit?: number; offset?: number }

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
const banner = ref('');
const bannerKind = ref<'info' | 'error' | 'success'>('info');
const submitting = ref(false);

const previewOpen = ref(false);
const previewFir = ref<FIR | null>(null);

const bannerClass = computed(() => {
  const base = 'rounded p-3 text-sm';
  if (bannerKind.value === 'error') return `${base} bg-rose-50 text-rose-800 border border-rose-200`;
  if (bannerKind.value === 'success') return `${base} bg-emerald-50 text-emerald-800 border border-emerald-200`;
  return `${base} bg-slate-50 text-slate-700 border border-slate-200`;
});

const counts = computed(() => ({
  vidimati: firs.value.filter((f) => f.state === 'vidimato').length,
  transito: firs.value.filter((f) => f.state === 'in_transito' || f.state === 'consegnato_trasportatore').length,
  daChiudere: firs.value.filter((f) => f.state === 'consegnato_destinatario').length,
}));

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

function stateLabel(state: string): string {
  return ({
    draft: 'bozza',
    vidimato: 'vidimato',
    consegnato_trasportatore: 'consegnato trasp.',
    in_transito: 'in transito',
    consegnato_destinatario: 'consegnato dest.',
    chiuso: 'chiuso',
    respinto: 'respinto',
    annullato: 'annullato',
  } as Record<string, string>)[state] ?? state;
}
function stateBadgeClass(state: string): string {
  if (state === 'draft' || state === 'annullato') return 'is-info';
  if (state === 'vidimato' || state === 'consegnato_trasportatore') return 'is-ok';
  if (state === 'in_transito' || state === 'consegnato_destinatario') return 'is-warn';
  if (state === 'respinto') return 'is-bad';
  if (state === 'chiuso') return 'is-info';
  return 'is-info';
}
function statoFisicoLabel(s: string): string {
  return ({
    solido_non_polverulento: 'Solido non polverulento',
    solido_polverulento: 'Solido polverulento',
    liquido: 'Liquido',
    fangoso_palabile: 'Fangoso palabile',
    aeriforme: 'Aeriforme',
    vischioso_sciropposo: 'Vischioso/sciropposo',
  } as Record<string, string>)[s] ?? s;
}

function comuneOf(a: { sede?: SedeShape }): string {
  return a.sede?.comune ?? '';
}
function produttoreLabel(id: string): string {
  return produttori.value.find((p) => p.id === id)?.ragioneSociale ?? '—';
}
function destinatarioLabel(id: string): string {
  return destinatari.value.find((d) => d.id === id)?.ragioneSociale ?? '—';
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
function openPreview(f: FIR): void {
  previewFir.value = f;
  previewOpen.value = true;
}

async function onCreate(): Promise<void> {
  if (!canSubmit.value || submitting.value) return;
  submitting.value = true;
  const payload = {
    produttoreId: form.produttoreId,
    trasportatoreId: form.trasportatoreId,
    destinatarioId: form.destinatarioId,
    cer: form.cer.trim(),
    state: 'draft',
    quantitaDichiarataGrammi: Math.round(form.quantitaKg * 1000),
    operazioneDestino: form.operazione,
    statoFisico: form.statoFisico,
    descrizioneRifiuto: form.descrizione || 'Rifiuto demo',
    adrClass: cerInfo.value?.pericoloso ? form.adrClass : undefined,
    numeroOnu: cerInfo.value?.pericoloso ? form.numeroONU : undefined,
    caratteristiche: cerInfo.value?.pericoloso ? [...form.hpClasses] : undefined,
  };
  try {
    await api.post<{ id: string; demo?: boolean }, typeof payload>('/rifiuti/fir', payload);
    showBanner('FIR demo accettato (le modifiche non sono persistite in modalità demo).', 'success');
    await reloadAll();
  } catch (err) {
    showBanner(`Creazione fallita: ${getErrorMessage(err)}`, 'error');
  } finally {
    submitting.value = false;
  }
}

async function onVidima(id?: string): Promise<void> {
  if (!id) return;
  try {
    await api.post<{ id: string }, Record<string, never>>(`/rifiuti/fir/${id}/vidima`, {});
    showBanner('FIR vidimato. Anteprima aggiornata con numero RENTRI.', 'success');
    await reloadAll();
  } catch (err) {
    showBanner(`Vidima fallita: ${getErrorMessage(err)}`, 'error');
  }
}
async function onTransition(id: string | undefined, to: string): Promise<void> {
  if (!id) return;
  try {
    await api.post<{ id: string }, { to: string }>(`/rifiuti/fir/${id}/transition`, { to });
    showBanner(`FIR → ${stateLabel(to)}`, 'success');
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
