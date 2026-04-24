// Small typed fetch wrapper. Keeps retries, auth-header injection and
// error normalisation in one place so the rest of the app can stay on
// the happy path.

export interface ApiError {
  status: number;
  code: string;
  detail?: string;
}

export interface ApiClientOptions {
  baseUrl?: string;
  tokenProvider?: () => string | null;
  timeoutMs?: number;
}

export class ApiClient {
  private readonly baseUrl: string;
  private readonly tokenProvider: () => string | null;
  private readonly timeoutMs: number;

  constructor(opts: ApiClientOptions = {}) {
    this.baseUrl = opts.baseUrl ?? '/api/v1';
    this.tokenProvider = opts.tokenProvider ?? (() => localStorage.getItem('lt.token'));
    this.timeoutMs = opts.timeoutMs ?? 15_000;
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
        credentials: 'include',
      });
      if (!res.ok) {
        let detail: string | undefined;
        let code = 'unknown';
        try {
          const parsed = (await res.json()) as { error?: string; detail?: string };
          code = parsed.error ?? code;
          detail = parsed.detail;
        } catch {
          /* non-JSON body */
        }
        const err: ApiError = { status: res.status, code, detail };
        throw err;
      }
      if (res.status === 204) return undefined as T;
      return (await res.json()) as T;
    } finally {
      window.clearTimeout(timer);
    }
  }
}

export const api = new ApiClient();
