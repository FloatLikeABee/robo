import { useCallback, useEffect, useRef, useState, type FormEvent } from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { useParams } from 'react-router-dom';
import { api, type AISheetUIBlock } from '../lib/api';

type ChatMsg = {
  role: 'user' | 'assistant';
  content: string;
  ui_blocks?: AISheetUIBlock[];
};

function sessionKey(slug: string) {
  return `sheetx_ai_sheet_session_${slug}`;
}

export function PublicAISheet() {
  const { slug = '' } = useParams();
  const [title, setTitle] = useState('AI Survey');
  const [metaError, setMetaError] = useState<string | null>(null);
  const [messages, setMessages] = useState<ChatMsg[]>([]);
  const [input, setInput] = useState('');
  const [busy, setBusy] = useState(false);
  const [state, setState] = useState<unknown>(null);
  const [done, setDone] = useState(false);
  const bottomRef = useRef<HTMLDivElement>(null);
  const sessionId = useRef('');

  useEffect(() => {
    document.documentElement.classList.add('public-form-page');
    return () => document.documentElement.classList.remove('public-form-page');
  }, []);

  useEffect(() => {
    try {
      let id = localStorage.getItem(sessionKey(slug));
      if (!id) {
        id = crypto.randomUUID();
        localStorage.setItem(sessionKey(slug), id);
      }
      sessionId.current = id;
    } catch {
      sessionId.current = `anon-${Date.now()}`;
    }
  }, [slug]);

  const scrollDown = () => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  const send = useCallback(
    async (message: string, nextState?: unknown) => {
      if (!slug || busy) return;
      setBusy(true);
      setMetaError(null);
      try {
        const res = await api.publicAISheet.chat(slug, {
          session_id: sessionId.current,
          message,
          state: nextState ?? state,
        });
        setState(res.state ?? null);
        setDone(Boolean(res.done));
        if (res.title) setTitle(res.title);
        setMessages((prev) => [
          ...prev,
          ...(message && message !== 'start'
            ? [{ role: 'user' as const, content: message }]
            : []),
          {
            role: 'assistant',
            content: res.message || '',
            ui_blocks: res.ui_blocks,
          },
        ]);
      } catch (e) {
        setMetaError(e instanceof Error ? e.message : 'Chat failed');
      } finally {
        setBusy(false);
        window.setTimeout(scrollDown, 50);
      }
    },
    [slug, busy, state]
  );

  useEffect(() => {
    if (!slug) return;
    let cancelled = false;
    (async () => {
      try {
        const meta = await api.publicAISheet.get(slug);
        if (cancelled) return;
        setTitle(meta.title || slug);
        await send('start', null);
      } catch (e) {
        if (!cancelled) setMetaError(e instanceof Error ? e.message : 'Sheet not found');
      }
    })();
    return () => {
      cancelled = true;
    };
    // start once per slug
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [slug]);

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault();
    const text = input.trim();
    if (!text || busy || done) return;
    setInput('');
    await send(text);
  };

  const onPick = async (block: AISheetUIBlock, value: string) => {
    const field = block.submit_as?.field || block.id;
    await send(`survey_bot_answer:${field}=${value}`);
  };

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col">
      <header className="border-b border-white/10 px-4 py-3 flex items-center justify-between gap-3">
        <div>
          <p className="text-[11px] uppercase tracking-[0.22em] text-violet-300/80">SurveyX · AI Survey</p>
          <h1 className="text-lg font-semibold">{title}</h1>
        </div>
        {done ? (
          <button
            type="button"
            className="text-sm rounded-lg px-3 py-1.5 bg-violet-600 hover:bg-violet-500"
            onClick={() => {
              setMessages([]);
              setState(null);
              setDone(false);
              void send('restart', null);
            }}
          >
            Restart
          </button>
        ) : null}
      </header>

      {metaError ? (
        <div className="m-4 rounded-lg border border-rose-500/40 bg-rose-950/40 px-3 py-2 text-sm text-rose-100">
          {metaError}
        </div>
      ) : null}

      <main className="flex-1 overflow-auto px-4 py-4 space-y-3 max-w-2xl w-full mx-auto">
        {messages.map((m, i) => (
          <div
            key={`${m.role}-${i}`}
            className={
              'rounded-2xl px-3.5 py-2.5 text-sm ' +
              (m.role === 'user'
                ? 'ml-8 bg-violet-700/40 border border-violet-500/30 whitespace-pre-wrap'
                : 'mr-8 bg-white/5 border border-white/10')
            }
          >
            {m.role === 'assistant' ? (
              <div className="ai-sheet-md">
                <ReactMarkdown remarkPlugins={[remarkGfm]}>{m.content || ''}</ReactMarkdown>
              </div>
            ) : (
              m.content
            )}
            {m.ui_blocks && m.ui_blocks.length > 0 && i === messages.length - 1 && !done ? (
              <div className="mt-3 flex flex-wrap gap-2">
                {m.ui_blocks.flatMap((block) =>
                  (block.options || []).map((opt) => (
                    <button
                      key={`${block.id}-${opt.value}`}
                      type="button"
                      disabled={busy}
                      onClick={() => void onPick(block, opt.value)}
                      className="rounded-lg px-3 py-1.5 text-sm bg-sky-600 hover:bg-sky-500 disabled:opacity-50"
                    >
                      {opt.label}
                    </button>
                  ))
                )}
              </div>
            ) : null}
          </div>
        ))}
        <div ref={bottomRef} />
      </main>

      <form
        onSubmit={onSubmit}
        className="border-t border-white/10 p-3 max-w-2xl w-full mx-auto flex gap-2"
      >
        <input
          className="flex-1 rounded-xl bg-black/40 border border-white/10 px-3 py-2 text-sm outline-none focus:border-violet-500/50"
          placeholder={done ? 'Survey complete' : 'Type your answer…'}
          value={input}
          disabled={busy || done}
          onChange={(e) => setInput(e.target.value)}
        />
        <button
          type="submit"
          disabled={busy || done || !input.trim()}
          className="rounded-xl px-4 py-2 text-sm font-medium bg-violet-600 hover:bg-violet-500 disabled:opacity-50"
        >
          Send
        </button>
      </form>
    </div>
  );
}
