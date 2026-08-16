/**
 * API client — same pattern as Booki: same-origin + Vite proxy in dev,
 * platform session from shared UsersPanel cookie.
 */
const SHARED_SESSION_COOKIE_KEY = 'userspanel_session_token'
const SESSION_COOKIE_MAX_AGE_SECONDS = 48 * 3600

function apiOrigin(): string {
  const fromEnv = import.meta.env.VITE_API_BASE_URL
  if (fromEnv !== undefined && String(fromEnv).trim() !== '') {
    return String(fromEnv).replace(/\/$/, '')
  }
  return ''
}

export function apiUrl(path: string): string {
  const p = path.startsWith('/') ? path : `/${path}`
  const o = apiOrigin()
  return o ? `${o}${p}` : p
}

const TOKEN_KEY = 'morph_engi_token'

const AUTH_REFRESH_SKIP = [
  '/api/v1/auth/login',
  '/api/v1/auth/platform-session',
  '/api/v1/auth/dev-login',
]

function shouldRefreshOn401(path: string): boolean {
  return !AUTH_REFRESH_SKIP.some((p) => path.includes(p))
}

let refreshInFlight: Promise<boolean> | null = null

async function refreshSession(): Promise<boolean> {
  if (!refreshInFlight) {
    refreshInFlight = performRefresh().finally(() => {
      refreshInFlight = null
    })
  }
  return refreshInFlight
}

async function performRefresh(): Promise<boolean> {
  setToken(null)
  const platform = resolvePlatformToken()
  if (platform) {
    try {
      await platformSessionFromToken(platform)
      return true
    } catch {
      /* fall through */
    }
  }
  if (import.meta.env.DEV) {
    return devLogin()
  }
  return false
}

async function validateStoredToken(): Promise<boolean> {
  const t = getToken()
  if (!t) return false
  try {
    await fetch(apiUrl('/api/v1/auth/me'), {
      cache: noStore,
      headers: { Authorization: `Bearer ${t}` },
    }).then((r) => {
      if (!r.ok) throw new Error('invalid')
    })
    return true
  } catch {
    setToken(null)
    return false
  }
}

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(token: string | null) {
  if (token) localStorage.setItem(TOKEN_KEY, token)
  else localStorage.removeItem(TOKEN_KEY)
}

function writeCookie(name: string, value: string, maxAgeSeconds: number): void {
  if (typeof document === 'undefined') return
  document.cookie = `${name}=${encodeURIComponent(value)}; Path=/; Max-Age=${maxAgeSeconds}; SameSite=Lax`
}

export function syncPlatformSessionCookie(platformToken: string | null | undefined): void {
  if (typeof document === 'undefined') return
  if (!platformToken) {
    writeCookie(SHARED_SESSION_COOKIE_KEY, '', 0)
    return
  }
  writeCookie(SHARED_SESSION_COOKIE_KEY, platformToken, SESSION_COOKIE_MAX_AGE_SECONDS)
}

function readCookie(name: string): string {
  if (typeof document === 'undefined') return ''
  const prefix = `${name}=`
  const row = document.cookie
    .split(';')
    .map((v) => v.trim())
    .find((v) => v.startsWith(prefix))
  if (!row) return ''
  return decodeURIComponent(row.slice(prefix.length))
}

export function readPlatformSessionCookie(): string {
  return readCookie(SHARED_SESSION_COOKIE_KEY)
}

/** Token from cookie, or ?userspanel_token= from Morph AI Apps menu. */
export function resolvePlatformToken(): string {
  if (typeof window !== 'undefined') {
    const params = new URLSearchParams(window.location.search)
    const fromUrl = params.get('userspanel_token') || params.get('token')
    if (fromUrl?.trim()) {
      syncPlatformSessionCookie(fromUrl.trim())
      params.delete('userspanel_token')
      params.delete('token')
      const next = `${window.location.pathname}${params.toString() ? `?${params}` : ''}${window.location.hash}`
      window.history.replaceState({}, '', next)
      return fromUrl.trim()
    }
  }
  return readPlatformSessionCookie().trim()
}

export function authHeaders(): Record<string, string> {
  const t = getToken()
  const h: Record<string, string> = { 'Content-Type': 'application/json' }
  if (t) h.Authorization = `Bearer ${t}`
  return h
}

const noStore: RequestCache = 'no-store'

export async function api<T = unknown>(path: string, init: RequestInit = {}): Promise<T> {
  const url = apiUrl(path)
  let attempt = 0

  while (attempt < 2) {
    let res: Response
    try {
      res = await fetch(url, {
        cache: noStore,
        credentials: 'include',
        ...init,
        headers: { ...authHeaders(), ...(init.headers as Record<string, string> | undefined) },
      })
    } catch (e) {
      const m = e instanceof Error ? e.message : 'Network error'
      throw new Error(
        m === 'Failed to fetch'
          ? `Cannot reach API (${url}). Start morph-engi-api and open via http://localhost:5179 (Vite proxies /api).`
          : m
      )
    }
    const text = await res.text()
    let data: unknown = null
    if (text) {
      try {
        data = JSON.parse(text)
      } catch {
        data = { raw: text }
      }
    }

    if (res.status === 401 && getToken() && attempt === 0 && shouldRefreshOn401(path)) {
      const renewed = await refreshSession().catch(() => false)
      if (renewed) {
        attempt += 1
        continue
      }
    }

    if (!res.ok) {
      const err = (data as { error?: string })?.error ?? res.statusText
      throw new Error(err)
    }
    return data as T
  }

  throw new Error('Unauthorized — session could not be renewed')
}

