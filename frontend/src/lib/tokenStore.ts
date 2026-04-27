// Token storage for the SPA.
//
// Threat model addressed:
//   - JWT-in-localStorage is XSS-readable. A single XSS sink anywhere
//     in the SPA (or in a third-party tile/asset library) exfiltrates
//     the bearer token to an attacker domain.
//
// Storage rule:
//   - Access token: in-memory only, plus a sessionStorage fallback so a
//     hard refresh inside the same browser tab does not re-prompt for
//     the password. sessionStorage is scoped to the tab, cleared on
//     close, and not shared across tabs.
//   - Refresh token: deliberately NOT stored client-side. The audit
//     posture is "backend-only refresh tokens, short-lived access
//     tokens"; the backend already issues both pairs and the SPA
//     re-authenticates on access-token expiry. A future refresh-via-
//     httpOnly-cookie path lands when the backend mints a Set-Cookie
//     on /auth/login and /auth/refresh.
//
// This module is the only file that should touch sessionStorage /
// localStorage for the access token. Every other consumer goes
// through getAccessToken() / setAccessToken() / clearAccessToken().

const SESSION_KEY = 'lt.access';

let memoryToken: string | null = null;

export function setAccessToken(token: string | null): void {
  memoryToken = token;
  try {
    if (token) {
      sessionStorage.setItem(SESSION_KEY, token);
    } else {
      sessionStorage.removeItem(SESSION_KEY);
    }
  } catch {
    // Private browsing modes deny sessionStorage; the in-memory
    // token still carries the SPA through the active tab.
  }
}

export function getAccessToken(): string | null {
  if (memoryToken) return memoryToken;
  try {
    const v = sessionStorage.getItem(SESSION_KEY);
    if (v) memoryToken = v;
    return memoryToken;
  } catch {
    return null;
  }
}

export function clearAccessToken(): void {
  setAccessToken(null);
}

export function hasAccessToken(): boolean {
  return getAccessToken() !== null;
}
