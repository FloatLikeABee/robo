import { writable } from 'svelte/store'

export type ToastItem = {
  id: number
  message: string
  type: 'success' | 'error'
}

export const toasts = writable<ToastItem[]>([])

let seq = 0

export function showToast(message: string, type: 'success' | 'error' = 'success', durationMs = 3500) {
  const id = ++seq
  toasts.update((items) => [...items, { id, message, type }])
  setTimeout(() => {
    toasts.update((items) => items.filter((t) => t.id !== id))
  }, durationMs)
}