function readAuthTokens(payload: unknown): { access_token: string } {
  if (payload == null || typeof payload !== 'object' || Array.isArray(payload)) {
    throw new Error('Auth response was not a JSON object')
  }
  const o = payload as Record<string, unknown>
  const a = o.access_token ?? o.accessToken
  if (typeof a === 'string' && a.length > 0) {
    return { access_token: a }
  }
  throw new Error('Auth response missing access_token')
}

/** Exchange shared UsersPanel JWT for Projects API token. */
export async function platformSessionFromToken(platform: string): Promise<{ access_token: string }> {
  const res = await fetch(apiUrl('/api/v1/auth/platform-session'), {
    method: 'POST',
    cache: noStore,
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${platform}`,
    },
    body: JSON.stringify({ token: platform }),
  })

  const payload = await res.json().catch(() => ({}))
  if (!res.ok) {
    throw new Error((payload as { error?: string }).error || res.statusText)
  }
  const tokens = readAuthTokens(payload)
  setToken(tokens.access_token)
  syncPlatformSessionCookie(platform)
  return tokens
}

export async function platformSessionFromCookie(): Promise<{ access_token: string } | null> {
  const platform = resolvePlatformToken()
  if (!platform) return null
  return platformSessionFromToken(platform)
}

export async function loginWithCredentials(email: string, password: string): Promise<boolean> {
  const out = await api<{ access_token: string }>('/api/v1/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email: email.trim(), password }),
  })
  setToken(out.access_token)
  return true
}

/** Dev fallback when cookie exchange fails but UsersPanel session exists server-side. */
export async function devLogin(): Promise<boolean> {
  try {
    const out = await api<{ access_token: string }>('/api/v1/auth/dev-login', {
      method: 'POST',
      body: '{}',
      credentials: 'include',
    })
    setToken(out.access_token)
    return true
  } catch {
    return false
  }
}

export type EnsureSessionResult =
  | { ok: true }
  | { ok: false; reason: string }

/** Try platform cookie/URL token first, then dev-login (same-origin cookie). */
export async function ensureSession(): Promise<EnsureSessionResult> {
  if (await validateStoredToken()) return { ok: true }

  const platform = resolvePlatformToken()
  if (platform) {
    try {
      await platformSessionFromToken(platform)
      return { ok: true }
    } catch (e) {
      const msg = e instanceof Error ? e.message : 'platform session failed'
      if (!import.meta.env.DEV) {
        return { ok: false, reason: msg }
      }
      // fall through to dev-login in dev
    }
  } else if (typeof window !== 'undefined' && window.location.hostname === '127.0.0.1') {
    return {
      ok: false,
      reason:
        'Open Projects at http://localhost:5179 (not 127.0.0.1) so the Morph AI session cookie is shared.',
    }
  }

  if (import.meta.env.DEV) {
    const devOk = await devLogin()
    if (devOk) return { ok: true }
  }

  if (!platform) {
    return {
      ok: false,
      reason:
        'No Morph AI session found. Sign in at http://localhost:3031, then open Projects from Morph Utils or reload.',
    }
  }

  return { ok: false, reason: 'Could not exchange Morph AI session. Is morph-api running on port 9090?' }
}

export const API_BASE = apiOrigin() || (typeof window !== 'undefined' ? window.location.origin : 'http://localhost:5179')

export async function uploadFile(path: string, file: File): Promise<{ file_url: string; file_name: string; source_type: string }> {
  const form = new FormData()
  form.append('file', file)
  const t = getToken()
  const res = await fetch(apiUrl(path), {
    method: 'POST',
    cache: noStore,
    credentials: 'include',
    headers: t ? { Authorization: `Bearer ${t}` } : {},
    body: form,
  })
  const payload = await res.json().catch(() => ({}))
  if (!res.ok) throw new Error((payload as { error?: string }).error || res.statusText)
  return payload as { file_url: string; file_name: string; source_type: string }
}

/** Upload several files as repeated `files` parts, plus optional text fields. */
export async function uploadFiles<T = unknown>(
  path: string,
  files: File[],
  fields: Record<string, string> = {}
): Promise<T> {
  const form = new FormData()
  for (const f of files) form.append('files', f)
  for (const [k, v] of Object.entries(fields)) {
    if (v) form.append(k, v)
  }
  const t = getToken()
  const res = await fetch(apiUrl(path), {
    method: 'POST',
    cache: noStore,
    credentials: 'include',
    headers: t ? { Authorization: `Bearer ${t}` } : {},
    body: form,
  })
  const payload = await res.json().catch(() => ({}))
  if (!res.ok) throw new Error((payload as { error?: string }).error || res.statusText)
  return payload as T
}
