import { useCallback, useEffect, useRef, useState, type ReactNode } from 'react';
import { api, type SurveyBotResult, type SurveyBotTemplate } from '../lib/api';
import { useConfirm } from '../context/ConfirmContext';

type Tab = 'surveys' | 'answers';
type EnlargePanel = 'survey' | 'answer' | null;

type SourcePreview = {
  name: string;
  kind: 'pdf' | 'image';
  text: string;
  truncated: boolean;
};

const DESCRIPTION_STARTER = `---
slug: weekend-parties
title: Weekend party preference survey
tags: [custom, ai-survey]
---

This is a preference survey for weekend parties. We want to know what staff wants to do for every weekend.
`;

export function SurveyBot() {
  const { confirm } = useConfirm();
  const fileRef = useRef<HTMLInputElement>(null);
  const [tab, setTab] = useState<Tab>('surveys');
  const [templates, setTemplates] = useState<SurveyBotTemplate[]>([]);
  const [results, setResults] = useState<SurveyBotResult[]>([]);
  const [resultsTotal, setResultsTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [refreshingAnswers, setRefreshingAnswers] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selectedResult, setSelectedResult] = useState<SurveyBotResult | null>(null);
  const [editing, setEditing] = useState<SurveyBotTemplate | null>(null);
  const [draftMd, setDraftMd] = useState('');
  const [draftTitle, setDraftTitle] = useState('');
  const [aiQuery, setAiQuery] = useState('');
  const [aiBusy, setAiBusy] = useState(false);
  const [fileBusy, setFileBusy] = useState(false);
  const [sourcePreview, setSourcePreview] = useState<SourcePreview | null>(null);
  const [compileBusy, setCompileBusy] = useState(false);
  const [publishBusy, setPublishBusy] = useState(false);
  const [saving, setSaving] = useState(false);
  const [shareUrl, setShareUrl] = useState('');
  const [answersQuery, setAnswersQuery] = useState('');
  const [answersSearch, setAnswersSearch] = useState('');
  const [enlarge, setEnlarge] = useState<EnlargePanel>(null);

  useEffect(() => {
    const t = window.setTimeout(() => setAnswersSearch(answersQuery.trim()), 250);
    return () => window.clearTimeout(t);
  }, [answersQuery]);

  useEffect(() => {
    if (!enlarge) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setEnlarge(null);
    };
    window.addEventListener('keydown', onKey);
    const prev = document.body.style.overflow;
    document.body.style.overflow = 'hidden';
    return () => {
      window.removeEventListener('keydown', onKey);
      document.body.style.overflow = prev;
    };
  }, [enlarge]);

  useEffect(() => {
    setEnlarge(null);
  }, [tab]);

  const load = useCallback(async (opts?: { answersOnly?: boolean; search?: string }) => {
    const search = opts?.search ?? answersSearch;
    if (opts?.answersOnly) {
      setRefreshingAnswers(true);
    } else {
      setLoading(true);
    }
    setError(null);
    try {
      if (opts?.answersOnly) {
        const r = await api.surveyBot.listResults(1, 200, search);
        setResults(r.results ?? []);
        setResultsTotal(r.total ?? (r.results ?? []).length);
      } else {
        const [t, r] = await Promise.all([
          api.surveyBot.listTemplates(1, 100),
          api.surveyBot.listResults(1, 200, search),
        ]);
        setTemplates(t.templates ?? []);
        setResults(r.results ?? []);
        setResultsTotal(r.total ?? (r.results ?? []).length);
      }
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load AI Surveys data');
    } finally {
      setLoading(false);
      setRefreshingAnswers(false);
    }
  }, [answersSearch]);

  useEffect(() => {
    void load();
  }, [load]);

  const publicUrlFor = (t: SurveyBotTemplate) => {
    if (t.public_url) return t.public_url;
    return `${window.location.origin}/s/${t.slug}`;
  };

  const openNewTemplate = () => {
    setEditing(null);
    setDraftTitle('Weekend party preference survey');
    setDraftMd(DESCRIPTION_STARTER);
    setShareUrl('');
  };

  const openEdit = (t: SurveyBotTemplate) => {
    setEditing(t);
    setDraftTitle(t.title);
    setDraftMd(t.markdown);
    setShareUrl(t.published ? publicUrlFor(t) : '');
    setSourcePreview(null);
  };

  const clearSurveyDraft = () => {
    setDraftMd('');
    setEditing(null);
    setShareUrl('');
    setEnlarge(null);
    setSourcePreview(null);
  };

  const saveTemplate = async () => {
    setSaving(true);
    setError(null);
    try {
      if (editing?.id) {
        const updated = await api.surveyBot.updateTemplate(editing.id, {
          title: draftTitle || editing.title,
          markdown: draftMd,
        });
        setEditing(updated);
      } else {
        const created = await api.surveyBot.createTemplate({
          title: draftTitle || undefined,
          markdown: draftMd,
        });
        setEditing(created);
      }
      await load();
      setTab('surveys');
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Save failed');
    } finally {
      setSaving(false);
    }
  };

  const runAiDraft = async () => {
    if (!aiQuery.trim()) return;
    setAiBusy(true);
    setError(null);
    try {
      const out = await api.surveyBot.aiDraft({ query: aiQuery.trim(), title_hint: aiQuery.trim() });
      setDraftMd(out.markdown);
      setDraftTitle(aiQuery.trim());
      setEditing(null);
      setShareUrl('');
    } catch (e) {
      setError(e instanceof Error ? e.message : 'AI draft failed');
    } finally {
      setAiBusy(false);
    }
  };

  const compileQuestions = async () => {
    setCompileBusy(true);
    setError(null);
    try {
      if (editing?.id) {
        const updated = await api.surveyBot.compileAndSave(editing.id);
        setEditing(updated);
        setDraftMd(updated.markdown);
        setDraftTitle(updated.title);
      } else {
        const out = await api.surveyBot.compile({
          markdown: draftMd,
          title_hint: draftTitle,
          use_web_search: true,
        });
        setDraftMd(out.markdown);
      }
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Compile failed');
    } finally {
      setCompileBusy(false);
    }
  };

  const togglePublish = async () => {
    if (!editing?.id) {
      setError('Save the survey before publishing');
      return;
    }
    setPublishBusy(true);
    setError(null);
    try {
      const updated = editing.published
        ? await api.surveyBot.unpublish(editing.id)
        : await api.surveyBot.publish(editing.id);
      setEditing(updated);
      const url = updated.published ? publicUrlFor(updated) : '';
      setShareUrl(url);
      if (url) {
        try {
          await navigator.clipboard.writeText(url);
        } catch {
          /* ignore */
        }
      }
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Publish failed');
    } finally {
      setPublishBusy(false);
    }
  };

  /** True for the file types the backend reads with AI rather than loading verbatim. */
  const needsAiReading = (file: File) =>
    /\.(pdf|jpe?g|png|gif|webp)$/i.test(file.name) ||
    file.type === 'application/pdf' ||
    file.type.startsWith('image/');

  const confirmReplaceDraft = async () => {
    if (!draftMd.trim()) return true;
    return confirm({
      title: 'Replace current draft?',
      message: 'The survey in the editor has not been saved. Generating from a file will replace it.',
    });
  };

  const onUploadFile = async (file: File | null | undefined) => {
    if (!file) return;
    if (!(await confirmReplaceDraft())) return;

    if (!needsAiReading(file)) {
      const text = await file.text();
      setDraftMd(text);
      if (!draftTitle) {
        setDraftTitle(file.name.replace(/\.(md|txt)$/i, ''));
      }
      setEditing(null);
      setShareUrl('');
      setSourcePreview(null);
      return;
    }

    setError(null);
    setFileBusy(true);
    try {
      const out = await api.surveyBot.fromFile(file, { title_hint: draftTitle.trim() || undefined });
      setDraftMd(out.markdown);
      setDraftTitle(out.title || file.name.replace(/\.[^.]+$/, ''));
      setEditing(null);
      setShareUrl('');
      setSourcePreview({
        name: out.source_name,
        kind: out.source_kind,
        text: out.source_text,
        truncated: Boolean(out.source_truncated),
      });
    } catch (e) {
      // A validation failure still carries markdown worth showing so it can be repaired.
      const rejected = (e as { markdown?: string }).markdown;
      if (rejected) {
        setDraftMd(rejected);
        setEditing(null);
      }
      setError(e instanceof Error ? e.message : 'Could not read that file');
    } finally {
      setFileBusy(false);
    }
  };

  const deleteTemplate = async (t: SurveyBotTemplate) => {
    const ok = await confirm({
      title: 'Delete AI Survey?',
      message: `Delete “${t.title}”?`,
    });
    if (!ok) return;
    await api.surveyBot.deleteTemplate(t.id);
    if (editing?.id === t.id) {
      clearSurveyDraft();
    }
    await load();
  };

  const deleteResult = async (r: SurveyBotResult) => {
    const ok = await confirm({
      title: 'Delete answer?',
      message: `Delete saved answer “${r.title}”?`,
    });
    if (!ok) return;
    await api.surveyBot.deleteResult(r.id);
    if (selectedResult?.id === r.id) {
      setSelectedResult(null);
      setEnlarge(null);
    }
    await load({ answersOnly: true });
  };

  const previewResult = async (r: SurveyBotResult) => {
    try {
      const full = await api.surveyBot.getResult(r.id);
      setSelectedResult(full);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load answer');
    }
  };

  const surveyEditor = (spacious: boolean, reserveEnlarge = false) =>
    draftMd ? (
      <>
        <input
          className={
            'rounded-lg border border-[#b7deee] dark:border-slate-600 bg-white dark:bg-slate-900 px-3 py-2 text-sm font-medium ' +
            (reserveEnlarge ? 'pr-12' : '')
          }
          placeholder="Title"
          value={draftTitle}
          onChange={(e) => setDraftTitle(e.target.value)}
        />
        <textarea
          className={
            'flex-1 font-mono rounded-lg border border-[#b7deee] dark:border-slate-600 bg-[#f8fcff] dark:bg-slate-900 px-3 py-2 ' +
            (spacious ? 'min-h-0 text-sm' : 'min-h-[240px] text-xs')
          }
          value={draftMd}
          onChange={(e) => setDraftMd(e.target.value)}
        />
        {shareUrl ? (
          <div className="rounded-lg border border-emerald-300/60 bg-emerald-50 dark:bg-emerald-950/30 dark:border-emerald-800 px-3 py-2 text-xs break-all">
            Public link:{' '}
            <a className="underline text-emerald-800 dark:text-emerald-200" href={shareUrl}>
              {shareUrl}
            </a>
          </div>
        ) : null}
        <div className="flex flex-wrap gap-2 justify-end">
          <button
            type="button"
            className="text-sm px-3 py-1.5 rounded-lg border border-slate-300 dark:border-slate-600"
            onClick={clearSurveyDraft}
          >
            Cancel
          </button>
          <button
            type="button"
            className="text-sm px-3 py-1.5 rounded-lg border border-sky-400 text-sky-800 dark:text-sky-200"
            disabled={compileBusy || !draftMd.trim()}
            onClick={() => void compileQuestions()}
          >
            {compileBusy ? 'Compiling…' : 'Compile questions'}
          </button>
          <button
            type="button"
            className="text-sm px-3 py-1.5 rounded-lg border border-emerald-500 text-emerald-800 dark:text-emerald-200"
            disabled={publishBusy || !editing?.id}
            onClick={() => void togglePublish()}
          >
            {publishBusy ? 'Working…' : editing?.published ? 'Unpublish' : 'Publish link'}
          </button>
          <button type="button" className={primaryBtn()} disabled={saving} onClick={() => void saveTemplate()}>
            {saving ? 'Saving…' : editing ? 'Update survey' : 'Save survey'}
          </button>
        </div>
      </>
    ) : (
      <div className="p-6 text-sm text-slate-500 space-y-2">
        <p>
          Upload a <strong>.md</strong> or <strong>.txt</strong> brief, paste a description (no questions required),
          then <strong>Compile questions</strong> (uses web search when helpful).
        </p>
        <p>
          Save, then <strong>Publish link</strong> so respondents open <code>/s/&#123;slug&#125;</code> and chat with a
          bot scoped to that survey.
        </p>
      </div>
    );

  const answerPreview = (
    selectedResult?.html ? (
      <iframe
        title="Survey answer"
        className="w-full h-full min-h-[420px] bg-white"
        srcDoc={selectedResult.html}
      />
    ) : (
      <div className="p-6 text-sm text-slate-500">Select an answer to preview its themed HTML.</div>
    )
  );

  return (
    <div className="flex-1 min-h-0 flex flex-col gap-3 overflow-hidden">
      <div className="flex flex-wrap items-center justify-between gap-2 shrink-0">
        <div>
          <h1 className="text-xl font-semibold text-sky-800 dark:text-sky-200">AI Surveys</h1>
          <p className="text-sm text-slate-600 dark:text-slate-400">
            MD/TXT survey briefs · compile questions · publish a bot link · collect answers
          </p>
        </div>
        <div className="flex gap-2">
          <button
            type="button"
            className={tabBtn(tab === 'surveys')}
            onClick={() => setTab('surveys')}
          >
            Surveys
          </button>
          <button type="button" className={tabBtn(tab === 'answers')} onClick={() => setTab('answers')}>
            Answers
          </button>
        </div>
      </div>

      {error ? (
        <div className="rounded-lg border border-rose-300 bg-rose-50 text-rose-800 px-3 py-2 text-sm dark:border-rose-800 dark:bg-rose-950/40 dark:text-rose-200">
          {error}
        </div>
      ) : null}

      {loading ? (
        <p className="text-sm text-slate-500">Loading…</p>
      ) : tab === 'answers' ? (
        <div className="flex-1 min-h-0 flex flex-col gap-2 overflow-hidden">
          <div className="shrink-0 flex flex-wrap items-center gap-2">
            <input
              className="flex-1 min-w-[12rem] rounded-lg border border-[#b7deee] dark:border-slate-600 bg-white dark:bg-slate-900 px-3 py-2 text-sm"
              placeholder="Search answers by title or survey slug…"
              value={answersQuery}
              onChange={(e) => setAnswersQuery(e.target.value)}
              aria-label="Search answers"
            />
            <button
              type="button"
              className="rounded-lg px-3 py-2 text-sm border border-[#b7deee] dark:border-slate-600 hover:bg-[#e8f5fb] dark:hover:bg-slate-800 disabled:opacity-50"
              disabled={refreshingAnswers}
              onClick={() => void load({ answersOnly: true, search: answersSearch })}
            >
              {refreshingAnswers ? 'Refreshing…' : 'Refresh'}
            </button>
            <span className="text-xs text-slate-500">
              {resultsTotal} answer{resultsTotal === 1 ? '' : 's'}
            </span>
          </div>
          <div className="flex-1 min-h-0 grid grid-cols-1 lg:grid-cols-2 gap-3 overflow-hidden">
            <div className="overflow-auto rounded-xl border border-[#b7deee] dark:border-slate-700 bg-white/80 dark:bg-slate-900/50">
              <ul className="divide-y divide-[#d8eef8] dark:divide-slate-800">
                {results.length === 0 ? (
                  <li className="p-4 text-sm text-slate-500">
                    {answersSearch ? (
                      'No answers match your search.'
                    ) : (
                      <>
                        No answers yet. Publish an AI Survey link, or in chat say <strong>survey bot</strong>.
                      </>
                    )}
                  </li>
                ) : (
                  results.map((r: SurveyBotResult) => (
                    <li key={r.id} className="p-3 flex items-start justify-between gap-2">
                      <button type="button" className="text-left flex-1" onClick={() => void previewResult(r)}>
                        <div className="font-medium text-[#0f4c66] dark:text-sky-200">{r.title}</div>
                        <div className="text-xs text-slate-500">
                          {r.template_slug || r.template_id} ·{' '}
                          {r.created_at ? new Date(r.created_at).toLocaleString() : ''}
                        </div>
                      </button>
                      <button
                        type="button"
                        className="text-xs text-rose-600 hover:underline"
                        onClick={() => void deleteResult(r)}
                      >
                        Delete
                      </button>
                    </li>
                  ))
                )}
              </ul>
            </div>
            <div className="relative min-h-0 flex flex-col overflow-hidden rounded-xl border border-[#b7deee] dark:border-slate-700 bg-white dark:bg-slate-950">
              {selectedResult?.html ? (
                <EnlargeIconButton
                  className="absolute top-3 right-3 z-10"
                  onClick={() => setEnlarge('answer')}
                />
              ) : null}
              <div className="flex-1 min-h-0 overflow-hidden">{answerPreview}</div>
            </div>
          </div>
        </div>
      ) : (
        <div className="flex-1 min-h-0 grid grid-cols-1 lg:grid-cols-2 gap-3 overflow-hidden">
          <div className="overflow-auto rounded-xl border border-[#b7deee] dark:border-slate-700 bg-white/80 dark:bg-slate-900/50 p-3 space-y-3">
            <div className="flex flex-wrap gap-2">
              <button type="button" className={primaryBtn()} onClick={openNewTemplate}>
                New survey
              </button>
              <button
                type="button"
                className="rounded-lg px-3 py-1.5 text-sm border border-[#b7deee] dark:border-slate-600 disabled:opacity-60"
                disabled={fileBusy}
                title="Markdown and text load straight into the editor. PDFs and images are read by AI."
                onClick={() => fileRef.current?.click()}
              >
                {fileBusy ? 'Reading file…' : 'Upload file'}
              </button>
              <input
                ref={fileRef}
                type="file"
                accept=".md,.txt,.pdf,.jpg,.jpeg,.png,.gif,.webp,text/markdown,text/plain,application/pdf,image/*"
                className="hidden"
                onChange={(e) => {
                  void onUploadFile(e.target.files?.[0]);
                  e.target.value = '';
                }}
              />
            </div>
            <p className="text-xs text-slate-500">
              Markdown and text open directly. PDFs and photos of a questionnaire are read by AI into a draft survey.
            </p>
            {sourcePreview ? (
              <details className="rounded-lg border border-[#d8eef8] dark:border-slate-800 bg-[#f8fcff] dark:bg-slate-900/60 px-3 py-2">
                <summary className="text-xs cursor-pointer text-[#0f4c66] dark:text-sky-200">
                  What the AI read from {sourcePreview.name} ({sourcePreview.kind === 'image' ? 'image' : 'PDF'})
                  {sourcePreview.truncated ? ' — truncated' : ''}
                </summary>
                <pre className="mt-2 max-h-52 overflow-auto whitespace-pre-wrap text-[11px] text-slate-600 dark:text-slate-300">
                  {sourcePreview.text}
                </pre>
              </details>
            ) : null}
            <div className="flex gap-2">
              <input
                className="flex-1 rounded-lg border border-[#b7deee] dark:border-slate-600 bg-white dark:bg-slate-900 px-3 py-2 text-sm"
                placeholder="Describe a survey to draft with AI…"
                value={aiQuery}
                onChange={(e) => setAiQuery(e.target.value)}
              />
              <button type="button" className={primaryBtn()} disabled={aiBusy} onClick={() => void runAiDraft()}>
                {aiBusy ? 'Drafting…' : 'Create with AI'}
              </button>
            </div>
            <ul className="divide-y divide-[#d8eef8] dark:divide-slate-800 border border-[#d8eef8] dark:border-slate-800 rounded-lg">
              {templates.map((t) => (
                <li key={t.id} className="p-3 flex items-start justify-between gap-2">
                  <button type="button" className="text-left flex-1" onClick={() => openEdit(t)}>
                    <div className="font-medium text-[#0f4c66] dark:text-sky-200">
                      {t.title}{' '}
                      {t.published ? (
                        <span className="text-[10px] uppercase tracking-wide text-emerald-600 dark:text-emerald-400">
                          published
                        </span>
                      ) : null}
                    </div>
                    <div className="text-xs text-slate-500">{t.slug}</div>
                  </button>
                  <button
                    type="button"
                    className="text-xs text-rose-600 hover:underline"
                    onClick={() => void deleteTemplate(t)}
                  >
                    Delete
                  </button>
                </li>
              ))}
            </ul>
          </div>
          <div className="relative min-h-0 flex flex-col gap-2 overflow-hidden rounded-xl border border-[#b7deee] dark:border-slate-700 bg-white dark:bg-slate-950 p-3">
            {draftMd ? (
              <EnlargeIconButton
                className="absolute top-3 right-3 z-10"
                onClick={() => setEnlarge('survey')}
              />
            ) : null}
            {surveyEditor(false, Boolean(draftMd))}
          </div>
        </div>
      )}

      {enlarge === 'survey' ? (
        <EnlargeModal
          title={draftTitle.trim() || (editing ? 'Edit survey' : 'Survey editor')}
          onClose={() => setEnlarge(null)}
        >
          <div className="flex flex-col flex-1 min-h-0 gap-2 p-3 overflow-hidden">{surveyEditor(true)}</div>
        </EnlargeModal>
      ) : null}

      {enlarge === 'answer' ? (
        <EnlargeModal
          title={selectedResult?.title || 'Answer preview'}
          onClose={() => setEnlarge(null)}
        >
          <div className="flex-1 min-h-0 overflow-hidden bg-white dark:bg-slate-950">{answerPreview}</div>
        </EnlargeModal>
      ) : null}
    </div>
  );
}

