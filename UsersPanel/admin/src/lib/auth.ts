import { get, writable } from 'svelte/store'

const TOKEN_KEY = 'users_panel_token'

function loadToken(): string | null {
  if (typeof localStorage === 'undefined') return null
  return localStorage.getItem(TOKEN_KEY)
}

export const token = writable<string | null>(loadToken())

export function setToken(t: string | null) {
  token.set(t)
  if (typeof localStorage === 'undefined') return
  if (t) localStorage.setItem(TOKEN_KEY, t)
  else localStorage.removeItem(TOKEN_KEY)
}

export function logout() {
  setToken(null)
}

export function authHeaders(): HeadersInit {
  const t = get(token)
  return t ? { Authorization: `Bearer ${t}` } : {}
}
