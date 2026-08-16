import { useCallback, useEffect, useMemo, useState } from 'react';
import { api } from '../api/client';
import { useAuth } from '../store/auth';

type Account = { id: number; code: string; name: string; type: string };
type JournalEntry = {
  id: number;
  reference: string;
  entry_date: string;
  description: string;
  status: string;
  source: string;
};
type TrialRow = { account_id: number; code: string; name: string; type: string; balance: number };

type TabId = 'ledger' | 'entries' | 'chart' | 'trial';

type EntriesPeriodMode = 'year' | 'month';

type LineDraft = { accountId: string; debit: string; credit: string; note: string };

function emptyLines(n: number): LineDraft[] {
  return Array.from({ length: n }, () => ({ accountId: '', debit: '', credit: '', note: '' }));
}

export function AccountingPage() {
  const token = useAuth((s) => s.accessToken);
  const [tab, setTab] = useState<TabId>('ledger');
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [glossary, setGlossary] = useState<{ term: string; definition: string }[]>([]);
  const [entries, setEntries] = useState<JournalEntry[]>([]);
  const [trialRows, setTrialRows] = useState<TrialRow[]>([]);
  const [asOf, setAsOf] = useState(() => new Date().toISOString().slice(0, 10));
  const [trialLoaded, setTrialLoaded] = useState(false);
  const [msg, setMsg] = useState('');

  const [jeDate, setJeDate] = useState(() => new Date().toISOString().slice(0, 10));
  const [jeRef, setJeRef] = useState('');
  const [jeDesc, setJeDesc] = useState('');
  const [lines, setLines] = useState<LineDraft[]>(() => emptyLines(2));
  const [glossaryOpen, setGlossaryOpen] = useState(false);

  const [entriesPeriodMode, setEntriesPeriodMode] = useState<EntriesPeriodMode>('month');
  const [entriesFilterYear, setEntriesFilterYear] = useState(() => new Date().getFullYear());
  const [entriesFilterMonth, setEntriesFilterMonth] = useState(() => new Date().toISOString().slice(0, 7));

  const loadAccounts = useCallback(async () => {
    if (!token) return;
    const r = await api<{ accounts: Account[] | null }>('/api/v1/accounts', { token });
    setAccounts(r.accounts ?? []);
  }, [token]);

  const loadGlossary = useCallback(async () => {
    if (!token) return;
    const r = await api<{ terms: { term: string; definition: string }[] | null }>('/api/v1/accounting/glossary', {
      token,
    });
    setGlossary(r.terms ?? []);
  }, [token]);

  const loadEntries = useCallback(async () => {
    if (!token) return;
    const r = await api<{ entries: JournalEntry[] | null }>('/api/v1/journal-entries', { token });
    setEntries(r.entries ?? []);
  }, [token]);

  const loadTrial = useCallback(async () => {
    if (!token) return;
    setMsg('');
    try {
      const r = await api<{ accounts: TrialRow[] | null; as_of?: string }>(
        `/api/v1/reports/trial-balance?as_of=${encodeURIComponent(asOf)}`,
        { token }
      );
      setTrialRows(r.accounts ?? []);
      setTrialLoaded(true);
    } catch (e) {
      setMsg(e instanceof Error ? e.message : 'Error');
    }
  }, [token, asOf]);

  useEffect(() => {
    loadAccounts().catch(() => {});
    loadGlossary().catch(() => {});
    loadEntries().catch(() => {});
  }, [loadAccounts, loadGlossary, loadEntries]);

  useEffect(() => {
    if (tab !== 'entries') return;
    loadEntries().catch(() => {});
  }, [tab, loadEntries]);

  useEffect(() => {
    if (tab !== 'trial') return;
    loadTrial().catch(() => {});
  }, [tab, loadTrial]);

  const recentEntries = useMemo(() => entries.slice(0, 10), [entries]);

  const entryYears = useMemo(() => {
    const y = new Set<number>();
    for (const en of entries) {
      const yy = parseInt(en.entry_date.slice(0, 4), 10);
      if (!Number.isNaN(yy)) y.add(yy);
    }
    y.add(new Date().getFullYear());
    y.add(entriesFilterYear);
    return Array.from(y).sort((a, b) => b - a);
  }, [entries, entriesFilterYear]);

  const filteredEntriesForTab = useMemo(() => {
    if (entriesPeriodMode === 'year') {
      return entries.filter((en) => parseInt(en.entry_date.slice(0, 4), 10) === entriesFilterYear);
    }
    return entries.filter((en) => en.entry_date.startsWith(entriesFilterMonth));
  }, [entries, entriesPeriodMode, entriesFilterYear, entriesFilterMonth]);

  useEffect(() => {
    if (!glossaryOpen) return;
    function onKey(e: KeyboardEvent) {
      if (e.key === 'Escape') setGlossaryOpen(false);
    }
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [glossaryOpen]);

  function updateLine(i: number, patch: Partial<LineDraft>) {
    setLines((prev) => prev.map((row, j) => (j === i ? { ...row, ...patch } : row)));
  }

  function addLine() {
    setLines((prev) => [...prev, { accountId: '', debit: '', credit: '', note: '' }]);
  }

  function removeLine(i: number) {
    setLines((prev) => (prev.length <= 2 ? prev : prev.filter((_, j) => j !== i)));
  }

  async function submitJournal(e: React.FormEvent) {
    e.preventDefault();
    if (!token) return;
    setMsg('');
    const parsed = lines.map((row) => ({
      account_id: row.accountId === '' ? 0 : Number(row.accountId),
      debit: parseFloat(row.debit) || 0,
      credit: parseFloat(row.credit) || 0,
      note: row.note.trim(),
    }));
    const valid = parsed.filter((ln) => ln.account_id > 0 && (ln.debit > 0 || ln.credit > 0));
    if (valid.length < 2) {
      setMsg('Add at least two lines with an account and either a debit or credit amount.');
      return;
    }
    let deb = 0;
    let cred = 0;
    for (const ln of valid) {
      if (ln.debit > 0 && ln.credit > 0) {
        setMsg('Each line must be either debit or credit, not both.');
        return;
      }
      deb += ln.debit;
      cred += ln.credit;
    }
    if (Math.abs(deb - cred) > 0.009) {
      setMsg(`Entry must balance (debits ${deb.toFixed(2)} ≠ credits ${cred.toFixed(2)}).`);
      return;
    }
    try {
      await api('/api/v1/journal-entries', {
        method: 'POST',
        token,
        body: JSON.stringify({
          reference: jeRef.trim(),
          entry_date: jeDate,
          description: jeDesc.trim(),
          status: 'posted',
          source: 'manual',
          lines: valid.map((ln) => ({
            account_id: ln.account_id,
            debit: ln.debit,
            credit: ln.credit,
            note: ln.note || undefined,
          })),
        }),
      });
      setJeRef('');
      setJeDesc('');
      setLines(emptyLines(2));
      await loadEntries();
      if (tab === 'trial') await loadTrial();
    } catch (err) {
      setMsg(err instanceof Error ? err.message : 'Error');
    }
  }

  let debSum = 0;
  let credSum = 0;
  for (const row of lines) {
    debSum += parseFloat(row.debit) || 0;
    credSum += parseFloat(row.credit) || 0;
  }
  const imbalance = Math.abs(debSum - credSum);

  const tabBtn = (id: TabId, label: string) => (
    <button
      key={id}
      type="button"
      onClick={() => setTab(id)}
      className={`px-4 py-2 rounded-xl text-sm font-medium transition-all ${
        tab === id ? 'bg-violet text-white shadow-[0_0_20px_rgba(91,63,214,0.25)]' : 'text-muted hover:text-text hover:bg-white/5'
      }`}
    >
      {label}
    </button>
  );

  return (
    <div className="flex flex-col gap-4 min-h-0">
      <div className="flex flex-wrap gap-1 p-1 rounded-2xl bg-surface/40 border border-white/5 w-fit max-w-full">
        {tabBtn('ledger', 'General ledger')}
        {tabBtn('entries', 'Entries')}
        {tabBtn('chart', 'Chart of accounts')}
        {tabBtn('trial', 'Trial balance')}
      </div>

      {msg && <p className="text-danger text-sm">{msg}</p>}

      {tab === 'ledger' && (
        <div className="space-y-4">
          <form onSubmit={submitJournal} className="rounded-[20px] bg-card border border-white/5 p-4 md:p-5 space-y-4">
            <p className="text-xs text-muted">Post a balanced manual journal (simple double-entry).</p>
            <div className="flex flex-wrap gap-3 items-end">
              <div>
                <label className="text-xs text-muted block mb-1">Date</label>
                <input
                  type="date"
                  required
                  value={jeDate}
                  onChange={(e) => setJeDate(e.target.value)}
                  className="rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm"
                />
              </div>
              <div className="flex-1 min-w-[140px]">
                <label className="text-xs text-muted block mb-1">Reference (optional)</label>
                <input
                  value={jeRef}
                  onChange={(e) => setJeRef(e.target.value)}
                  placeholder="e.g. INV-12"
                  className="w-full rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm"
                />
              </div>
              <div className="w-full min-w-[200px] flex-[2]">
                <label className="text-xs text-muted block mb-1">Description</label>
                <input
                  value={jeDesc}
                  onChange={(e) => setJeDesc(e.target.value)}
                  placeholder="What this entry records"
                  className="w-full rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm"
                />
              </div>
            </div>

            <div className="space-y-2">
              <div className="text-xs text-muted uppercase tracking-wide">Lines</div>
              {lines.map((row, i) => (
                <div key={i} className="flex flex-wrap gap-2 items-end">
                  <select
                    value={row.accountId}
                    onChange={(e) => updateLine(i, { accountId: e.target.value })}
                    className="flex-1 min-w-[200px] rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm"
                  >
                    <option value="">Account</option>
                    {accounts.map((a) => (
                      <option key={a.id} value={a.id}>
                        {a.code} — {a.name}
                      </option>
                    ))}
                  </select>
                  <input
                    type="number"
                    min={0}
                    step="0.01"
                    placeholder="Debit"
                    value={row.debit}
                    onChange={(e) => updateLine(i, { debit: e.target.value, credit: '' })}
                    className="w-28 rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm tabular-nums"
                  />
                  <input
                    type="number"
                    min={0}
                    step="0.01"
                    placeholder="Credit"
                    value={row.credit}
                    onChange={(e) => updateLine(i, { credit: e.target.value, debit: '' })}
                    className="w-28 rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm tabular-nums"
                  />
                  <input
                    placeholder="Note"
                    value={row.note}
                    onChange={(e) => updateLine(i, { note: e.target.value })}
                    className="flex-1 min-w-[120px] rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm"
                  />
                  {lines.length > 2 && (
                    <button type="button" onClick={() => removeLine(i)} className="text-xs text-muted hover:text-danger px-2 py-2">
                      Remove
                    </button>
                  )}
                </div>
              ))}
            </div>

            <div className="flex flex-wrap items-center justify-between gap-3">
              <div className="text-xs text-muted tabular-nums">
                Debits {debSum.toFixed(2)} · Credits {credSum.toFixed(2)}
                {imbalance > 0.009 && <span className="text-yellow ml-2">Off by {imbalance.toFixed(2)}</span>}
              </div>
              <div className="flex gap-2">
                <button type="button" onClick={addLine} className="rounded-xl border border-white/10 px-4 py-2 text-sm text-muted hover:text-text">
                  Add line
                </button>
                <button type="submit" className="rounded-xl bg-violet px-4 py-2 text-sm font-medium text-white">
                  Post journal
                </button>
              </div>
            </div>
          </form>

          <div className="min-h-0 shrink-0">
            <h2 className="text-sm font-medium text-muted mb-2 uppercase tracking-wide">Recent entries</h2>
            <div className="rounded-[20px] border border-white/5 max-h-[min(29rem,42vh)] overflow-x-auto overflow-y-auto overscroll-y-contain isolate [scrollbar-gutter:stable]">
              <table className="w-full text-sm min-w-[520px]">
                  <thead className="bg-surface text-left text-muted text-xs uppercase sticky top-0 z-[1] shadow-[inset_0_-1px_0_rgba(255,255,255,0.06)]">
                    <tr>
                      <th className="p-3">Date</th>
                      <th className="p-3">Ref</th>
                      <th className="p-3">Description</th>
                      <th className="p-3">Source</th>
                      <th className="p-3">Status</th>
                    </tr>
                  </thead>
                  <tbody>
                    {recentEntries.length === 0 && (
                      <tr>
                        <td colSpan={5} className="p-6 text-muted text-center">
                          No journal entries yet — post one above or post a booking from Bookings.
                        </td>
                      </tr>
                    )}
                    {recentEntries.map((en) => (
                      <tr key={en.id} className="border-t border-white/5">
                        <td className="p-3 whitespace-nowrap">{en.entry_date}</td>
                        <td className="p-3 font-mono text-xs">{en.reference || '—'}</td>
                        <td className="p-3">{en.description || '—'}</td>
                        <td className="p-3 capitalize text-muted">{en.source}</td>
                        <td className="p-3 capitalize">{en.status}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
            </div>
          </div>
        </div>
      )}

      {tab === 'entries' && (
        <div className="space-y-4 min-h-0 flex flex-col">
          <div className="flex flex-wrap items-end gap-3">
            <div>
              <div className="text-xs text-muted mb-1">Period</div>
              <div className="flex rounded-xl border border-white/10 p-0.5 bg-surface/40">
                <button
                  type="button"
                  onClick={() => {
                    setEntriesPeriodMode('year');
                    const m = entriesFilterMonth.slice(0, 4);
                    const y = parseInt(m, 10);
                    if (!Number.isNaN(y)) setEntriesFilterYear(y);
                  }}
                  className={`px-3 py-1.5 rounded-lg text-xs font-medium transition-all ${
                    entriesPeriodMode === 'year' ? 'bg-violet text-white' : 'text-muted hover:text-text'
                  }`}
                >
                  Yearly
                </button>
                <button
                  type="button"
                  onClick={() => {
                    setEntriesPeriodMode('month');
                    const now = new Date();
                    const ym =
                      entriesFilterYear === now.getFullYear()
                        ? `${entriesFilterYear}-${String(now.getMonth() + 1).padStart(2, '0')}`
                        : `${entriesFilterYear}-01`;
                    setEntriesFilterMonth(ym);
                  }}
                  className={`px-3 py-1.5 rounded-lg text-xs font-medium transition-all ${
                    entriesPeriodMode === 'month' ? 'bg-violet text-white' : 'text-muted hover:text-text'
                  }`}
                >
                  Monthly
                </button>
              </div>
            </div>
            {entriesPeriodMode === 'year' && (
              <div>
                <label className="text-xs text-muted block mb-1">Year</label>
                <select
                  value={entriesFilterYear}
                  onChange={(e) => setEntriesFilterYear(Number(e.target.value))}
                  className="rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm min-w-[7rem]"
                >
                  {entryYears.map((y) => (
                    <option key={y} value={y}>
                      {y}
                    </option>
                  ))}
                </select>
              </div>
            )}
            {entriesPeriodMode === 'month' && (
              <div>
                <label className="text-xs text-muted block mb-1">Month</label>
                <input
                  type="month"
                  value={entriesFilterMonth}
                  onChange={(e) => {
                    const v = e.target.value;
                    setEntriesFilterMonth(v);
                    const y = parseInt(v.slice(0, 4), 10);
                    if (!Number.isNaN(y)) setEntriesFilterYear(y);
                  }}
                  className="rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm"
                />
              </div>
            )}
          </div>

          <div className="rounded-[20px] border border-white/5 overflow-hidden flex flex-col flex-1 min-h-0 max-h-[min(70dvh,36rem)]">
            <div className="overflow-x-auto overflow-y-auto overscroll-y-contain min-h-0 flex-1 [scrollbar-gutter:stable]">
              <table className="w-full text-sm min-w-[520px]">
                <thead className="bg-surface text-left text-muted text-xs uppercase sticky top-0 z-[1] shadow-[inset_0_-1px_0_rgba(255,255,255,0.06)]">
                  <tr>
                    <th className="p-3">Date</th>
                    <th className="p-3">Ref</th>
                    <th className="p-3">Description</th>
                    <th className="p-3">Source</th>
                    <th className="p-3">Status</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredEntriesForTab.length === 0 && (
                    <tr>
                      <td colSpan={5} className="p-6 text-muted text-center">
                        No journal entries in this period.
                      </td>
                    </tr>
                  )}
                  {filteredEntriesForTab.map((en) => (
                    <tr key={en.id} className="border-t border-white/5">
                      <td className="p-3 whitespace-nowrap">{en.entry_date}</td>
                      <td className="p-3 font-mono text-xs">{en.reference || '—'}</td>
                      <td className="p-3">{en.description || '—'}</td>
                      <td className="p-3 capitalize text-muted">{en.source}</td>
                      <td className="p-3 capitalize">{en.status}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
          <p className="text-xs text-muted">
            Showing {filteredEntriesForTab.length} entr{filteredEntriesForTab.length === 1 ? 'y' : 'ies'} in selected period (loaded up to 200 most recent).
          </p>
        </div>
      )}

      {tab === 'chart' && (
        <div className="rounded-[20px] border border-white/5 overflow-x-auto max-h-[min(560px,60vh)] overflow-y-auto">
          <table className="w-full text-sm min-w-[480px]">
            <thead className="bg-surface text-left text-muted text-xs uppercase sticky top-0">
              <tr>
                <th className="p-3">Code</th>
                <th className="p-3">Name</th>
                <th className="p-3">Type</th>
              </tr>
            </thead>
            <tbody>
              {accounts.length === 0 && (
                <tr>
                  <td colSpan={3} className="p-6 text-muted text-center">
                    No accounts — complete onboarding or import a chart.
                  </td>
                </tr>
              )}
              {accounts.map((a) => (
                <tr key={a.id} className="border-t border-white/5">
                  <td className="p-3 font-mono">{a.code}</td>
                  <td className="p-3">{a.name}</td>
                  <td className="p-3 capitalize text-muted">{a.type}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {tab === 'trial' && (
        <div className="space-y-3">
          <div className="flex flex-wrap gap-3 items-end">
            <div>
              <label className="text-xs text-muted block mb-1">As of</label>
              <input
                type="date"
                value={asOf}
                onChange={(e) => setAsOf(e.target.value)}
                className="rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm"
              />
            </div>
            <button type="button" onClick={() => loadTrial()} className="rounded-xl bg-violet px-4 py-2 text-sm text-white">
              Refresh
            </button>
          </div>
          <div className="rounded-[20px] border border-white/5 overflow-x-auto max-h-[min(560px,55vh)] overflow-y-auto">
            <table className="w-full text-sm min-w-[480px]">
              <thead className="bg-surface text-left text-muted text-xs uppercase sticky top-0">
                <tr>
                  <th className="p-3">Code</th>
                  <th className="p-3">Name</th>
                  <th className="p-3 text-right">Balance</th>
                </tr>
              </thead>
              <tbody>
                {!trialLoaded && (
                  <tr>
                    <td colSpan={3} className="p-6 text-muted text-center">
                      Loading…
                    </td>
                  </tr>
                )}
                {trialLoaded &&
                  trialRows.map((r) => (
                    <tr key={r.account_id} className="border-t border-white/5">
                      <td className="p-3 font-mono">{r.code}</td>
                      <td className="p-3">{r.name}</td>
                      <td className="p-3 text-right tabular-nums">{r.balance.toFixed(2)}</td>
                    </tr>
                  ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      <button
        type="button"
        onClick={() => setGlossaryOpen(true)}
        className="fixed bottom-24 right-4 md:bottom-8 md:right-8 z-[55] flex h-9 w-9 items-center justify-center rounded-full bg-violet/90 text-white text-sm font-semibold shadow-[0_4px_20px_rgba(91,63,214,0.5)] border border-white/15 hover:bg-violet active:scale-95 transition-transform"
        aria-label="Accounting glossary"
        title="Glossary"
      >
        ?
      </button>

      {glossaryOpen && (
        <div className="fixed inset-0 z-[70] flex items-center justify-center p-4 pointer-events-none">
          <button
            type="button"
            className="absolute inset-0 z-0 bg-black/55 backdrop-blur-[2px] pointer-events-auto cursor-default border-0 p-0"
            aria-label="Close glossary"
            onClick={() => setGlossaryOpen(false)}
          />
          <div
            role="dialog"
            aria-modal="true"
            aria-labelledby="accounting-glossary-title"
            className="relative z-[1] pointer-events-auto w-full max-w-sm max-h-[min(72dvh,440px)] glass rounded-2xl border border-white/10 shadow-[0_20px_50px_rgba(0,0,0,0.45)] flex flex-col overflow-hidden"
          >
            <div className="flex items-center justify-between gap-3 px-4 py-3 border-b border-white/5 shrink-0 bg-surface/40">
              <h2 id="accounting-glossary-title" className="text-sm font-medium text-text">
                Quick glossary
              </h2>
              <button
                type="button"
                onClick={() => setGlossaryOpen(false)}
                className="rounded-lg px-2 py-1 text-muted hover:text-text hover:bg-white/5 text-lg leading-none"
                aria-label="Close"
              >
                ×
              </button>
            </div>
            <div className="overflow-y-auto p-3 space-y-2.5 text-xs">
              {glossary.length === 0 ? (
                <p className="text-muted text-center py-6">Loading terms…</p>
              ) : (
                glossary.map((t) => (
                  <div key={t.term} className="rounded-xl bg-card/60 border border-white/5 p-2.5">
                    <div className="font-medium text-yellow/90">{t.term}</div>
                    <p className="mt-1 text-muted leading-snug">{t.definition}</p>
                  </div>
                ))
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
