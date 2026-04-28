// Tiny factory so demo-mode and production-mode SPA can share the same
// WebSocket call sites. ShipmentMap.vue and TrackingTimeline.vue use
// `createTrackingSocket` instead of `new WebSocket(...)` directly.
//
// In production builds: returns a real WebSocket.
// In demo builds (VITE_DEMO_MODE=true): returns a MockWebSocket that
// emits canned events.

import { MockWebSocket } from '@/demo/wsMock';

export type WSLike = WebSocket | MockWebSocket;

export function createTrackingSocket(url: string, protocols: string[]): WSLike {
  if (import.meta.env.VITE_DEMO_MODE === 'true') {
    return new MockWebSocket(url, protocols);
  }
  return new WebSocket(url, protocols);
}
