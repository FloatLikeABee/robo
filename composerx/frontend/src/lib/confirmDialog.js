/** @typedef {{ title?: string, message: string, confirmLabel?: string, cancelLabel?: string, danger?: boolean, alert?: boolean }} ConfirmOptions */

/** @type {ConfirmOptions & { resolve: (v: boolean) => void } | null} */
let pending = null

/** @type {Set<() => void>} */
const listeners = new Set()

function notify() {
  for (const fn of listeners) fn()
}

export function subscribeConfirm(fn) {
  listeners.add(fn)
  fn()
  return () => listeners.delete(fn)
}

export function getConfirmPending() {
  return pending
}

/**
 * @param {ConfirmOptions | string} opts
 * @returns {Promise<boolean>}
 */
export function confirm(opts) {
  const options = typeof opts === 'string' ? { message: opts } : opts
  return new Promise((resolve) => {
    pending = {
      title: options.title ?? 'Confirm',
      message: options.message ?? 'Are you sure?',
      confirmLabel: options.confirmLabel ?? 'OK',
      cancelLabel: options.cancelLabel ?? 'Cancel',
      danger: options.danger ?? false,
      alert: false,
      resolve,
    }
    notify()
  })
}

/**
 * @param {ConfirmOptions | string} opts
 * @returns {Promise<void>}
 */
export function alertDialog(opts) {
  const options = typeof opts === 'string' ? { message: opts } : opts
  return new Promise((resolve) => {
    pending = {
      title: options.title ?? 'Notice',
      message: options.message ?? '',
      confirmLabel: options.confirmLabel ?? 'OK',
      cancelLabel: 'Cancel',
      danger: false,
      alert: true,
      resolve: () => resolve(),
    }
    notify()
  })
}

/** @param {boolean} ok */
export function finishConfirm(ok) {
  if (!pending) return
  const { resolve, alert } = pending
  pending = null
  notify()
  resolve(alert ? true : ok)
}
