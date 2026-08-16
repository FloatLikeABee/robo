import { useEffect } from 'react';
import { MarkdownDetailEditor } from './MarkdownDetailEditor';

export type EventInfoSubmitValues = {
  title: string;
  detail: string;
  reporter: string;
  time: string;
};

export function defaultEventInfoTimeLocal(): string {
  const d = new Date();
  d.setMinutes(d.getMinutes() - d.getTimezoneOffset());
  return d.toISOString().slice(0, 16);
}

type EventInfoSubmitCardProps = {
  variant?: 'modal' | 'public';
  title?: string;
  subtitle?: string;
  /** When true, omit the outer dialog chrome (used when nested inside a parent modal). */
  bare?: boolean;
  values: EventInfoSubmitValues;
  onChange: (patch: Partial<EventInfoSubmitValues>) => void;
  onSubmit: () => void;
  onCancel?: () => void;
  submitting?: boolean;
  error?: string | null;
  successMessage?: string | null;
  showCancel?: boolean;
};

export function EventInfoSubmitCard({
  variant = 'modal',
  title = 'New entry',
  subtitle = 'Add an operational note. Title is required. Detail supports Markdown.',
  bare = false,
  values,
  onChange,
  onSubmit,
  onCancel,
  submitting = false,
  error = null,
  successMessage = null,
  showCancel = true,
}: EventInfoSubmitCardProps) {
  const isPublic = variant === 'public';

  useEffect(() => {
    if (isPublic) {
      document.documentElement.classList.add('dark');
      return () => document.documentElement.classList.remove('dark');
    }
    return undefined;
  }, [isPublic]);

  const shellClass = bare
    ? 'flex flex-col flex-1 min-h-0 overflow-hidden'
    : isPublic
      ? 'w-full max-w-2xl rounded-2xl border border-slate-600 bg-slate-900 shadow-2xl shadow-black/40 overflow-hidden'
      : 'w-full max-w-lg max-h-[min(90dvh,44rem)] flex flex-col rounded-2xl border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-900 shadow-2xl shadow-slate-900/25 overflow-hidden';

  const headerClass = isPublic
    ? 'px-6 py-4 border-b border-slate-700 bg-slate-800/80'
    : 'px-4 py-3 border-b border-slate-200 dark:border-slate-700 bg-slate-50/90 dark:bg-slate-800/80';

  const bodyClass = isPublic ? 'px-6 py-5' : 'flex-1 min-h-0 overflow-y-auto px-4 py-3';

  const footerClass = isPublic
    ? 'px-6 py-4 border-t border-slate-700 bg-slate-800/60 flex justify-end gap-2'
    : 'px-4 py-3 border-t border-slate-200 dark:border-slate-700 bg-slate-50/80 dark:bg-slate-800/60 flex flex-wrap justify-end gap-2';

  const inputClass = isPublic
    ? 'w-full px-3 py-2.5 rounded-lg bg-slate-800 border border-slate-600 text-slate-100 text-sm'
    : 'w-full px-2.5 py-2 rounded-md bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-600 text-slate-900 dark:text-white text-sm';

  const labelClass = isPublic
    ? 'block text-xs text-slate-400 mb-1.5'
    : 'block text-xs text-slate-500 dark:text-slate-400 mb-1';

  const showHeader = !bare && Boolean(title);

  return (
    <div
      className={shellClass}
      role={isPublic || bare ? undefined : 'dialog'}
      aria-modal={isPublic || bare ? undefined : true}
    >
      {showHeader ? (
        <div className={headerClass}>
          <h2 className={`font-semibold text-white leading-tight ${isPublic ? 'text-xl' : 'text-base'}`}>{title}</h2>
          {subtitle && <p className={`text-slate-400 mt-1 ${isPublic ? 'text-sm' : 'text-xs'}`}>{subtitle}</p>}
        </div>
      ) : null}
      <div className={bodyClass}>
        {error && (
          <div className="mb-3 p-3 rounded-lg bg-red-900/40 border border-red-700/50 text-red-200 text-sm">{error}</div>
        )}
        {successMessage && (
          <div className="mb-3 p-3 rounded-lg bg-emerald-900/35 border border-emerald-700/50 text-emerald-100 text-sm">
            {successMessage}
          </div>
        )}
        <div className={`grid grid-cols-1 sm:grid-cols-2 ${isPublic ? 'gap-4' : 'gap-3'}`}>
          <div className="sm:col-span-2">
            <label className={labelClass}>Title *</label>
            <input
              type="text"
              value={values.title}
              onChange={(e) => onChange({ title: e.target.value })}
              className={inputClass}
              placeholder="Short title"
              autoFocus={!bare}
            />
          </div>
          <div>
            <label className={labelClass}>Reporter</label>
            <input
              type="text"
              value={values.reporter}
              onChange={(e) => onChange({ reporter: e.target.value })}
              className={inputClass}
              placeholder="Name or role"
            />
          </div>
          <div>
            <label className={labelClass}>Time</label>
            <input
              type="datetime-local"
              value={values.time}
              onChange={(e) => onChange({ time: e.target.value })}
              className={inputClass}
            />
          </div>
          <div className="sm:col-span-2">
            <label className={labelClass}>Detail (Markdown)</label>
            <MarkdownDetailEditor
              value={values.detail}
              onChange={(detail) => onChange({ detail })}
              rows={isPublic ? 7 : 5}
              isPublic={isPublic}
              placeholder="Longer description, links, lists, headings…"
            />
          </div>
        </div>
      </div>
      <div className={footerClass}>
        {showCancel && onCancel && (
          <button
            type="button"
            onClick={onCancel}
            className={
              isPublic
                ? 'px-4 py-2 rounded-lg bg-slate-700 text-slate-100 text-sm font-medium'
                : 'px-4 py-2 rounded-lg bg-slate-200 dark:bg-slate-700 text-slate-800 dark:text-slate-100 text-sm font-medium'
            }
          >
            Cancel
          </button>
        )}
        <button
          type="button"
          disabled={submitting || Boolean(successMessage)}
          onClick={onSubmit}
          className="px-4 py-2 rounded-lg bg-violet-600 hover:bg-violet-500 text-white text-sm font-medium disabled:opacity-50"
        >
          {submitting ? 'Saving…' : successMessage ? 'Submitted' : 'Save event'}
        </button>
      </div>
    </div>
  );
}
