import { useCallback, useEffect, useMemo, useState } from 'react';
import { api } from '../api/client';
import { useAuth } from '../store/auth';

type BookingRow = {
  id: number;
  booking_number: string;
  status: string;
  currency: string;
  subtotal: number;
  tax: number;
  total: number;
  booking_date: string;
};

type BookingItemRow = {
  id?: number;
  description: string;
  quantity: number;
  unit_price: number;
  line_total?: number;
};

type BookingDetail = BookingRow & {
  notes?: string;
  items: BookingItemRow[];
};

type LineForm = { description: string; quantity: string; unit_price: string };

type BookingsTabId = 'new' | 'entries';

function emptyLine(): LineForm {
  return { description: '', quantity: '1', unit_price: '' };
}

export function BookingsPage() {
  const token = useAuth((s) => s.accessToken);
  const [tab, setTab] = useState<BookingsTabId>('new');
  const [bookings, setBookings] = useState<BookingRow[]>([]);
  const [msg, setMsg] = useState('');
  const [loadingDetail, setLoadingDetail] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [bookingNumber, setBookingNumber] = useState('');
  const [bookingDate, setBookingDate] = useState(() => new Date().toISOString().slice(0, 10));
  const [currency, setCurrency] = useState('USD');
  const [notes, setNotes] = useState('');
  const [lines, setLines] = useState<LineForm[]>(() => [emptyLine()]);
  const [entriesFilterFrom, setEntriesFilterFrom] = useState('');
  const [entriesFilterTo, setEntriesFilterTo] = useState('');

  const load = useCallback(async () => {
    if (!token) return;
    const res = await api<{ bookings: BookingRow[] | null }>('/api/v1/bookings', { token });
    setBookings(res.bookings ?? []);
  }, [token]);

  useEffect(() => {
    load().catch((e) => setMsg(String(e)));
  }, [load]);

  useEffect(() => {
    if (tab !== 'entries') return;
    load().catch((e) => setMsg(String(e)));
  }, [tab, load]);

  const recentBookings = useMemo(() => bookings.slice(0, 10), [bookings]);

  const entriesDateRangeInvalid =
    entriesFilterFrom !== '' && entriesFilterTo !== '' && entriesFilterFrom > entriesFilterTo;

  const filteredEntriesBookings = useMemo(() => {
    if (entriesDateRangeInvalid) return bookings;
    return bookings.filter((b) => {
      const d = b.booking_date;
      if (entriesFilterFrom && d < entriesFilterFrom) return false;
      if (entriesFilterTo && d > entriesFilterTo) return false;
      return true;
    });
  }, [bookings, entriesFilterFrom, entriesFilterTo, entriesDateRangeInvalid]);

  function resetForm() {
    setEditingId(null);
    setBookingNumber('');
    setBookingDate(new Date().toISOString().slice(0, 10));
    setCurrency('USD');
    setNotes('');
    setLines([emptyLine()]);
  }

  function updateLine(i: number, patch: Partial<LineForm>) {
    setLines((prev) => prev.map((row, j) => (j === i ? { ...row, ...patch } : row)));
  }

  function addLine() {
    setLines((prev) => [...prev, emptyLine()]);
  }

  function removeLine(i: number) {
    setLines((prev) => (prev.length <= 1 ? prev : prev.filter((_, j) => j !== i)));
  }

  async function startEdit(id: number) {
    if (!token) return;
    setMsg('');
    setLoadingDetail(true);
    try {
      const res = await api<{ booking: BookingDetail }>(`/api/v1/bookings/${id}`, { token });
      const b = res.booking;
      if (b.status === 'posted') {
        setMsg('Posted bookings cannot be edited.');
        setLoadingDetail(false);
        return;
      }
      setEditingId(id);
      setBookingNumber(b.booking_number);
      setBookingDate(b.booking_date);
      setCurrency(b.currency || 'USD');
      setNotes(b.notes ?? '');
      setLines(
        b.items?.length
          ? b.items.map((it) => ({
              description: it.description ?? '',
              quantity: String(it.quantity),
              unit_price: String(it.unit_price),
            }))
          : [emptyLine()]
      );
    } catch (e) {
      setMsg(e instanceof Error ? e.message : 'Error');
    } finally {
      setLoadingDetail(false);
    }
  }

  async function removeBooking(id: number) {
    if (!token) return;
    const row = bookings.find((b) => b.id === id);
    if (row?.status === 'posted') {
      setMsg('Posted bookings cannot be deleted.');
      return;
    }
    if (!window.confirm('Delete this draft booking? This cannot be undone.')) return;
    setMsg('');
    try {
      await api(`/api/v1/bookings/${id}`, { method: 'DELETE', token });
      if (editingId === id) resetForm();
      await load();
    } catch (e) {
      setMsg(e instanceof Error ? e.message : 'Error');
    }
  }

  async function post(id: number) {
    if (!token) return;
    setMsg('');
    try {
      await api(`/api/v1/bookings/${id}/post`, { method: 'POST', token, body: '{}' });
      await load();
    } catch (e) {
      setMsg(e instanceof Error ? e.message : 'Error');
    }
  }

  async function submitBooking(e: React.FormEvent) {
    e.preventDefault();
    if (!token) return;
    setMsg('');
    const parsedItems: { description: string; quantity: number; unit_price: number }[] = [];
    for (const row of lines) {
      const q = parseFloat(row.quantity);
      const price = parseFloat(row.unit_price);
      if (!row.description.trim() || Number.isNaN(q) || q <= 0 || Number.isNaN(price)) {
        setMsg('Each line needs a description, quantity greater than 0, and a unit price.');
        return;
      }
      parsedItems.push({ description: row.description.trim(), quantity: q, unit_price: price });
    }
    try {
      if (editingId != null) {
        await api(`/api/v1/bookings/${editingId}`, {
          method: 'PATCH',
          token,
          body: JSON.stringify({
            booking_number: bookingNumber.trim(),
            booking_date: bookingDate,
            currency: currency.trim() || 'USD',
            notes: notes.trim(),
            items: parsedItems,
          }),
        });
      } else {
        await api('/api/v1/bookings', {
          method: 'POST',
          token,
          body: JSON.stringify({
            booking_date: bookingDate,
            currency: currency.trim() || 'USD',
            status: 'draft',
            notes: notes.trim() || undefined,
            items: parsedItems,
          }),
        });
      }
      resetForm();
      await load();
    } catch (err) {
      setMsg(err instanceof Error ? err.message : 'Error');
    }
  }

  function bookingRows(list: BookingRow[], emptyLabel: string) {
    if (list.length === 0) {
      return (
        <tr>
          <td colSpan={5} className="p-6 text-muted text-center">
            {emptyLabel}
          </td>
        </tr>
      );
    }
    return list.map((b) => (
      <tr key={b.id} className="border-t border-white/5">
        <td className="p-3 font-mono">{b.booking_number}</td>
        <td className="p-3 capitalize">{b.status}</td>
        <td className="p-3">{b.booking_date}</td>
        <td className="p-3 text-right tabular-nums">
          {b.currency} {b.total.toFixed(2)}
        </td>
        <td className="p-3">
          <div className="flex flex-wrap justify-end gap-2">
            {b.status !== 'posted' && (
              <>
                <button
                  type="button"
                  onClick={() => {
                    setTab('new');
                    void startEdit(b.id);
                  }}
                  disabled={loadingDetail}
                  className="text-xs px-3 py-1.5 rounded-lg border border-white/15 text-muted hover:text-text hover:bg-white/5 disabled:opacity-50"
                >
                  Edit
                </button>
                <button
                  type="button"
                  onClick={() => removeBooking(b.id)}
                  className="text-xs px-3 py-1.5 rounded-lg border border-danger/40 text-danger hover:bg-danger/10"
                >
                  Delete
                </button>
                <button
                  type="button"
                  onClick={() => post(b.id)}
                  className="text-xs px-3 py-1.5 rounded-lg bg-violet text-white"
                >
                  Post
                </button>
              </>
            )}
          </div>
        </td>
      </tr>
    ));
  }

  const tabBtn = (id: BookingsTabId, label: string) => (
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
      {msg && <p className="text-danger text-sm">{msg}</p>}

      <div className="flex flex-wrap gap-1 p-1 rounded-2xl bg-surface/40 border border-white/5 w-fit max-w-full">
        {tabBtn('new', 'New booking')}
        {tabBtn('entries', 'Entries')}
      </div>

      {tab === 'new' && (
        <>
          <form onSubmit={submitBooking} className="rounded-[20px] bg-card border border-white/5 p-4 md:p-5 space-y-3 max-w-3xl">
        <p className="text-xs text-muted">
          {editingId != null
            ? `Editing booking ${bookingNumber} — lines replace existing line items. Tax uses your org rate.`
            : 'Draft sale — add one or more lines. Tax uses your org rate; post when ready.'}
        </p>
        <div className="flex flex-wrap gap-3 items-end">
          {editingId != null && (
            <div>
              <label className="text-xs text-muted block mb-1">Booking #</label>
              <input
                value={bookingNumber}
                onChange={(e) => setBookingNumber(e.target.value)}
                required
                className="w-36 rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm font-mono"
              />
            </div>
          )}
          <div>
            <label className="text-xs text-muted block mb-1">Date</label>
            <input
              type="date"
              required
              value={bookingDate}
              onChange={(e) => setBookingDate(e.target.value)}
              className="rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm"
            />
          </div>
          <div>
            <label className="text-xs text-muted block mb-1">Currency</label>
            <input
              value={currency}
              onChange={(e) => setCurrency(e.target.value.toUpperCase())}
              maxLength={8}
              className="w-24 rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm font-mono uppercase"
            />
          </div>
        </div>

        <div className="space-y-2">
          <div className="text-xs text-muted uppercase tracking-wide">Line items</div>
          {lines.map((row, i) => (
            <div key={i} className="flex flex-wrap gap-2 items-end">
              <div className="flex-1 min-w-[180px]">
                <label className="text-xs text-muted block mb-1">Description</label>
                <input
                  value={row.description}
                  onChange={(e) => updateLine(i, { description: e.target.value })}
                  placeholder="e.g. Consulting — May"
                  required
                  className="w-full rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm"
                />
              </div>
              <div>
                <label className="text-xs text-muted block mb-1">Qty</label>
                <input
                  type="number"
                  min={0}
                  step="any"
                  value={row.quantity}
                  onChange={(e) => updateLine(i, { quantity: e.target.value })}
                  className="w-24 rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm tabular-nums"
                />
              </div>
              <div>
                <label className="text-xs text-muted block mb-1">Unit price</label>
                <input
                  type="number"
                  min={0}
                  step="0.01"
                  value={row.unit_price}
                  onChange={(e) => updateLine(i, { unit_price: e.target.value })}
                  placeholder="0.00"
                  required
                  className="w-32 rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm tabular-nums"
                />
              </div>
              {lines.length > 1 && (
                <button
                  type="button"
                  onClick={() => removeLine(i)}
                  className="text-xs text-muted hover:text-danger px-2 py-2 mb-0.5"
                >
                  Remove
                </button>
              )}
            </div>
          ))}
          <button type="button" onClick={addLine} className="text-xs text-muted hover:text-text underline-offset-2 hover:underline">
            Add line
          </button>
        </div>

        <div>
          <label className="text-xs text-muted block mb-1">Notes (optional)</label>
          <input
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
            className="w-full rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm"
            placeholder="Internal note"
          />
        </div>

        <div className="flex flex-wrap gap-2">
          <button type="submit" disabled={loadingDetail} className="rounded-xl bg-violet px-4 py-2 text-sm font-medium text-white disabled:opacity-50">
            {editingId != null ? 'Save changes' : 'Save draft booking'}
          </button>
          {editingId != null && (
            <button type="button" onClick={() => resetForm()} className="rounded-xl border border-white/15 px-4 py-2 text-sm text-muted hover:text-text">
              Cancel edit
            </button>
          )}
        </div>
      </form>

          <div className="min-h-0 shrink-0">
            <h2 className="text-sm font-medium text-muted mb-2 uppercase tracking-wide">Recent bookings</h2>
            <div className="rounded-[20px] border border-white/5 max-h-[min(29rem,42vh)] overflow-x-auto overflow-y-auto overscroll-y-contain isolate [scrollbar-gutter:stable]">
              <table className="w-full text-sm min-w-[640px]">
                <thead className="bg-surface text-left text-muted text-xs uppercase sticky top-0 z-[1] shadow-[inset_0_-1px_0_rgba(255,255,255,0.06)]">
                  <tr>
                    <th className="p-3">#</th>
                    <th className="p-3">Status</th>
                    <th className="p-3">Date</th>
                    <th className="p-3 text-right">Total</th>
                    <th className="p-3 text-right">Actions</th>
                  </tr>
                </thead>
                <tbody>{bookingRows(recentBookings, 'No bookings yet — add one above.')}</tbody>
              </table>
            </div>
          </div>
        </>
      )}

      {tab === 'entries' && (
        <div className="space-y-3 min-h-0 flex flex-col flex-1">
          <div className="flex flex-wrap items-end gap-3">
            <div>
              <label className="text-xs text-muted block mb-1">From</label>
              <input
                type="date"
                value={entriesFilterFrom}
                onChange={(e) => setEntriesFilterFrom(e.target.value)}
                className="rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="text-xs text-muted block mb-1">To</label>
              <input
                type="date"
                value={entriesFilterTo}
                onChange={(e) => setEntriesFilterTo(e.target.value)}
                className="rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm"
              />
            </div>
            <button
              type="button"
              onClick={() => {
                setEntriesFilterFrom('');
                setEntriesFilterTo('');
              }}
              className="rounded-xl border border-white/15 px-3 py-2 text-sm text-muted hover:text-text hover:bg-white/5"
            >
              Clear dates
            </button>
          </div>
          {entriesDateRangeInvalid && (
            <p className="text-danger text-xs">From date must be on or before To date. Adjust the range to filter.</p>
          )}
          <div className="rounded-[20px] border border-white/5 overflow-hidden flex flex-col flex-1 min-h-0 max-h-[min(70dvh,36rem)]">
            <div className="overflow-x-auto overflow-y-auto overscroll-y-contain min-h-0 flex-1 [scrollbar-gutter:stable]">
              <table className="w-full text-sm min-w-[640px]">
                <thead className="bg-surface text-left text-muted text-xs uppercase sticky top-0 z-[1] shadow-[inset_0_-1px_0_rgba(255,255,255,0.06)]">
                  <tr>
                    <th className="p-3">#</th>
                    <th className="p-3">Status</th>
                    <th className="p-3">Date</th>
                    <th className="p-3 text-right">Total</th>
                    <th className="p-3 text-right">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {bookingRows(
                    filteredEntriesBookings,
                    entriesFilterFrom || entriesFilterTo
                      ? 'No bookings in this date range.'
                      : 'No bookings yet — add one on the New booking tab.'
                  )}
                </tbody>
              </table>
            </div>
          </div>
          <p className="text-xs text-muted">
            {entriesFilterFrom || entriesFilterTo
              ? !entriesDateRangeInvalid
                ? `Showing ${filteredEntriesBookings.length} of ${bookings.length} loaded bookings (up to 100 from server).`
                : 'Invalid date range — showing all loaded bookings.'
              : 'All bookings loaded from the server (up to 100). Set From / To to narrow by booking date.'}
          </p>
        </div>
      )}
    </div>
  );
}
