<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="lt-print-portal"
      role="dialog"
      aria-modal="true"
      :aria-label="`Anteprima FIR ${fir?.numeroProgressivo ?? ''}`"
    >
      <!-- Backdrop (hidden in print) -->
      <div
        class="lt-print-backdrop fixed inset-0 z-40 bg-slate-900/60"
        @click="$emit('close')"
      ></div>

      <!-- Centered scrollable container -->
      <div class="fixed inset-0 z-50 flex items-start justify-center overflow-auto p-4 pointer-events-none">
        <div class="lt-print-document w-full max-w-[210mm] pointer-events-auto">
          <!-- Toolbar (hidden in print) -->
          <div class="lt-print-toolbar bg-white rounded-t-md shadow flex items-center justify-between px-4 py-2 border-b border-slate-200">
            <div class="text-sm text-slate-600">
              Anteprima FIR <span class="font-mono">{{ fir?.numeroProgressivo }}</span>
              — uscita stampa A4 verticale
            </div>
            <div class="flex gap-2">
              <button
                type="button"
                class="text-sm px-3 py-1.5 rounded border border-slate-300 hover:bg-slate-50"
                @click="$emit('close')"
              >Chiudi</button>
              <button
                type="button"
                class="text-sm px-3 py-1.5 rounded bg-logitrack-blue text-white hover:bg-blue-700"
                @click="onPrint"
              >Stampa / PDF</button>
            </div>
          </div>

          <!-- The printed page itself -->
          <article v-if="fir" class="lt-fir-page bg-white shadow p-8 text-[10pt] leading-snug text-slate-900">
        <!-- Intestazione -->
        <header class="border-b-2 border-slate-900 pb-2 mb-3">
          <div class="flex items-start justify-between">
            <div>
              <p class="text-[8pt] uppercase tracking-wider text-slate-600">Modello conforme RENTRI v1.0</p>
              <h1 class="text-lg font-bold uppercase tracking-tight mt-0.5">
                Formulario di Identificazione del Rifiuto
              </h1>
              <p class="text-[9pt] text-slate-700 mt-0.5">
                D.Lgs. 152/2006 art. 193 — D.M. MASE 04/04/2023 n. 59
              </p>
            </div>
            <div class="text-right text-[9pt]">
              <p>
                <span class="text-slate-600">N°</span>
                <span class="font-mono font-semibold text-base ml-1">{{ fir.numeroProgressivo }}</span>
              </p>
              <p class="mt-0.5">
                <span class="text-slate-600">Data emissione:</span>
                <span class="font-medium ml-1">{{ formatDate(fir.dataEmissione) }}</span>
              </p>
              <p class="mt-0.5">
                <span class="text-slate-600">Stato:</span>
                <span class="font-medium ml-1 uppercase">{{ stateLabel(fir.state) }}</span>
              </p>
            </div>
          </div>
          <p v-if="fir.numeroRENTRI" class="text-[9pt] mt-2">
            <span class="text-slate-600">Riferimento RENTRI:</span>
            <span class="font-mono ml-1">{{ fir.numeroRENTRI }}</span>
            <span v-if="fir.vidimatoAt" class="ml-3 text-slate-600">vidimato il</span>
            <span v-if="fir.vidimatoAt" class="ml-1 font-medium">{{ formatDate(fir.vidimatoAt) }}</span>
          </p>
        </header>

        <!-- Section A — Produttore / Detentore -->
        <section class="grid grid-cols-2 gap-3 mb-3">
          <div class="border border-slate-300 p-2 rounded">
            <h2 class="text-[9pt] font-semibold uppercase tracking-wide bg-slate-100 -m-2 mb-1.5 px-2 py-1 border-b border-slate-300">
              A — Produttore / Detentore
            </h2>
            <p class="font-semibold">{{ produttore?.ragioneSociale ?? '—' }}</p>
            <p>{{ produttore?.sede.via }}</p>
            <p>{{ produttore?.sede.cap }} {{ produttore?.sede.comune }} ({{ produttore?.sede.provincia }})</p>
            <p class="mt-1">CF: <span class="font-mono">{{ produttore?.codiceFiscale ?? '—' }}</span></p>
            <p>P.IVA: <span class="font-mono">{{ produttore?.partitaIva ?? '—' }}</span></p>
            <p v-if="produttore?.numeroRegistroProduttori" class="text-[9pt]">
              Reg. produttori: <span class="font-mono">{{ produttore.numeroRegistroProduttori }}</span>
            </p>
          </div>

          <!-- Section B — Destinatario -->
          <div class="border border-slate-300 p-2 rounded">
            <h2 class="text-[9pt] font-semibold uppercase tracking-wide bg-slate-100 -m-2 mb-1.5 px-2 py-1 border-b border-slate-300">
              B — Destinatario / Impianto
            </h2>
            <p class="font-semibold">{{ destinatario?.ragioneSociale ?? '—' }}</p>
            <p>{{ destinatario?.sede.via }}</p>
            <p>{{ destinatario?.sede.cap }} {{ destinatario?.sede.comune }} ({{ destinatario?.sede.provincia }})</p>
            <p class="mt-1">CF: <span class="font-mono">{{ destinatario?.codiceFiscale ?? '—' }}</span></p>
            <p>P.IVA: <span class="font-mono">{{ destinatario?.partitaIva ?? '—' }}</span></p>
            <p v-if="autorizzazioneRilevante" class="text-[9pt] mt-1">
              Aut.: {{ autorizzazioneRilevante.provvedimento }}
              — scad. {{ formatDate(autorizzazioneRilevante.scadenza) }}
            </p>
          </div>
        </section>

        <!-- Section C — Trasportatore -->
        <section class="border border-slate-300 p-2 rounded mb-3">
          <h2 class="text-[9pt] font-semibold uppercase tracking-wide bg-slate-100 -m-2 mb-1.5 px-2 py-1 border-b border-slate-300">
            C — Trasportatore
          </h2>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <p class="font-semibold">{{ trasportatore?.ragioneSociale ?? '—' }}</p>
              <p>{{ trasportatore?.sede.via }}</p>
              <p>{{ trasportatore?.sede.cap }} {{ trasportatore?.sede.comune }} ({{ trasportatore?.sede.provincia }})</p>
              <p class="mt-1">CF: <span class="font-mono">{{ trasportatore?.codiceFiscale ?? '—' }}</span></p>
              <p>P.IVA: <span class="font-mono">{{ trasportatore?.partitaIva ?? '—' }}</span></p>
            </div>
            <div>
              <p>
                Albo cat. <span class="font-semibold">{{ trasportatore?.alboCategoria }}</span>
                classe <span class="font-semibold">{{ trasportatore?.alboClasse }}</span>
              </p>
              <p>
                Iscrizione: <span class="font-mono">{{ trasportatore?.alboNumeroIscrizione }}</span>
              </p>
              <p>Scadenza Albo: {{ formatDate(trasportatore?.alboScadenza) }}</p>
              <p class="mt-1 text-[9pt]">Mezzo: <span class="font-mono">{{ fir.vehicle?.targa }}</span> — {{ fir.vehicle?.modello }}</p>
              <p class="text-[9pt]">Conducente: {{ fir.driver?.nome }} {{ fir.driver?.cognome }} (pat. {{ fir.driver?.patente }})</p>
            </div>
          </div>
        </section>

        <!-- Section D — Caratteristiche del rifiuto -->
        <section class="border border-slate-300 p-2 rounded mb-3">
          <h2 class="text-[9pt] font-semibold uppercase tracking-wide bg-slate-100 -m-2 mb-1.5 px-2 py-1 border-b border-slate-300">
            D — Caratteristiche del rifiuto
          </h2>
          <table class="w-full text-[9.5pt] border-collapse">
            <tbody>
              <tr class="border-b border-slate-200">
                <td class="py-1 pr-2 text-slate-600 w-1/4">Codice EER (CER)</td>
                <td class="py-1 font-mono font-semibold">{{ fir.cer }}</td>
              </tr>
              <tr class="border-b border-slate-200">
                <td class="py-1 pr-2 text-slate-600 align-top">Descrizione</td>
                <td class="py-1">{{ fir.descrizioneRifiuto }}</td>
              </tr>
              <tr class="border-b border-slate-200">
                <td class="py-1 pr-2 text-slate-600">Stato fisico</td>
                <td class="py-1">{{ statoFisicoLabel(fir.statoFisico) }}</td>
              </tr>
              <tr class="border-b border-slate-200">
                <td class="py-1 pr-2 text-slate-600">Pericolosità</td>
                <td class="py-1">
                  <template v-if="fir.pericoloso">
                    <span class="font-semibold text-red-700">Pericoloso</span>
                    <span v-if="fir.classiPericolo?.length" class="ml-2">
                      ({{ fir.classiPericolo.join(', ') }})
                    </span>
                  </template>
                  <template v-else>Non pericoloso</template>
                </td>
              </tr>
              <tr v-if="fir.pericoloso" class="border-b border-slate-200">
                <td class="py-1 pr-2 text-slate-600">ADR</td>
                <td class="py-1">
                  Classe <span class="font-mono">{{ fir.adrClasse }}</span>
                  · ONU <span class="font-mono">{{ fir.adrNumeroONU }}</span>
                  · Gr. imballaggio <span class="font-mono">{{ fir.adrGruppoImballaggio }}</span>
                </td>
              </tr>
            </tbody>
          </table>
        </section>

        <!-- Section E + F — Quantità + Operazione -->
        <section class="grid grid-cols-2 gap-3 mb-3">
          <div class="border border-slate-300 p-2 rounded">
            <h2 class="text-[9pt] font-semibold uppercase tracking-wide bg-slate-100 -m-2 mb-1.5 px-2 py-1 border-b border-slate-300">
              E — Quantità e imballaggi
            </h2>
            <table class="w-full text-[9.5pt]">
              <tbody>
                <tr><td class="text-slate-600 py-0.5">Peso lordo</td><td class="text-right font-mono">{{ (fir.pesoLordoKg ?? 0).toLocaleString('it-IT') }} kg</td></tr>
                <tr><td class="text-slate-600 py-0.5">Peso netto</td><td class="text-right font-mono">{{ (fir.pesoNettoKg ?? 0).toLocaleString('it-IT') }} kg</td></tr>
                <tr v-if="fir.pesoArrivoKg"><td class="text-slate-600 py-0.5">Peso a destino</td><td class="text-right font-mono">{{ fir.pesoArrivoKg.toLocaleString('it-IT') }} kg</td></tr>
                <tr><td class="text-slate-600 py-0.5">Numero colli</td><td class="text-right font-mono">{{ fir.numeroColli ?? '—' }}</td></tr>
                <tr><td class="text-slate-600 py-0.5">Imballaggio</td><td class="text-right">{{ imballaggioLabel(fir.tipoImballaggio) }}</td></tr>
              </tbody>
            </table>
          </div>
          <div class="border border-slate-300 p-2 rounded">
            <h2 class="text-[9pt] font-semibold uppercase tracking-wide bg-slate-100 -m-2 mb-1.5 px-2 py-1 border-b border-slate-300">
              F — Operazione di destino
            </h2>
            <p class="font-mono font-bold text-base">{{ fir.operazioneDestino }}</p>
            <p class="text-[9pt] text-slate-700 mt-1">{{ operazioneLabel(fir.operazioneDestino) }}</p>
            <p class="text-[8pt] text-slate-500 mt-2">
              Allegato {{ fir.operazioneDestino.startsWith('R') ? 'C' : 'B' }}
              D.Lgs. 152/2006 (recepimento Dir. 2008/98/CE)
            </p>
          </div>
        </section>

        <!-- Section G — Percorso -->
        <section class="border border-slate-300 p-2 rounded mb-3">
          <h2 class="text-[9pt] font-semibold uppercase tracking-wide bg-slate-100 -m-2 mb-1.5 px-2 py-1 border-b border-slate-300">
            G — Percorso
          </h2>
          <div class="grid grid-cols-3 gap-3">
            <div>
              <p class="text-slate-600 text-[8.5pt] uppercase tracking-wide">Da</p>
              <p class="font-medium">{{ fir.percorso?.da ?? '—' }}</p>
            </div>
            <div>
              <p class="text-slate-600 text-[8.5pt] uppercase tracking-wide">A</p>
              <p class="font-medium">{{ fir.percorso?.a ?? '—' }}</p>
            </div>
            <div class="text-right">
              <p class="text-slate-600 text-[8.5pt] uppercase tracking-wide">Distanza stimata</p>
              <p class="font-mono">{{ fir.percorso?.kmStimati ?? '—' }} km</p>
            </div>
          </div>
        </section>

        <!-- Section H — Annotazioni -->
        <section v-if="fir.annotazioni" class="border border-slate-300 p-2 rounded mb-3">
          <h2 class="text-[9pt] font-semibold uppercase tracking-wide bg-slate-100 -m-2 mb-1.5 px-2 py-1 border-b border-slate-300">
            H — Annotazioni
          </h2>
          <p class="text-[9.5pt]">{{ fir.annotazioni }}</p>
          <p v-if="fir.motivoRespingimento" class="text-[9pt] mt-2 text-red-700">
            <strong>Motivo respingimento:</strong> {{ fir.motivoRespingimento }}
          </p>
          <p v-if="fir.motivoAnnullamento" class="text-[9pt] mt-2 text-slate-600">
            <strong>Motivo annullamento:</strong> {{ fir.motivoAnnullamento }}
          </p>
        </section>

        <!-- Section I — Firme -->
        <section class="grid grid-cols-3 gap-3 mb-3">
          <div class="border border-slate-300 p-2 rounded">
            <h2 class="text-[9pt] font-semibold uppercase tracking-wide bg-slate-100 -m-2 mb-1.5 px-2 py-1 border-b border-slate-300">
              I.1 — Firma produttore
            </h2>
            <p class="text-[8.5pt] text-slate-600 min-h-[2.2em]">
              {{ fir.firmaProduttoreAt ? formatDate(fir.firmaProduttoreAt) : 'In attesa' }}
            </p>
            <div class="border-b border-slate-400 mt-6"></div>
            <p class="text-[8pt] text-slate-500 mt-1">Firma e timbro</p>
          </div>
          <div class="border border-slate-300 p-2 rounded">
            <h2 class="text-[9pt] font-semibold uppercase tracking-wide bg-slate-100 -m-2 mb-1.5 px-2 py-1 border-b border-slate-300">
              I.2 — Firma trasportatore
            </h2>
            <p class="text-[8.5pt] text-slate-600 min-h-[2.2em]">
              {{ fir.firmaTrasportatoreAt ? formatDate(fir.firmaTrasportatoreAt) : 'In attesa' }}
            </p>
            <div class="border-b border-slate-400 mt-6"></div>
            <p class="text-[8pt] text-slate-500 mt-1">Firma e timbro</p>
          </div>
          <div class="border border-slate-300 p-2 rounded">
            <h2 class="text-[9pt] font-semibold uppercase tracking-wide bg-slate-100 -m-2 mb-1.5 px-2 py-1 border-b border-slate-300">
              I.3 — Firma destinatario
            </h2>
            <p class="text-[8.5pt] text-slate-600 min-h-[2.2em]">
              {{ fir.firmaDestinatarioAt ? formatDate(fir.firmaDestinatarioAt) : 'In attesa' }}
            </p>
            <div class="border-b border-slate-400 mt-6"></div>
            <p class="text-[8pt] text-slate-500 mt-1">Firma e timbro</p>
          </div>
        </section>

        <!-- Section L — RENTRI / QR -->
        <section v-if="fir.numeroRENTRI" class="border-2 border-slate-900 p-2 rounded mb-3 flex items-center gap-3">
          <div
            class="shrink-0 w-20 h-20 grid place-items-center bg-white border border-slate-300 rounded"
            :aria-label="`QR code RENTRI ${fir.numeroRENTRI}`"
          >
            <!-- Stylised QR placeholder (deterministic-looking) -->
            <svg viewBox="0 0 8 8" class="w-16 h-16" aria-hidden="true">
              <g fill="currentColor">
                <rect x="0" y="0" width="3" height="3" />
                <rect x="5" y="0" width="3" height="3" />
                <rect x="0" y="5" width="3" height="3" />
                <rect x="1" y="1" width="1" height="1" fill="white" />
                <rect x="6" y="1" width="1" height="1" fill="white" />
                <rect x="1" y="6" width="1" height="1" fill="white" />
                <rect x="4" y="0" width="1" height="1" />
                <rect x="3" y="3" width="1" height="1" />
                <rect x="4" y="4" width="1" height="1" />
                <rect x="5" y="3" width="1" height="1" />
                <rect x="4" y="5" width="1" height="1" />
                <rect x="6" y="5" width="1" height="1" />
                <rect x="7" y="4" width="1" height="1" />
                <rect x="5" y="6" width="1" height="1" />
                <rect x="6" y="7" width="1" height="1" />
                <rect x="3" y="6" width="1" height="1" />
                <rect x="2" y="4" width="1" height="1" />
              </g>
            </svg>
          </div>
          <div class="flex-1">
            <p class="text-[8pt] uppercase tracking-wider text-slate-600">Riferimento RENTRI</p>
            <p class="font-mono font-bold text-base">{{ fir.numeroRENTRI }}</p>
            <p class="text-[8pt] text-slate-600 mt-1">
              Vidimato {{ formatDate(fir.vidimatoAt) }} — payload firmato:
              <span class="font-mono">{{ fir.qrCodePayload }}</span>
            </p>
          </div>
        </section>

        <!-- Footer -->
        <footer class="text-[7.5pt] text-slate-500 border-t border-slate-200 pt-2 mt-2">
          <p>
            Stampato da LogiTrack Rifiuti — istanza {{ DEMO_TENANT_ID }}.
            Documento conforme al D.Lgs. 152/2006 art. 193 e al D.M. MASE 04/04/2023 n. 59.
            La copia produttore (art. 188-bis c. 4) deve essere restituita entro 90 giorni dalla consegna.
          </p>
        </footer>
          </article>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import {
  DEMO_DESTINATARI,
  DEMO_PRODUTTORI,
  DEMO_TENANT_ID,
  DEMO_TRASPORTATORI,
} from '@/demo/data';

