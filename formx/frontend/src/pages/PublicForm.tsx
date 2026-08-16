import { useMemo, useEffect, useState, useCallback, useRef, type ClipboardEvent, type ChangeEvent } from 'react';
import { useParams } from 'react-router-dom';
import { normalizeDefaultForQuestion } from '../lib/defaultValue';
import { QuestionPromptAttachment } from '../components/QuestionPromptAttachment';
import { api, type Form, type FormPage, type Question, type QuestionRule } from '../lib/api';

const PUBLIC_THEME_KEY = 'sheetx-public-theme';
const LEGACY_PUBLIC_THEME_KEY = 'formsx-public-theme';

function useRespondentTheme() {
  const [theme, setTheme] = useState<'light' | 'dark'>(() => {
    try {
      const saved =
        localStorage.getItem(PUBLIC_THEME_KEY) ?? localStorage.getItem(LEGACY_PUBLIC_THEME_KEY);
      return saved === 'light' ? 'light' : 'dark';
    } catch {
      return 'dark';
    }
  });
  useEffect(() => {
    try {
      localStorage.setItem(PUBLIC_THEME_KEY, theme);
    } catch {
      /* ignore */
    }
    document.documentElement.classList.toggle('dark', theme === 'dark');
  }, [theme]);
  useEffect(() => {
    return () => {
      document.documentElement.classList.remove('dark');
    };
  }, []);
  return {
    isDark: theme === 'dark',
    toggle: () => setTheme((t) => (t === 'dark' ? 'light' : 'dark')),
  };
}

/** Stable per-browser ID so single-response forms can enforce one submission. */
function getOrCreateRespondentId(slug: string): string {
  const key = `sheetx_respondent_${slug}`;
  const legacyKey = `formsx_respondent_${slug}`;
  try {
    let id = localStorage.getItem(key) ?? localStorage.getItem(legacyKey);
    if (!id) {
      id = crypto.randomUUID();
    }
    localStorage.setItem(key, id);
    return id;
  } catch {
    return `anon-${slug}-${Date.now()}`;
  }
}

function isAnswerFilled(value: unknown): boolean {
  if (value === undefined || value === null) return false;
  if (typeof value === 'string') return value.trim() !== '';
  if (typeof value === 'boolean') return true;
  if (typeof value === 'number') return true;
  if (Array.isArray(value)) return value.length > 0;
  return true;
}

function visibleQuestions(
  questions: Question[],
  rules: QuestionRule[],
  answers: Record<number, unknown>
): Question[] {
  const rulesByQuestion = new Map<number, QuestionRule[]>();
  for (const r of rules) {
    const list = rulesByQuestion.get(r.question_id) ?? [];
    list.push(r);
    rulesByQuestion.set(r.question_id, list);
  }
  return questions.filter((q) => {
    const ruleList = rulesByQuestion.get(q.id);
    if (!ruleList || ruleList.length === 0) return true;
    return ruleList.every((r) => {
      const filled = isAnswerFilled(answers[r.depends_on_question_id]);
      return r.condition === 'answered' ? filled : !filled;
    });
  });
}

function examStorageKey(slug: string): string {
  return `sheetx_exam_started_${slug}`;
}

