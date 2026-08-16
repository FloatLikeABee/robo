import { useEffect, useState } from 'react';
import { api } from '../api/client';
import { useAuth } from '../store/auth';

export function DashboardPage() {
  const token = useAuth((s) => s.accessToken);
  const [data, setData] = useState<Record<string, number> | null>(null);
  const [err, setErr] = useState('');

  useEffect(() => {
    if (!token) return;
    api<{ bookings_total: number; bookings_pending: number; products: number; assets: number; warehouses: number; low_stock_skus: number }>(
      '/api/v1/dashboard',
      { token }
    )
      .then(setData)
      .catch((e) => setErr(e instanceof Error ? e.message : 'Error'));
  }, [token]);

  if (!data && !err) return <p className="text-muted">Loading…</p>;
  if (err) return <p className="text-danger">{err}</p>;

  const cards = [
    ['Bookings', data!.bookings_total],
    ['Pending', data!.bookings_pending],
    ['Products', data!.products],
    ['Assets', data!.assets],
    ['Warehouses', data!.warehouses],
    ['Low stock', data!.low_stock_skus],
  ];

  return (
    <div>
      <p className="text-muted text-sm mb-6">Operational snapshot for your organization</p>
      <div className="grid grid-cols-2 lg:grid-cols-3 gap-4">
        {cards.map(([label, val]) => (
          <div key={String(label)} className="rounded-[20px] bg-card border border-white/5 p-5 shadow-[0_10px_30px_rgba(0,0,0,0.2)]">
            <div className="text-xs text-muted uppercase tracking-wide">{label}</div>
            <div className="text-2xl font-semibold mt-2 tabular-nums">{val}</div>
          </div>
        ))}
      </div>
    </div>
  );
}
