// MockWebSocket — duck-types the global WebSocket interface enough for
// ShipmentMap.vue and TrackingTimeline.vue. Emits canned position
// updates for the live demo shipment along DEMO_LIVE_POLYLINE on a
// timer, then loops.

import { DEMO_LIVE_POLYLINE, DEMO_STEP_MS, DEMO_TIMELINE_TYPES } from './data';

const LIVE_SHIPMENT_ID = 'demo-shp-001';

type Handler = ((ev: { data: string }) => void) | null;
type CloseHandler = ((ev: { code: number }) => void) | null;

export class MockWebSocket {
  // Match WebSocket.readyState constants closely enough for any caller
  // that checks `readyState === 1`.
  static readonly CONNECTING = 0;
  static readonly OPEN = 1;
  static readonly CLOSING = 2;
  static readonly CLOSED = 3;

  readyState = MockWebSocket.CONNECTING;

  onopen: ((ev: Event) => void) | null = null;
  onmessage: Handler = null;
  onclose: CloseHandler = null;
  onerror: ((ev: Event) => void) | null = null;

  private subscribedShipmentId: string | null = null;
  private timer: ReturnType<typeof setInterval> | null = null;
  private waypointIndex = 0;

  constructor(_url: string, _protocols?: string | string[]) {
    // Open on next tick so the caller can attach handlers first.
    setTimeout(() => {
      this.readyState = MockWebSocket.OPEN;
      this.onopen?.(new Event('open'));
      this.startEmission();
    }, 50);
  }

  send(data: string): void {
    try {
      const msg = JSON.parse(data) as { op?: string; shipmentId?: string };
      if (msg.op === 'subscribe' && msg.shipmentId) {
        this.subscribedShipmentId = msg.shipmentId;
      }
    } catch {
      /* ignore — demo mode is permissive */
    }
  }

  close(): void {
    this.readyState = MockWebSocket.CLOSED;
    if (this.timer) {
      clearInterval(this.timer);
      this.timer = null;
    }
    this.onclose?.({ code: 1000 });
  }

  private startEmission(): void {
    // Always emit; if the subscribed shipment doesn't match the live
    // demo, the consumer simply filters and the timeline stays empty
    // (which is the "no events yet" state).
    this.emitOne();
    this.timer = setInterval(() => this.emitOne(), DEMO_STEP_MS);
  }

  private emitOne(): void {
    if (this.readyState !== MockWebSocket.OPEN) return;
    if (this.subscribedShipmentId && this.subscribedShipmentId !== LIVE_SHIPMENT_ID) return;
    const idx = this.waypointIndex % DEMO_LIVE_POLYLINE.length;
    const [lon, lat] = DEMO_LIVE_POLYLINE[idx];
    const now = new Date().toISOString();
    const type = DEMO_TIMELINE_TYPES[idx] ?? 'position_update';
    const payload = {
      id: `demo-evt-${Date.now()}-${idx}`,
      shipmentId: LIVE_SHIPMENT_ID,
      type,
      occurredAt: now,
      recordedAt: now,
      position: { type: 'Point', coordinates: [lon, lat] },
      source: 'demo',
    };
    this.onmessage?.({ data: JSON.stringify(payload) });
    this.waypointIndex++;
  }
}

export type WSLike = WebSocket | MockWebSocket;