function EnlargeIconButton({
  onClick,
  className = '',
}: {
  onClick: () => void;
  className?: string;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-label="Enlarge"
      title="Enlarge"
      className={
        'inline-flex h-8 w-8 items-center justify-center rounded-lg border border-[#b7deee]/90 ' +
        'bg-white/90 text-[#0f4c66] shadow-sm backdrop-blur-sm ' +
        'hover:bg-[#e8f5fb] dark:border-slate-600 dark:bg-slate-900/90 dark:text-sky-200 dark:hover:bg-slate-800 ' +
        className
      }
    >
      <svg width="15" height="15" viewBox="0 0 24 24" fill="none" aria-hidden="true">
        <path
          d="M8 3H3v5M16 3h5v5M8 21H3v-5M21 16v5h-5"
          stroke="currentColor"
          strokeWidth="1.8"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
    </button>
  );
}

function EnlargeModal({
  title,
  onClose,
  children,
}: {
  title: string;
  onClose: () => void;
  children: ReactNode;
}) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-3 sm:p-6" role="presentation">
      <button
        type="button"
        className="absolute inset-0 bg-slate-950/70 backdrop-blur-[2px]"
        aria-label="Close enlarged view"
        onClick={onClose}
      />
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="surveyx-enlarge-modal-title"
        className="relative flex flex-col w-full max-w-5xl h-[min(92vh,960px)] rounded-2xl border border-[#b7deee] dark:border-slate-600 bg-slate-100 dark:bg-slate-900 shadow-2xl overflow-hidden"
        onMouseDown={(e) => e.stopPropagation()}
      >
        <div className="shrink-0 flex items-center justify-between gap-3 px-4 py-3 border-b border-[#d8eef8] dark:border-slate-700 bg-white dark:bg-slate-900/90">
          <h2 id="surveyx-enlarge-modal-title" className="text-base font-semibold text-slate-900 dark:text-white truncate">
            {title}
          </h2>
          <button
            type="button"
            onClick={onClose}
            className="shrink-0 px-3 py-1.5 rounded-lg border border-slate-300 dark:border-slate-600 text-slate-700 dark:text-slate-200 text-sm hover:bg-slate-100 dark:hover:bg-slate-800"
          >
            Close
          </button>
        </div>
        <div className="flex flex-col flex-1 min-h-0 overflow-hidden">{children}</div>
      </div>
    </div>
  );
}

function tabBtn(active: boolean) {
  return (
    'rounded-lg px-3 py-1.5 text-sm font-medium border transition-colors ' +
    (active
      ? 'bg-[#d8eef8] text-[#0f4c66] border-[#b7deee] dark:bg-sky-900/40 dark:text-sky-100 dark:border-sky-700'
      : 'bg-transparent text-slate-600 border-transparent hover:bg-[#e8f5fb] dark:text-slate-300 dark:hover:bg-slate-800')
  );
}

function primaryBtn() {
  return 'rounded-lg px-3 py-1.5 text-sm font-medium bg-[#0f4c66] text-white hover:bg-[#0c3d52] disabled:opacity-50';
}
