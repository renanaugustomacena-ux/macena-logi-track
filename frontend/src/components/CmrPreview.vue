<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="lt-print-portal"
      role="dialog"
      aria-modal="true"
      :aria-label="`Anteprima CMR ${shipment?.reference ?? ''}`"
    >
      <div class="lt-print-backdrop fixed inset-0 z-40 bg-slate-900/60" @click="$emit('close')"></div>

      <div class="fixed inset-0 z-50 flex items-start justify-center overflow-auto p-4 pointer-events-none">
        <div class="lt-print-document w-full max-w-[210mm] pointer-events-auto">
          <!-- Toolbar -->
          <div class="lt-print-toolbar bg-white rounded-t-md shadow flex items-center justify-between px-4 py-2 border-b border-slate-200">
            <div class="text-sm text-slate-600">
              Anteprima CMR <span class="font-mono">{{ shipment?.reference }}</span>
              — Convenzione di Ginevra 19/05/1956
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

          <article v-if="shipment" class="lt-cmr-page bg-white shadow p-8 text-[10pt] leading-snug text-slate-900">
            <!-- Header -->
            <header class="border-b-2 border-slate-900 pb-2 mb-3">
              <div class="flex items-start justify-between">
                <div>
                  <p class="text-[8pt] uppercase tracking-wider text-slate-600">
                    LETTERA DI VETTURA INTERNAZIONALE
                  </p>
                  <h1 class="text-lg font-bold uppercase tracking-tight mt-0.5">
                    CMR — Convention relative au contrat de transport international
                  </h1>
                  <p class="text-[9pt] text-slate-700 mt-0.5">
                    Convenzione di Ginevra 19/05/1956 — Protocollo 05/07/1978
                  </p>
                </div>
                <div class="text-right text-[9pt]">
                  <p>
                    <span class="text-slate-600">Riferimento:</span>
                    <span class="font-mono font-semibold text-base ml-1">{{ shipment.reference }}</span>
                  </p>
                  <p class="mt-0.5">
                    <span class="text-slate-600">Modalità:</span>
                    <span class="font-medium ml-1 uppercase">{{ modeLabel(shipment.mode) }}</span>
                  </p>
                  <p class="mt-0.5">
                    <span class="text-slate-600">Stato:</span>
                    <span class="font-medium ml-1 uppercase">{{ statusLabel(shipment.status) }}</span>
                  </p>
                </div>
              </div>
            </header>

            <!-- Sezioni 1-3 — speditore / destinatario / luogo presa in carico -->
            <section class="grid grid-cols-2 gap-3 mb-3">
              <div class="border border-slate-300 p-2 rounded">
                <h2 class="text-[9pt] font-semibold uppercase tracking-wide bg-slate-100 -m-2 mb-1.5 px-2 py-1 border-b border-slate-300">
                  1 — Speditore (Mittente)
                </h2>
                <p class="font-semibold">{{ shipment.consignor.name }}</p>
                <p>{{ shipment.consignor.city }} — {{ shipment.consignor.country }}</p>
              </div>
              <div class="border border-slate-300 p-2 rounded">
                <h2 class="text-[9pt] font-semibold uppercase tracking-wide bg-slate-100 -m-2 mb-1.5 px-2 py-1 border-b border-slate-300">
                  2 — Destinatario
                </h2>
                <p class="font-semibold">{{ shipment.consignee.name }}</p>
                <p>{{ shipment.consignee.city }} — {{ shipment.consignee.country }}</p>
              </div>
            </section>

            <section class="grid grid-cols-2 gap-3 mb-3">
              <div class="border border-slate-300 p-2 rounded">
                <h2 class="text-[9pt] font-semibold uppercase tracking-wide bg-slate-100 -m-2 mb-1.5 px-2 py-1 border-b border-slate-300">
                  3 — Luogo previsto di consegna
                </h2>
                <p>{{ shipment.consignee.city }}, {{ shipment.consignee.country }}</p>
                <p class="text-[8.5pt] text-slate-600 mt-1">
                  Coord.: {{ formatCoord(shipment.destination.coordinates) }}
                </p>
              </div>
              <div class="border border-slate-300 p-2 rounded">
                <h2 class="text-[9pt] font-semibold uppercase tracking-wide bg-slate-100 -m-2 mb-1.5 px-2 py-1 border-b border-slate-300">
                  4 — Luogo + data presa in carico
                </h2>
                <p>{{ shipment.consignor.city }}, {{ shipment.consignor.country }}</p>
                <p class="text-[8.5pt] text-slate-600 mt-1">
                  Coord.: {{ formatCoord(shipment.origin.coordinates) }}
                </p>
                <p class="text-[8.5pt] text-slate-600 mt-1">
                  Data: <span class="font-medium">{{ formatDate(shipment.etd) }}</span>
                </p>
              </div>
            </section>

            <!-- Sezione 5-6-7 — vettore -->
            <section class="border border-slate-300 p-2 rounded mb-3">
              <h2 class="text-[9pt] font-semibold uppercase tracking-wide bg-slate-100 -m-2 mb-1.5 px-2 py-1 border-b border-slate-300">
                5 — Vettore
              </h2>
              <div class="grid grid-cols-2 gap-3">
                <div>
                  <p class="font-semibold">{{ shipment.carrier }}</p>
                  <p class="text-[9pt]">Targa mezzo: <span class="font-mono">{{ shipment.vehiclePlate ?? '—' }}</span></p>
                </div>
                <div class="text-[9pt]">
                  <p>ETD: <span class="tabular">{{ formatDate(shipment.etd) }}</span></p>
                  <p>ETA: <span class="tabular">{{ formatDate(shipment.eta) }}</span></p>
                </div>
              </div>
            </section>

            <!-- Sezione 6-9 — descrizione merce -->
            <section class="border border-slate-300 p-2 rounded mb-3">
              <h2 class="text-[9pt] font-semibold uppercase tracking-wide bg-slate-100 -m-2 mb-1.5 px-2 py-1 border-b border-slate-300">
                6-9 — Descrizione merce
              </h2>
              <table class="w-full text-[9.5pt] border-collapse">
                <thead>
                  <tr class="bg-slate-50 text-[8.5pt] uppercase text-slate-600">
                    <th class="text-left py-1 px-2 border border-slate-200">Marche e numeri</th>
                    <th class="text-left py-1 px-2 border border-slate-200">Numero colli</th>
                    <th class="text-left py-1 px-2 border border-slate-200">Imballaggio</th>
                    <th class="text-left py-1 px-2 border border-slate-200">Natura merce</th>
                    <th class="text-left py-1 px-2 border border-slate-200">N° statistico</th>
                  </tr>
                </thead>
                <tbody>
                  <tr>
                    <td class="py-1 px-2 border border-slate-200 font-mono">{{ shipment.reference }}</td>
                    <td class="py-1 px-2 border border-slate-200 tabular">—</td>
                    <td class="py-1 px-2 border border-slate-200">Pallet / Cassoni</td>
                    <td class="py-1 px-2 border border-slate-200">Merce {{ shipment.adrClass ? 'pericolosa ADR cl. ' + shipment.adrClass : 'generale' }}</td>
                    <td class="py-1 px-2 border border-slate-200 font-mono">—</td>
                  </tr>
                </tbody>
              </table>
            </section>

            <!-- Sezione 13 — istruzioni mittente -->
            <section class="border border-slate-300 p-2 rounded mb-3">
              <h2 class="text-[9pt] font-semibold uppercase tracking-wide bg-slate-100 -m-2 mb-1.5 px-2 py-1 border-b border-slate-300">
                13 — Istruzioni del mittente (formalità doganali, etc.)
              </h2>
              <p class="text-[9pt]">
                <template v-if="shipment.adrClass">
                  Trasporto ADR classe {{ shipment.adrClass }} — applicare etichette UN, controllare DGSA del mittente.
                </template>
                <template v-else-if="shipment.atpClass">
                  Trasporto ATP classe {{ shipment.atpClass }} — mantenere temperatura stabilita per tutto il percorso.
                </template>
                <template v-else>Nessuna istruzione speciale.</template>
              </p>
            </section>

            <!-- Sezione catena di custodia -->
            <section v-if="custody && custody.length" class="border border-slate-300 p-2 rounded mb-3">
              <h2 class="text-[9pt] font-semibold uppercase tracking-wide bg-slate-100 -m-2 mb-1.5 px-2 py-1 border-b border-slate-300">
                Catena di custodia digitale (firma SHA-256)
              </h2>
              <table class="w-full text-[9pt] border-collapse">
                <tbody>
                  <tr v-for="rec in custody.slice(0, 6)" :key="rec.id" class="border-b border-slate-200">
                    <td class="py-1 px-2 font-mono text-slate-500">#{{ rec.sequence }}</td>
                    <td class="py-1 px-2 capitalize">{{ rec.action }}</td>
                    <td class="py-1 px-2">{{ rec.actor.name }}</td>
                    <td class="py-1 px-2 tabular">{{ formatDate(rec.occurredAt) }}</td>
                    <td class="py-1 px-2 font-mono text-[8pt] text-slate-500">{{ rec.hash.slice(0, 12) }}…</td>
                  </tr>
                </tbody>
              </table>
            </section>

            <!-- Sezioni 22-23-24 — firme -->
            <section class="grid grid-cols-3 gap-3 mb-3">
              <div class="border border-slate-300 p-2 rounded">
                <h2 class="text-[9pt] font-semibold uppercase tracking-wide bg-slate-100 -m-2 mb-1.5 px-2 py-1 border-b border-slate-300">
                  22 — Firma e timbro mittente
                </h2>
                <p class="text-[8.5pt] text-slate-600 min-h-[2.2em]">{{ formatDate(shipment.etd) }}</p>
                <div class="border-b border-slate-400 mt-6"></div>
              </div>
              <div class="border border-slate-300 p-2 rounded">
                <h2 class="text-[9pt] font-semibold uppercase tracking-wide bg-slate-100 -m-2 mb-1.5 px-2 py-1 border-b border-slate-300">
                  23 — Firma e timbro vettore
                </h2>
                <p class="text-[8.5pt] text-slate-600 min-h-[2.2em]">In transito</p>
                <div class="border-b border-slate-400 mt-6"></div>
              </div>
              <div class="border border-slate-300 p-2 rounded">
                <h2 class="text-[9pt] font-semibold uppercase tracking-wide bg-slate-100 -m-2 mb-1.5 px-2 py-1 border-b border-slate-300">
                  24 — Firma e timbro destinatario
                </h2>
                <p class="text-[8.5pt] text-slate-600 min-h-[2.2em]">
                  {{ shipment.status === 'delivered' ? formatDate(shipment.updatedAt) : 'In attesa' }}
                </p>
                <div class="border-b border-slate-400 mt-6"></div>
              </div>
            </section>

            <footer class="text-[7.5pt] text-slate-500 border-t border-slate-200 pt-2 mt-2">
              <p>
                Stampato da LogiTrack — Convenzione di Ginevra 19/05/1956 (CMR), Protocollo 05/07/1978.
                Documento generato automaticamente dalla piattaforma; firma manuale richiesta sulle copie cartacee.
              </p>
            </footer>
          </article>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
