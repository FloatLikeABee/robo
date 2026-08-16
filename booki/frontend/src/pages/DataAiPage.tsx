import { useEffect, useMemo, useRef, useState } from 'react';
import { runAiProgress } from '@robo/platform-chat/aiProgress';
import { api } from '../api/client';
import { useAuth } from '../store/auth';

type ChatMessage = { role: 'user' | 'assistant'; content: string };
type AssistantState = { intent?: string; fields?: Record<string, string> };
type AssistantResponse = {
  assistant_message?: string;
  reply?: string;
  state?: AssistantState;
};

type ReferenceDoc = {
  id: string;
  name: string;
  size: number;
  kind: 'pdf' | 'txt' | 'json';
  uploadedAt: string;
  content: string;
};

const STORAGE_KEY = 'booki-data-ai-references-v1';
const QUICK_PROMPTS = [
  'Give me this month profit and loss summary.',
  'Run trial balance as of today and check if balanced.',
  'Analyze account 1000 for this month.',
  'List unusual accounting risks based on my references.',
];

const INTRO_MESSAGE =
  'I am **Morph Booki AI** for bookings and flow. I can help with bookings, accounts, and operational flow analysis.';

function formatBytes(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes <= 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB'];
  let value = bytes;
  let unitIdx = 0;
  while (value >= 1024 && unitIdx < units.length - 1) {
    value /= 1024;
    unitIdx += 1;
  }
  return `${value.toFixed(unitIdx === 0 ? 0 : 1)} ${units[unitIdx]}`;
}

function normalizeText(input: string, maxChars = 40000): string {
  const text = input.replace(/\u0000/g, '').trim();
  if (text.length <= maxChars) return text;
  return `${text.slice(0, maxChars)}\n\n[truncated to ${maxChars} chars]`;
}

async function parsePdfText(file: File): Promise<string> {
  const pdfjs = await import('pdfjs-dist');
  const raw = await file.arrayBuffer();
  const loadingTask = pdfjs.getDocument({ data: raw, disableWorker: true } as any);
  const pdf = await loadingTask.promise;
  const pages: string[] = [];
  const maxPages = Math.min(pdf.numPages, 24);
  for (let p = 1; p <= maxPages; p += 1) {
    const page = await pdf.getPage(p);
    const textContent = await page.getTextContent();
    const line = textContent.items
      .map((item) => ('str' in item ? item.str : ''))
      .join(' ')
      .replace(/\s+/g, ' ')
      .trim();
    if (line) pages.push(`[Page ${p}] ${line}`);
  }
  const merged = pages.join('\n');
  if (!merged.trim()) throw new Error('No readable text found in PDF.');
  return normalizeText(merged);
}

async function parseReferenceFile(file: File): Promise<{ kind: ReferenceDoc['kind']; content: string }> {
  const name = file.name.toLowerCase();
  if (name.endsWith('.pdf') || file.type.includes('pdf')) {
    return { kind: 'pdf', content: await parsePdfText(file) };
  }
  if (name.endsWith('.json') || file.type.includes('json')) {
    const raw = await file.text();
    try {
      return { kind: 'json', content: normalizeText(JSON.stringify(JSON.parse(raw), null, 2)) };
    } catch {
      return { kind: 'json', content: normalizeText(raw) };
    }
  }
  if (name.endsWith('.txt') || name.endsWith('.md') || file.type.startsWith('text/')) {
    return { kind: 'txt', content: normalizeText(await file.text()) };
  }
  throw new Error(`Unsupported file type for ${file.name}. Use PDF, TXT, or JSON.`);
}

function buildPrompt(userInput: string, selectedDocs: ReferenceDoc[]): string {
  const trimmed = userInput.trim();
  if (!selectedDocs.length) return trimmed;

  const cappedDocs = selectedDocs.slice(0, 6);
  const docBlocks = cappedDocs.map((doc) => {
    const maxDocChars = 6000;
    const clipped = doc.content.length > maxDocChars ? `${doc.content.slice(0, maxDocChars)}\n...[truncated]` : doc.content;
    return `Reference file: ${doc.name} (${doc.kind})\n${clipped}`;
  });

  return [
    'Accounting reference files provided by user (ground truth for this request):',
    ...docBlocks.map((b) => `---\n${b}`),
    '---',
    'Use these references together with live Booki ledger data when relevant.',
    `User request: ${trimmed}`,
  ].join('\n');
}

