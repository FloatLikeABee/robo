import { useEffect, useState, useRef } from 'react';
import { Link, useParams, useNavigate } from 'react-router-dom';
import { useConfirm } from '../context/ConfirmContext';
import { QuestionDefaultValueFields } from '../components/QuestionDefaultValueFields';
import { FormTemplateAIPanel } from '../components/FormTemplateAIPanel';
import { api, type Form, type FormPage, type Question, type QuestionRule, uploadsUrl, MAX_QUESTION_PROMPT_IMAGE_BYTES, MAX_QUESTION_PROMPT_VIDEO_BYTES } from '../lib/api';
import { LEGACY_QUESTION_TYPES, QUESTION_TYPES } from '../lib/questionTypes';

export function EditForm() {
  const { confirm } = useConfirm();
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const isNew = !id || id === 'new';
  const formId = isNew ? 0 : Number(id);

  const [form, setForm] = useState<Form | null>(null);
  const [pages, setPages] = useState<FormPage[]>([]);
  const [activePageId, setActivePageId] = useState<number | null>(null);
  const [questions, setQuestions] = useState<Question[]>([]);
  const [loading, setLoading] = useState(!isNew);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [saved, setSaved] = useState(false);

  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [slug, setSlug] = useState('');
  const [singleResponseOnly, setSingleResponseOnly] = useState(false);
  const [examMode, setExamMode] = useState(false);
  const [questionsModalOpen, setQuestionsModalOpen] = useState(false);

  useEffect(() => {
    if (!questionsModalOpen) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setQuestionsModalOpen(false);
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [questionsModalOpen]);

  useEffect(() => {
    if (isNew) {
      setLoading(false);
      return;
    }
    api.forms.get(formId)
      .then((f) => {
        setForm(f);
        setName(f.name);
        setDescription(f.description || '');
        setSlug(f.slug);
        setSingleResponseOnly(f.single_response_only);
        setExamMode(f.exam_mode ?? false);
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, [formId, isNew]);

  const [rules, setRules] = useState<QuestionRule[]>([]);

  useEffect(() => {
    if (isNew || !formId) return;
    api.pages.list(formId)
      .then((p) => {
        setPages(p);
        setActivePageId((prev) => (prev && p.some((x) => x.id === prev) ? prev : p[0]?.id ?? null));
      })
      .catch(() => {});
  }, [formId, isNew]);

  useEffect(() => {
    if (isNew || !formId) return;
    api.questions.list(formId)
      .then(setQuestions)
      .catch(() => {});
  }, [formId, isNew]);

  useEffect(() => {
    if (isNew || !formId) return;
    api.rules.listByForm(formId)
      .then(setRules)
      .catch(() => {});
  }, [formId, isNew]);

  const saveForm = async () => {
    setSaving(true);
    setError(null);
    setSaved(false);
    try {
      if (isNew) {
        const created = await api.forms.create({
          name,
          description,
          slug,
          single_response_only: singleResponseOnly,
          exam_mode: examMode,
        });
        navigate(`/forms/${created.id}/edit`, { replace: true });
        setForm(created);
        setExamMode(Boolean(created.exam_mode));
        const ps = await api.pages.list(created.id);
        setPages(ps);
        setActivePageId(ps[0]?.id ?? null);
        const qs = await api.questions.list(created.id);
        setQuestions(qs);
      } else {
        const updated = await api.forms.update(formId, {
          name,
          description,
          slug,
          single_response_only: singleResponseOnly,
          exam_mode: examMode,
        });
        setForm(updated);
      }
      setSaved(true);
      setTimeout(() => setSaved(false), 2500);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Save failed');
    } finally {
      setSaving(false);
    }
  };

  const addQuestion = async () => {
    if (!formId || !activePageId) return;
    const pageQuestions = questions.filter((q) => q.page_id === activePageId);
    try {
      const q = await api.questions.create(formId, {
        title: 'New question',
        type: 'text',
        required: false,
        page_id: activePageId,
        sort_order: pageQuestions.length,
        config: {},
      });
      setQuestions((prev) => [...prev, q]);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Add question failed');
    }
  };

  const addPage = async () => {
    if (!formId) return;
    try {
      const page = await api.pages.create(formId, { name: '', sort_order: pages.length });
      setPages((prev) => [...prev, page]);
      setActivePageId(page.id);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Add page failed');
    }
  };

  const updatePage = async (page: FormPage, updates: Partial<{ name: string; sort_order: number }>) => {
    try {
      const updated = await api.pages.update(formId, page.id, updates);
      setPages((prev) => prev.map((p) => (p.id === page.id ? updated : p)));
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Update page failed');
    }
  };

  const deletePage = async (page: FormPage) => {
    if (pages.length <= 1) return;
    const ok = await confirm({
      title: 'Remove page',
      message: 'Questions on this page will move to another page. Continue?',
      confirmLabel: 'Remove',
      danger: true,
    });
    if (!ok) return;
    try {
      await api.pages.delete(formId, page.id);
      const remaining = pages.filter((p) => p.id !== page.id);
      setPages(remaining);
      if (activePageId === page.id) {
        setActivePageId(remaining[0]?.id ?? null);
      }
      const qs = await api.questions.list(formId);
      setQuestions(qs);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Delete page failed');
    }
  };

  const updateQuestion = async (q: Question, updates: Partial<Question>) => {
    try {
      const updated = await api.questions.update(formId, q.id, updates);
      setQuestions((prev) => prev.map((x) => (x.id === q.id ? updated : x)));
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Update failed');
    }
  };

  const replaceQuestion = (updated: Question) => {
    setQuestions((prev) => prev.map((x) => (x.id === updated.id ? updated : x)));
  };

  const deleteQuestion = async (q: Question) => {
    const ok = await confirm({
      title: 'Remove question',
      message: 'Are you sure you want to remove this question?',
      confirmLabel: 'Remove',
      danger: true,
    });
    if (!ok) return;
    try {
      await api.questions.delete(formId, q.id);
      setQuestions((prev) => prev.filter((x) => x.id !== q.id));
      setRules((prev) => prev.filter((r) => r.question_id !== q.id && r.depends_on_question_id !== q.id));
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Delete failed');
    }
  };

  const addRule = async (questionId: number, dependsOnQuestionId: number, condition: 'answered' | 'not_answered') => {
    try {
      const rule = await api.rules.create(formId, questionId, { depends_on_question_id: dependsOnQuestionId, condition });
      setRules((prev) => [...prev, rule]);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Add rule failed');
    }
  };

  const deleteRule = async (questionId: number, ruleId: number) => {
    try {
      await api.rules.delete(formId, questionId, ruleId);
      setRules((prev) => prev.filter((r) => r.id !== ruleId));
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Delete rule failed');
    }
  };

  if (loading) {
    return (
      <div className="flex flex-col flex-1 min-h-0 overflow-hidden">
        <div className="text-slate-500 dark:text-slate-400 text-sm">Loading…</div>
      </div>
    );
  }
  if (!isNew && !form) {
    return (
      <div className="flex flex-col flex-1 min-h-0 overflow-hidden">
        <div className="text-red-600 dark:text-red-400 text-sm">Form not found</div>
      </div>
    );
  }

  return (
    <div className="flex flex-col flex-1 min-h-0 h-full overflow-hidden w-full max-w-full">
      <div className="flex-shrink-0 flex flex-wrap items-center justify-between gap-2 mb-2">
        <div className="flex items-center gap-2 min-w-0">
          <Link
            to="/forms"
            className="p-1.5 rounded-lg text-slate-500 dark:text-slate-400 hover:bg-slate-200 dark:hover:bg-slate-700/50 shrink-0"
          >
            ←
          </Link>
          <div className="min-w-0">
            <h1 className="text-lg font-semibold text-slate-900 dark:text-white leading-tight">
              {isNew ? 'New Form' : 'Edit Form'}
            </h1>
            <p className="text-slate-600 dark:text-slate-400 text-xs truncate">{isNew ? 'Create a form' : 'Update your form'}</p>
          </div>
        </div>
        <button
          type="button"
          onClick={saveForm}
          disabled={saving}
          className="shrink-0 inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-violet-600 hover:bg-violet-500 text-white text-xs font-medium disabled:opacity-50"
        >
          {isNew ? 'Create Form' : '📄 Update Form'}
        </button>
      </div>

      {error && (
        <div className="flex-shrink-0 mb-2 p-2 rounded-lg bg-red-100 dark:bg-red-900/30 text-red-800 dark:text-red-300 text-xs">
          {error}
        </div>
      )}
      {!error && saved && (
        <div className="flex-shrink-0 mb-2 p-2 rounded-lg bg-emerald-100 dark:bg-emerald-900/30 text-emerald-800 dark:text-emerald-300 text-xs">
          Changes saved.
        </div>
      )}

      <div className="flex-shrink-0 rounded-lg bg-slate-100 dark:bg-slate-800/40 border border-slate-200 dark:border-slate-700/50 p-2 mb-2">
        <h2 className="text-xs font-medium text-slate-800 dark:text-white mb-2">Form settings</h2>
        <div className="space-y-2">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-2">
            <div>
              <label className="block text-[10px] text-slate-500 dark:text-slate-400 mb-0.5">Form name *</label>
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="w-full px-2 py-1.5 rounded-md bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-600 text-slate-900 dark:text-white text-xs"
                placeholder="Survey title"
              />
            </div>
            <div className="md:col-span-1 lg:col-span-1">
              <label className="block text-[10px] text-slate-500 dark:text-slate-400 mb-0.5">Description</label>
              <input
                type="text"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                className="w-full px-2 py-1.5 rounded-md bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-600 text-slate-900 dark:text-white text-xs"
                placeholder="Short description"
              />
            </div>
            <div>
              <label className="block text-[10px] text-slate-500 dark:text-slate-400 mb-0.5">Form URL *</label>
              <div className="flex items-stretch gap-0 rounded-md overflow-hidden border border-slate-300 dark:border-slate-600">
                <span className="px-2 py-1.5 bg-slate-200 dark:bg-slate-700 text-slate-600 dark:text-slate-400 text-xs shrink-0">
                  /f/
                </span>
                <input
                  type="text"
                  value={slug}
                  onChange={(e) => setSlug(e.target.value)}
                  className="flex-1 min-w-0 px-2 py-1.5 bg-white dark:bg-slate-800 text-slate-900 dark:text-white text-xs font-mono"
                  placeholder="slug"
                />
              </div>
            </div>
            <div className="md:col-span-1 lg:col-span-1 flex items-end">
              <p className="text-[10px] text-slate-500 dark:text-slate-500">
                Public link: /f/{slug || '…'}
              </p>
            </div>
            <div className="md:col-span-1 lg:col-span-2 flex items-center gap-2 pt-1 border-t border-slate-200 dark:border-slate-700/50">
              <span className="text-xs text-slate-700 dark:text-slate-300">Single response only</span>
              <button
                type="button"
                role="switch"
                aria-checked={singleResponseOnly}
                onClick={() => setSingleResponseOnly((v) => !v)}
                className={`relative w-9 h-5 rounded-full shrink-0 transition-colors ${singleResponseOnly ? 'bg-violet-600' : 'bg-slate-400 dark:bg-slate-600'}`}
              >
                <span
                  className={`absolute top-0.5 w-4 h-4 rounded-full bg-white transition-transform ${singleResponseOnly ? 'left-4' : 'left-0.5'}`}
                />
              </button>
            </div>
            <div className="md:col-span-1 lg:col-span-2 flex items-center gap-2 flex-wrap pt-1 border-t border-slate-200 dark:border-slate-700/50">
              <span className="text-xs text-slate-700 dark:text-slate-300 shrink-0">Exam Mode</span>
              <button
                type="button"
                role="switch"
                aria-checked={examMode}
                onClick={() => setExamMode((v) => !v)}
                className={`relative w-9 h-5 rounded-full shrink-0 transition-colors ${examMode ? 'bg-violet-600' : 'bg-slate-400 dark:bg-slate-600'}`}
              >
                <span
                  className={`absolute top-0.5 w-4 h-4 rounded-full bg-white transition-transform ${examMode ? 'left-4' : 'left-0.5'}`}
                />
              </button>
              <span className="text-[10px] text-slate-500 dark:text-slate-400 min-w-0">
                Respondent taps Start before the form; time is recorded on submit.
              </span>
            </div>
          </div>
        </div>
      </div>

      <FormTemplateAIPanel
        formId={formId}
        formName={name}
        formDescription={description}
        initialHtml={form?.landing_html || ''}
        onQuestionsApplied={(created) => setQuestions((prev) => [...prev, ...created])}
        activePageId={activePageId}
        disabled={isNew}
      />

      <div className="flex flex-col flex-1 min-h-0 rounded-lg bg-slate-100 dark:bg-slate-800/40 border border-slate-200 dark:border-slate-700/50 overflow-hidden">
        <QuestionsSectionHeader
          isNew={isNew}
          activePageId={activePageId}
          onAddPage={addPage}
          onAddQuestion={addQuestion}
          onExpand={() => setQuestionsModalOpen(true)}
        />
        <QuestionsEditorBody
          compact
          isNew={isNew}
          pages={pages}
          activePageId={activePageId}
          setActivePageId={setActivePageId}
          questions={questions}
          rules={rules}
          onUpdatePage={updatePage}
          onDeletePage={deletePage}
          onUpdateQuestion={updateQuestion}
          onReplaceQuestion={replaceQuestion}
          onDeleteQuestion={deleteQuestion}
          onAddRule={addRule}
          onDeleteRule={deleteRule}
        />
      </div>

      {questionsModalOpen && !isNew && (
        <QuestionsEditorModal onClose={() => setQuestionsModalOpen(false)}>
          <QuestionsSectionHeader
            isNew={isNew}
            activePageId={activePageId}
            onAddPage={addPage}
            onAddQuestion={addQuestion}
            onExpand={undefined}
            modal
          />
          <QuestionsEditorBody
            compact={false}
            isNew={isNew}
            pages={pages}
            activePageId={activePageId}
            setActivePageId={setActivePageId}
            questions={questions}
            rules={rules}
            onUpdatePage={updatePage}
            onDeletePage={deletePage}
            onUpdateQuestion={updateQuestion}
            onReplaceQuestion={replaceQuestion}
            onDeleteQuestion={deleteQuestion}
            onAddRule={addRule}
            onDeleteRule={deleteRule}
          />
        </QuestionsEditorModal>
      )}
    </div>
  );
}

type QuestionsEditorHandlers = {
  isNew: boolean;
  pages: FormPage[];
  activePageId: number | null;
  setActivePageId: (id: number | null) => void;
  questions: Question[];
  rules: QuestionRule[];
  onUpdatePage: (page: FormPage, updates: Partial<{ name: string; sort_order: number }>) => void;
  onDeletePage: (page: FormPage) => void;
  onUpdateQuestion: (q: Question, updates: Partial<Question>) => void;
  onReplaceQuestion: (q: Question) => void;
  onDeleteQuestion: (q: Question) => void;
  onAddRule: (questionId: number, dependsOnQuestionId: number, condition: 'answered' | 'not_answered') => void;
  onDeleteRule: (questionId: number, ruleId: number) => void;
};

function QuestionsSectionHeader({
  isNew,
  activePageId,
  onAddPage,
  onAddQuestion,
  onExpand,
  modal,
}: {
  isNew: boolean;
  activePageId: number | null;
  onAddPage: () => void;
  onAddQuestion: () => void;
  onExpand?: () => void;
  modal?: boolean;
}) {
  return (
    <div className="flex-shrink-0 flex flex-wrap items-center justify-between gap-2 px-2.5 py-2 border-b border-slate-200 dark:border-slate-700/50">
      <div className="min-w-0">
        <h2 className={`font-medium text-slate-800 dark:text-white ${modal ? 'text-sm' : 'text-xs'}`}>Questions</h2>
        {modal && (
          <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">Expanded editor — press Esc to close</p>
        )}
      </div>
      <div className="flex items-center gap-1.5">
        {!modal && onExpand && (
          <button
            type="button"
            onClick={onExpand}
            disabled={isNew}
            className="px-2 py-1 rounded-md border border-slate-300 dark:border-slate-600 text-slate-700 dark:text-slate-300 text-xs disabled:opacity-50 disabled:cursor-not-allowed hover:bg-slate-200 dark:hover:bg-slate-700/50"
            title="Open expanded question editor"
          >
            Expand editor
          </button>
        )}
        <button
          type="button"
          onClick={onAddPage}
          disabled={isNew}
          className="px-2 py-1 rounded-md border border-slate-300 dark:border-slate-600 text-slate-700 dark:text-slate-300 text-xs disabled:opacity-50 disabled:cursor-not-allowed"
        >
          + Add page
        </button>
        <button
          type="button"
          onClick={onAddQuestion}
          disabled={isNew || !activePageId}
          className="px-2.5 py-1 rounded-md bg-violet-600 hover:bg-violet-500 text-white text-xs disabled:opacity-50 disabled:cursor-not-allowed"
        >
          + Add question
        </button>
      </div>
    </div>
  );
}

function QuestionsEditorBody({
  compact,
  isNew,
  pages,
  activePageId,
  setActivePageId,
  questions,
  rules,
  onUpdatePage,
  onDeletePage,
  onUpdateQuestion,
  onReplaceQuestion,
  onDeleteQuestion,
  onAddRule,
  onDeleteRule,
}: QuestionsEditorHandlers & { compact: boolean }) {
  const activePage = activePageId != null ? pages.find((p) => p.id === activePageId) : undefined;
  const pageQuestions = questions.filter((q) => q.page_id === activePageId);

  return (
    <div className="flex flex-col flex-1 min-h-0 overflow-hidden">
      {!isNew && pages.length > 0 && (
        <div className="flex-shrink-0 flex flex-wrap items-center gap-1 px-2.5 py-1.5 border-b border-slate-200 dark:border-slate-700/50 bg-slate-50 dark:bg-slate-900/20">
          {pages.map((page, idx) => (
            <button
              key={page.id}
              type="button"
              onClick={() => setActivePageId(page.id)}
              className={`px-2 py-0.5 rounded-md text-xs transition-colors ${
                activePageId === page.id
                  ? 'bg-violet-600 text-white'
                  : 'bg-slate-200 dark:bg-slate-700/70 text-slate-700 dark:text-slate-300 hover:bg-slate-300 dark:hover:bg-slate-600/70'
              }`}
            >
              {page.name.trim() || String(idx + 1)}
            </button>
          ))}
        </div>
      )}
      {!isNew && activePage && (
        <PageEditor
          page={activePage}
          canDelete={pages.length > 1}
          onUpdate={(updates) => onUpdatePage(activePage, updates)}
          onDelete={() => onDeletePage(activePage)}
          spacious={!compact}
        />
      )}
      <div
        className={`flex-1 min-h-0 overflow-y-auto overflow-x-hidden space-y-2 ${compact ? 'p-2' : 'p-4 space-y-4'}`}
      >
        {isNew ? (
          <div className="rounded-md border border-dashed border-slate-300 dark:border-slate-600 p-3 text-xs text-slate-600 dark:text-slate-400">
            Create the form first, then add questions in this panel.
          </div>
        ) : pageQuestions.length === 0 ? (
          <div className="rounded-md border border-dashed border-slate-300 dark:border-slate-600 p-4 text-sm text-slate-600 dark:text-slate-400 text-center">
            No questions on this page yet. Click &quot;+ Add question&quot; to create one.
          </div>
        ) : (
          pageQuestions.map((q) => (
            <QuestionCard
              key={q.id}
              compact={compact}
              question={q}
              allQuestions={questions}
              rules={rules.filter((r) => r.question_id === q.id)}
              onUpdate={(updates) => onUpdateQuestion(q, updates)}
              onQuestionReplaced={onReplaceQuestion}
              onDelete={() => onDeleteQuestion(q)}
              onAddRule={(dependsOnId, condition) => onAddRule(q.id, dependsOnId, condition)}
              onDeleteRule={(ruleId) => onDeleteRule(q.id, ruleId)}
            />
          ))
        )}
      </div>
    </div>
  );
}

function QuestionsEditorModal({ children, onClose }: { children: React.ReactNode; onClose: () => void }) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-3 sm:p-6" role="presentation">
      <button
        type="button"
        className="absolute inset-0 bg-slate-950/70 backdrop-blur-[2px]"
        aria-label="Close expanded editor"
        onClick={onClose}
      />
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="questions-editor-modal-title"
        className="relative flex flex-col w-full max-w-4xl h-[min(92vh,920px)] rounded-2xl border border-slate-300 dark:border-slate-600 bg-slate-100 dark:bg-slate-900 shadow-2xl overflow-hidden"
        onMouseDown={(e) => e.stopPropagation()}
      >
        <div className="flex-shrink-0 flex items-center justify-between gap-3 px-4 py-3 border-b border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900/90">
          <h2 id="questions-editor-modal-title" className="text-base font-semibold text-slate-900 dark:text-white">
            Question editor
          </h2>
          <button
            type="button"
            onClick={onClose}
            className="px-3 py-1.5 rounded-lg border border-slate-300 dark:border-slate-600 text-slate-700 dark:text-slate-200 text-sm hover:bg-slate-100 dark:hover:bg-slate-800"
          >
            Close
          </button>
        </div>
        <div className="flex flex-col flex-1 min-h-0 overflow-hidden">{children}</div>
      </div>
    </div>
  );
}

function PageEditor({
  page,
  canDelete,
  onUpdate,
  onDelete,
  spacious,
}: {
  page: FormPage;
  canDelete: boolean;
  onUpdate: (updates: Partial<{ name: string; sort_order: number }>) => void;
  onDelete: () => void;
  spacious?: boolean;
}) {
  const [name, setName] = useState(page.name);

  useEffect(() => {
    setName(page.name);
  }, [page.id, page.updated_at]);

  const blurName = () => {
    const trimmed = name.trim();
    if (trimmed !== page.name) {
      setName(trimmed);
      onUpdate({ name: trimmed });
    }
  };

  return (
    <div className={`flex-shrink-0 flex items-center gap-2 border-b border-slate-200 dark:border-slate-700/50 ${spacious ? 'px-4 py-2.5' : 'px-2.5 py-1.5'}`}>
      <input
        type="text"
        value={name}
        onChange={(e) => setName(e.target.value)}
        onBlur={blurName}
        placeholder="Page name (optional)"
        className={`flex-1 min-w-0 rounded-md bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-600 text-slate-900 dark:text-white ${spacious ? 'px-3 py-2 text-sm' : 'px-2 py-1 text-xs'}`}
      />
      {canDelete && (
        <button
          type="button"
          onClick={onDelete}
          className="px-2 py-1 rounded-md text-red-600 dark:text-red-400 hover:bg-red-100 dark:hover:bg-red-900/30 text-xs shrink-0"
        >
          Remove page
        </button>
      )}
    </div>
  );
}

function QuestionPromptMediaEditor({
  question,
  compact,
  lbl,
  onReplaced,
}: {
  question: Question;
  compact?: boolean;
  lbl: string;
  onReplaced: (q: Question) => void;
}) {
  const [busy, setBusy] = useState(false);
  const [err, setErr] = useState<string | null>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const media = question.config?.question_prompt_media;
  const url = media?.relative_path ? uploadsUrl(media.relative_path) : '';

  const maxBytesForFile = (f: File) => {
    const t = f.type.toLowerCase();
    if (t === 'video/mp4' || t === 'video/webm') return MAX_QUESTION_PROMPT_VIDEO_BYTES;
    if (t.startsWith('image/')) return MAX_QUESTION_PROMPT_IMAGE_BYTES;
    const n = f.name.toLowerCase();
    if (n.endsWith('.mp4') || n.endsWith('.webm')) return MAX_QUESTION_PROMPT_VIDEO_BYTES;
    return MAX_QUESTION_PROMPT_IMAGE_BYTES;
  };

  const onPick = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const f = e.target.files?.[0];
    e.target.value = '';
    if (!f) return;
    setErr(null);
    const max = maxBytesForFile(f);
    if (f.size > max) {
      setErr(`File too large (max ${max / (1024 * 1024)} MB for this type).`);
      return;
    }
    setBusy(true);
    try {
      const q = await api.questions.uploadPromptMedia(question.form_id, question.id, f);
      onReplaced(q);
    } catch (ex) {
      setErr(ex instanceof Error ? ex.message : 'Upload failed');
    } finally {
      setBusy(false);
    }
  };

  const onRemove = async () => {
    if (!media) return;
    setErr(null);
    setBusy(true);
    try {
      const q = await api.questions.deletePromptMedia(question.form_id, question.id);
      onReplaced(q);
    } catch (ex) {
      setErr(ex instanceof Error ? ex.message : 'Remove failed');
    } finally {
      setBusy(false);
    }
  };

  const hintCls = compact ? 'text-[10px] text-slate-500 dark:text-slate-500' : 'text-xs text-slate-500 dark:text-slate-500';

  return (
    <div className="space-y-1">
      <label className={`block text-slate-500 dark:text-slate-400 ${lbl}`}>Prompt image or video (optional)</label>
      <p className={hintCls}>
        One attachment per question (image or video). Up to 10 MB for images, 100 MB for video. Types: JPEG, PNG, GIF,
        WebP, MP4, WebM.
      </p>
      {media && (
        <div className="rounded-md overflow-hidden border border-slate-300 dark:border-slate-600 max-w-xs">
          {media.kind === 'video' ? (
            <video src={url} controls className="w-full max-h-40 bg-black" />
          ) : (
            <img src={url} alt="Question prompt" className="w-full max-h-40 object-contain bg-slate-900" />
          )}
        </div>
      )}
      <div className="flex flex-wrap items-center gap-2">
        <input
          ref={inputRef}
          type="file"
          accept="image/jpeg,image/png,image/gif,image/webp,video/mp4,video/webm"
          className="hidden"
          disabled={busy}
          onChange={onPick}
        />
        <button
          type="button"
          disabled={busy}
          onClick={() => inputRef.current?.click()}
          className={`rounded-md bg-violet-600 hover:bg-violet-500 text-white disabled:opacity-50 ${compact ? 'px-2 py-1 text-[11px]' : 'px-2 py-1.5 text-sm'}`}
        >
          {busy ? '…' : media ? 'Replace…' : 'Add image or video…'}
        </button>
        {media && (
          <button
            type="button"
            disabled={busy}
            onClick={onRemove}
            className={`rounded-md border border-slate-300 dark:border-slate-600 text-slate-700 dark:text-slate-300 ${compact ? 'px-2 py-1 text-[11px]' : 'px-2 py-1.5 text-sm'}`}
          >
            Remove
          </button>
        )}
      </div>
      {err && <p className="text-red-600 dark:text-red-400 text-[11px]">{err}</p>}
    </div>
  );
}
function QuestionCard({
  compact,
  question,
  allQuestions,
  rules,
  onUpdate,
  onQuestionReplaced,
  onDelete,
  onAddRule,
  onDeleteRule,
}: {
  compact?: boolean;
  question: Question;
  allQuestions: Question[];
  rules: QuestionRule[];
  onUpdate: (u: Partial<Question>) => void;
  onQuestionReplaced: (q: Question) => void;
  onDelete: () => void;
  onAddRule: (dependsOnQuestionId: number, condition: 'answered' | 'not_answered') => void;
  onDeleteRule: (ruleId: number) => void;
}) {
  const [title, setTitle] = useState(question.title);
  const [type, setType] = useState(question.type);
  const [required, setRequired] = useState(question.required);
  const [optionsStr, setOptionsStr] = useState(
    question.config?.options?.map((o) => `${o.value}:${o.label}`).join('\n') || ''
  );
  const [addDependsOn, setAddDependsOn] = useState<number>(0);
  const [addCondition, setAddCondition] = useState<'answered' | 'not_answered'>('answered');

  useEffect(() => {
    setTitle(question.title);
    setType(question.type);
    setRequired(question.required);
    setOptionsStr(question.config?.options?.map((o) => `${o.value}:${o.label}`).join('\n') || '');
    // Re-sync when this question is replaced from the server (e.g. after PUT), not only when id changes
  }, [question.id, question.updated_at]);

  const typeOptions = (() => {
    const known = new Set<string>(QUESTION_TYPES.map((t) => t.value));
    if (known.has(type)) return [...QUESTION_TYPES];
    const legacy = LEGACY_QUESTION_TYPES.find((t) => t.value === type);
    return legacy ? [...QUESTION_TYPES, legacy] : [...QUESTION_TYPES, { value: type, label: type }];
  })();

  const blurTitle = () => {
    const t = title.trim();
    if (!t) {
      setTitle(question.title);
      return;
    }
    if (t !== question.title) {
      setTitle(t);
      onUpdate({ title: t });
    }
  };

  const applyOptions = () => {
    const options: { value: number; label: string }[] = [];
    optionsStr.split('\n').forEach((line) => {
      const m = line.trim().match(/^(\d+):(.+)$/);
      if (m) options.push({ value: parseInt(m[1], 10), label: m[2].trim() });
    });
    onUpdate({ title, type, required, config: { ...question.config, options } });
  };

  const needsOptions = type === 'select' || type === 'multiselect';

  const otherQuestions = allQuestions.filter((q) => q.id !== question.id);
  const canAddRule = otherQuestions.length > 0 && addDependsOn > 0;
  const handleAddRule = () => {
    if (!canAddRule) return;
    onAddRule(addDependsOn, addCondition);
    setAddDependsOn(0);
  };

  const getQuestionTitle = (id: number) => allQuestions.find((q) => q.id === id)?.title ?? `#${id}`;

  const pad = compact ? 'p-1.5 gap-1.5' : 'p-4 gap-3';
  const space = compact ? 'space-y-1' : 'space-y-3';
  const lbl = compact ? 'text-[10px] mb-0.5' : 'text-xs mb-1';
  const inp = compact
    ? 'px-2 py-1.5 rounded-md bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-600 text-slate-900 dark:text-white text-xs'
    : 'px-3 py-2 rounded-lg bg-slate-800 border border-slate-600 text-white text-sm';
  const optRows = compact ? 2 : 6;

  return (
    <div
      className={`rounded-lg border border-slate-300 dark:border-slate-600/50 bg-white dark:bg-slate-800/60 flex ${pad}`}
    >
      <div className={`flex flex-col text-slate-400 cursor-grab shrink-0 ${compact ? 'pt-1 text-xs' : 'pt-2'}`}>⋮</div>
      <div className={`flex-1 min-w-0 ${space}`}>
        {compact ? (
          <div className="grid grid-cols-1 sm:grid-cols-[1fr_10rem] gap-1.5">
            <div>
              <label className={`block text-slate-500 dark:text-slate-400 ${lbl}`}>Question</label>
              <input
                type="text"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                onBlur={blurTitle}
                className={`w-full ${inp}`}
                placeholder="What is your name?"
              />
            </div>
            <div>
              <label className={`block text-slate-500 dark:text-slate-400 ${lbl}`}>Type</label>
              <select
                value={type}
                onChange={(e) => {
                  const next = e.target.value;
                  setType(next);
                  onUpdate({ type: next });
                }}
                className={`w-full ${inp}`}
              >
                {typeOptions.map((t) => (
                  <option key={t.value} value={t.value}>{t.label}</option>
                ))}
              </select>
            </div>
          </div>
        ) : (
          <>
            <div>
              <label className={`block text-slate-500 dark:text-slate-400 ${lbl}`}>Question</label>
              <input
                type="text"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                onBlur={blurTitle}
                className={`w-full ${inp}`}
                placeholder="What is your name?"
              />
            </div>
            <div>
              <label className={`block text-slate-500 dark:text-slate-400 ${lbl}`}>Question type</label>
              <select
                value={type}
                onChange={(e) => {
                  const next = e.target.value;
                  setType(next);
                  onUpdate({ type: next });
                }}
                className={`w-full ${inp}`}
              >
                {typeOptions.map((t) => (
                  <option key={t.value} value={t.value}>{t.label}</option>
                ))}
              </select>
            </div>
          </>
        )}
        {needsOptions && (
          <div>
            <label className={`block text-slate-500 dark:text-slate-400 ${lbl}`}>Options (value:label / line)</label>
            <textarea
              value={optionsStr}
              onChange={(e) => setOptionsStr(e.target.value)}
              onBlur={applyOptions}
              rows={optRows}
              className={`w-full font-mono resize-y ${compact ? 'min-h-14 max-h-40' : 'min-h-[8rem] max-h-72'} ${inp} ${compact ? 'text-[11px]' : ''}`}
              placeholder="1:Male&#10;2:Female"
            />
          </div>
        )}
        <QuestionPromptMediaEditor
          question={question}
          compact={compact}
          lbl={lbl}
          onReplaced={onQuestionReplaced}
        />
        <QuestionDefaultValueFields
          questionType={type}
          config={question.config}
          value={question.config?.default_value}
          compact={compact}
          onConfigChange={(cfg) => onUpdate({ config: cfg })}
        />
        <div className="flex items-center gap-2">
          <span className={compact ? 'text-[10px] text-slate-500 dark:text-slate-400' : 'text-sm text-slate-400'}>
            Required
          </span>
          <button
            type="button"
            role="switch"
            aria-checked={required}
            onClick={() => {
              const next = !required;
              setRequired(next);
              onUpdate({ required: next });
            }}
            className={`relative rounded-full transition-colors shrink-0 ${compact ? 'w-9 h-5' : 'w-10 h-5'} ${required ? 'bg-violet-600' : 'bg-slate-400 dark:bg-slate-600'}`}
          >
            <span
              className={`absolute top-0.5 w-4 h-4 rounded-full bg-white transition-transform ${required ? (compact ? 'left-4' : 'left-5') : 'left-0.5'}`}
            />
          </button>
        </div>

        <div className={`border-t border-slate-200 dark:border-slate-600/50 ${compact ? 'pt-1.5 mt-1.5' : 'pt-3 mt-3'}`}>
          <label className={`block text-slate-500 dark:text-slate-400 ${lbl}`}>Visibility rules</label>
          <p className={`text-slate-500 dark:text-slate-500 ${compact ? 'text-[10px] mb-1' : 'text-xs mb-2'}`}>
            AND all rules to show this question.
          </p>
          {rules.length > 0 && (
            <ul className={`mb-1.5 ${compact ? 'space-y-0.5' : 'space-y-1'}`}>
              {rules.map((r) => (
                <li
                  key={r.id}
                  className={`flex items-center justify-between gap-2 text-slate-700 dark:text-slate-300 bg-slate-100 dark:bg-slate-700/50 rounded px-2 ${compact ? 'py-0.5 text-[11px]' : 'py-1 text-sm'}`}
                >
                  <span className="min-w-0 truncate">
                    When &quot;{getQuestionTitle(r.depends_on_question_id)}&quot; is{' '}
                    {r.condition === 'answered' ? 'answered' : 'not answered'}
                  </span>
                  <button
                    type="button"
                    onClick={() => onDeleteRule(r.id)}
                    className="text-slate-400 hover:text-red-400 shrink-0"
                    title="Remove rule"
                  >
                    ×
                  </button>
                </li>
              ))}
            </ul>
          )}
          {otherQuestions.length > 0 && (
            <div className={`flex flex-wrap items-center ${compact ? 'gap-1' : 'gap-2'}`}>
              <select
                value={addDependsOn}
                onChange={(e) => setAddDependsOn(Number(e.target.value))}
                className={`rounded-md bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-600 text-slate-900 dark:text-white ${compact ? 'px-1.5 py-1 text-[11px]' : 'px-2 py-1.5 text-sm'}`}
              >
                <option value={0}>Select question…</option>
                {otherQuestions.map((q) => (
                  <option key={q.id} value={q.id}>{q.title || `#${q.id}`}</option>
                ))}
              </select>
              <select
                value={addCondition}
                onChange={(e) => setAddCondition(e.target.value as 'answered' | 'not_answered')}
                className={`rounded-md bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-600 text-slate-900 dark:text-white ${compact ? 'px-1.5 py-1 text-[11px]' : 'px-2 py-1.5 text-sm'}`}
              >
                <option value="answered">is answered</option>
                <option value="not_answered">is not answered</option>
              </select>
              <button
                type="button"
                onClick={handleAddRule}
                disabled={!canAddRule}
                className={`rounded-md bg-violet-600 hover:bg-violet-500 text-white disabled:opacity-50 ${compact ? 'px-2 py-1 text-[11px]' : 'px-2 py-1.5 text-sm'}`}
              >
                Add rule
              </button>
            </div>
          )}
        </div>
      </div>
      <button
        type="button"
        onClick={onDelete}
        className={`rounded-lg text-slate-400 hover:bg-red-900/30 hover:text-red-400 shrink-0 h-fit ${compact ? 'p-1.5' : 'p-2'}`}
        title="Delete question"
      >
        🗑
      </button>
    </div>
  );
}
