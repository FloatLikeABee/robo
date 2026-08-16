import { useEffect, useMemo, useState } from 'react';
import { api } from '../api/client';
import { useAuth } from '../store/auth';
import { runAiProgress } from '@robo/platform-chat/aiProgress';

type FlowEntry = {
  id: number;
  entry_date: string;
  direction: 'income' | 'expense';
  amount: number;
  currency: string;
  category: string;
  status: string;
  title: string;
  notes: string;
  tags: string[];
};

type FlowSummary = {
  from: string;
  to: string;
  income: number;
  expense: number;
  net: number;
  entry_count: number;
  by_category: Record<string, number>;
  by_status: Record<string, number>;
};

const STATUS_SUGGESTIONS = ['logged', 'pending', 'cleared', 'reimbursed', 'planned', 'ignored'];
const CATEGORY_SUGGESTIONS = ['Food', 'Travel', 'Salary', 'Rent', 'Software', 'Marketing', 'Other'];

const AI_PROMPTS = [
  'Summarize my spending and income this month.',
  'Which categories drove the most expense?',
  'Flag unusual or large entries.',
];

function todayISO(): string {
  return new Date().toISOString().slice(0, 10);
}

function monthStartISO(): string {
  const d = new Date();
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-01`;
}

export function FlowLogPage() {
  const token = useAuth((s) => s.accessToken);
  const [entries, setEntries] = useState<FlowEntry[]>([]);
  const [summary, setSummary] = useState<FlowSummary | null>(null);
  const [msg, setMsg] = useState('');
  const [filterFrom, setFilterFrom] = useState(monthStartISO());
  const [filterTo, setFilterTo] = useState(todayISO());

  const [entryDate, setEntryDate] = useState(todayISO());
  const [direction, setDirection] = useState<'income' | 'expense'>('expense');
  const [amount, setAmount] = useState('');
  const [title, setTitle] = useState('');
  const [category, setCategory] = useState('');
  const [status, setStatus] = useState('logged');
  const [notes, setNotes] = useState('');

  const [aiPrompt, setAiPrompt] = useState('');
  const [aiReport, setAiReport] = useState('');
  const [aiLoading, setAiLoading] = useState(false);
  const [aiProgressStatus, setAiProgressStatus] = useState('');

  async function load() {
    if (!token) return;
    const q = new URLSearchParams({ from: filterFrom, to: filterTo, limit: '300' });
    const [listRes, sumRes] = await Promise.all([
      api<{ entries: FlowEntry[] | null }>(`/api/v1/flow-log/entries?${q}`, { token }),
      api<FlowSummary>(`/api/v1/flow-log/summary?from=${encodeURIComponent(filterFrom)}&to=${encodeURIComponent(filterTo)}`, {
        token,
      }),
    ]);
    setEntries(listRes.entries ?? []);
    setSummary(sumRes);
  }

  useEffect(() => {
    load()
      .then(() => setMsg(''))
      .catch((e) => setMsg(e instanceof Error ? e.message : 'Failed to load Flow Log.'));
  }, [token, filterFrom, filterTo]);

  const topCategories = useMemo(() => {
    if (!summary?.by_category) return [];
    return Object.entries(summary.by_category)
      .sort((a, b) => Math.abs(b[1]) - Math.abs(a[1]))
      .slice(0, 6);
  }, [summary]);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    if (!token) return;
    setMsg('');
    const amt = parseFloat(amount);
    if (!Number.isFinite(amt) || amt <= 0) {
      setMsg('Enter a positive amount.');
      return;
    }
    try {
      await api('/api/v1/flow-log/entries', {
        method: 'POST',
        token,
        body: JSON.stringify({
          entry_date: entryDate,
          direction,
          amount: amt,
          title: title.trim() || (direction === 'income' ? 'Income' : 'Expense'),
          category: category.trim(),
          status: status.trim() || 'logged',
          notes: notes.trim(),
        }),
      });
      setAmount('');
      setTitle('');
      setNotes('');
      await load();
    } catch (err) {
      setMsg(err instanceof Error ? err.message : 'Failed to save entry.');
    }
  }

  async function removeEntry(id: number) {
    if (!token || !window.confirm('Delete this Flow Log entry?')) return;
    try {
      await api(`/api/v1/flow-log/entries/${id}`, { method: 'DELETE', token });
      await load();
    } catch (err) {
      setMsg(err instanceof Error ? err.message : 'Delete failed.');
    }
  }

  async function runAnalysis(seed?: string) {
    if (!token) return;
    const prompt = (seed ?? aiPrompt).trim();
    if (!prompt) return;
    setAiLoading(true);
    setAiProgressStatus('Reading your question…');
    const stopProgress = runAiProgress({ app: 'booki', userText: prompt }, setAiProgressStatus);
    setMsg('');
    try {
      const res = await api<{ report: string; ai_enabled?: boolean }>('/api/v1/flow-log/analyze', {
        method: 'POST',
        token,
        body: JSON.stringify({ from: filterFrom, to: filterTo, prompt }),
      });
      setAiReport(res.report ?? '');
      if (!seed) setAiPrompt(prompt);
    } catch (err) {
      setMsg(err instanceof Error ? err.message : 'Analysis failed.');
    } finally {
      stopProgress();
      setAiLoading(false);
      setAiProgressStatus('');
    }
  }

  return (
    <div className="h-full min-h-0 flex flex-col gap-3">
      <div className="shrink-0 flex flex-wrap items-center justify-between gap-x-4 gap-y-2 border-b border-white/8 pb-2">
        <div className="min-w-0">
          <h1 className="text-base font-semibold text-text leading-tight">Flow Log</h1>
          <p className="text-xs text-muted leading-snug">
            Free-form money notes — no double-entry rules. You choose status; AI can analyze anytime.
          </p>
        </div>
        <div className="flex flex-wrap items-center gap-2 text-xs shrink-0">
          <div className="rounded-lg border border-white/8 bg-card/80 px-2.5 py-1.5">
            <span className="text-muted">Income </span>
            <span className="tabular-nums font-medium text-emerald-300">{(summary?.income ?? 0).toFixed(2)}</span>
          </div>
          <div className="rounded-lg border border-white/8 bg-card/80 px-2.5 py-1.5">
            <span className="text-muted">Expense </span>
            <span className="tabular-nums font-medium text-rose-300">{(summary?.expense ?? 0).toFixed(2)}</span>
          </div>
          <div className="rounded-lg border border-white/8 bg-card/80 px-2.5 py-1.5">
            <span className="text-muted">Net </span>
            <span className="tabular-nums font-medium text-yellow">{(summary?.net ?? 0).toFixed(2)}</span>
          </div>
        </div>
      </div>

      {msg ? <p className="shrink-0 text-danger text-xs">{msg}</p> : null}

      <div className="flex-1 min-h-0 grid gap-4 xl:grid-cols-[minmax(0,1fr)_300px] overflow-hidden">
        <div className="min-h-0 flex flex-col gap-3 overflow-hidden">
          <form onSubmit={submit} className="shrink-0 rounded-2xl bg-card border border-white/5 p-3 space-y-2">
            <div className="flex flex-wrap gap-2 items-end">
              <div>
                <label className="text-[11px] text-muted block mb-0.5">Date</label>
                <input
                  type="date"
                  className="rounded-lg bg-bg border border-white/10 px-2 py-1.5 text-sm"
                  value={entryDate}
                  onChange={(e) => setEntryDate(e.target.value)}
                  required
                />
              </div>
              <div>
                <label className="text-[11px] text-muted block mb-0.5">Type</label>
                <select
                  className="rounded-lg bg-bg border border-white/10 px-2 py-1.5 text-sm"
                  value={direction}
                  onChange={(e) => setDirection(e.target.value as 'income' | 'expense')}
                >
                  <option value="expense">Expense</option>
                  <option value="income">Income</option>
                </select>
              </div>
              <div>
                <label className="text-[11px] text-muted block mb-0.5">Amount</label>
                <input
                  className="rounded-lg bg-bg border border-white/10 px-2 py-1.5 text-sm w-24"
                  value={amount}
                  onChange={(e) => setAmount(e.target.value)}
                  type="number"
                  step="0.01"
                  min="0.01"
                  required
                />
              </div>
              <div className="min-w-[8rem] flex-1">
                <label className="text-[11px] text-muted block mb-0.5">Title</label>
                <input
                  className="rounded-lg bg-bg border border-white/10 px-2 py-1.5 text-sm w-full"
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                  placeholder="What was this?"
                />
              </div>
              <div>
                <label className="text-[11px] text-muted block mb-0.5">Category</label>
                <input
                  list="flow-categories"
                  className="rounded-lg bg-bg border border-white/10 px-2 py-1.5 text-sm w-28"
                  value={category}
                  onChange={(e) => setCategory(e.target.value)}
                />
                <datalist id="flow-categories">
                  {CATEGORY_SUGGESTIONS.map((c) => (
                    <option key={c} value={c} />
                  ))}
                </datalist>
              </div>
              <div>
                <label className="text-[11px] text-muted block mb-0.5">Status</label>
                <input
                  list="flow-statuses"
                  className="rounded-lg bg-bg border border-white/10 px-2 py-1.5 text-sm w-24"
                  value={status}
                  onChange={(e) => setStatus(e.target.value)}
                />
                <datalist id="flow-statuses">
                  {STATUS_SUGGESTIONS.map((s) => (
                    <option key={s} value={s} />
                  ))}
                </datalist>
              </div>
              <div className="min-w-[8rem] flex-1">
                <label className="text-[11px] text-muted block mb-0.5">Notes</label>
                <input
                  className="rounded-lg bg-bg border border-white/10 px-2 py-1.5 text-sm w-full"
                  value={notes}
                  onChange={(e) => setNotes(e.target.value)}
                />
              </div>
              <button type="submit" className="rounded-lg bg-violet px-3 py-1.5 text-sm text-white">
                Add
              </button>
            </div>
          </form>

          <div className="shrink-0 flex flex-wrap items-end gap-2">
            <input
              type="date"
              aria-label="Filter from"
              className="rounded-lg bg-bg border border-white/10 px-2 py-1.5 text-sm"
              value={filterFrom}
              onChange={(e) => setFilterFrom(e.target.value)}
            />
            <span className="text-xs text-muted pb-1">→</span>
            <input
              type="date"
              aria-label="Filter to"
              className="rounded-lg bg-bg border border-white/10 px-2 py-1.5 text-sm"
              value={filterTo}
              onChange={(e) => setFilterTo(e.target.value)}
            />
            <span className="text-xs text-muted pb-1">{summary?.entry_count ?? 0} entries</span>
          </div>

          <div className="flex-1 min-h-0 rounded-2xl border border-white/5 overflow-auto">
            <table className="w-full text-sm">
              <thead className="bg-surface text-left text-muted text-xs uppercase sticky top-0 z-10">
                <tr>
                  <th className="p-2">Date</th>
                  <th className="p-2">Type</th>
                  <th className="p-2">Title</th>
                  <th className="p-2 hidden md:table-cell">Category</th>
                  <th className="p-2 hidden sm:table-cell">Status</th>
                  <th className="p-2 text-right">Amount</th>
                  <th className="p-2 w-12" />
                </tr>
              </thead>
              <tbody>
                {entries.map((e) => (
                  <tr key={e.id} className="border-t border-white/5">
                    <td className="p-2 whitespace-nowrap text-xs">{e.entry_date}</td>
                    <td className="p-2 capitalize text-xs">
                      <span className={e.direction === 'income' ? 'text-emerald-300' : 'text-rose-300'}>
                        {e.direction}
                      </span>
                    </td>
                    <td className="p-2">
                      <div className="text-sm">{e.title}</div>
                      {e.notes ? <div className="text-xs text-muted">{e.notes}</div> : null}
                    </td>
                    <td className="p-2 text-muted text-xs hidden md:table-cell">{e.category || '—'}</td>
                    <td className="p-2 hidden sm:table-cell">
                      <span className="rounded-full border border-white/10 px-1.5 py-0.5 text-[10px]">{e.status}</span>
                    </td>
                    <td className="p-2 text-right tabular-nums text-sm whitespace-nowrap">
                      {e.direction === 'expense' ? '−' : '+'}
                      {e.amount.toFixed(2)}
                    </td>
                    <td className="p-2">
                      <button
                        type="button"
                        className="text-[11px] text-muted hover:text-rose-300"
                        onClick={() => void removeEntry(e.id)}
                      >
                        ×
                      </button>
                    </td>
                  </tr>
                ))}
                {entries.length === 0 ? (
                  <tr>
                    <td colSpan={7} className="p-4 text-center text-muted text-sm">
                      No entries yet.
                    </td>
                  </tr>
                ) : null}
              </tbody>
            </table>
          </div>
        </div>

        <aside className="min-h-0 flex flex-col gap-2 overflow-hidden">
          {topCategories.length > 0 ? (
            <div className="shrink-0 rounded-2xl border border-white/5 bg-card p-3">
              <h3 className="text-xs font-semibold mb-2">Top categories</h3>
              <ul className="space-y-1 text-xs">
                {topCategories.map(([cat, val]) => (
                  <li key={cat} className="flex justify-between gap-2">
                    <span className="text-muted truncate">{cat}</span>
                    <span className="tabular-nums shrink-0">{val.toFixed(2)}</span>
                  </li>
                ))}
              </ul>
            </div>
          ) : null}

          <div className="flex-1 min-h-0 rounded-2xl border border-violet/25 bg-card p-3 flex flex-col overflow-hidden">
            <h3 className="text-xs font-semibold shrink-0">AI report</h3>
            <div className="mt-2 shrink-0 flex flex-wrap gap-1">
              {AI_PROMPTS.map((p) => (
                <button
                  key={p}
                  type="button"
                  disabled={aiLoading}
                  onClick={() => void runAnalysis(p)}
                  className="rounded-full border border-white/12 bg-white/[0.03] px-2 py-0.5 text-[10px] text-muted hover:text-text disabled:opacity-50"
                >
                  {p.length > 36 ? `${p.slice(0, 34)}…` : p}
                </button>
              ))}
            </div>
            <textarea
              value={aiPrompt}
              onChange={(e) => setAiPrompt(e.target.value)}
              placeholder="Ask about your Flow Log…"
              className="mt-2 shrink-0 min-h-[52px] rounded-lg border border-white/10 bg-bg/60 px-2 py-1.5 text-xs placeholder:text-muted"
            />
            <button
              type="button"
              disabled={aiLoading || !aiPrompt.trim()}
              onClick={() => void runAnalysis()}
              className="mt-1.5 shrink-0 rounded-lg bg-violet px-3 py-1.5 text-xs text-white disabled:opacity-45"
            >
              {aiLoading ? 'Analyzing…' : 'Generate report'}
            </button>
            {aiLoading && aiProgressStatus ? (
              <p className="mt-1.5 shrink-0 text-[10px] text-violet-200/90">{aiProgressStatus}</p>
            ) : null}
            <div className="mt-2 flex-1 min-h-0 overflow-y-auto rounded-lg border border-white/8 bg-[#1a1a23] p-2 text-xs whitespace-pre-wrap leading-relaxed">
              {aiReport || (aiLoading ? aiProgressStatus || 'Working…' : 'Reports appear here.')}
            </div>
          </div>
        </aside>
      </div>
    </div>
  );
}
