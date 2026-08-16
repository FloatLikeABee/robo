import { get } from 'svelte/store'
import { token } from './auth'

/** In dev, Vite proxies /api; use relative URLs. OAuth needs absolute API origin. */
export function apiOrigin(): string {
  return import.meta.env.VITE_API_ORIGIN ?? ''
}

export async function apiFetch(path: string, init: RequestInit = {}): Promise<Response> {
  const url = path.startsWith('http') ? path : `${apiOrigin() || ''}${path}`
  const headers = new Headers(init.headers)
  const t = get(token)
  if (t) headers.set('Authorization', `Bearer ${t}`)
  if (!headers.has('Content-Type') && init.body && typeof init.body === 'string') {
    headers.set('Content-Type', 'application/json')
  }
  return fetch(url, { ...init, headers })
}

export async function apiJson<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await apiFetch(path, init)
  const text = await res.text()
  const data = text ? (JSON.parse(text) as unknown) : null
  if (!res.ok) {
    const err = (data as { error?: string })?.error ?? res.statusText
    throw new Error(err)
  }
  return data as T
}