interface Vehicle {
  targa: string;
  modello: string;
  adr?: boolean;
  atp?: boolean;
}
interface Driver {
  nome: string;
  cognome: string;
  patente: string;
  cqcMerci?: string;
}
interface Percorso {
  da: string;
  a: string;
  kmStimati: number;
}

interface FirDocument {
  id?: string;
  state: string;
  numeroProgressivo?: string;
  dataEmissione?: string;
  produttoreId: string;
  trasportatoreId: string;
  destinatarioId: string;
  cer: string;
  descrizioneRifiuto?: string;
  statoFisico?: string;
  pericoloso?: boolean;
  classiPericolo?: string[];
  adrClasse?: string;
  adrNumeroONU?: string;
  adrGruppoImballaggio?: string;
  pesoLordoKg?: number;
  pesoNettoKg?: number;
  pesoArrivoKg?: number;
  numeroColli?: number;
  tipoImballaggio?: string;
  operazioneDestino: string;
  percorso?: Percorso;
  vehicle?: Vehicle;
  driver?: Driver;
  annotazioni?: string;
  motivoRespingimento?: string;
  motivoAnnullamento?: string;
  numeroRENTRI?: string;
  qrCodePayload?: string;
  vidimatoAt?: string;
  firmaProduttoreAt?: string | null;
  firmaTrasportatoreAt?: string | null;
  firmaDestinatarioAt?: string | null;
}

