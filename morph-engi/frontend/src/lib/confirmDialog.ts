export type ConfirmOptions = {
  title?: string
  message: string
  confirmLabel?: string
  cancelLabel?: string
  danger?: boolean
}

type Pending = ConfirmOptions & {
  alert: boolean
  resolve: (value: boolean) => void
}

let pending: Pending | null = null
const listeners = new Set<() => void>()

function notify() {
  for (const fn of listeners) fn()
}

export function subscribeConfirm(fn: () => void) {
  listeners.add(fn)
  fn()
  return () => listeners.delete(fn)
}

export function getConfirmPending(): Pending | null {
  return pending
}

export function confirm(opts: ConfirmOptions | string): Promise<boolean> {
  const options = typeof opts === 'string' ? { message: opts } : opts
  return new Promise((resolve) => {
    pending = {
      title: options.title ?? 'Confirm',
      message: options.message ?? 'Are you sure?',
      confirmLabel: options.confirmLabel ?? 'Confirm',
      cancelLabel: options.cancelLabel ?? 'Cancel',
      danger: options.danger ?? false,
      alert: false,
      resolve,
    }
    notify()
  })
}

export function alertDialog(opts: ConfirmOptions | string): Promise<void> {
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

export function finishConfirm(ok: boolean) {
  if (!pending) return
  const { resolve, alert } = pending
  pending = null
  notify()
  resolve(alert ? true : ok)
}
