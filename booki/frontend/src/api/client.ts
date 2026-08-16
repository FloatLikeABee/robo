/**
 * API origin for fetch().
 *
 * - Default: same-origin (`''`) — in dev, `vite.config.ts` proxies `/api` to the Go backend.
 * - Set `VITE_API_BASE_URL` when the API is on another origin (e.g. direct `http://127.0.0.1:9095` without Vite).
 */
import { useAuth } from '../store/auth';
/** Empty = same-origin; Vite proxies `/api/messages` to UsersPanel in dev/preview. */
const USERS_PANEL_BASE = (import.meta.env.VITE_USERS_PANEL_API_URL ?? '').replace(/\/$/, '');
const SHARED_SESSION_COOKIE_KEY = 'userspanel_session_token';
const SESSION_COOKIE_MAX_AGE_SECONDS = 48 * 3600;

function writeCookie(name: string, value: string, maxAgeSeconds: number): void {
  if (typeof document === 'undefined') return;
  document.cookie = `${name}=${encodeURIComponent(value)}; Path=/; Max-Age=${maxAgeSeconds}; SameSite=Lax`;
}

function writeSessionCookie(name: string, value: string): void {
  if (typeof document === 'undefined') return;
  document.cookie = `${name}=${encodeURIComponent(value)}; Path=/; SameSite=Lax`;
}

/** Keep UsersPanel JWT in the shared cookie used across Morph / Tran* apps. */
export function syncPlatformSessionCookie(platformToken: string | null | undefined, remember = true): void {
  if (typeof document === 'undefined') return;
  if (!platformToken) {
    writeCookie(SHARED_SESSION_COOKIE_KEY, '', 0);
    return;
  }
  if (remember) {
    writeCookie(SHARED_SESSION_COOKIE_KEY, platformToken, SESSION_COOKIE_MAX_AGE_SECONDS);
  } else {
    writeSessionCookie(SHARED_SESSION_COOKIE_KEY, platformToken);
  }
}

function readCookie(name: string): string {
  if (typeof document === 'undefined') return '';
  const prefix = `${name}=`;
  const row = document.cookie.split(';').map((v) => v.trim()).find((v) => v.startsWith(prefix));
  if (!row) return '';
  return decodeURIComponent(row.slice(prefix.length));
}

export function readPlatformSessionCookie(): string {
  return readCookie(SHARED_SESSION_COOKIE_KEY);
}

function apiOrigin(): string {
  const fromEnv = import.meta.env.VITE_API_BASE_URL;
  if (fromEnv !== undefined && String(fromEnv).trim() !== '') {
    return String(fromEnv).replace(/\/$/, '');
  }
  return '';
}

export function apiUrl(path: string): string {
  const p = path.startsWith('/') ? path : `/${path}`;
  const o = apiOrigin();
  return o ? `${o}${p}` : p;
}

/** Avoid stale cached API responses (can yield empty or wrong bodies while status stays 200). */
const noStore: RequestCache = 'no-store';

/** Don’t recurse refresh on these routes (caller sends no/expired bearer intentionally). */
function shouldAttemptRefreshOn401(path: string): boolean {
  const blocked = [
    '/api/v1/auth/login',
    '/api/v1/auth/register',
    '/api/v1/auth/refresh',
    '/api/v1/auth/platform-session',
  ];
  return !blocked.some((b) => path.includes(b));
}

let refreshInFlight: Promise<boolean> | null = null;

/** Try Redis-backed refresh, then dev-login when allowed (Redis often absent locally). */
function refreshSession(): Promise<boolean> {
  if (!refreshInFlight) {
    refreshInFlight = performRefresh().finally(() => {
      refreshInFlight = null;
    });
  }
  return refreshInFlight;
}

async function performRefresh(): Promise<boolean> {
  try {
    const { refreshToken, setTokens } = useAuth.getState();

    if (refreshToken) {
      try {
        const res = await fetch(apiUrl('/api/v1/auth/refresh'), {
          method: 'POST',
          cache: noStore,
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ refresh_token: refreshToken }),
        });
        if (res.ok) {
          let payload: unknown = null;
          try {
            payload = await res.json();
          } catch {
            payload = null;
          }
          if (payload !== null && typeof payload === 'object') {
            try {
              const t = readAuthTokens(payload);
              setTokens(t.access_token, t.refresh_token);
              return true;
            } catch {
              /* fall through */
            }
          }
        }
      } catch {
        /* network — try dev-login below */
      }
    }

    if (import.meta.env.DEV) {
      try {
        const fromCookie = await platformSessionFromCookie().catch(() => null);
        if (fromCookie) {
          setTokens(fromCookie.access_token, fromCookie.refresh_token, fromCookie.platform_token);
          return true;
        }
      } catch {
        /* fall through */
      }
    }

    return false;
  } catch {
    return false;
  }
}

