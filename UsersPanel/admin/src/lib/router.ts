import { writable } from 'svelte/store'

function parse(): { path: string; query: URLSearchParams } {
  const raw = window.location.hash.replace(/^#/, '') || '/'
  const q = raw.indexOf('?')
  if (q === -1) {
    return { path: raw || '/', query: new URLSearchParams() }
  }
  return {
    path: raw.slice(0, q) || '/',
    query: new URLSearchParams(raw.slice(q + 1)),
  }
}

/** Legacy dashboard URL: send to primary admin landing. */
function redirectDashboardHash() {
  if (typeof window === 'undefined') return
  if (parse().path === '/dashboard') {
    window.location.hash = '#/users'
  }
}

redirectDashboardHash()

export const route = writable(
  typeof window !== 'undefined' ? parse() : { path: '/', query: new URLSearchParams() }
)

if (typeof window !== 'undefined') {
  window.addEventListener('hashchange', () => {
    redirectDashboardHash()
    route.set(parse())
  })
}

export function navigate(path: string) {
  window.location.hash = path.startsWith('#') ? path : `#${path}`
}
