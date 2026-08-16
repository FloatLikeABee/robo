import { useCallback, useEffect, useRef, useState } from 'react';
import { runAiProgress } from '@robo/platform-chat/aiProgress';
import { api, type Question } from '../lib/api';

const API_BASE = import.meta.env.VITE_API_URL ?? '';

function formatResponseTime(ms: number | null) {
  if (ms == null) return '';
  if (ms < 1000) return `${ms} ms`;
  return `${(ms / 1000).toFixed(ms < 10000 ? 1 : 0)} s`;
}

export interface FormTemplateSource {
  title: string;
  type: string;
  url?: string;
}

export interface ProposedFormQuestion {
  title: string;
  type: string;
  required?: boolean;
  config?: Record<string, unknown>;
}

interface FormTemplateAIPanelProps {
  formId: number;
  formName: string;
  formDescription: string;
  /** Optional initial HTML shown in the preview (not persisted from this panel). */
  initialHtml?: string;
  onQuestionsApplied: (questions: Question[]) => void;
  activePageId: number | null;
  disabled?: boolean;
}

type AIMessage = {
  id: number;
  role: 'user' | 'assistant';
  content: string;
  proposedHtml?: string | null;
  proposedQuestions?: ProposedFormQuestion[];
  sources?: FormTemplateSource[];
};

function authHeaders(): Record<string, string> {
  const token = localStorage.getItem('tranform_auth_token') || sessionStorage.getItem('tranform_auth_token') || '';
  const h: Record<string, string> = { 'Content-Type': 'application/json' };
  if (token) h.Authorization = `Bearer ${token}`;
  return h;
}