export async function api<T>(
  path: string,
  options: RequestInit & { token?: string | null } = {}
): Promise<T> {
  let attempt = 0;
  let tok = options.token;

  while (attempt < 2) {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...(options.headers as Record<string, string>),
    };
    if (tok) headers['Authorization'] = `Bearer ${tok}`;
    const { token: _t, ...rest } = options;

    let res: Response;
    try {
      res = await fetch(apiUrl(path), { cache: noStore, ...rest, headers });
    } catch (e) {
      const m = e instanceof Error ? e.message : 'Network error';
      throw new Error(
        m === 'Failed to fetch'
          ? `Cannot reach API (${apiUrl(path)}). Use “npm run dev” and open the Vite URL so /api proxies to the backend, or set VITE_API_BASE_URL.`
          : m
      );
    }
    const text = await res.text();
    let data: unknown = null;
    const trimmed = text.replace(/^\uFEFF/, '');
    try {
      data = trimmed ? JSON.parse(trimmed) : null;
    } catch {
      data = trimmed;
    }

    if (
      res.status === 401 &&
      tok &&
      attempt === 0 &&
      shouldAttemptRefreshOn401(path)
    ) {
      const renewed = await refreshSession().catch(() => false);
      if (renewed) {
        const next = useAuth.getState().accessToken;
        if (next) {
          tok = next;
          attempt += 1;
          continue;
        }
      }
    }

    if (!res.ok) {
      const err = (data as { error?: string })?.error || res.statusText;
      throw new Error(err);
    }
    if (data !== null && typeof data !== 'object') {
      throw new Error(
        typeof data === 'string' && data.trimStart().startsWith('<')
          ? 'API returned HTML instead of JSON — check VITE_API_BASE_URL or open the app via the Vite dev URL so /api is proxied.'
          : 'API returned non-JSON body'
      );
    }
    return data as T;
  }

  throw new Error('Unauthorized — session could not be renewed');
}

/** Read access/refresh tokens from login, register, or refresh responses. */
export function readAuthTokens(payload: unknown): { access_token: string; refresh_token: string | null } {
  if (payload == null || typeof payload !== 'object' || Array.isArray(payload)) {
    throw new Error('Auth response was not a JSON object');
  }

  const tryRead = (x: Record<string, unknown>) => {
    const a = x.access_token;
    const r = x.refresh_token;
    if (typeof a === 'string' && a.length > 0) {
      return { access_token: a, refresh_token: typeof r === 'string' && r.length > 0 ? r : null };
    }
    const a2 = x.accessToken;
    const r2 = x.refreshToken;
    if (typeof a2 === 'string' && a2.length > 0) {
      return { access_token: a2, refresh_token: typeof r2 === 'string' && r2.length > 0 ? r2 : null };
    }
    return null;
  };

  const o = payload as Record<string, unknown>;
  const direct = tryRead(o);
  if (direct) return direct;

  const inner = o.data;
  if (inner && typeof inner === 'object' && !Array.isArray(inner)) {
    const nested = tryRead(inner as Record<string, unknown>);
    if (nested) return nested;
  }

  throw new Error(`Auth response missing tokens (keys: ${Object.keys(o).join(', ')})`);
}

/** Apply login / platform-session payload: Booki JWTs + shared UsersPanel cookie. */
export function applyAuthResponse(payload: unknown): {
  access_token: string;
  refresh_token: string | null;
  platform_token: string | null;
} {
  const tokens = readAuthTokens(payload);
  const platform =
    payload != null && typeof payload === 'object' && !Array.isArray(payload)
      ? (payload as Record<string, unknown>).token
      : undefined;
  let platform_token: string | null = null;
  if (typeof platform === 'string' && platform.length > 0) {
    syncPlatformSessionCookie(platform, true);
    platform_token = platform;
  }
  return { ...tokens, platform_token };
}

export async function platformSessionFromCookie(): Promise<{
  access_token: string;
  refresh_token: string | null;
  platform_token: string | null;
} | null> {
  const platform = readPlatformSessionCookie();
  if (!platform) return null;
  const raw = await postJson<Record<string, unknown>>('/api/v1/auth/platform-session', {}, platform);
  return applyAuthResponse(raw);
}

