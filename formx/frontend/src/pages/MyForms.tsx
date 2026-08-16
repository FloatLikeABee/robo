import { useCallback, useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { useConfirm } from '../context/ConfirmContext';
import { api, type Form } from '../lib/api';

const LIST_LIMIT = 100;

export function MyForms() {
  const { confirm } = useConfirm();
  const [forms, setForms] = useState<Form[]>([]);
  const [total, setTotal] = useState(0);
  const [search, setSearch] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [copyToast, setCopyToast] = useState<{ message: string; id: string } | null>(null);

  useEffect(() => {
    const t = window.setTimeout(() => setDebouncedSearch(search.trim()), 300);
    return () => window.clearTimeout(t);
  }, [search]);

  useEffect(() => {
    if (!copyToast) return;
    const tid = window.setTimeout(() => setCopyToast(null), 2500);
    return () => window.clearTimeout(tid);
  }, [copyToast]);

  const load = useCallback(() => {
    setLoading(true);
    api.forms
      .list(1, LIST_LIMIT, debouncedSearch)
      .then((r) => {
        setForms(r.forms);
        setTotal(r.total);
        setError(null);
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, [debouncedSearch]);

  useEffect(() => {
    load();
  }, [load]);

  const handleDelete = async (id: number, e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    const ok = await confirm({
      title: 'Delete form',
      message: 'Are you sure you want to delete this form? This cannot be undone.',
      confirmLabel: 'Delete',
      danger: true,
    });
    if (!ok) return;
    try {
      await api.forms.delete(id);
      setForms((prev) => prev.filter((f) => f.id !== id));
      setTotal((n) => Math.max(0, n - 1));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Delete failed');
    }
  };

  return (
    <div className="flex flex-col flex-1 min-h-0">
      {/* Stable header — no scroll */}
      <div className="flex-shrink-0 flex flex-wrap items-center gap-3 mb-3">
        <h1 className="text-xl font-semibold text-slate-900 dark:text-white shrink-0">Forms</h1>
        <input
          type="search"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search forms…"
          className="flex-1 min-w-[12rem] max-w-md px-3 py-2 rounded-lg bg-white dark:bg-slate-800/80 border border-slate-300 dark:border-slate-600 text-slate-900 dark:text-white text-sm placeholder-slate-500 focus:ring-2 focus:ring-violet-500 focus:border-transparent"
        />
        <Link
          to="/forms/new"
          className="shrink-0 inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-violet-600 hover:bg-violet-500 text-white text-sm font-medium"
        >
          + Create Form
        </Link>
      </div>

      {error && (
        <div className="flex-shrink-0 mb-3 p-3 rounded-lg bg-red-100 dark:bg-red-900/30 text-red-800 dark:text-red-300 text-sm">{error}</div>
      )}

      {/* Scrollable list only */}
      <div className="flex-1 min-h-0 overflow-y-auto rounded-xl border border-slate-200 dark:border-slate-700/50 bg-white dark:bg-slate-900/20">
        {loading ? (
          <div className="p-6 text-slate-500 dark:text-slate-400 text-sm">Loading…</div>
        ) : forms.length === 0 ? (
          <div className="p-8 text-center text-slate-500 dark:text-slate-400 text-sm">
            No forms match. {debouncedSearch ? 'Try another search.' : 'Create your first form.'}
          </div>
        ) : (
          <ul className="grid grid-cols-1 sm:grid-cols-2 gap-3 p-3">
            {forms.map((form) => (
              <FormCard
                key={form.id}
                form={form}
                onDelete={handleDelete}
                onCopyMessage={(message) =>
                  setCopyToast({ message, id: typeof crypto !== 'undefined' ? crypto.randomUUID() : String(Date.now()) })
                }
              />
            ))}
          </ul>
        )}
      </div>

      {/* Footer — stable */}
      <div className="flex-shrink-0 mt-3 pt-3 border-t border-slate-200 dark:border-slate-700/50 text-sm text-slate-600 dark:text-slate-400">
        {total === 1 ? '1 form' : `${total} forms`}
        {debouncedSearch && forms.length !== total && (
          <span className="text-slate-500 dark:text-slate-500"> · showing {forms.length} on this page</span>
        )}
      </div>

      {copyToast && (
        <div
          className="fixed bottom-6 left-1/2 z-50 max-w-[min(90vw,20rem)] -translate-x-1/2 rounded-lg border border-violet-500/30 bg-violet-600 px-4 py-2 text-center text-sm font-medium text-white shadow-lg"
          role="status"
          aria-live="polite"
          key={copyToast.id}
        >
          {copyToast.message}
        </div>
      )}
    </div>
  );
}

function FormCard({
  form,
  onDelete,
  onCopyMessage,
}: {
  form: Form;
  onDelete: (id: number, e: React.MouseEvent) => void;
  onCopyMessage: (message: string) => void;
}) {
  const [questionCount, setQuestionCount] = useState<number | null>(null);
  const [responseCount, setResponseCount] = useState<number | null>(null);

  useEffect(() => {
    api.questions.list(form.id).then((q) => setQuestionCount(q.length));
    api.responses.list(form.id, 1, 1).then((r) => setResponseCount(r.total));
  }, [form.id]);

  const formUrl = `${window.location.origin}/f/${form.slug}`;
  const copyUrl = async () => {
    try {
      await navigator.clipboard.writeText(formUrl);
      onCopyMessage('Link copied to clipboard');
    } catch {
      onCopyMessage('Could not copy — check browser permissions');
    }
  };
  const desc = (form.description || '').trim();

  return (
    <li className="flex flex-col min-h-0 min-w-0 rounded-xl border border-[#c6e3ef] bg-[#eef7fc] p-3 shadow-sm hover:bg-[#e6f3fa] dark:border-slate-700/60 dark:bg-slate-800/45 dark:hover:bg-slate-800/65 transition-colors">
      <div className="flex-1 min-w-0">
        <h2 className="text-sm font-semibold text-slate-900 dark:text-white truncate" title={form.name}>
          {form.name}
        </h2>
        <p
          className="text-xs text-slate-600 dark:text-slate-500 mt-1 line-clamp-2 leading-snug break-words"
          title={desc || 'No description'}
        >
          {desc || 'No description'}
        </p>
        <div className="flex flex-wrap gap-x-3 gap-y-0.5 mt-2 text-xs text-slate-500 dark:text-slate-500">
          <span>Responses {responseCount ?? '…'}</span>
          <span>Questions {questionCount ?? '…'}</span>
        </div>
      </div>
      <div className="mt-3 pt-2 border-t border-[#c6e3ef]/70 dark:border-slate-600/40 flex items-center gap-2 min-w-0">
        <button
          type="button"
          onClick={copyUrl}
          className="flex-1 min-w-0 text-violet-600 dark:text-violet-400 hover:underline font-mono text-xs truncate text-left shrink"
          title={formUrl}
        >
          /f/{form.slug}
        </button>
        <div className="flex items-center gap-1 shrink-0">
          <Link
            to={`/forms/${form.id}/edit`}
            className="px-2 py-1 rounded-md bg-slate-200 dark:bg-slate-700/70 hover:bg-slate-300 dark:hover:bg-slate-600/70 text-slate-900 dark:text-white text-xs"
          >
            Edit
          </Link>
          <Link
            to={`/forms/${form.id}/results`}
            className="px-2 py-1 rounded-md bg-slate-200 dark:bg-slate-700/70 hover:bg-slate-300 dark:hover:bg-slate-600/70 text-slate-900 dark:text-white text-xs"
          >
            Results
          </Link>
          <button
            type="button"
            onClick={copyUrl}
            className="p-1.5 rounded-md bg-slate-200 dark:bg-slate-700/70 hover:bg-slate-300 dark:hover:bg-slate-600/70 text-slate-700 dark:text-slate-300 text-xs"
            title="Copy URL"
          >
            ↗
          </button>
          <button
            type="button"
            onClick={(e) => onDelete(form.id, e)}
            className="p-1.5 rounded-md bg-red-900/40 hover:bg-red-800/50 text-red-400 text-xs"
            title="Delete"
          >
            🗑
          </button>
        </div>
      </div>
    </li>
  );
}