export function FormTemplateAIPanel({
  formId,
  formName,
  formDescription,
  initialHtml = '',
  onQuestionsApplied,
  activePageId,
  disabled,
}: FormTemplateAIPanelProps) {
  const [open, setOpen] = useState(false);
  const [aiLoading, setAiLoading] = useState(false);
  const [aiLoadingStatus, setAiLoadingStatus] = useState('');
  const [responseTimeMs, setResponseTimeMs] = useState<number | null>(null);
  const [aiInput, setAiInput] = useState('');
  const [aiMessages, setAiMessages] = useState<AIMessage[]>([]);
  const [previewHtml, setPreviewHtml] = useState(initialHtml);
  const msgSeq = useRef(0);
  const logRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    setPreviewHtml(initialHtml);
  }, [initialHtml]);

  const endpoint = `${API_BASE}/api/v1/ai/form-template-chat`.replace(/([^:]\/)\/+/g, '$1');

  useEffect(() => {
    if (open) {
      logRef.current?.scrollTo({ top: logRef.current.scrollHeight, behavior: 'smooth' });
    }
  }, [aiMessages, open, aiLoading]);

  const runAI = useCallback(async (userMessage: string) => {
    setAiLoading(true);
    setResponseTimeMs(null);
    setAiLoadingStatus('Reading your question…');
    const startedAt = performance.now();
    const stopProgress = runAiProgress(
      { app: 'sheetx', userText: userMessage, webSearch: true },
      setAiLoadingStatus,
    );
    const userId = ++msgSeq.current;
    const nextMsgs: AIMessage[] = [...aiMessages, { id: userId, role: 'user', content: userMessage }];
    setAiMessages(nextMsgs);
    try {
      const res = await fetch(endpoint, {
        method: 'POST',
        headers: authHeaders(),
        body: JSON.stringify({
          messages: nextMsgs.map((m) => ({ role: m.role, content: m.content })),
          form_name: formName,
          form_description: formDescription,
          current_html: previewHtml,
          use_web_search: true,
        }),
      });
      if (!res.ok) {
        const err = await res.json().catch(() => ({}));
        throw new Error((err as { error?: string }).error || `HTTP ${res.status}`);
      }
      const data = await res.json();
      setResponseTimeMs(Math.round(performance.now() - startedAt));
      const assistant = String(data.assistant_message || 'Template ready.');
      const proposed = data.proposed_form_html ?? null;
      const sources = (data.sources || []) as FormTemplateSource[];
      if (proposed && String(proposed).trim()) {
        setPreviewHtml(String(proposed));
      }
      setAiMessages([
        ...nextMsgs,
        {
          id: ++msgSeq.current,
          role: 'assistant',
          content: assistant,
          proposedHtml: proposed,
          proposedQuestions: (data.proposed_questions || []) as ProposedFormQuestion[],
          sources,
        },
      ]);
    } catch (e) {
      setResponseTimeMs(Math.round(performance.now() - startedAt));
      setAiMessages([
        ...nextMsgs,
        {
          id: ++msgSeq.current,
          role: 'assistant',
          content: e instanceof Error ? e.message : 'AI request failed',
        },
      ]);
    } finally {
      stopProgress();
      setAiLoading(false);
      setAiLoadingStatus('');
    }
  }, [aiMessages, endpoint, formDescription, formName, previewHtml]);

  const applyQuestions = async (items: ProposedFormQuestion[]) => {
    if (!formId || !activePageId || items.length === 0) return;
    const created: Question[] = [];
    for (let i = 0; i < items.length; i++) {
      const pq = items[i];
      if (!pq.title || !pq.type) continue;
      try {
        const q = await api.questions.create(formId, {
          title: pq.title,
          type: pq.type,
          required: Boolean(pq.required),
          page_id: activePageId,
          sort_order: i,
          config: pq.config || {},
        });
        created.push(q);
      } catch {
        /* skip invalid */
      }
    }
    if (created.length > 0) {
      onQuestionsApplied(created);
    }
  };

  const sendAI = async () => {
    const msg = aiInput.trim();
    if (!msg || aiLoading || disabled) return;
    setAiInput('');
    await runAI(msg);
  };

  const showPreview = (html: string) => {
    const next = html.trim();
    if (!next) return;
    setPreviewHtml(next);
  };

  return (
    <div className="flex-shrink-0 rounded-lg bg-slate-100 dark:bg-slate-800/40 border border-slate-200 dark:border-slate-700/50 mb-2 overflow-hidden">
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className="w-full flex items-center justify-between gap-2 px-2.5 py-2 text-left"
      >
        <span className="text-xs font-medium text-slate-800 dark:text-white">
          ✦ AI form template (web search + HTML)
        </span>
        <span className="text-[10px] text-slate-500">{open ? 'Hide' : 'Show'}</span>
      </button>
      {open && (
        <div className="border-t border-slate-200 dark:border-slate-700/50 p-2 space-y-2">
          <p className="text-[10px] text-slate-500 dark:text-slate-400">
            Morph AI searches the web for examples, generates an HTML template preview, and suggests questions.
          </p>
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-2 min-h-[140px]">
            <div className="flex flex-col min-h-0">
              <label className="text-[10px] text-slate-500 mb-0.5">Template preview</label>
              <div className="flex-1 min-h-[120px] w-full rounded-md bg-white dark:bg-slate-900 border border-slate-300 dark:border-slate-600 overflow-hidden">
                {previewHtml.trim() ? (
                  <iframe
                    title="Form template preview"
                    sandbox="allow-same-origin"
                    srcDoc={previewHtml}
                    className="w-full h-[200px] border-0 bg-white"
                  />
                ) : (
                  <p className="p-3 text-[11px] text-slate-500 dark:text-slate-400">
                    Generate a template to see a live preview here.
                  </p>
                )}
              </div>
            </div>
            <div className="flex flex-col min-h-0">
              <div ref={logRef} className="flex-1 min-h-[100px] max-h-40 overflow-y-auto space-y-1.5 mb-1.5 p-1.5 rounded-md bg-slate-50 dark:bg-slate-900/40 border border-slate-200 dark:border-slate-700/50">
                {aiMessages.length === 0 && (
                  <p className="text-[10px] text-slate-500">Try: &quot;Create a patient intake form with contact fields&quot;</p>
                )}
                {aiMessages.map((m) => (
                  <div key={m.id} className={`text-[11px] ${m.role === 'user' ? 'text-slate-700 dark:text-slate-200' : 'text-slate-600 dark:text-slate-300'}`}>
                    <strong>{m.role === 'user' ? 'You' : 'AI'}:</strong> {m.content}
                    {m.proposedHtml && (
                      <button
                        type="button"
                        onClick={() => showPreview(m.proposedHtml!)}
                        className="ml-2 text-violet-600 dark:text-violet-400 underline text-[10px]"
                      >
                        Show preview
                      </button>
                    )}
                    {m.proposedQuestions && m.proposedQuestions.length > 0 && (
                      <button
                        type="button"
                        onClick={() => applyQuestions(m.proposedQuestions!)}
                        className="ml-2 text-violet-600 dark:text-violet-400 underline text-[10px]"
                      >
                        Apply {m.proposedQuestions.length} questions
                      </button>
                    )}
                    {m.sources && m.sources.length > 0 && (
                      <ul className="mt-0.5 text-[10px] text-slate-500 list-disc pl-4">
                        {m.sources.map((s, i) => (
                          <li key={i}>
                            {s.url ? (
                              <a href={s.url} target="_blank" rel="noreferrer" className="underline">
                                {s.title}
                              </a>
                            ) : (
                              s.title
                            )}
                          </li>
                        ))}
                      </ul>
                    )}
                  </div>
                ))}
                {aiLoading ? (
                  <p className="text-[10px] text-violet-600 dark:text-violet-400">{aiLoadingStatus || 'Working…'}</p>
                ) : null}
              </div>
              <div className="flex gap-1">
                <input
                  type="text"
                  value={aiInput}
                  onChange={(e) => setAiInput(e.target.value)}
                  onKeyDown={(e) => e.key === 'Enter' && !e.shiftKey && sendAI()}
                  disabled={disabled || aiLoading}
                  placeholder={disabled ? 'Create the form first…' : 'Describe the form to create…'}
                  className="flex-1 min-w-0 px-2 py-1 rounded-md bg-white dark:bg-slate-900 border border-slate-300 dark:border-slate-600 text-xs"
                />
                <button
                  type="button"
                  onClick={sendAI}
                  disabled={disabled || aiLoading || !aiInput.trim()}
                  className="px-2 py-1 rounded-md bg-violet-600 hover:bg-violet-500 text-white text-xs disabled:opacity-50"
                >
                  {aiLoading ? '…' : 'Generate'}
                </button>
              </div>
              {responseTimeMs != null && (
                <div className="mt-1 text-right text-[10px] text-violet-500 tabular-nums">
                  Response {formatResponseTime(responseTimeMs)}
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