/** POST JSON and parse with res.json() — fewer edge cases than text()+parse for API bodies. */
export async function postJson<TReturn, TBody = unknown>(
  path: string,
  body: TBody,
  token?: string | null
): Promise<TReturn> {
  let attempt = 0;
  let tok = token;

  while (attempt < 2) {
    const headers: Record<string, string> = { 'Content-Type': 'application/json' };
    if (tok) headers['Authorization'] = `Bearer ${tok}`;

    let res: Response;
    try {
      res = await fetch(apiUrl(path), {
        method: 'POST',
        cache: noStore,
        headers,
        body: JSON.stringify(body),
      });
    } catch (e) {
      const m = e instanceof Error ? e.message : 'Network error';
      throw new Error(
        m === 'Failed to fetch'
          ? `Cannot reach API (${apiUrl(path)}). Use “npm run dev” and open the Vite URL so /api proxies to the backend, or set VITE_API_BASE_URL.`
          : m
      );
    }

    const ct = res.headers.get('content-type') || '';
    let payload: unknown = null;
    if (ct.includes('application/json')) {
      try {
        payload = await res.json();
      } catch {
        payload = null;
      }
    } else {
      const t = await res.text();
      if (!res.ok) throw new Error(t || res.statusText);
      throw new Error(
        `Expected JSON from ${path}, got ${ct}: ${t.slice(0, 200)} (is the Go API running at ${apiOrigin() || 'same origin'}?)`
      );
    }

    if (
      res.status === 401 &&
      tok &&
      attempt === 0 &&
      shouldAttemptRefreshOn401(path)
    ) {
      const renewed = await refreshSession().catch(() => false);
      if (renewed) {
        const next = useAuth.getState().accessToken;
        if (next) {
          tok = next;
          attempt += 1;
          continue;
        }
      }
    }

    if (!res.ok) {
      throw new Error((payload as { error?: string })?.error || res.statusText);
    }
    if (payload == null || typeof payload !== 'object') {
      throw new Error('Empty or invalid JSON from API');
    }
    return payload as TReturn;
  }

  throw new Error('Unauthorized — session could not be renewed');
}

export function apiFile(
  path: string,
  form: FormData,
  token: string | null
): Promise<unknown> {
  const headers: Record<string, string> = {};
  if (token) headers['Authorization'] = `Bearer ${token}`;
  return fetch(apiUrl(path), { method: 'POST', cache: noStore, headers, body: form }).then(async (res) => {
    const text = await res.text();
    const data = text ? JSON.parse(text) : null;
    if (!res.ok) throw new Error((data as { error?: string })?.error || res.statusText);
    return data;
  });
}

/** Backfill persisted platform JWT from the shared cookie (sessions from before `platformJwt` was stored). */
function ensurePlatformJwtHydrated(): void {
  if (typeof document === 'undefined') return;
  if (useAuth.getState().platformJwt?.trim()) return;
  const c = readCookie(SHARED_SESSION_COOKIE_KEY).trim();
  if (!c) return;
  useAuth.setState({ platformJwt: c });
}

function usersPanelAuthToken(): string {
  ensurePlatformJwtHydrated();
  /** UsersPanel validates the platform JWT from login — not Booki's API access_token. */
  const fromStore = useAuth.getState().platformJwt?.trim();
  const shared = readCookie(SHARED_SESSION_COOKIE_KEY).trim();
  return fromStore || shared || '';
}

async function renewPlatformSessionForMessaging(): Promise<boolean> {
  try {
    const renewed = await platformSessionFromCookie();
    if (!renewed?.platform_token) return false;
    useAuth.getState().setTokens(renewed.access_token, renewed.refresh_token, renewed.platform_token);
    return true;
  } catch {
    return false;
  }
}

export async function usersPanelJson<T>(path: string, options: RequestInit = {}, attempt = 0): Promise<T> {
  const token = usersPanelAuthToken();
  const url = `${USERS_PANEL_BASE}${path}`;
  let res: Response;
  try {
    res = await fetch(url, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...(options.headers as Record<string, string>),
      },
    });
  } catch (e) {
    const m = e instanceof Error ? e.message : 'Network error';
    throw new Error(
      m === 'Failed to fetch'
        ? `Cannot reach messaging API (${url || path}). Start UsersPanel (port 5001) and use the Vite dev URL so /api/messages is proxied, or set VITE_USERS_PANEL_API_URL.`
        : m
    );
  }
  const data = await res.json().catch(() => ({}));
  if (res.status === 401 && attempt === 0) {
    const renewed = await renewPlatformSessionForMessaging();
    if (renewed) return usersPanelJson<T>(path, options, 1);
  }
  if (!res.ok) {
    const err = (data as { error?: string }).error || res.statusText;
    if (res.status === 401) {
      throw new Error(`${err} — sign out and sign in again to restore messaging.`);
    }
    throw new Error(err);
  }
  return data as T;
}
