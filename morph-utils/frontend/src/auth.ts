/** Shared Morph JWT cookie used across Morph AI / Morph Utils / embedded apps. */
export const SHARED_SESSION_COOKIE = 'userspanel_session_token';
/** Persist signed-in sessions without a short logout TTL (~100 years). Cleared only on Sign out. */
const SESSION_MAX_AGE_SECONDS = 100 * 365 * 24 * 3600;

const MORPH_API = (import.meta.env.VITE_MORPH_API_URL ?? import.meta.env.VITE_USERS_PANEL_API_URL ?? '').replace(
  /\/$/,
  '',
);

/** Prefer same-origin Vite proxy to Morph (`/api` → :9090). */
function morphAuthUrl(path: string): string {
  const p = path.startsWith('/') ? path : `/${path}`;
  if (MORPH_API) return `${MORPH_API}${p}`;
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
  document.cookie = `${name}=${encodeURIComponent(value)}; Path=/; Max-Age=${maxAgeSeconds}; SameSite=Lax`;
}

function writeSessionCookie(name: string, value: string): void {
  if (typeof document === 'undefined') return;
  document.cookie = `${name}=${encodeURIComponent(value)}; Path=/; SameSite=Lax`;
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
  const t = readCookie(SHARED_SESSION_COOKIE);
  if (!t) return '';
  if (isJwtExpired(t)) {
    clearSharedToken();
    return '';
  }
  return t;
}

export function setSharedToken(token: string, rememberMe = true): void {
  if (!token) {
    clearSharedToken();
    return;
  }
  if (rememberMe) writeCookie(SHARED_SESSION_COOKIE, token, SESSION_MAX_AGE_SECONDS);
  else writeSessionCookie(SHARED_SESSION_COOKIE, token);
}

export function clearSharedToken(): void {
  writeCookie(SHARED_SESSION_COOKIE, '', 0);
}

/** Pull MorphAI / launcher handoff token into the shared cookie, then strip it from the URL. */
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
    const url = new URL(baseUrl, window.location.origin);
    url.searchParams.set('userspanel_token', token);
    return url.toString();
  } catch {
    return baseUrl;
  }
}

export async function loginWithMorph(
  email: string,
  password: string,
  rememberMe = true,
): Promise<{ token: string; email: string }> {
  const res = await fetch(morphAuthUrl('/api/auth/login'), {
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

/** @deprecated use loginWithMorph */
export const loginWithUsersPanel = loginWithMorph;

/**
 * Soft-check the Morph AI shared cookie. Utils itself has no login UI —
 * sign in on Morph AI; this only validates / refreshes handoff tokens.
 */
export async function ensureSharedSession(): Promise<{ ok: boolean; reason: string }> {
  consumeUrlSessionToken();
  const token = getSharedToken();
  if (!token) return { ok: false, reason: 'No Morph AI session yet.' };
  try {
    const res = await fetch(morphAuthUrl('/api/auth/user'), {
      headers: { Authorization: `Bearer ${token}` },
    });
    if (!res.ok) {
      clearSharedToken();
      return { ok: false, reason: 'Morph AI session expired.' };
    }
    return { ok: true, reason: '' };
  } catch {
    // Network blip — keep cookie so embeds can still try.
    return { ok: true, reason: '' };
  }
}
