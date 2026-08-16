import { writable } from 'svelte/store'

const KEY = 'users_panel_theme'

export type Theme = 'light' | 'dark'

export function readTheme(): Theme {
  if (typeof localStorage === 'undefined') return 'dark'
  const v = localStorage.getItem(KEY)
  if (v === 'light' || v === 'dark') return v
  if (typeof window !== 'undefined' && window.matchMedia) {
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
  }
  return 'dark'
}

export const theme = writable<Theme>(readTheme())

export function applyTheme(t: Theme) {
  if (typeof document === 'undefined') return
  document.documentElement.dataset.theme = t
  localStorage.setItem(KEY, t)
}

export function toggleTheme() {
  theme.update((cur) => {
    const next: Theme = cur === 'dark' ? 'light' : 'dark'
    applyTheme(next)
    return next
  })
}
