import { useEffect, useState } from 'react';
import { api } from '../api/client';
import { useAuth } from '../store/auth';

export function SettingsPage() {
  const token = useAuth((s) => s.accessToken);
  const logout = useAuth((s) => s.logout);
  const [country, setCountry] = useState('');
  const [currency, setCurrency] = useState('USD');
  const [tax, setTax] = useState('0');
  const [method, setMethod] = useState('accrual');
  const [fiscal, setFiscal] = useState('');
  const [taxSystem, setTaxSystem] = useState('');
  const [msg, setMsg] = useState('');
  const [loadErr, setLoadErr] = useState('');

  useEffect(() => {
    if (!token) return;
    setLoadErr('');
    api<{
      country: string;
      currency: string;
      tax_percent: number;
      accounting_method: string;
      fiscal_year_start: string;
      tax_system: string;
    }>('/api/v1/organization', { token })
      .then((o) => {
        setCountry(o.country || '');
        setCurrency(o.currency || 'USD');
        setTax(String(o.tax_percent ?? 0));
        setMethod(o.accounting_method || 'accrual');
        setFiscal(o.fiscal_year_start?.slice(0, 10) || '');
        setTaxSystem(o.tax_system || '');
      })
      .catch((e) => setLoadErr(e instanceof Error ? e.message : 'Failed to load'));
  }, [token]);

  async function save(e: React.FormEvent) {
    e.preventDefault();
    if (!token) return;
    setMsg('');
    setLoadErr('');
    try {
      const taxPercent = parseFloat(tax) || 0;
      await api('/api/v1/organization', {
        method: 'PATCH',
        token,
        body: JSON.stringify({
          country,
          currency,
          tax_percent: taxPercent,
          accounting_method: method,
          fiscal_year_start: fiscal || null,
          tax_system: taxSystem,
          base_currency: currency,
        }),
      });
      setMsg('Saved');
    } catch (e) {
      setMsg(e instanceof Error ? e.message : 'Error');
    }
  }

  return (
    <div className="max-w-xl space-y-6">
      <p className="text-muted text-sm">Where you operate from drives tax labels and booking defaults</p>
      {loadErr && <p className="text-danger text-sm">{loadErr}</p>}
      <form onSubmit={save} className="rounded-[20px] bg-card border border-white/5 p-6 space-y-4">
        <div>
          <label className="text-xs text-muted">Country / region</label>
          <input
            className="mt-1 w-full rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm"
            value={country}
            onChange={(e) => setCountry(e.target.value)}
            placeholder="e.g. US, GB, SG"
          />
        </div>
        <div>
          <label className="text-xs text-muted">Base currency</label>
          <input
            className="mt-1 w-full rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm"
            value={currency}
            onChange={(e) => setCurrency(e.target.value)}
            maxLength={3}
          />
        </div>
        <div>
          <label className="text-xs text-muted">Tax system label</label>
          <input
            className="mt-1 w-full rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm"
            value={taxSystem}
            onChange={(e) => setTaxSystem(e.target.value)}
            placeholder="VAT, GST, Sales tax…"
          />
        </div>
        <div>
          <label className="text-xs text-muted">Default tax % (for bookings)</label>
          <input
            className="mt-1 w-full rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm"
            value={tax}
            onChange={(e) => setTax(e.target.value)}
            type="number"
            step="0.0001"
          />
        </div>
        <div>
          <label className="text-xs text-muted">Fiscal year start</label>
          <input
            className="mt-1 w-full rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm"
            value={fiscal}
            onChange={(e) => setFiscal(e.target.value)}
            type="date"
          />
        </div>
        <div>
          <label className="text-xs text-muted">Accounting method</label>
          <select
            className="mt-1 w-full rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm"
            value={method}
            onChange={(e) => setMethod(e.target.value)}
          >
            <option value="accrual">Accrual</option>
            <option value="cash">Cash</option>
          </select>
        </div>
        {msg && (
          <p className={`text-sm ${msg === 'Saved' ? 'text-success' : 'text-danger'}`}>{msg}</p>
        )}
        <button type="submit" className="rounded-xl bg-violet px-5 py-2.5 text-sm font-medium text-white">
          Save fiscal profile
        </button>
      </form>
      <button
        type="button"
        onClick={logout}
        className="text-sm text-muted underline"
      >
        Sign out
      </button>
    </div>
  );
}
