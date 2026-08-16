/**
 * Morph / TranDemo auth: aligns with TranForm + TranMail (shared SSO cookie name).
 */
import { API_BASE_URL } from '../apiBase';

const AUTH_TOKEN_KEY = 'morph_auth_token';
const AUTH_REMEMBER_KEY = 'morph_auth_remember';
const AUTH_SNAPSHOT_KEY = 'morph_auth_snapshot';
export const SHARED_SESSION_COOKIE = 'userspanel_session_token';
/** Persist signed-in sessions without a short logout TTL (~100 years). Cleared only on Sign out. */
const SESSION_MAX_AGE_SECONDS = 100 * 365 * 24 * 3600;

let authSnapshotRefreshPromise = null;

function readCookie(name) {
  const prefix = `${name}=`;
  const row = document.cookie
    .split(';')
    .map((v) => v.trim())
    .find((v) => v.startsWith(prefix));
  if (!row) return '';
  return decodeURIComponent(row.slice(prefix.length));
}

function writeCookie(name, value, maxAgeSeconds) {
  document.cookie = `${name}=${encodeURIComponent(value)}; Path=/; Max-Age=${maxAgeSeconds}; SameSite=Lax`;
}

function writeSessionCookie(name, value) {
  document.cookie = `${name}=${encodeURIComponent(value)}; Path=/; SameSite=Lax`;
}

function rememberEnabled() {
  return localStorage.getItem(AUTH_REMEMBER_KEY) !== '0';
}

export function getMorphToken() {
  let t = readCookie(SHARED_SESSION_COOKIE);
  if (t) return t;
  const remember = rememberEnabled();
  if (remember) return localStorage.getItem(AUTH_TOKEN_KEY) || '';
  return sessionStorage.getItem(AUTH_TOKEN_KEY) || '';
}

/**
 * @param {string} token
 * @param {boolean} [rememberMe=true]
 */
export function setMorphToken(token, rememberMe = true) {
  if (!token) {
    clearMorphSession();
    return;
  }
  try {
    localStorage.setItem(AUTH_REMEMBER_KEY, rememberMe ? '1' : '0');
  } catch {
    /* ignore */
  }
  if (rememberMe) {
    localStorage.setItem(AUTH_TOKEN_KEY, token);
    sessionStorage.removeItem(AUTH_TOKEN_KEY);
    writeCookie(SHARED_SESSION_COOKIE, token, SESSION_MAX_AGE_SECONDS);
  } else {
    sessionStorage.setItem(AUTH_TOKEN_KEY, token);
    localStorage.removeItem(AUTH_TOKEN_KEY);
    writeSessionCookie(SHARED_SESSION_COOKIE, token);
  }
}

export function clearMorphSession() {
  try {
    localStorage.removeItem(AUTH_TOKEN_KEY);
  } catch {
    /* ignore */
  }
  try {
    sessionStorage.removeItem(AUTH_TOKEN_KEY);
  } catch {
    /* ignore */
  }
  writeCookie(SHARED_SESSION_COOKIE, '', 0);
  try {
    localStorage.removeItem(AUTH_SNAPSHOT_KEY);
  } catch {
    /* ignore */
  }
  try {
    window.dispatchEvent(new CustomEvent('morph-auth-updated'));
  } catch {
    /* ignore */
  }
}

export function jwtSubjectUnsafe(token) {
  if (!token || typeof token !== 'string') return '';
  const parts = token.split('.');
  if (parts.length !== 3) return '';
  try {
    let payload = parts[1].replace(/-/g, '+').replace(/_/g, '/');
    const pad = payload.length % 4;
    if (pad) payload += '='.repeat(4 - pad);
    const body = JSON.parse(atob(payload));
    return String(body.sub || '').trim();
  } catch {
    return '';
  }
}

function dispatchMorphAuthUpdated() {
  try {
    window.dispatchEvent(new CustomEvent('morph-auth-updated'));
  } catch {
    /* ignore */
  }
}

/** Cache UsersPanel user + permissions (from login or /api/auth/me). */
export function setMorphAuthSnapshot(payload) {
  if (!payload || typeof payload !== 'object') return;
  try {
    const snap = {
      user: payload.user || null,
      permissions: Array.isArray(payload.permissions) ? payload.permissions : [],
    };
    localStorage.setItem(AUTH_SNAPSHOT_KEY, JSON.stringify(snap));
  } catch {
    /* ignore */
  }
  dispatchMorphAuthUpdated();
}

export function getMorphAuthSnapshot() {
  try {
    const raw = localStorage.getItem(AUTH_SNAPSHOT_KEY);
    if (!raw) return null;
    const o = JSON.parse(raw);
    if (!o || typeof o !== 'object') return null;
    return o;
  } catch {
    return null;
  }
}

/** Fetches /api/auth/me and refreshes the cached user snapshot (deduped). */
export async function refreshMorphAuthSnapshot() {
  if (authSnapshotRefreshPromise) return authSnapshotRefreshPromise;
  const token = getMorphToken();
  if (!token) {
    return null;
  }
  const prefix = (API_BASE_URL || '').replace(/\/$/, '');
  authSnapshotRefreshPromise = fetch(`${prefix}/api/auth/me`, {
    headers: { Authorization: `Bearer ${token}` },
  })
    .then(async (res) => {
      const data = await res.json().catch(() => ({}));
      if (!res.ok) return null;
      setMorphAuthSnapshot({ user: data.user, permissions: data.permissions });
      return data;
    })
    .catch(() => null)
    .finally(() => {
      authSnapshotRefreshPromise = null;
    });
  return authSnapshotRefreshPromise;
}

/**
 * @param {string} apiBase — same as REACT_APP_API_URL ('' = same origin)
 * @param {{ email: string, password: string }} body
 */
export async function loginMorph(apiBase, body) {
  const base = (apiBase || '').replace(/\/$/, '');
  const url = `${base}/api/auth/login`;
  const res = await fetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error(data.error || `Login failed (${res.status})`);
  }
  const token = data.token || '';
  if (!token) throw new Error('Missing token in login response');
  return data;
}
