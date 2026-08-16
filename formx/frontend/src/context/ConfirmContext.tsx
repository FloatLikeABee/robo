import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from 'react';

export type ConfirmOptions = {
  title?: string;
  message: string;
  confirmLabel?: string;
  cancelLabel?: string;
  /** Destructive action styling for the confirm button */
  danger?: boolean;
};

type PendingConfirm = ConfirmOptions & {
  resolve: (value: boolean) => void;
};

type ConfirmContextValue = {
  confirm: (options: ConfirmOptions) => Promise<boolean>;
};

const ConfirmContext = createContext<ConfirmContextValue | null>(null);

export function ConfirmProvider({ children }: { children: ReactNode }) {
  const [pending, setPending] = useState<PendingConfirm | null>(null);

  const confirm = useCallback((options: ConfirmOptions) => {
    return new Promise<boolean>((resolve) => {
      setPending({ ...options, resolve });
    });
  }, []);

  const finish = useCallback((result: boolean) => {
    setPending((p) => {
      if (p) p.resolve(result);
      return null;
    });
  }, []);

  useEffect(() => {
    if (!pending) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') finish(false);
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [pending, finish]);

  const title = pending?.title ?? 'Confirm';
  const message = pending?.message ?? '';
  const confirmLabel = pending?.confirmLabel ?? 'OK';
  const cancelLabel = pending?.cancelLabel ?? 'Cancel';
  const danger = pending?.danger ?? false;

  return (
    <ConfirmContext.Provider value={{ confirm }}>
      {children}
      {pending && (
        <div
          className="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-slate-900/50 dark:bg-slate-950/60 backdrop-blur-[2px]"
          role="presentation"
          onClick={() => finish(false)}
        >
          <div
            role="alertdialog"
            aria-modal="true"
            aria-labelledby="confirm-dialog-title"
            aria-describedby="confirm-dialog-desc"
            className="w-full max-w-md rounded-xl border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 shadow-[0_25px_80px_-12px_rgba(0,0,0,0.35)] ring-1 ring-black/5 dark:ring-white/10"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="px-4 py-3 border-b border-slate-200 dark:border-slate-700">
              <h2 id="confirm-dialog-title" className="font-semibold text-sm sm:text-base">
                {title}
              </h2>
            </div>
            <p id="confirm-dialog-desc" className="px-4 py-3 text-sm text-slate-600 dark:text-slate-300">
              {message}
            </p>
            <div className="px-4 py-3 border-t border-slate-200 dark:border-slate-700 flex justify-end gap-2">
              <button
                type="button"
                onClick={() => finish(false)}
                className="px-3 py-1.5 rounded-lg bg-slate-200 dark:bg-slate-700 text-slate-800 dark:text-slate-100 text-sm hover:bg-slate-300 dark:hover:bg-slate-600"
              >
                {cancelLabel}
              </button>
              <button
                type="button"
                onClick={() => finish(true)}
                className={
                  danger
                    ? 'px-3 py-1.5 rounded-lg bg-red-600 hover:bg-red-500 text-white text-sm'
                    : 'px-3 py-1.5 rounded-lg bg-violet-600 hover:bg-violet-500 text-white text-sm'
                }
              >
                {confirmLabel}
              </button>
            </div>
          </div>
        </div>
      )}
    </ConfirmContext.Provider>
  );
}

export function useConfirm() {
  const ctx = useContext(ConfirmContext);
  if (!ctx) {
    throw new Error('useConfirm must be used within ConfirmProvider');
  }
  return ctx;
}
