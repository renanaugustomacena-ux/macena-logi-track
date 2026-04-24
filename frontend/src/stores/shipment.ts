import { defineStore } from 'pinia';
import { api } from '@/api/client';

export type ShipmentMode = 'road' | 'rail' | 'multimodal' | 'sea' | 'air';
export type ShipmentStatus =
  | 'draft'
  | 'booked'
  | 'picked_up'
  | 'in_transit'
  | 'delayed'
  | 'at_customs'
  | 'delivered'
  | 'cancelled';

export interface GeoPoint {
  type: 'Point';
  coordinates: [number, number];
}

export interface Shipment {
  id: string;
  reference: string;
  carrier: string;
  mode: ShipmentMode;
  status: ShipmentStatus;
  consignor: { name: string; city: string; country: string };
  consignee: { name: string; city: string; country: string };
  origin: GeoPoint;
  destination: GeoPoint;
  currentPosition?: GeoPoint;
  routePolyline?: string;
  etd: string;
  eta: string;
  updatedAt: string;
  adrClass?: string;
  atpClass?: string;
  vehiclePlate?: string;
}

export interface ShipmentFilters {
  status?: ShipmentStatus | '';
  carrier?: string;
  from?: string;
  to?: string;
}

interface ListResponse {
  items: Shipment[];
  limit: number;
  offset: number;
}

// Pinia store covering the "current list + current detail" needs of
// the dashboard. State is kept small and immutable where reasonable;
// mutations are always funnelled through actions.
export const useShipmentStore = defineStore('shipments', {
  state: () => ({
    items: [] as Shipment[],
    selected: null as Shipment | null,
    filters: { status: '' } as ShipmentFilters,
    loading: false,
    error: null as string | null,
  }),
  getters: {
    inTransit: (s) => s.items.filter((x) => x.status === 'in_transit' || x.status === 'delayed'),
    delayed: (s) => s.items.filter((x) => x.status === 'delayed'),
    count: (s) => s.items.length,
  },
  actions: {
    async fetchAll() {
      this.loading = true;
      this.error = null;
      try {
        const data = await api.get<ListResponse>('/shipments', this.filters as Record<string, string>);
        this.items = data.items ?? [];
      } catch (err) {
        this.error = (err as { detail?: string; code?: string }).detail ?? 'errore_di_rete';
      } finally {
        this.loading = false;
      }
    },
    async fetchOne(id: string) {
      this.loading = true;
      this.error = null;
      try {
        this.selected = await api.get<Shipment>(`/shipments/${id}`);
      } catch (err) {
        this.error = (err as { detail?: string; code?: string }).detail ?? 'errore_di_rete';
        this.selected = null;
      } finally {
        this.loading = false;
      }
    },
    setFilter<K extends keyof ShipmentFilters>(key: K, value: ShipmentFilters[K]) {
      this.filters[key] = value;
    },
  },
});
