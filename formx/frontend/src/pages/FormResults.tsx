import { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { useConfirm } from '../context/ConfirmContext';
import { QuestionPromptAttachment } from '../components/QuestionPromptAttachment';
import { api, type Form, type FormPage, type FormResponse, type Question } from '../lib/api';

function formatResponseExamMs(ms?: number): string | null {
  if (ms == null || ms <= 0) return null;
  const s = Math.max(0, Math.floor(ms / 1000));
  const m = Math.floor(s / 60);
  const r = s % 60;
  if (m >= 60) {
    const h = Math.floor(m / 60);
    const mr = m % 60;
    return `${h}h ${mr}m ${r}s`;
  }
  return m > 0 ? `${m}m ${String(r).padStart(2, '0')}s` : `${r}s`;
}

export function FormResults() {
  const { confirm } = useConfirm();
  const { id } = useParams<{ id: string }>();
  const formId = id ? Number(id) : NaN;
  const [form, setForm] = useState<Form | null>(null);
  const [pages, setPages] = useState<FormPage[]>([]);
  const [questions, setQuestions] = useState<Question[]>([]);
  const [responses, setResponses] = useState<FormResponse[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [openResponse, setOpenResponse] = useState<FormResponse | null>(null);
  const [emailOpen, setEmailOpen] = useState(false);
  const [emailTo, setEmailTo] = useState('');
  const [emailSubject, setEmailSubject] = useState('');
  const [emailSending, setEmailSending] = useState(false);
  const [emailError, setEmailError] = useState<string | null>(null);

  useEffect(() => {
    if (!id || Number.isNaN(formId)) {
      setError('Invalid form id');
      setLoading(false);
      return;
    }
    Promise.all([
      api.forms.get(formId),
      api.pages.list(formId),
      api.questions.list(formId),
      api.responses.list(formId, 1, 100),
    ])
      .then(([f, p, q, r]) => {
        setForm(f);
        setPages(p);
        setQuestions(q);
        setResponses(r.responses);
        setTotal(r.total);
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, [formId, id]);

  if (loading) {
    return (
      <div className="flex flex-col flex-1 min-h-0 overflow-hidden">
        <div className="text-slate-500 dark:text-slate-400 text-sm p-2">Loading…</div>
      </div>
    );
  }
  if (error || !form) {
    return (
      <div className="flex flex-col flex-1 min-h-0 overflow-hidden">
        <div className="text-red-600 dark:text-red-400 text-sm p-2">{error || 'Form not found'}</div>
      </div>
    );
  }

  const qList = questions || [];
  const rList = responses || [];
  const pageGroups = groupQuestionsByPage(qList, pages);
  const formIcon = iconForForm(form?.id ?? 0);
  const previewQuestions = qList.slice(0, 2);
  const shortDescription = (form.description || '').trim();
  const deleteResponse = async (responseId: string) => {
    const ok = await confirm({
      title: 'Delete response',
      message: 'Are you sure you want to delete this individual response? This cannot be undone.',
      confirmLabel: 'Delete',
      danger: true,
    });
    if (!ok) return;
    try {
      await api.responses.delete(formId, responseId);
      setResponses((prev) => prev.filter((r) => r.id !== responseId));
      setTotal((prev) => Math.max(0, prev - 1));
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Delete failed');
    }
  };

  const submitEmail = async () => {
    if (!openResponse) return;
    const to = emailTo
      .split(/[\n,;]+/)
      .map((s) => s.trim())
      .filter(Boolean);
    if (to.length === 0) {
      setEmailError('Enter at least one recipient email');
      return;
    }
    setEmailSending(true);
    setEmailError(null);
    try {
      await api.responses.email(formId, openResponse.id, { to, subject: emailSubject.trim() || undefined });
      setEmailOpen(false);
      setEmailTo('');
      setEmailSubject('');
    } catch (e) {
      setEmailError(e instanceof Error ? e.message : 'Email failed');
    } finally {
      setEmailSending(false);
    }
  };

  return (
    <div className="flex flex-col flex-1 min-h-0 overflow-hidden w-full max-w-full">
      {/* Fixed dashboard: no scroll */}
      <div className="shrink-0 flex flex-col gap-2 min-w-0">
        <div className="flex flex-wrap items-center gap-2">
          <div className="flex items-center gap-2 min-w-0">
            <Link
              to="/forms"
              className="p-1.5 rounded-lg text-slate-500 dark:text-slate-400 hover:bg-slate-200 dark:hover:bg-slate-700/50 shrink-0"
            >
              ←
            </Link>
            <div className="min-w-0">
              <h1 className="text-lg font-semibold text-slate-900 dark:text-white leading-tight truncate">
                Form Results
              </h1>
              <p className="text-slate-600 dark:text-slate-400 text-xs truncate">{form.name}</p>
            </div>
          </div>
        </div>

        <div className="rounded-lg border border-slate-200 dark:border-slate-700/50 bg-white dark:bg-slate-800/40 p-2">
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-2">
            <div className="rounded-md bg-slate-100 dark:bg-slate-900/30 px-2 py-1">
              <p className="text-slate-500 dark:text-slate-400 text-[10px] uppercase tracking-wide leading-none">Total</p>
              <p className="text-sm font-semibold text-slate-900 dark:text-white tabular-nums">{total}</p>
            </div>
            <div className="rounded-md bg-slate-100 dark:bg-slate-900/30 px-2 py-1">
              <p className="text-slate-500 dark:text-slate-400 text-[10px] uppercase tracking-wide leading-none">Questions</p>
              <p className="text-sm font-semibold text-slate-900 dark:text-white tabular-nums">{qList.length}</p>
            </div>
            <div className="rounded-md bg-slate-100 dark:bg-slate-900/30 px-2 py-1">
              <p className="text-slate-500 dark:text-slate-400 text-[10px] uppercase tracking-wide leading-none">Type</p>
              <p className="text-sm font-medium text-slate-900 dark:text-white">{form.single_response_only ? 'Single' : 'Multiple'}</p>
            </div>
            <div className="rounded-md bg-slate-100 dark:bg-slate-900/30 px-2 py-1">
              <p className="text-slate-500 dark:text-slate-400 text-[10px] uppercase tracking-wide leading-none">Exam</p>
              <p className="text-sm font-medium text-slate-900 dark:text-white">{form.exam_mode ? 'Timed' : 'Off'}</p>
            </div>
          </div>
          <div className="mt-2 rounded-md bg-slate-100 dark:bg-slate-900/30 px-2 py-1">
            <p className="text-slate-500 dark:text-slate-400 text-[10px] uppercase tracking-wide leading-none">Description</p>
            <p className="text-xs text-slate-700 dark:text-slate-200 truncate">
              {shortDescription || '—'}
            </p>
          </div>
        </div>
      </div>

      {/* Individual responses: only this area scrolls vertically */}
      <div className="flex flex-col flex-1 min-h-0 mt-2 border-t border-slate-200 dark:border-slate-700/50 pt-2">
        <h2 className="shrink-0 text-xs font-medium text-slate-800 dark:text-white mb-1.5">
          Individual responses
        </h2>
        <div className="flex-1 min-h-0 overflow-y-auto overflow-x-hidden rounded-lg border border-slate-200 dark:border-slate-700/50 bg-slate-50/50 dark:bg-slate-900/20 p-2">
          {rList.length === 0 ? (
            <p className="text-center text-slate-500 dark:text-slate-400 text-xs py-4">No individual rows.</p>
          ) : (
            <ul className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              {rList.map((r) => {
              const submitted = new Date(r.submitted_at).toLocaleString();
              const metaLabel =
                formatResponseExamMs(r.exam_duration_ms) != null
                  ? `Submitting time: ${submitted} · Exam: ${formatResponseExamMs(r.exam_duration_ms)}`
                  : `Submitting time: ${submitted}`;
              return (
                <li
                  key={r.id}
                  onClick={() => setOpenResponse(r)}
                  className="flex flex-col min-h-0 min-w-0 rounded-xl border border-slate-200 dark:border-slate-600/40 bg-white dark:bg-slate-800/60 overflow-hidden shadow-sm hover:border-violet-400/40 dark:hover:border-violet-500/30 cursor-pointer transition-colors"
                >
                  <div className="flex justify-between items-center gap-2 px-2 py-1.5 border-b border-[#c6e3ef] dark:border-slate-700/80 bg-[#eaf6fb]/95 dark:bg-slate-800/95 shrink-0">
                    <div className="flex items-center gap-1.5 min-w-0 flex-1">
                      <span className="text-base leading-none shrink-0" title="Form icon" aria-hidden>
                        {formIcon}
                      </span>
                      <span
                        className="text-[10px] text-slate-600 dark:text-slate-300 tabular-nums truncate"
                        title={metaLabel}
                      >
                        {metaLabel}
                      </span>
                    </div>
                    <button
                      type="button"
                      onClick={(e) => {
                        e.stopPropagation();
                        deleteResponse(r.id);
                      }}
                      className="shrink-0 text-red-600 dark:text-red-400 hover:bg-red-100/70 dark:hover:bg-red-900/30 rounded p-1 leading-none"
                      title="Delete response"
                      aria-label="Delete response"
                    >
                      🗑
                    </button>
                  </div>
                  <div className="flex flex-col gap-1.5 text-xs px-2 py-2 min-w-0 flex-1">
                    {previewQuestions.map((q) => {
                      const a = r.answers.find((x) => x.question_id === q.id);
                      const ansStr = answerDisplay(a?.value);
                      return (
                        <div key={q.id} className="space-y-0.5 min-w-0">
                          <div className="flex gap-1 items-start min-w-0">
                            <span
                              className="text-slate-500 dark:text-slate-400 shrink-0 truncate max-w-[42%]"
                              title={q.title}
                            >
                              {q.title}:
                            </span>
                            <span
                              className="text-slate-900 dark:text-slate-100 min-w-0 flex-1 line-clamp-2 break-words text-left"
                              title={ansStr}
                            >
                              {ansStr}
                            </span>
                          </div>
                          <QuestionPromptAttachment question={q} tone="sheet" compact />
                        </div>
                      );
                    })}
                    {qList.length > 2 && (
                      <span className="text-left text-[11px] text-violet-600 dark:text-violet-400">···</span>
                    )}
                  </div>
                </li>
              );
              })}
            </ul>
          )}
        </div>
      </div>

      {openResponse && (
        <div className="print-sheet-modal fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm">
          <div className="print-sheet-paper flex h-[min(90vh,720px)] w-[min(92vw,42rem)] flex-col overflow-hidden rounded-2xl border border-slate-200 bg-white text-slate-900 shadow-2xl dark:border-white/10 dark:bg-[#0a0a12] dark:text-slate-100">
            <div className="sheet-header flex shrink-0 items-start justify-between gap-3 border-b border-slate-200 bg-slate-50/90 px-5 py-4 dark:border-white/10 dark:bg-white/[0.04]">
              <div className="min-w-0">
                <h2 className="truncate text-base font-semibold leading-tight text-slate-900 dark:text-white">
                  {form.name}
                </h2>
                {form.description?.trim() && (
                  <p className="truncate text-xs text-slate-600 dark:text-slate-400">{form.description}</p>
                )}
                <p className="mt-1 text-xs text-slate-500 dark:text-slate-400">
                  Submitted {new Date(openResponse.submitted_at).toLocaleString()}
                </p>
                {formatResponseExamMs(openResponse.exam_duration_ms) != null && (
                  <p className="mt-0.5 text-xs text-violet-600 dark:text-violet-300">
                    Exam time: {formatResponseExamMs(openResponse.exam_duration_ms)}
                  </p>
                )}
              </div>
              <div className="sheet-actions flex shrink-0 items-center gap-1">
                <button
                  type="button"
                  onClick={() => {
                    setEmailError(null);
                    setEmailOpen(true);
                    setEmailSubject(`SurveyX response: ${form.name}`);
                  }}
                  className="rounded-lg p-2 text-slate-600 transition hover:bg-slate-200 dark:text-slate-300 dark:hover:bg-white/10"
                  title="Email"
                  aria-label="Email"
                >
                  ✉️
                </button>
                <button
                  type="button"
                  onClick={() => setOpenResponse(null)}
                  className="rounded-lg p-2 text-lg leading-none text-slate-500 transition hover:bg-slate-200 hover:text-slate-800 dark:text-slate-400 dark:hover:bg-white/10 dark:hover:text-white"
                  aria-label="Close"
                >
                  ×
                </button>
              </div>
            </div>
            <div className="sheet-body min-h-0 flex-1 overflow-y-auto px-5 py-4">
              <div className="space-y-6">
                {pageGroups.map(({ page, questions: pageQs }, pageIdx) => (
                  <section key={page.id} className="space-y-3">
                    {pages.length > 1 && pageIdx > 0 && (
                      <div className="h-px bg-slate-200 dark:bg-white/10" aria-hidden />
                    )}
                    {pages.length > 1 && page.name.trim() && (
                      <h3 className="text-sm font-medium text-slate-800 dark:text-slate-200">
                        {page.name.trim()}
                      </h3>
                    )}
                    <ol className="sheet-qa-list space-y-3 text-sm">
                      {pageQs.map((q, qIdx) => {
                        const a = openResponse.answers.find((x) => x.question_id === q.id);
                        return (
                          <li
                            key={q.id}
                            className="sheet-qa-item rounded-xl border border-slate-200 bg-slate-50/80 p-4 dark:border-white/[0.07] dark:bg-white/[0.03]"
                          >
                            <p className="font-medium text-slate-900 dark:text-slate-100">
                              <span className="mr-1.5 text-slate-400 dark:text-slate-500 tabular-nums">
                                {qIdx + 1}.
                              </span>
                              {q.title}
                            </p>
                            <QuestionPromptAttachment question={q} tone="sheet" />
                            <p className="mt-2 wrap-break-word text-slate-700 dark:text-slate-300">
                              {answerDisplay(a?.value)}
                            </p>
                          </li>
                        );
                      })}
                    </ol>
                  </section>
                ))}
              </div>
            </div>
          </div>
        </div>
      )}

      {openResponse && emailOpen && (
        <div className="fixed inset-0 z-60 flex items-center justify-center p-4 bg-black/50">
          <div className="w-full max-w-lg rounded-xl border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 overflow-hidden">
            <div className="px-4 py-3 border-b border-slate-200 dark:border-slate-700 flex items-center justify-between">
              <h3 className="font-semibold text-sm">Email response</h3>
              <button type="button" onClick={() => setEmailOpen(false)} className="text-slate-500 dark:text-slate-300 hover:text-slate-800 dark:hover:text-white text-lg leading-none">
                ×
              </button>
            </div>
            <div className="p-4 space-y-3">
              {emailError && (
                <div className="p-2 rounded-md bg-red-100 dark:bg-red-900/30 text-red-800 dark:text-red-200 text-xs">
                  {emailError}
                </div>
              )}
              <div>
                <label className="block text-xs text-slate-600 dark:text-slate-300 mb-1">To (comma/newline separated)</label>
                <textarea
                  value={emailTo}
                  onChange={(e) => setEmailTo(e.target.value)}
                  rows={2}
                  placeholder="a@example.com, b@example.com"
                  className="w-full px-3 py-2 rounded-lg bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-600 text-sm"
                />
              </div>
              <div>
                <label className="block text-xs text-slate-600 dark:text-slate-300 mb-1">Subject</label>
                <input
                  value={emailSubject}
                  onChange={(e) => setEmailSubject(e.target.value)}
                  className="w-full px-3 py-2 rounded-lg bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-600 text-sm"
                />
              </div>
              <p className="text-[11px] text-slate-500 dark:text-slate-400">
                This sends the currently open response as an HTML sheet (requires SMTP configured on the server).
              </p>
            </div>
            <div className="px-4 py-3 border-t border-slate-200 dark:border-slate-700 flex justify-end gap-2">
              <button
                type="button"
                onClick={() => setEmailOpen(false)}
                className="px-3 py-2 rounded-lg bg-slate-200 dark:bg-slate-700 text-slate-800 dark:text-slate-100 text-sm"
              >
                Cancel
              </button>
              <button
                type="button"
                onClick={submitEmail}
                disabled={emailSending}
                className="px-4 py-2 rounded-lg bg-violet-600 hover:bg-violet-500 text-white text-sm disabled:opacity-50"
              >
                {emailSending ? 'Sending…' : 'Send email'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

function iconForForm(formId: number): string {
  const icons = ['📄', '🧾', '📝', '📋', '📌', '📎', '🗂️', '🧩'];
  return icons[Math.abs(formId) % icons.length];
}

function orderPages(pages: FormPage[]): FormPage[] {
  return [...pages].sort((a, b) => a.sort_order - b.sort_order || a.id - b.id);
}

function groupQuestionsByPage(
  questions: Question[],
  pages: FormPage[]
): { page: FormPage; questions: Question[] }[] {
  if (pages.length === 0) {
    return questions.length > 0
      ? [{ page: { id: 0, form_id: 0, name: '', sort_order: 0, created_at: '', updated_at: '' }, questions }]
      : [];
  }
  return orderPages(pages)
    .map((page) => ({
      page,
      questions: questions.filter((q) => q.page_id === page.id),
    }))
    .filter((g) => g.questions.length > 0);
}

function answerDisplay(value: unknown): string {
  if (value == null) return '—';
  if (typeof value === 'string') return value;
  if (typeof value === 'number') return String(value);
  if (typeof value === 'boolean') return value ? 'Yes' : 'No';
  if (Array.isArray(value)) return value.map(String).join(', ');
  if (typeof value === 'object' && value !== null && 'value' in value) return String((value as { value: unknown }).value);
  return String(value);
}