const props = defineProps<{
  open: boolean;
  fir: FirDocument | null;
}>();
defineEmits<{ (e: 'close'): void }>();

const produttore = computed(() =>
  DEMO_PRODUTTORI.find((p) => p.id === props.fir?.produttoreId) ?? null,
);
const trasportatore = computed(() =>
  DEMO_TRASPORTATORI.find((t) => t.id === props.fir?.trasportatoreId) ?? null,
);
const destinatario = computed(() =>
  DEMO_DESTINATARI.find((d) => d.id === props.fir?.destinatarioId) ?? null,
);
const autorizzazioneRilevante = computed(() => {
  if (!destinatario.value || !props.fir) return null;
  return destinatario.value.autorizzazioni.find(
    (a) => a.cer === props.fir!.cer && a.operazione === props.fir!.operazioneDestino,
  ) ?? destinatario.value.autorizzazioni[0];
});

function formatDate(iso?: string | null): string {
  if (!iso) return '—';
  try {
    return new Intl.DateTimeFormat('it-IT', {
      dateStyle: 'medium',
      timeStyle: 'short',
    }).format(new Date(iso));
  } catch {
    return iso;
  }
}

function stateLabel(state: string): string {
  const map: Record<string, string> = {
    draft: 'Bozza',
    vidimato: 'Vidimato',
    consegnato_trasportatore: 'Consegnato al trasportatore',
    in_transito: 'In transito',
    consegnato_destinatario: 'Consegnato al destinatario',
    chiuso: 'Chiuso',
    respinto: 'Respinto',
    annullato: 'Annullato',
  };
  return map[state] ?? state;
}