function formatExamElapsed(ms: number): string {
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

function ThemeToggle({ isDark, onToggle }: { isDark: boolean; onToggle: () => void }) {
  return (
    <button
      type="button"
      onClick={onToggle}
      aria-label={isDark ? 'Switch to light mode' : 'Switch to dark mode'}
      className={
        isDark
          ? 'fixed top-4 right-4 z-20 flex h-10 w-10 items-center justify-center rounded-full border border-white/10 bg-white/5 text-lg backdrop-blur-md transition hover:border-violet-400/40 hover:bg-white/10'
          : 'fixed top-4 right-4 z-20 flex h-10 w-10 items-center justify-center rounded-full border border-slate-200 bg-white/80 text-lg shadow-sm backdrop-blur-md transition hover:border-violet-300 hover:shadow-md'
      }
    >
      {isDark ? '☀' : '☾'}
    </button>
  );
}

function PublicFormShell({
  isDark,
  onToggleTheme,
  children,
}: {
  isDark: boolean;
  onToggleTheme: () => void;
  children: React.ReactNode;
}) {
  return (
    <div
      className={
        isDark
          ? 'relative min-h-dvh bg-[#06060b] text-slate-100'
          : 'relative min-h-dvh bg-gradient-to-b from-slate-50 via-white to-violet-50/30 text-slate-900'
      }
    >
      {isDark && (
        <>
          <div className="pointer-events-none fixed inset-0 bg-[radial-gradient(ellipse_90%_60%_at_50%_-10%,rgba(139,92,246,0.22),transparent)]" />
          <div className="pointer-events-none fixed inset-0 bg-[radial-gradient(ellipse_50%_40%_at_100%_80%,rgba(34,211,238,0.07),transparent)]" />
          <div className="pointer-events-none fixed inset-0 bg-[radial-gradient(ellipse_40%_30%_at_0%_60%,rgba(217,70,239,0.06),transparent)]" />
        </>
      )}
      <ThemeToggle isDark={isDark} onToggle={onToggleTheme} />
      <div className="relative mx-auto max-w-xl px-4 py-12 pb-24 sm:py-14 sm:pb-28">{children}</div>
    </div>
  );
}

function cardClass(isDark: boolean) {
  return isDark
    ? 'rounded-2xl border border-white/[0.08] bg-white/[0.04] p-6 shadow-[0_8px_32px_rgba(0,0,0,0.4)] backdrop-blur-xl'
    : 'rounded-2xl border border-slate-200/80 bg-white/90 p-6 shadow-[0_8px_30px_rgba(15,23,42,0.06)] backdrop-blur-sm';
}

function primaryBtnClass(isDark: boolean) {
  const base =
    'flex-1 rounded-xl py-3.5 text-sm font-semibold tracking-wide transition-all duration-200 disabled:cursor-not-allowed disabled:opacity-50';
  if (isDark) {
    return `${base} bg-gradient-to-r from-violet-600 via-fuchsia-600 to-violet-600 bg-[length:200%_auto] text-white shadow-[0_4px_24px_rgba(139,92,246,0.35)] hover:bg-right hover:shadow-[0_6px_28px_rgba(168,85,247,0.45)]`;
  }
  return `${base} bg-gradient-to-r from-violet-600 to-fuchsia-600 text-white shadow-lg shadow-violet-500/25 hover:shadow-violet-500/40 hover:brightness-105`;
}

function secondaryBtnClass(isDark: boolean) {
  return isDark
    ? 'flex-1 rounded-xl border border-white/15 bg-white/5 py-3.5 text-sm font-medium text-slate-200 backdrop-blur-sm transition hover:border-white/25 hover:bg-white/10'
    : 'flex-1 rounded-xl border border-slate-200 bg-white py-3.5 text-sm font-medium text-slate-700 shadow-sm transition hover:border-slate-300 hover:bg-slate-50';
}

export function PublicForm() {
  const { slug } = useParams<{ slug: string }>();
  const { isDark, toggle: toggleTheme } = useRespondentTheme();
  const [form, setForm] = useState<Form | null>(null);
  const [pages, setPages] = useState<FormPage[]>([]);
  const [questions, setQuestions] = useState<Question[]>([]);
  const [rules, setRules] = useState<QuestionRule[]>([]);
  const [currentPageIndex, setCurrentPageIndex] = useState(0);
  const [pageError, setPageError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [answers, setAnswers] = useState<Record<number, unknown>>({});
  const [submitStatus, setSubmitStatus] = useState<'idle' | 'sending' | 'done' | 'error'>('idle');
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [examPhase, setExamPhase] = useState<'prestart' | 'active'>('active');
  const [examElapsedMs, setExamElapsedMs] = useState(0);
  const [recordedExamMs, setRecordedExamMs] = useState<number | null>(null);

  useEffect(() => {
    document.documentElement.classList.add('public-form-page');
    return () => document.documentElement.classList.remove('public-form-page');
  }, []);

  useEffect(() => {
    if (!slug) return;
    api.public
      .getForm(slug)
      .then(({ form: f, pages: p, questions: q, rules: r }) => {
        setForm(f);
        setPages(p ?? []);
        setQuestions(q);
        setRules(r ?? []);
        setCurrentPageIndex(0);
        setPageError(null);
        const initial: Record<number, unknown> = {};
        for (const qq of q) {
          const d = normalizeDefaultForQuestion(qq);
          if (d !== undefined) initial[qq.id] = d;
        }
        setAnswers(initial);
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, [slug]);

  useEffect(() => {
    setRecordedExamMs(null);
    setExamElapsedMs(0);
    if (!form?.exam_mode || !slug) {
      setExamPhase('active');
      return;
    }
    try {
      const raw = sessionStorage.getItem(examStorageKey(slug));
      setExamPhase(raw != null && !Number.isNaN(Date.parse(raw)) ? 'active' : 'prestart');
    } catch {
      setExamPhase('prestart');
    }
  }, [form?.exam_mode, slug]);

  useEffect(() => {
    if (!form?.exam_mode || examPhase !== 'active' || !slug) {
      setExamElapsedMs(0);
      return;
    }
    let startMs = 0;
    try {
      const raw = sessionStorage.getItem(examStorageKey(slug));
      if (!raw || Number.isNaN(Date.parse(raw))) {
        setExamElapsedMs(0);
        return;
      }
      startMs = Date.parse(raw);
    } catch {
      setExamElapsedMs(0);
      return;
    }
    const tick = () => setExamElapsedMs(Math.max(0, Date.now() - startMs));
    tick();
    const id = window.setInterval(tick, 1000);
    return () => clearInterval(id);
  }, [form?.exam_mode, examPhase, slug]);

  const visible = useMemo(
    () => visibleQuestions(questions, rules, answers),
    [questions, rules, answers]
  );

  const orderedPages = useMemo(() => {
    if (pages.length === 0) {
      return [{ id: 0, form_id: 0, name: '', sort_order: 0, created_at: '', updated_at: '' }];
    }
    return [...pages].sort((a, b) => a.sort_order - b.sort_order || a.id - b.id);
  }, [pages]);

  const isMultiPage = pages.length > 1;
  const currentPage = orderedPages[currentPageIndex] ?? orderedPages[0];
  const isLastPage = !isMultiPage || currentPageIndex >= orderedPages.length - 1;

  const pageQuestions = useMemo(() => {
    if (pages.length === 0) return visible;
    return visible.filter((q) => q.page_id === currentPage?.id);
  }, [visible, pages.length, currentPage?.id]);

  const validatePage = useCallback(
    (pageId: number): string | null => {
      const qs = visible.filter((q) => q.page_id === pageId);
      for (const q of qs) {
        if (q.required && !isAnswerFilled(answers[q.id])) {
          return `"${q.title}" is required before continuing.`;
        }
      }
      return null;
    },
    [visible, answers]
  );

  const goNextPage = useCallback(() => {
    if (!currentPage || isLastPage) return;
    const err = validatePage(currentPage.id);
    if (err) {
      setPageError(err);
      return;
    }
    setPageError(null);
    setSubmitError(null);
    setCurrentPageIndex((i) => Math.min(i + 1, orderedPages.length - 1));
  }, [currentPage, isLastPage, validatePage, orderedPages.length]);

  const goPrevPage = useCallback(() => {
    setPageError(null);
    setCurrentPageIndex((i) => Math.max(i - 1, 0));
  }, []);

  const setAnswer = (questionId: number, value: unknown) => {
    setAnswers((prev) => ({ ...prev, [questionId]: value }));
    setPageError(null);
  };

  const handleSubmit = async () => {
    if (!slug || !form) return;

    if (isMultiPage) {
      for (const page of orderedPages) {
        const err = validatePage(page.id);
        if (err) {
          setSubmitError(err);
          const idx = orderedPages.findIndex((p) => p.id === page.id);
          if (idx >= 0) setCurrentPageIndex(idx);
          return;
        }
      }
    } else if (currentPage) {
      const err = validatePage(currentPage.id);
      if (err) {
        setSubmitError(err);
        return;
      }
    }

    setSubmitStatus('sending');
    setSubmitError(null);
    setPageError(null);
    const visibleIds = new Set(visible.map((q) => q.id));
    let examStartedAtISO: string | undefined;
    if (form.exam_mode) {
      try {
        const raw = sessionStorage.getItem(examStorageKey(slug));
        if (!raw || Number.isNaN(Date.parse(raw))) {
          setSubmitStatus('error');
          setSubmitError('Exam session expired or missing. Refresh the page and press Start.');
          return;
        }
        examStartedAtISO = new Date(Date.parse(raw)).toISOString();
      } catch {
        setSubmitStatus('error');
        setSubmitError('Could not read exam timer. Refresh the page.');
        return;
      }
    }
    try {
      const resp = await api.public.submit(slug, {
        respondent_id: form.single_response_only ? getOrCreateRespondentId(slug) : undefined,
        ...(examStartedAtISO ? { exam_started_at: examStartedAtISO } : {}),
        answers: Object.entries(answers)
          .filter(([qid]) => visibleIds.has(parseInt(qid, 10)))
          .map(([qid, value]) => ({
            question_id: parseInt(qid, 10),
            value,
          })),
      });
      if (form.exam_mode) {
        try {
          sessionStorage.removeItem(examStorageKey(slug));
        } catch {
          /* ignore */
        }
      }
      setRecordedExamMs(resp.exam_duration_ms ?? null);
      setSubmitStatus('done');
    } catch (e) {
      setSubmitStatus('error');
      setSubmitError(e instanceof Error ? e.message : 'Submit failed');
    }
  };

  if (loading) {
    return (
      <PublicFormShell isDark={isDark} onToggleTheme={toggleTheme}>
        <div className={`text-center text-sm ${isDark ? 'text-slate-500' : 'text-slate-400'}`}>
          Loading form…
        </div>
      </PublicFormShell>
    );
  }

  if (error || !form) {
    return (
      <PublicFormShell isDark={isDark} onToggleTheme={toggleTheme}>
        <div className={`text-center text-sm ${isDark ? 'text-red-400' : 'text-red-600'}`}>
          {error || 'Form not found'}
        </div>
      </PublicFormShell>
    );
  }

  if (slug && form.exam_mode && examPhase === 'prestart') {
    return (
      <PublicFormShell isDark={isDark} onToggleTheme={toggleTheme}>
        <div className={`${cardClass(isDark)} mb-5`}>
          <h1 className={`text-xl font-semibold tracking-tight ${isDark ? 'text-white' : 'text-slate-900'}`}>
            {form.name}
          </h1>
          {form.description && (
            <p className={`mt-2 text-sm leading-relaxed ${isDark ? 'text-slate-400' : 'text-slate-600'}`}>
              {form.description}
            </p>
          )}
        </div>
        <div className={`${cardClass(isDark)} space-y-5`}>
          <p className={`text-sm leading-relaxed ${isDark ? 'text-slate-300' : 'text-slate-600'}`}>
            This form uses{' '}
            <span className={isDark ? 'font-medium text-violet-300' : 'font-medium text-violet-600'}>
              Exam Mode
            </span>
            . When you are ready, tap Start — the timer begins and stops when you submit.
          </p>
          <button
            type="button"
            className={primaryBtnClass(isDark)}
            onClick={() => {
              sessionStorage.setItem(examStorageKey(slug), new Date().toISOString());
              setExamPhase('active');
            }}
          >
            Start exam
          </button>
        </div>
      </PublicFormShell>
    );
  }

  if (submitStatus === 'done') {
    return (
      <PublicFormShell isDark={isDark} onToggleTheme={toggleTheme}>
        <div className={`${cardClass(isDark)} text-center`}>
          <div
            className={`mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-full text-2xl ${
              isDark
                ? 'bg-gradient-to-br from-violet-500/30 to-fuchsia-500/20 text-violet-300'
                : 'bg-gradient-to-br from-violet-100 to-fuchsia-100 text-violet-600'
            }`}
          >
            ✓
          </div>
          <p className={`text-xl font-semibold ${isDark ? 'text-violet-300' : 'text-violet-700'}`}>Thank you!</p>
          <p className={`mt-2 text-sm ${isDark ? 'text-slate-400' : 'text-slate-600'}`}>
            Your response has been submitted.
          </p>
          {recordedExamMs != null && recordedExamMs > 0 && (
            <p className={`mt-3 text-sm ${isDark ? 'text-slate-400' : 'text-slate-600'}`}>
              Recorded exam time:{' '}
              <span className={`font-medium tabular-nums ${isDark ? 'text-white' : 'text-slate-900'}`}>
                {formatExamElapsed(recordedExamMs)}
              </span>
            </p>
          )}
        </div>
      </PublicFormShell>
    );
  }

  const showExamTimer = !!(slug && form.exam_mode && examPhase === 'active');

  return (
    <PublicFormShell isDark={isDark} onToggleTheme={toggleTheme}>
      <div className={`${cardClass(isDark)} mb-5`}>
        <h1 className={`text-xl font-semibold tracking-tight ${isDark ? 'text-white' : 'text-slate-900'}`}>
          {form.name}
        </h1>
        {form.description && (
          <p className={`mt-2 text-sm leading-relaxed ${isDark ? 'text-slate-400' : 'text-slate-600'}`}>
            {form.description}
          </p>
        )}
      </div>

      {showExamTimer && (
        <div
          className={
            isDark
              ? 'mb-5 flex items-center justify-between gap-3 rounded-xl border border-violet-500/25 bg-violet-500/10 px-4 py-3 backdrop-blur-sm'
              : 'mb-5 flex items-center justify-between gap-3 rounded-xl border border-violet-200 bg-violet-50 px-4 py-3'
          }
        >
          <span className={`text-sm ${isDark ? 'text-slate-300' : 'text-slate-600'}`}>Exam in progress</span>
          <span
            className={`font-mono text-sm font-semibold tabular-nums ${
              isDark ? 'text-fuchsia-300' : 'text-violet-700'
            }`}
          >
            {formatExamElapsed(examElapsedMs)}
          </span>
        </div>
      )}

      {isMultiPage && (
        <div className="mb-5 space-y-2">
          <div className="flex items-center justify-between text-xs">
            <span className={isDark ? 'text-slate-500' : 'text-slate-500'}>
              Step {currentPageIndex + 1} of {orderedPages.length}
            </span>
          </div>
          <div className="flex gap-1.5">
            {orderedPages.map((_, i) => (
              <div
                key={i}
                className={`h-1 flex-1 rounded-full transition-all duration-300 ${
                  i <= currentPageIndex
                    ? isDark
                      ? 'bg-gradient-to-r from-violet-500 to-fuchsia-500 shadow-[0_0_12px_rgba(168,85,247,0.5)]'
                      : 'bg-gradient-to-r from-violet-500 to-fuchsia-500'
                    : isDark
                      ? 'bg-white/10'
                      : 'bg-slate-200'
                }`}
              />
            ))}
          </div>
        </div>
      )}

      {/* No <form> — prevents Enter key from submitting on intermediate pages */}
      <div className="space-y-4">
        {currentPage?.name.trim() && (
          <h2 className={`text-lg font-medium tracking-tight ${isDark ? 'text-white' : 'text-slate-900'}`}>
            {currentPage.name.trim()}
          </h2>
        )}

        {pageQuestions.map((q) => (
          <div
            key={q.id}
            className={
              isDark
                ? 'rounded-2xl border border-white/[0.07] bg-white/[0.03] p-5 backdrop-blur-sm transition hover:border-white/10'
                : 'rounded-2xl border border-slate-200/80 bg-white p-5 shadow-sm'
            }
          >
            <label className={`mb-3 block font-medium ${isDark ? 'text-slate-100' : 'text-slate-800'}`}>
              {q.title} {q.required && <span className="text-red-400">*</span>}
            </label>
            <QuestionPromptAttachment question={q} tone="public" />
            <QuestionInput
              question={q}
              value={answers[q.id]}
              isDark={isDark}
              onChange={(v) => setAnswer(q.id, v)}
            />
          </div>
        ))}

        {pageQuestions.length === 0 && isMultiPage && (
          <p className={`text-center text-sm ${isDark ? 'text-slate-500' : 'text-slate-400'}`}>
            No questions on this page.
          </p>
        )}

        {(pageError || (submitStatus === 'error' && submitError)) && (
          <div
            className={
              isDark
                ? 'rounded-xl border border-red-500/30 bg-red-950/40 px-4 py-3 text-sm text-red-300'
                : 'rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700'
            }
          >
            {pageError || submitError}
          </div>
        )}

        <div className="flex gap-3 pt-2">
          {isMultiPage && currentPageIndex > 0 && (
            <button type="button" onClick={goPrevPage} className={secondaryBtnClass(isDark)}>
              Back
            </button>
          )}
          {isMultiPage && !isLastPage ? (
            <button type="button" onClick={goNextPage} className={primaryBtnClass(isDark)}>
              Continue
            </button>
          ) : (
            <button
              type="button"
              disabled={submitStatus === 'sending'}
              onClick={() => void handleSubmit()}
              className={primaryBtnClass(isDark)}
            >
              {submitStatus === 'sending' ? 'Submitting…' : 'Submit'}
            </button>
          )}
        </div>
      </div>
    </PublicFormShell>
  );
}

function QuestionInput({
  question,
  value,
  isDark,
  onChange,
}: {
  question: Question;
  value: unknown;
  isDark: boolean;
  onChange: (v: unknown) => void;
}) {
  const inputClass = isDark
    ? 'w-full rounded-xl border border-white/10 bg-black/30 px-3.5 py-2.5 text-white placeholder-slate-500 transition focus:border-violet-500/50 focus:outline-none focus:ring-2 focus:ring-violet-500/20'
    : 'w-full rounded-xl border border-slate-200 bg-slate-50 px-3.5 py-2.5 text-slate-900 placeholder-slate-400 transition focus:border-violet-400 focus:bg-white focus:outline-none focus:ring-2 focus:ring-violet-500/15';

  const opts = question.config?.options || [];
  switch (question.type) {
    case 'text':
      return (
        <input
          type="text"
          value={(value as string) || ''}
          onChange={(e) => onChange(e.target.value)}
          className={inputClass}
        />
      );
    case 'integer':
      return (
        <input
          type="number"
          step={1}
          value={value === undefined || value === null ? '' : String(value)}
          onChange={(e) => {
            const v = e.target.value;
            if (v === '') {
              onChange(undefined);
              return;
            }
            const n = parseInt(v, 10);
            if (!Number.isNaN(n)) onChange(n);
          }}
          className={inputClass}
        />
      );
    case 'select':
      return (
        <select
          value={value === undefined || value === null ? '' : String(value)}
          onChange={(e) => {
            const v = e.target.value;
            onChange(v === '' ? undefined : parseInt(v, 10));
          }}
          className={inputClass}
        >
          <option value="">Select…</option>
          {opts.map((o) => (
            <option key={o.value} value={o.value}>
              {o.label}
            </option>
          ))}
        </select>
      );
    case 'multiselect': {
      const selected = Array.isArray(value) ? (value as number[]) : [];
      const maxSel = question.config?.max_selections;
      return (
        <div className="space-y-2.5">
          {opts.map((o) => (
            <label key={o.value} className="flex cursor-pointer items-center gap-2.5">
              <input
                type="checkbox"
                className="rounded border-slate-500 bg-transparent text-violet-500 focus:ring-violet-500/30"
                checked={selected.includes(o.value)}
                onChange={() => {
                  if (selected.includes(o.value)) {
                    onChange(selected.filter((x) => x !== o.value));
                  } else {
                    const next = [...selected, o.value];
                    if (maxSel && next.length > maxSel) return;
                    onChange(next);
                  }
                }}
              />
              <span className={isDark ? 'text-slate-300' : 'text-slate-700'}>{o.label}</span>
            </label>
          ))}
          {opts.length === 0 && (
            <p className={`text-sm ${isDark ? 'text-slate-500' : 'text-slate-400'}`}>
              No options configured for this question.
            </p>
          )}
        </div>
      );
    }
    case 'boolean':
      return (
        <label className="flex cursor-pointer items-center gap-2.5">
          <input
            type="checkbox"
            checked={Boolean(value)}
            onChange={(e) => onChange(e.target.checked)}
            className="rounded border-slate-500 bg-transparent text-violet-500 focus:ring-violet-500/30"
          />
          <span className={isDark ? 'text-slate-400' : 'text-slate-600'}>Yes</span>
        </label>
      );
    case 'date':
      return (
        <input
          type="date"
          value={(value as string) || ''}
          onChange={(e) => onChange(e.target.value)}
          className={inputClass}
        />
      );
    case 'datetime':
      return (
        <input
          type="datetime-local"
          value={(value as string) || ''}
          onChange={(e) => onChange(e.target.value)}
          className={inputClass}
        />
      );
    case 'float':
      return (
        <input
          type="number"
          step="any"
          value={(value as number) ?? ''}
          onChange={(e) => onChange(e.target.value ? parseFloat(e.target.value) : null)}
          className={inputClass}
        />
      );
    case 'qrcode':
      return (
        <input
          type="text"
          value={(value as string) || ''}
          onChange={(e) => onChange(e.target.value)}
          placeholder="Scan or enter code"
          className={inputClass}
        />
      );
    case 'image':
      return (
        <ImageAnswerInput value={value} isDark={isDark} inputClass={inputClass} onChange={onChange} />
      );
    case 'document':
      return (
        <FileAnswerInput value={value} isDark={isDark} inputClass={inputClass} onChange={onChange} />
      );
    default:
      return (
        <input
          type="text"
          value={(value as string) || ''}
          onChange={(e) => onChange(e.target.value)}
          className={inputClass}
        />
      );
  }
}

function readFileAsDataURL(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(String(reader.result || ''));
    reader.onerror = () => reject(reader.error ?? new Error('Failed to read file'));
    reader.readAsDataURL(file);
  });
}

function ImageAnswerInput({
  value,
  isDark,
  inputClass,
  onChange,
}: {
  value: unknown;
  isDark: boolean;
  inputClass: string;
  onChange: (v: unknown) => void;
}) {
  const preview = typeof value === 'string' && value.startsWith('data:image') ? value : '';
  const urlValue = typeof value === 'string' && !value.startsWith('data:') ? value : '';
  const pasteRef = useRef<HTMLDivElement>(null);

  const applyFile = async (file: File | null | undefined) => {
    if (!file || !file.type.startsWith('image/')) return;
    const dataUrl = await readFileAsDataURL(file);
    onChange(dataUrl);
  };

  const onPaste = async (e: ClipboardEvent<HTMLDivElement>) => {
    const items = e.clipboardData?.items;
    if (!items) return;
    for (const item of Array.from(items)) {
      if (item.type.startsWith('image/')) {
        e.preventDefault();
        await applyFile(item.getAsFile());
        return;
      }
    }
  };

  const onFileChange = async (e: ChangeEvent<HTMLInputElement>) => {
    await applyFile(e.target.files?.[0]);
    e.target.value = '';
  };

  return (
    <div className="space-y-2">
      <div
        ref={pasteRef}
        tabIndex={0}
        onPaste={onPaste}
        className={
          'rounded-xl border border-dashed px-3 py-3 text-sm outline-none focus:ring-2 focus:ring-violet-500/25 ' +
          (isDark
            ? 'border-white/15 bg-black/20 text-slate-300'
            : 'border-slate-300 bg-slate-50 text-slate-600')
        }
      >
        Paste an image here (Ctrl/Cmd+V), or choose a file below.
      </div>
      <input type="file" accept="image/*" onChange={onFileChange} className="block w-full text-sm" />
      <input
        type="text"
        value={urlValue}
        onChange={(e) => onChange(e.target.value)}
        placeholder="Or paste an image URL"
        className={inputClass}
      />
      {preview ? (
        <img src={preview} alt="Uploaded preview" className="max-h-48 rounded-lg border border-white/10 object-contain" />
      ) : null}
    </div>
  );
}

function FileAnswerInput({
  value,
  isDark,
  inputClass,
  onChange,
}: {
  value: unknown;
  isDark: boolean;
  inputClass: string;
  onChange: (v: unknown) => void;
}) {
  const label =
    typeof value === 'object' && value && 'filename' in (value as object)
      ? String((value as { filename?: string }).filename || 'File attached')
      : typeof value === 'string' && value
        ? value.startsWith('data:')
          ? 'File attached'
          : value
        : '';

  const onFileChange = async (e: ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    e.target.value = '';
    if (!file) return;
    const dataUrl = await readFileAsDataURL(file);
    onChange({ url: dataUrl, filename: file.name, size: file.size });
  };

  return (
    <div className="space-y-2">
      <input type="file" onChange={onFileChange} className="block w-full text-sm" />
      <input
        type="text"
        value={typeof value === 'string' && !value.startsWith('data:') ? value : ''}
        onChange={(e) => onChange(e.target.value)}
        placeholder="Or paste a file URL"
        className={inputClass}
      />
      {label ? (
        <p className={`text-xs ${isDark ? 'text-slate-400' : 'text-slate-500'}`}>{label}</p>
      ) : null}
    </div>
  );
}