interface CustodyRecord {
  id: string;
  sequence: number;
  action: string;
  occurredAt: string;
  actor: { name: string; role: string; organisation: string };
  hash: string;
}
interface ShipmentLite {
  id: string;
  reference: string;
  carrier: string;
  mode: string;
  status: string;
  consignor: { name: string; city: string; country: string };
  consignee: { name: string; city: string; country: string };
  origin: { coordinates: [number, number] };
  destination: { coordinates: [number, number] };
  etd: string;
  eta: string;
  updatedAt: string;
  vehiclePlate?: string;
  adrClass?: string;
  atpClass?: string;
}

defineProps<{
  open: boolean;
  shipment: ShipmentLite | null;
  custody?: CustodyRecord[];
}>();
defineEmits<{ (e: 'close'): void }>();

function formatDate(iso?: string): string {
  if (!iso) return '—';
  try {
    return new Intl.DateTimeFormat('it-IT', { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(iso));
  } catch {
    return iso;
  }
}
function formatCoord([lon, lat]: [number, number]): string {
  return `${lat.toFixed(4)}°N, ${lon.toFixed(4)}°E`;
}
function modeLabel(mode: string): string {
  return ({
    road: 'Strada',
    rail: 'Ferrovia',
    sea: 'Mare',
    air: 'Aria',
    multimodal: 'Multimodale',
  } as Record<string, string>)[mode] ?? mode;
}
function statusLabel(status: string): string {
  return ({
    in_transit: 'in transito',
    delayed: 'in ritardo',
    at_customs: 'in dogana',
    delivered: 'consegnata',
    cancelled: 'annullata',
    booked: 'prenotata',
    picked_up: 'caricata',
    draft: 'bozza',
  } as Record<string, string>)[status] ?? status;
}
function onPrint(): void {
  window.print();
}
</script>