function statoFisicoLabel(s?: string): string {
  if (!s) return '—';
  return (
    {
      solido_non_polverulento: 'Solido non polverulento',
      solido_polverulento: 'Solido polverulento',
      liquido: 'Liquido',
      fangoso_palabile: 'Fangoso palabile',
      aeriforme: 'Aeriforme',
      vischioso_sciropposo: 'Vischioso/sciropposo',
    } as Record<string, string>
  )[s] ?? s;
}

function imballaggioLabel(s?: string): string {
  if (!s) return '—';
  return (
    {
      casse_legno: 'Casse di legno',
      pallet_legno: 'Pallet di legno',
      big_bag: 'Big bag',
      cassone_scarrabile: 'Cassone scarrabile',
      fusti_metallici_200l: 'Fusti metallici 200 L',
      fusti_plastica_60l: 'Fusti plastica 60 L',
      fusti_plastica_200l: 'Fusti plastica 200 L',
      sacconi: 'Sacconi',
    } as Record<string, string>
  )[s] ?? s;
}

function operazioneLabel(op: string): string {
  return (
    {
      R3: 'Riciclaggio/recupero sostanze organiche non solventi',
      R4: 'Riciclaggio/recupero metalli e composti metallici',
      R5: 'Riciclaggio/recupero altri inorganici',
      R13: 'Messa in riserva',
      D1: 'Deposito sul/nel suolo (discarica)',
      D9: 'Trattamento fisico-chimico',
      D10: 'Incenerimento a terra',
      D15: 'Deposito preliminare',
    } as Record<string, string>
  )[op] ?? op;
}

function onPrint(): void {
  window.print();
}
</script>