export function DataAiPage() {
  const token = useAuth((s) => s.accessToken);
  const [messages, setMessages] = useState<ChatMessage[]>([{ role: 'assistant', content: INTRO_MESSAGE }]);
  const [assistantState, setAssistantState] = useState<AssistantState>({});
  const [input, setInput] = useState('');
  const [loading, setLoading] = useState(false);
  const [loadingStatus, setLoadingStatus] = useState('');
  const [error, setError] = useState('');

  const [docs, setDocs] = useState<ReferenceDoc[]>([]);
  const [selectedDocIds, setSelectedDocIds] = useState<string[]>([]);
  const [importing, setImporting] = useState(false);
  const [importError, setImportError] = useState('');
  const scrollRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    try {
      const raw = localStorage.getItem(STORAGE_KEY);
      if (!raw) return;
      const parsed = JSON.parse(raw) as ReferenceDoc[];
      if (!Array.isArray(parsed)) return;
      const valid = parsed.filter((d) => d && typeof d.name === 'string' && typeof d.content === 'string');
      setDocs(valid);
      setSelectedDocIds(valid.map((d) => d.id));
    } catch {
      // Ignore broken local cache.
    }
  }, []);

  useEffect(() => {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(docs));
  }, [docs]);

  useEffect(() => {
    const el = scrollRef.current;
    if (!el) return;
    el.scrollTop = el.scrollHeight;
  }, [messages, loading]);

  const selectedDocs = useMemo(
    () => docs.filter((d) => selectedDocIds.includes(d.id)),
    [docs, selectedDocIds]
  );

  const sendMessage = async (seed?: string) => {
    const text = (seed ?? input).trim();
    if (!text || !token || loading) return;

    setError('');
    const userRow: ChatMessage = {
      role: 'user',
      content:
        selectedDocs.length > 0
          ? `${text}\n\n[Using references: ${selectedDocs.map((d) => d.name).join(', ')}]`
          : text,
    };
    const nextMessages = [...messages, userRow];
    setMessages(nextMessages);
    setInput('');
    setLoading(true);
    setLoadingStatus('Reading your question…');
    const stopProgress = runAiProgress(
      {
        app: 'booki',
        userText: text,
        hasReferenceDocs: selectedDocs.length > 0,
      },
      setLoadingStatus,
    );

    try {
      const promptWithRefs = buildPrompt(text, selectedDocs);
      const payloadMessages = [...nextMessages.slice(0, -1), { role: 'user' as const, content: promptWithRefs }];
      const res = await api<AssistantResponse>('/api/v1/assistant/chat', {
        method: 'POST',
        token,
        body: JSON.stringify({ messages: payloadMessages, state: assistantState }),
      });
      const assistantText = (res.assistant_message ?? res.reply ?? '').trim() || 'Done.';
      setMessages((prev) => [...prev, { role: 'assistant', content: assistantText }]);
      setAssistantState(res.state ?? {});
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to reach Booki AI.';
      setError(message);
      setMessages((prev) => [...prev, { role: 'assistant', content: `**Error:** ${message}` }]);
    } finally {
      stopProgress();
      setLoading(false);
      setLoadingStatus('');
    }
  };

  const onUpload = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(event.target.files ?? []);
    event.target.value = '';
    if (!files.length) return;
    setImporting(true);
    setImportError('');
    try {
      const imported: ReferenceDoc[] = [];
      for (const file of files) {
        const parsed = await parseReferenceFile(file);
        imported.push({
          id: `${Date.now()}-${Math.random().toString(36).slice(2)}`,
          name: file.name,
          size: file.size,
          kind: parsed.kind,
          uploadedAt: new Date().toISOString(),
          content: parsed.content,
        });
      }
      setDocs((prev) => [...imported, ...prev].slice(0, 40));
      setSelectedDocIds((prev) => [...new Set([...prev, ...imported.map((d) => d.id)])]);
    } catch (err) {
      setImportError(err instanceof Error ? err.message : 'Failed to import references.');
    } finally {
      setImporting(false);
    }
  };

  const toggleDoc = (id: string) => {
    setSelectedDocIds((prev) => (prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]));
  };

  return (
    <div className="h-full min-h-0 flex flex-col gap-3">
      <div className="grid flex-1 min-h-0 gap-4 xl:grid-cols-[320px_minmax(0,1fr)]">
        <section className="h-full min-h-0 rounded-[22px] border border-white/10 bg-card/85 p-4 md:p-5 space-y-4 flex flex-col">
          <div className="flex items-center justify-between gap-2">
            <h2 className="text-sm font-semibold tracking-wide">Reference Library</h2>
            <span className="text-xs text-muted">{docs.length} files</span>
          </div>
          <p className="text-xs text-muted">
            Import accounting references (`.pdf`, `.txt`, `.json`) and tick what should be included in the next AI request.
          </p>
          <label className="block">
            <input
              type="file"
              multiple
              accept=".pdf,.txt,.md,.json,application/pdf,text/plain,application/json"
              className="hidden"
              onChange={onUpload}
              disabled={importing}
            />
            <span className="inline-flex w-full cursor-pointer items-center justify-center rounded-xl border border-violet/35 bg-violet/20 px-3 py-2 text-sm font-medium text-text hover:bg-violet/30">
              {importing ? 'Importing…' : 'Import reference files'}
            </span>
          </label>
          {importError ? <p className="text-xs text-danger">{importError}</p> : null}
          <div className="flex flex-wrap gap-2 text-[11px]">
            <button
              type="button"
              className="rounded-lg border border-white/15 px-2 py-1 text-muted hover:text-text"
              onClick={() => setSelectedDocIds(docs.map((d) => d.id))}
            >
              Include all
            </button>
            <button
              type="button"
              className="rounded-lg border border-white/15 px-2 py-1 text-muted hover:text-text"
              onClick={() => setSelectedDocIds([])}
            >
              Clear selection
            </button>
            <button
              type="button"
              className="rounded-lg border border-rose-400/35 px-2 py-1 text-rose-300 hover:text-rose-200"
              onClick={() => {
                setDocs([]);
                setSelectedDocIds([]);
              }}
            >
              Reset library
            </button>
          </div>
          <div className="flex-1 min-h-0 space-y-2 overflow-y-auto pr-1">
            {docs.map((doc) => {
              const included = selectedDocIds.includes(doc.id);
              return (
                <label
                  key={doc.id}
                  className={`block rounded-xl border p-3 transition-colors ${included ? 'border-violet/45 bg-violet/10' : 'border-white/10 bg-bg/35'}`}
                >
                  <div className="flex items-start gap-2">
                    <input
                      type="checkbox"
                      checked={included}
                      onChange={() => toggleDoc(doc.id)}
                      className="mt-1 accent-yellow"
                    />
                    <div className="min-w-0">
                      <p className="truncate text-sm font-medium text-text">{doc.name}</p>
                      <p className="mt-0.5 text-[11px] uppercase tracking-wide text-muted">
                        {doc.kind} · {formatBytes(doc.size)}
                      </p>
                    </div>
                  </div>
                </label>
              );
            })}
            {docs.length === 0 ? <p className="text-xs text-muted">No references imported yet.</p> : null}
          </div>
        </section>

        <section className="h-full min-h-0 rounded-[22px] border border-white/10 bg-card/85 p-4 md:p-5 flex flex-col">
          <div className="flex flex-wrap items-center justify-between gap-2 border-b border-white/10 pb-3">
            <div>
              <h2 className="text-base font-semibold">Booki AI Chat</h2>
              <p className="text-xs text-muted">Accounting only · Morph Booki scoped assistant</p>
            </div>
            <div className="text-xs text-muted">{selectedDocs.length} references selected</div>
          </div>

          <div className="mt-3 flex flex-wrap gap-2">
            {QUICK_PROMPTS.map((prompt) => (
              <button
                key={prompt}
                type="button"
                onClick={() => setInput(prompt)}
                className="rounded-full border border-white/12 bg-white/[0.03] px-3 py-1 text-xs text-muted hover:text-text hover:bg-white/[0.07]"
              >
                {prompt}
              </button>
            ))}
          </div>

          <div ref={scrollRef} className="mt-4 flex-1 min-h-0 overflow-y-auto space-y-3 rounded-2xl border border-white/8 bg-[#1a1a23] p-3 md:p-4">
            {messages.map((m, idx) => (
              <div key={`${m.role}-${idx}`} className={`flex ${m.role === 'user' ? 'justify-end' : 'justify-start'}`}>
                <div
                  className={`max-w-[92%] rounded-2xl px-3 py-2.5 text-sm leading-relaxed whitespace-pre-wrap ${
                    m.role === 'user'
                      ? 'bg-violet/35 border border-violet/40 text-text'
                      : 'bg-white/[0.04] border border-white/10 text-text'
                  }`}
                >
                  {m.content}
                </div>
              </div>
            ))}
            {loading ? <p className="text-xs text-muted">{loadingStatus || 'Working…'}</p> : null}
            {error ? <p className="text-xs text-danger">{error}</p> : null}
          </div>

          {selectedDocs.length > 0 ? (
            <div className="mt-3 flex flex-wrap gap-2">
              {selectedDocs.map((doc) => (
                <span
                  key={doc.id}
                  className="inline-flex items-center rounded-full border border-violet/35 bg-violet/10 px-2.5 py-1 text-[11px] text-violet-bright"
                  title={doc.name}
                >
                  {doc.name}
                </span>
              ))}
            </div>
          ) : null}

          <div className="mt-3 flex items-end gap-2">
            <textarea
              value={input}
              onChange={(e) => setInput(e.target.value)}
              placeholder="Ask Booki AI about bookings, flow log, or operational analysis…"
              className="min-h-[88px] flex-1 rounded-2xl border border-white/10 bg-bg/60 px-3 py-2.5 text-sm text-text placeholder:text-muted focus:border-violet/45 focus:outline-none"
              disabled={!token || loading}
            />
            <button
              type="button"
              onClick={() => void sendMessage()}
              disabled={!token || loading || !input.trim()}
              className="h-11 rounded-xl bg-violet px-5 text-sm font-medium text-white disabled:opacity-45"
            >
              Send
            </button>
          </div>
        </section>
      </div>
    </div>
  );
}
