// Small typed fetch wrapper. Keeps retries, auth-header injection and
// error normalisation in one place so the rest of the app can stay on
// the happy path.
//
// Hardening applied during the 2026-04-27 audit follow-up:
//  - Token now sourced from tokenStore (in-memory + sessionStorage),
//    not localStorage.
//  - credentials: 'include' removed — the SPA no longer relies on
//    cookies for auth, so dropping it eliminates the CSRF surface
//    that came with cookie+Bearer cohabitation.
//  - 401 raises a typed Unauthorized error and clears the token
//    so the router guard can route to /login on the next navigation.
//  - Error parser reads the RFC 7807 ProblemDetails envelope the
//    backend actually emits ({type, title, status, detail, instance,
//    code}) instead of the legacy {error, detail} shape.

import { clearAccessToken, getAccessToken } from '@/lib/tokenStore';

// Custom event dispatched on the window when any in-flight request
// receives a 401. Decouples this module from vue-router; main.ts
// subscribes and pushes to /login.
export const UNAUTHORIZED_EVENT = 'lt:unauthorized';

export interface ApiError {
  status: number;
  code: string;
  detail?: string;
}

export class Unauthorized extends Error {
  constructor(public readonly detail?: string) {
    super(detail ?? 'unauthorized');
    this.name = 'Unauthorized';
  }
}

export interface ApiClientOptions {
  baseUrl?: string;
  tokenProvider?: () => string | null;
  timeoutMs?: number;
  onUnauthorized?: () => void;
}

export class ApiClient {
  private readonly baseUrl: string;
  private readonly tokenProvider: () => string | null;
  private readonly timeoutMs: number;
  private readonly onUnauthorized: () => void;

  constructor(opts: ApiClientOptions = {}) {
    this.baseUrl = opts.baseUrl ?? '/api/v1';
    this.tokenProvider = opts.tokenProvider ?? getAccessToken;
    this.timeoutMs = opts.timeoutMs ?? 15_000;
    this.onUnauthorized =
      opts.onUnauthorized ??
      (() => {
        clearAccessToken();
        try {
          window.dispatchEvent(new Event(UNAUTHORIZED_EVENT));
        } catch {
          /* no DOM (SSR / tests) */
        }
      });
  }

  async get<T>(path: string, query?: Record<string, string | number | undefined>): Promise<T> {
    const qs = query
      ? '?' +
        Object.entries(query)
          .filter(([, v]) => v !== undefined && v !== '')
          .map(([k, v]) => `${encodeURIComponent(k)}=${encodeURIComponent(String(v))}`)
          .join('&')
      : '';
    return this.request<T>('GET', `${path}${qs}`);
  }

  async post<T, B = unknown>(path: string, body: B): Promise<T> {
    return this.request<T>('POST', path, body);
  }

  private async request<T>(method: string, path: string, body?: unknown): Promise<T> {
    const controller = new AbortController();
    const timer = window.setTimeout(() => controller.abort(), this.timeoutMs);
    const token = this.tokenProvider();
    const headers: Record<string, string> = {
      Accept: 'application/json',
    };
    if (body !== undefined) headers['Content-Type'] = 'application/json';
    if (token) headers.Authorization = `Bearer ${token}`;

    try {
      const res = await fetch(`${this.baseUrl}${path}`, {
        method,
        headers,
        body: body !== undefined ? JSON.stringify(body) : undefined,
        signal: controller.signal,
      });
      if (!res.ok) {
        const parsed = await this.parseProblem(res);
        if (res.status === 401) {
          this.onUnauthorized();
          throw new Unauthorized(parsed.detail);
        }
        throw parsed satisfies ApiError;
      }
      if (res.status === 204) return undefined as T;
      return (await res.json()) as T;
    } finally {
      window.clearTimeout(timer);
    }
  }

  // RFC 7807 ProblemDetails envelope. The backend's problem package
  // emits {type, title, status, detail, instance, code, traceId}.
  private async parseProblem(res: Response): Promise<ApiError> {
    let detail: string | undefined;
    let code = 'unknown';
    try {
      const parsed = (await res.json()) as {
        code?: string;
        detail?: string;
        title?: string;
      };
      code = parsed.code ?? code;
      detail = parsed.detail ?? parsed.title;
    } catch {
      /* non-JSON body — keep the defaults */
    }
    return { status: res.status, code, detail };
  }
}

export const api = new ApiClient();
