import {
  appendSessionToken,
  readRuntimeConfig,
  resolvePublicUrl,
  sessionCookieAttributes,
} from './publicUrl';

/** Shared Dorph JWT cookie used across Dorph AI / DorphUtils / embedded apps. */
export const SHARED_SESSION_COOKIE = 'userspanel_session_token';
const AUTH_TOKEN_KEY = 'dorph_auth_token';
const LEGACY_AUTH_TOKEN_KEY = 'morph_auth_token';
/** Persist signed-in sessions without a short logout TTL (~100 years). Cleared only on Sign out. */
const SESSION_MAX_AGE_SECONDS = 100 * 365 * 24 * 3600;

const runtimeConfig = readRuntimeConfig();
const DORPH_API = resolvePublicUrl({
  runtime: runtimeConfig.dorphApiUrl,
  runtimeAlias: runtimeConfig.usersPanelApiUrl,
  built: import.meta.env.VITE_DORPH_API_URL || import.meta.env.VITE_MORPH_API_URL,
  builtAlias: import.meta.env.VITE_USERS_PANEL_API_URL,
  devFallback: '',
  dev: false,
});

/** Prefer same-origin Vite proxy to Dorph (`/api` → :9090). */
function dorphAuthUrl(path: string): string {
  const p = path.startsWith('/') ? path : `/${path}`;
  if (DORPH_API) return `${DORPH_API}${p}`;
  return p;
}

function readCookie(name: string): string {
  if (typeof document === 'undefined') return '';
  const prefix = `${name}=`;
  const row = document.cookie
    .split(';')
    .map((v) => v.trim())
    .find((v) => v.startsWith(prefix));
  if (!row) return '';
  return decodeURIComponent(row.slice(prefix.length));
}

function writeCookie(name: string, value: string, maxAgeSeconds: number): void {
  if (typeof document === 'undefined') return;
  document.cookie = `${name}=${encodeURIComponent(value)}; ${sessionCookieAttributes(maxAgeSeconds)}`;
}

function readLocalToken(): string {
  if (typeof localStorage === 'undefined') return '';
  try {
    return localStorage.getItem(AUTH_TOKEN_KEY) || localStorage.getItem(LEGACY_AUTH_TOKEN_KEY) || '';
  } catch {
    return '';
  }
}

function writeLocalToken(token: string): void {
  if (typeof localStorage === 'undefined') return;
  try {
    if (token) {
      localStorage.setItem(AUTH_TOKEN_KEY, token);
      localStorage.removeItem(LEGACY_AUTH_TOKEN_KEY);
    } else {
      localStorage.removeItem(AUTH_TOKEN_KEY);
      localStorage.removeItem(LEGACY_AUTH_TOKEN_KEY);
    }
  } catch {
    /* ignore */
  }
}

export function isJwtExpired(token: string, skewMs = 60_000): boolean {
  const parts = token.split('.');
  if (parts.length !== 3) return false;
  try {
    let b64 = parts[1].replace(/-/g, '+').replace(/_/g, '/');
    const pad = b64.length % 4;
    if (pad) b64 += '='.repeat(4 - pad);
    const payload = JSON.parse(atob(b64)) as { exp?: number };
    if (typeof payload.exp !== 'number') return false;
    return Date.now() >= payload.exp * 1000 - skewMs;
  } catch {
    return false;
  }
}

export function getSharedToken(): string {
  const t = readCookie(SHARED_SESSION_COOKIE) || readLocalToken();
  if (!t) return '';
  if (isJwtExpired(t)) {
    clearSharedToken();
    return '';
  }
  return t;
}

export function setSharedToken(token: string, _rememberMe = true): void {
  if (!token) {
    clearSharedToken();
    return;
  }
  writeLocalToken(token);
  writeCookie(SHARED_SESSION_COOKIE, token, SESSION_MAX_AGE_SECONDS);
}

export function clearSharedToken(): void {
  writeCookie(SHARED_SESSION_COOKIE, '', 0);
  writeLocalToken('');
}

/** Pull DorphAI / launcher handoff token into the shared cookie, then strip it from the URL. */
export function consumeUrlSessionToken(): string {
  if (typeof window === 'undefined') return '';
  const params = new URLSearchParams(window.location.search);
  const fromUrl = (params.get('userspanel_token') || '').trim();
  if (fromUrl) {
    setSharedToken(fromUrl, true);
    params.delete('userspanel_token');
    const url = new URL(window.location.href);
    url.search = params.toString();
    window.history.replaceState({}, '', url.toString());
    return fromUrl;
  }
  return getSharedToken();
}

export function withSessionToken(baseUrl: string | undefined): string | undefined {
  if (!baseUrl) return undefined;
  const token = getSharedToken();
  if (!token) return baseUrl;
  try {
    return appendSessionToken(new URL(baseUrl, window.location.origin).toString(), token);
  } catch {
    return baseUrl;
  }
}

export async function loginWithDorph(
  email: string,
  password: string,
  rememberMe = true,
): Promise<{ token: string; email: string }> {
  const res = await fetch(dorphAuthUrl('/api/auth/login'), {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email: email.trim(), password }),
  });
  const text = await res.text();
  let data: { token?: string; user?: { email?: string }; error?: string } = {};
  try {
    data = text ? JSON.parse(text) : {};
  } catch {
    /* ignore */
  }
  if (!res.ok || !data.token) {
    throw new Error(data.error || (res.ok ? 'Missing token' : `Login failed (${res.status})`));
  }
  setSharedToken(data.token, rememberMe);
  return { token: data.token, email: data.user?.email || email.trim() };
}

/** @deprecated use loginWithDorph */
export const loginWithUsersPanel = loginWithDorph;

/**
 * Soft-check the Dorph AI shared cookie. Utils itself has no login UI —
 * sign in on Dorph AI; this only validates / refreshes handoff tokens.
 */
export async function ensureSharedSession(): Promise<{ ok: boolean; reason: string }> {
  consumeUrlSessionToken();
  const token = getSharedToken();
  if (!token) return { ok: false, reason: 'No Dorph AI session yet.' };
  try {
    const res = await fetch(dorphAuthUrl('/api/auth/user'), {
      headers: { Authorization: `Bearer ${token}` },
    });
    if (res.status === 401 || res.status === 403) {
      clearSharedToken();
      return { ok: false, reason: 'Dorph AI session expired.' };
    }
    // 5xx / unexpected statuses: keep the token so a proxy blip does not sign the user out.
    return { ok: true, reason: '' };
  } catch {
    return { ok: true, reason: '' };
  }
}

export type AuthUser = { id?: string; email?: string; username?: string; is_admin?: boolean };

export async function fetchAuthMe(): Promise<AuthUser> {
  const token = getSharedToken();
  if (!token) throw new Error('Not signed in');
  const res = await fetch(dorphAuthUrl('/api/auth/me'), {
    headers: { Authorization: `Bearer ${token}` },
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(data.error || 'Failed to load profile');
  return (data.user || {}) as AuthUser;
}

export async function patchAuthMe(body: {
  username?: string;
  password?: string;
  current_password?: string;
}): Promise<AuthUser> {
  const token = getSharedToken();
  if (!token) throw new Error('Not signed in');
  const res = await fetch(dorphAuthUrl('/api/auth/me'), {
    method: 'PATCH',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify(body),
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(data.error || 'Update failed');
  return (data.user || {}) as AuthUser;
}
