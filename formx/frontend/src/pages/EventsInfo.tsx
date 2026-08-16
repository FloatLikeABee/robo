import { useCallback, useEffect, useState } from 'react';
import { useConfirm } from '../context/ConfirmContext';
import {
  EventInfoSubmitCard,
  defaultEventInfoTimeLocal,
  type EventInfoSubmitValues,
} from '../components/EventInfoSubmitCard';
import { MarkdownDetailView } from '../components/MarkdownDetailEditor';
import { api, type EventInfo, type EventInfoCollectionInfo } from '../lib/api';

function formatDisplayTime(iso: string): string {
  try {
    const d = new Date(iso);
    if (Number.isNaN(d.getTime())) return iso;
    return d.toLocaleString(undefined, {
      dateStyle: 'medium',
      timeStyle: 'short',
    });
  } catch {
    return iso;
  }
}

function parseRecipients(raw: string): string[] {
  return raw
    .split(/[\n,;]+/)
    .map((s) => s.trim())
    .filter(Boolean);
}

export function EventsInfo() {
  const { confirm } = useConfirm();
  const [events, setEvents] = useState<EventInfo[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [drawerEvent, setDrawerEvent] = useState<EventInfo | null>(null);

  const [addModalOpen, setAddModalOpen] = useState(false);
  const [formValues, setFormValues] = useState<EventInfoSubmitValues>(() => ({
    title: '',
    detail: '',
    reporter: '',
    time: defaultEventInfoTimeLocal(),
  }));
  const [submitting, setSubmitting] = useState(false);
  const [aiPrompt, setAiPrompt] = useState('');
  const [aiBusy, setAiBusy] = useState(false);
  const [aiHint, setAiHint] = useState<string | null>(null);

  const [shareOpen, setShareOpen] = useState(false);
  const [collectionInfo, setCollectionInfo] = useState<EventInfoCollectionInfo | null>(null);
  const [shareLoading, setShareLoading] = useState(false);
  const [shareError, setShareError] = useState<string | null>(null);
  const [shareSuccess, setShareSuccess] = useState<string | null>(null);
  const [emailTo, setEmailTo] = useState('');
  const [emailMessage, setEmailMessage] = useState('');
  const [emailSending, setEmailSending] = useState<'page' | 'api' | null>(null);
  const [copied, setCopied] = useState(false);

  const [ingestOpen, setIngestOpen] = useState(false);
  const [ingestFiles, setIngestFiles] = useState<File[]>([]);
  const [ingestUrl, setIngestUrl] = useState('');
  const [ingestPaste, setIngestPaste] = useState('');
  const [ingestBusy, setIngestBusy] = useState(false);
  const [ingestHint, setIngestHint] = useState<string | null>(null);
  const [ingestDrafts, setIngestDrafts] = useState<
    Array<{ title: string; detail: string; reporter: string; time: string; selected: boolean }>
  >([]);
  const [ingestSaving, setIngestSaving] = useState(false);

  const resetForm = useCallback(() => {
    setFormValues({
      title: '',
      detail: '',
      reporter: '',
      time: defaultEventInfoTimeLocal(),
    });
    setAiPrompt('');
    setAiHint(null);
  }, []);

  const openIngestModal = () => {
    setIngestFiles([]);
    setIngestUrl('');
    setIngestPaste('');
    setIngestHint(null);
    setIngestDrafts([]);
    setIngestOpen(true);
  };

  const closeIngestModal = () => {
    if (ingestBusy || ingestSaving) return;
    setIngestOpen(false);
  };

  const hasIngestSource =
    ingestFiles.length > 0 || Boolean(ingestUrl.trim()) || Boolean(ingestPaste.trim());

  const runIngest = async () => {
    if (!hasIngestSource || ingestBusy) return;
    setIngestBusy(true);
    setError(null);
    setIngestHint(null);
    setIngestDrafts([]);
    try {
      const out = await api.eventsInfo.aiIngest({
        files: ingestFiles,
        url: ingestUrl,
        paste: ingestPaste,
      });
      const drafts = Array.isArray(out.drafts) ? out.drafts : [];
      setIngestDrafts(
        drafts.map((d) => ({
          title: d.title || '',
          detail: d.detail || '',
          reporter: d.reporter || '',
          time: d.time || '',
          selected: true,
        }))
      );
      setIngestHint(out.assistant_message || `Extracted ${drafts.length} draft(s).`);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Import failed');
    } finally {
      setIngestBusy(false);
    }
  };

  const saveSelectedIngestDrafts = async () => {
    const selected = ingestDrafts.filter((d) => d.selected && d.title.trim());
    if (!selected.length || ingestSaving) return;
    setIngestSaving(true);
    setError(null);
    try {
      for (const d of selected) {
        const t = d.time ? new Date(d.time).toISOString() : new Date().toISOString();
        await api.eventsInfo.create({
          title: d.title.trim(),
          detail: d.detail,
          reporter: d.reporter.trim(),
          time: Number.isNaN(new Date(t).getTime()) ? new Date().toISOString() : t,
        });
      }
      closeIngestModal();
      load();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to save drafts');
    } finally {
      setIngestSaving(false);
    }
  };

  const openAddModal = () => {
    resetForm();
    setAddModalOpen(true);
  };

  const closeAddModal = () => {
    if (aiBusy) return;
    setAddModalOpen(false);
  };

  /** Convert an RFC3339 draft time into datetime-local value for the form. */
  const toLocalDatetimeInput = (iso: string): string => {
    try {
      const d = new Date(iso);
      if (Number.isNaN(d.getTime())) return defaultEventInfoTimeLocal();
      d.setMinutes(d.getMinutes() - d.getTimezoneOffset());
      return d.toISOString().slice(0, 16);
    } catch {
      return defaultEventInfoTimeLocal();
    }
  };

  const runAiDraft = async () => {
    if (!aiPrompt.trim() || aiBusy) return;
    setAiBusy(true);
    setError(null);
    setAiHint(null);
    try {
      const out = await api.eventsInfo.aiDraft({ prompt: aiPrompt.trim() });
      setFormValues({
        title: out.title || '',
        detail: out.detail || '',
        reporter: out.reporter || '',
        time: out.time ? toLocalDatetimeInput(out.time) : defaultEventInfoTimeLocal(),
      });
      setAiHint(out.assistant_message || 'Draft ready — review and save.');
    } catch (e) {
      setError(e instanceof Error ? e.message : 'AI draft failed');
    } finally {
      setAiBusy(false);
    }
  };

  const load = useCallback(() => {
    setLoading(true);
    api.eventsInfo
      .list(1, 200)
      .then((r) => {
        setEvents(r.events);
        setTotal(r.total);
        setError(null);
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  useEffect(() => {
    if (!addModalOpen) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') closeAddModal();
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [addModalOpen]);

  useEffect(() => {
    if (!shareOpen) return;
    setShareLoading(true);
    setShareError(null);
    setShareSuccess(null);
    api.eventsInfo
      .collectionInfo()
      .then(setCollectionInfo)
      .catch((e) => setShareError(e instanceof Error ? e.message : 'Failed to load share info'))
      .finally(() => setShareLoading(false));
  }, [shareOpen]);

  const openRow = (ev: EventInfo) => setDrawerEvent(ev);
  const closeDrawer = () => setDrawerEvent(null);

  const patchForm = useCallback((patch: Partial<EventInfoSubmitValues>) => {
    setFormValues((prev) => ({ ...prev, ...patch }));
  }, []);

  const createEvent = async () => {
    if (!formValues.title.trim()) {
      setError('Title is required');
      return;
    }
    const t = formValues.time ? new Date(formValues.time).toISOString() : new Date().toISOString();
    setSubmitting(true);
    setError(null);
    try {
      await api.eventsInfo.create({
        title: formValues.title.trim(),
        detail: formValues.detail,
        reporter: formValues.reporter.trim(),
        time: t,
      });
      closeAddModal();
      resetForm();
      load();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Create failed');
    } finally {
      setSubmitting(false);
    }
  };

  const copySubmitUrl = async () => {
    const url = collectionInfo?.submit_page_url ?? `${window.location.origin}/events-info/submit`;
    try {
      await navigator.clipboard.writeText(url);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      setShareError('Could not copy to clipboard');
    }
  };

  const sendShareEmail = async (kind: 'page' | 'api') => {
    const to = parseRecipients(emailTo);
    if (to.length === 0) {
      setShareError('Enter at least one email address');
      return;
    }
    setEmailSending(kind);
    setShareError(null);
    setShareSuccess(null);
    try {
      const res = await api.eventsInfo.shareEmail({ to, kind, message: emailMessage.trim() || undefined });
      setShareSuccess(`Sent to ${res.sent_to.join(', ')}`);
      setEmailTo('');
      setEmailMessage('');
    } catch (e) {
      setShareError(e instanceof Error ? e.message : 'Email failed');
    } finally {
      setEmailSending(null);
    }
  };

  const remove = async (ev: EventInfo) => {
    const ok = await confirm({
      title: 'Delete event',
      message: 'Remove this Events & Info record? This cannot be undone.',
      confirmLabel: 'Delete',
      danger: true,
    });
    if (!ok) return;
    try {
      await api.eventsInfo.delete(ev.id);
      if (drawerEvent?.id === ev.id) closeDrawer();
      load();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Delete failed');
    }
  };

  return (
    <div className="flex flex-col flex-1 min-h-0">
      <div className="flex-shrink-0 mb-3 flex flex-col sm:flex-row sm:items-start sm:justify-between gap-3">
        <div className="min-w-0">
          <h1 className="text-xl font-semibold text-slate-900 dark:text-white">Events &amp; Info</h1>
          <p className="text-slate-600 dark:text-slate-400 text-sm mt-1">
            Operational notes. Import from files, a URL, or paste — AI extracts drafts you confirm. Or add an entry by
            hand. Open a row for full detail.
          </p>
        </div>
        <div className="shrink-0 flex flex-wrap gap-2">
          <button
            type="button"
            onClick={openIngestModal}
            className="inline-flex items-center gap-1.5 px-4 py-2 rounded-lg bg-violet-600 hover:bg-violet-500 text-white text-sm font-medium"
          >
            Import
          </button>
          <button
            type="button"
            onClick={openAddModal}
            className="inline-flex items-center gap-1.5 px-4 py-2 rounded-lg bg-slate-200 dark:bg-slate-700 hover:bg-slate-300 dark:hover:bg-slate-600 text-slate-800 dark:text-slate-100 text-sm font-medium"
          >
            + Add New Entry
          </button>
          <button
            type="button"
            onClick={() => setShareOpen(true)}
            className="inline-flex items-center gap-1.5 px-3 py-2 rounded-lg text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 text-sm"
          >
            Share
          </button>
        </div>
      </div>

      {error && (
        <div className="flex-shrink-0 mb-3 p-3 rounded-lg bg-red-100 dark:bg-red-900/30 text-red-800 dark:text-red-300 text-sm">
          {error}
        </div>
      )}

      <div className="flex-1 min-h-0 overflow-y-auto rounded-xl border border-slate-200 dark:border-slate-700/50 bg-white dark:bg-slate-900/20">
        {loading ? (
          <div className="p-6 text-slate-500 dark:text-slate-400 text-sm">Loading…</div>
        ) : events.length === 0 ? (
          <div className="p-8 text-center text-slate-500 dark:text-slate-400 text-sm">No events yet.</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm text-left">
              <thead className="sticky top-0 z-10 bg-slate-100 dark:bg-slate-800/95 border-b border-slate-200 dark:border-slate-700">
                <tr>
                  <th className="px-3 py-2 font-medium text-slate-700 dark:text-slate-300">Title</th>
                  <th className="px-3 py-2 font-medium text-slate-700 dark:text-slate-300 w-40">Reporter</th>
                  <th className="px-3 py-2 font-medium text-slate-700 dark:text-slate-300 w-44">Time</th>
                </tr>
              </thead>
              <tbody>
                {events.map((ev) => (
                  <tr
                    key={ev.id}
                    role="button"
                    tabIndex={0}
                    onClick={() => openRow(ev)}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter' || e.key === ' ') {
                        e.preventDefault();
                        openRow(ev);
                      }
                    }}
                    className="border-b border-slate-100 dark:border-slate-800 cursor-pointer hover:bg-slate-50 dark:hover:bg-slate-800/50"
                  >
                    <td className="px-3 py-2 text-slate-900 dark:text-slate-100">
                      <div className="line-clamp-2 font-medium">{ev.title}</div>
                    </td>
                    <td className="px-3 py-2 text-slate-700 dark:text-slate-300 whitespace-nowrap">
                      {ev.reporter || '—'}
                    </td>
                    <td className="px-3 py-2 text-slate-600 dark:text-slate-400 whitespace-nowrap">
                      {formatDisplayTime(ev.time)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      <div className="flex-shrink-0 mt-3 pt-3 border-t border-slate-200 dark:border-slate-700/50 text-sm text-slate-600 dark:text-slate-400">
        {total === 1 ? '1 event' : `${total} events`}
      </div>

      {ingestOpen && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6 bg-slate-900/55 backdrop-blur-[2px]"
          role="presentation"
          onClick={closeIngestModal}
        >
          <div
            role="dialog"
            aria-modal="true"
            className="w-full max-w-2xl max-h-[min(92dvh,52rem)] flex flex-col rounded-2xl border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-900 shadow-2xl overflow-hidden"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="shrink-0 px-4 py-3 border-b border-slate-200 dark:border-slate-700">
              <h2 className="text-base font-semibold text-slate-900 dark:text-white">Import Events &amp; Info</h2>
              <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
                Upload TXT/MD/PDF, paste a URL, and/or paste text. At least one source is required. Nothing is saved
                until you confirm.
              </p>
            </div>
            <div className="flex-1 min-h-0 overflow-y-auto px-4 py-4 space-y-4 text-sm">
              <label className="block text-xs text-slate-500 dark:text-slate-400">
                Files (TXT, MD, PDF)
                <input
                  type="file"
                  accept=".txt,.md,.markdown,.pdf,text/plain,text/markdown,application/pdf"
                  multiple
                  disabled={ingestBusy || ingestSaving}
                  className="mt-1 block w-full text-sm text-slate-700 dark:text-slate-200"
                  onChange={(e) => setIngestFiles(Array.from(e.target.files || []))}
                />
              </label>
              {ingestFiles.length > 0 ? (
                <ul className="text-xs text-slate-600 dark:text-slate-400 list-disc pl-5">
                  {ingestFiles.map((f) => (
                    <li key={`${f.name}-${f.size}`}>{f.name}</li>
                  ))}
                </ul>
              ) : null}
              <label className="block text-xs text-slate-500 dark:text-slate-400">
                URL
                <input
                  type="url"
                  value={ingestUrl}
                  onChange={(e) => setIngestUrl(e.target.value)}
                  disabled={ingestBusy || ingestSaving}
                  placeholder="https://…"
                  className="mt-1 w-full px-2.5 py-2 rounded-md bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-600 text-slate-900 dark:text-white text-sm"
                />
              </label>
              <label className="block text-xs text-slate-500 dark:text-slate-400">
                Paste content
                <textarea
                  value={ingestPaste}
                  onChange={(e) => setIngestPaste(e.target.value)}
                  rows={4}
                  disabled={ingestBusy || ingestSaving}
                  placeholder="Paste notes, a report excerpt, or meeting minutes…"
                  className="mt-1 w-full px-2.5 py-2 rounded-md bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-600 text-slate-900 dark:text-white text-sm"
                />
              </label>
              <div className="flex flex-wrap items-center gap-2">
                <button
                  type="button"
                  disabled={ingestBusy || ingestSaving || !hasIngestSource}
                  onClick={() => void runIngest()}
                  className="px-3 py-1.5 rounded-lg bg-violet-600 hover:bg-violet-500 text-white text-xs font-medium disabled:opacity-50"
                >
                  {ingestBusy ? 'Extracting…' : 'Extract with AI'}
                </button>
                {ingestHint ? <p className="text-xs text-emerald-700 dark:text-emerald-300">{ingestHint}</p> : null}
              </div>
              {ingestDrafts.length > 0 ? (
                <div className="space-y-2 border-t border-slate-200 dark:border-slate-700 pt-3">
                  <p className="text-xs font-medium text-slate-700 dark:text-slate-300">
                    Review drafts ({ingestDrafts.filter((d) => d.selected).length} selected)
                  </p>
                  {ingestDrafts.map((d, idx) => (
                    <label
                      key={`${d.title}-${idx}`}
                      className="flex gap-2 items-start p-2 rounded-lg border border-slate-200 dark:border-slate-700"
                    >
                      <input
                        type="checkbox"
                        className="mt-1"
                        checked={d.selected}
                        disabled={ingestSaving}
                        onChange={() =>
                          setIngestDrafts((prev) =>
                            prev.map((row, i) => (i === idx ? { ...row, selected: !row.selected } : row))
                          )
                        }
                      />
                      <span className="min-w-0">
                        <span className="block font-medium text-slate-900 dark:text-slate-100">{d.title}</span>
                        {d.detail ? (
                          <span className="block text-xs text-slate-500 dark:text-slate-400 line-clamp-3 mt-0.5">
                            {d.detail}
                          </span>
                        ) : null}
                      </span>
                    </label>
                  ))}
                </div>
              ) : null}
            </div>
            <div className="shrink-0 flex justify-end gap-2 px-4 py-3 border-t border-slate-200 dark:border-slate-700">
              <button
                type="button"
                onClick={closeIngestModal}
                disabled={ingestBusy || ingestSaving}
                className="px-3 py-2 rounded-lg text-sm text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800"
              >
                Cancel
              </button>
              <button
                type="button"
                disabled={
                  ingestBusy ||
                  ingestSaving ||
                  !ingestDrafts.some((d) => d.selected && d.title.trim())
                }
                onClick={() => void saveSelectedIngestDrafts()}
                className="px-3 py-2 rounded-lg bg-violet-600 hover:bg-violet-500 text-white text-sm font-medium disabled:opacity-50"
              >
                {ingestSaving ? 'Saving…' : 'Save selected'}
              </button>
            </div>
          </div>
        </div>
      )}

      {addModalOpen && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6 bg-slate-900/55 backdrop-blur-[2px]"
          role="presentation"
          onClick={closeAddModal}
        >
          <div
            className="w-full max-w-lg max-h-[min(92dvh,48rem)] flex flex-col rounded-2xl border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-900 shadow-2xl shadow-slate-900/25 overflow-hidden"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="shrink-0 px-4 py-3 border-b border-slate-200 dark:border-slate-700 bg-slate-50/90 dark:bg-slate-800/80 space-y-2">
              <div>
                <h2 className="text-base font-semibold text-slate-900 dark:text-white">New entry</h2>
                <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
                  Describe what happened and let AI draft the fields, or fill them in by hand.
                </p>
              </div>
              <label className="block text-xs text-slate-500 dark:text-slate-400">
                AI prompt
                <textarea
                  value={aiPrompt}
                  onChange={(e) => setAiPrompt(e.target.value)}
                  rows={2}
                  disabled={aiBusy}
                  placeholder="e.g. Site B pump room flood this morning — pumps offline, Delta Mechanical notified, waiting on parts"
                  className="mt-1 w-full px-2.5 py-2 rounded-md bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-600 text-slate-900 dark:text-white text-sm disabled:opacity-60"
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
                      e.preventDefault();
                      void runAiDraft();
                    }
                  }}
                />
              </label>
              <div className="flex flex-wrap items-center gap-2">
                <button
                  type="button"
                  disabled={aiBusy || !aiPrompt.trim()}
                  onClick={() => void runAiDraft()}
                  className="px-3 py-1.5 rounded-lg bg-violet-600 hover:bg-violet-500 text-white text-xs font-medium disabled:opacity-50"
                >
                  {aiBusy ? 'Generating…' : 'Generate with AI'}
                </button>
                {aiHint ? <p className="text-xs text-emerald-700 dark:text-emerald-300">{aiHint}</p> : null}
              </div>
            </div>
            <EventInfoSubmitCard
              variant="modal"
              bare
              values={formValues}
              onChange={patchForm}
              onSubmit={createEvent}
              onCancel={closeAddModal}
              submitting={submitting || aiBusy}
            />
          </div>
        </div>
      )}

      {shareOpen && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6 bg-slate-900/55 backdrop-blur-[2px]"
          role="presentation"
          onClick={() => setShareOpen(false)}
        >
          <div
            role="dialog"
            aria-modal="true"
            className="w-full max-w-xl max-h-[min(90dvh,44rem)] flex flex-col rounded-2xl border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-900 shadow-2xl overflow-hidden"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="shrink-0 flex items-start justify-between gap-3 px-4 py-3 border-b border-slate-200 dark:border-slate-700">
              <div>
                <h2 className="text-base font-semibold text-slate-900 dark:text-white">Collect from others</h2>
                <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
                  Share a public submit page or the no-auth API with external contributors.
                </p>
              </div>
              <button
                type="button"
                onClick={() => setShareOpen(false)}
                className="shrink-0 rounded-lg p-2 text-slate-500 hover:bg-slate-100 dark:hover:bg-slate-800"
                aria-label="Close"
              >
                ×
              </button>
            </div>
            <div className="flex-1 min-h-0 overflow-y-auto px-4 py-4 space-y-5 text-sm">
              {shareLoading && <p className="text-slate-500">Loading…</p>}
              {shareError && (
                <div className="p-3 rounded-lg bg-red-100 dark:bg-red-900/30 text-red-800 dark:text-red-300">{shareError}</div>
              )}
              {shareSuccess && (
                <div className="p-3 rounded-lg bg-emerald-100 dark:bg-emerald-900/30 text-emerald-800 dark:text-emerald-200">
                  {shareSuccess}
                </div>
              )}
              {collectionInfo && (
                <>
                  <section>
                    <h3 className="font-medium text-slate-900 dark:text-white mb-2">Public submit page</h3>
                    <p className="text-slate-600 dark:text-slate-400 text-xs mb-2">
                      Dark-mode form — no login required. Send this link by email or chat.
                    </p>
                    <div className="flex gap-2">
                      <input
                        readOnly
                        value={collectionInfo.submit_page_url}
                        className="flex-1 min-w-0 px-2.5 py-2 rounded-md bg-slate-50 dark:bg-slate-800 border border-slate-300 dark:border-slate-600 text-xs font-mono"
                      />
                      <button
                        type="button"
                        onClick={copySubmitUrl}
                        className="shrink-0 px-3 py-2 rounded-lg bg-violet-600 hover:bg-violet-500 text-white text-xs font-medium"
                      >
                        {copied ? 'Copied' : 'Copy'}
                      </button>
                    </div>
                  </section>

                  <section>
                    <h3 className="font-medium text-slate-900 dark:text-white mb-2">Public API</h3>
                    <p className="text-slate-600 dark:text-slate-400 text-xs mb-2">
                      POST JSON without a token. Same fields as the admin form.
                    </p>
                    <code className="block text-xs font-mono bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-md p-2 break-all">
                      POST {collectionInfo.public_api_url}
                    </code>
                    <pre className="mt-2 text-[11px] font-mono bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-md p-2 overflow-x-auto whitespace-pre-wrap">
                      {collectionInfo.sample_curl}
                    </pre>
                  </section>

                  <section className="border-t border-slate-200 dark:border-slate-700 pt-4">
                    <h3 className="font-medium text-slate-900 dark:text-white mb-2">Email instructions</h3>
                    <label className="block text-xs text-slate-500 dark:text-slate-400 mb-1">Recipients</label>
                    <textarea
                      value={emailTo}
                      onChange={(e) => setEmailTo(e.target.value)}
                      rows={2}
                      placeholder="one@example.com, other@example.com"
                      className="w-full px-2.5 py-2 rounded-md bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-600 text-sm mb-2"
                    />
                    <label className="block text-xs text-slate-500 dark:text-slate-400 mb-1">Optional message</label>
                    <textarea
                      value={emailMessage}
                      onChange={(e) => setEmailMessage(e.target.value)}
                      rows={2}
                      placeholder="Hi team — please submit your notes using the link below."
                      className="w-full px-2.5 py-2 rounded-md bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-600 text-sm mb-3"
                    />
                    <div className="flex flex-wrap gap-2">
                      <button
                        type="button"
                        disabled={emailSending !== null}
                        onClick={() => sendShareEmail('page')}
                        className="px-3 py-2 rounded-lg bg-violet-600 hover:bg-violet-500 text-white text-xs font-medium disabled:opacity-50"
                      >
                        {emailSending === 'page' ? 'Sending…' : 'Email submit page link'}
                      </button>
                      <button
                        type="button"
                        disabled={emailSending !== null}
                        onClick={() => sendShareEmail('api')}
                        className="px-3 py-2 rounded-lg bg-slate-700 hover:bg-slate-600 text-white text-xs font-medium disabled:opacity-50"
                      >
                        {emailSending === 'api' ? 'Sending…' : 'Email API instructions'}
                      </button>
                    </div>
                    <p className="text-[11px] text-slate-500 dark:text-slate-400 mt-2">Requires SMTP configured in Settings.</p>
                  </section>
                </>
              )}
            </div>
          </div>
        </div>
      )}

      {drawerEvent && (
        <div className="fixed inset-0 z-50 flex justify-end" aria-modal="true" role="dialog">
          <button
            type="button"
            className="absolute inset-0 appearance-none border-0 p-0 m-0 cursor-default bg-[rgb(2,8,23)]/45 backdrop-blur-[2px] dark:bg-black/40"
            aria-label="Close drawer"
            onClick={closeDrawer}
          />
          <div className="relative z-10 h-full w-full max-w-xl md:max-w-2xl shadow-2xl border-l border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 flex flex-col">
            <div className="shrink-0 flex items-start justify-between gap-3 px-4 py-3 border-b border-slate-200 dark:border-slate-700">
              <div className="min-w-0">
                <h2 className="text-lg font-semibold text-slate-900 dark:text-white leading-snug break-words">
                  {drawerEvent.title}
                </h2>
                <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">
                  {drawerEvent.reporter && (
                    <>
                      <span className="font-medium text-slate-600 dark:text-slate-300">Reporter:</span> {drawerEvent.reporter}
                      <span className="mx-2">·</span>
                    </>
                  )}
                  <span className="font-medium text-slate-600 dark:text-slate-300">Time:</span> {formatDisplayTime(drawerEvent.time)}
                </p>
                <p className="text-[10px] text-slate-400 dark:text-slate-500 font-mono mt-1 break-all">id: {drawerEvent.id}</p>
              </div>
              <button
                type="button"
                onClick={closeDrawer}
                className="shrink-0 p-2 rounded-lg text-slate-500 hover:bg-slate-100 dark:hover:bg-slate-800"
                aria-label="Close"
              >
                ✕
              </button>
            </div>
            <div className="flex-1 min-h-0 overflow-y-auto px-4 py-3 text-slate-800 dark:text-slate-200 text-sm break-words">
              <MarkdownDetailView value={drawerEvent.detail || ''} />
            </div>
            <div className="shrink-0 border-t border-slate-200 dark:border-slate-700 px-4 py-3 flex justify-end gap-2">
              <button
                type="button"
                onClick={() => remove(drawerEvent)}
                className="px-3 py-1.5 rounded-lg text-sm font-medium bg-red-600 hover:bg-red-500 text-white"
              >
                Delete
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
