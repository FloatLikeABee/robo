import { useEffect, useState } from 'react';
import { api } from '../api/client';
import { useAuth } from '../store/auth';

export function AssetsPage() {
  const token = useAuth((s) => s.accessToken);
  const [assets, setAssets] = useState<
    { id: number; asset_tag: string; name: string; purchase_value: number; status: string }[]
  >([]);
  const [tag, setTag] = useState('');
  const [name, setName] = useState('');
  const [value, setValue] = useState('');
  const [msg, setMsg] = useState('');
  const [syncBusy, setSyncBusy] = useState(false);

  async function load() {
    if (!token) return;
    const r = await api<{ assets: typeof assets | null }>('/api/v1/assets', { token });
    setAssets(r.assets ?? []);
  }

  useEffect(() => {
    load().catch((e) => setMsg(String(e)));
  }, [token]);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    if (!token) return;
    setMsg('');
    try {
      const pv = parseFloat(value) || 0;
      await api('/api/v1/assets', {
        method: 'POST',
        token,
        body: JSON.stringify({ asset_tag: tag, name, purchase_value: pv, current_value: pv }),
      });
      setTag('');
      setName('');
      setValue('');
      await load();
    } catch (e) {
      setMsg(e instanceof Error ? e.message : 'Error');
    }
  }

  async function syncFromMorph() {
    if (!token) return;
    setSyncBusy(true);
    setMsg('');
    try {
      const res = await api<{ created: number; updated: number; total: number }>('/api/v1/assets/sync-morph', {
        method: 'POST',
        token,
      });
      await load();
      setMsg(`Synced from MorphData: ${res.created} added, ${res.updated} updated (${res.total} total).`);
    } catch (e) {
      setMsg(e instanceof Error ? e.message : 'Sync failed');
    } finally {
      setSyncBusy(false);
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <p className="text-muted text-sm">Fixed asset register — sync from MorphData or add manually</p>
        <button
          type="button"
          disabled={syncBusy}
          onClick={() => void syncFromMorph()}
          className="rounded-xl border border-white/15 px-4 py-2 text-sm font-medium hover:bg-white/5 disabled:opacity-50"
        >
          {syncBusy ? 'Syncing…' : 'Sync from MorphData'}
        </button>
      </div>
      {msg && <p className={`text-sm ${msg.includes('failed') || msg.includes('Error') ? 'text-danger' : 'text-emerald-300'}`}>{msg}</p>}
      <form onSubmit={submit} className="rounded-[20px] bg-card border border-white/5 p-5 flex flex-wrap gap-3 items-end max-w-2xl">
        <div>
          <label className="text-xs text-muted block mb-1">Tag</label>
          <input
            className="rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm"
            value={tag}
            onChange={(e) => setTag(e.target.value)}
            required
          />
        </div>
        <div>
          <label className="text-xs text-muted block mb-1">Name</label>
          <input
            className="rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
          />
        </div>
        <div>
          <label className="text-xs text-muted block mb-1">Value</label>
          <input
            className="rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm w-28"
            value={value}
            onChange={(e) => setValue(e.target.value)}
            type="number"
            step="0.01"
          />
        </div>
        <button type="submit" className="rounded-xl bg-violet px-4 py-2 text-sm text-white">
          Add
        </button>
      </form>
      <div className="rounded-[20px] border border-white/5 overflow-x-auto">
        <table className="w-full text-sm">
          <thead className="bg-surface text-left text-muted text-xs uppercase">
            <tr>
              <th className="p-3">Tag</th>
              <th className="p-3">Name</th>
              <th className="p-3 text-right">Value</th>
              <th className="p-3">Status</th>
            </tr>
          </thead>
          <tbody>
            {assets.map((a) => (
              <tr key={a.id} className="border-t border-white/5">
                <td className="p-3 font-mono">{a.asset_tag}</td>
                <td className="p-3">{a.name}</td>
                <td className="p-3 text-right tabular-nums">{a.purchase_value}</td>
                <td className="p-3 capitalize">{a.status}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
