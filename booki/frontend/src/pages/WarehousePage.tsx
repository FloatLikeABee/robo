import { useEffect, useMemo, useState } from 'react';
import { api } from '../api/client';
import { useAuth } from '../store/auth';

function todayISO() {
  return new Date().toISOString().slice(0, 10);
}

type WarehouseRow = {
  id: number;
  name: string;
  code: string;
  location?: string;
  detail: string | null;
  record_date: string | null;
};

type ProductRow = {
  id: number;
  sku: string;
  name: string;
  description: string;
  category: string;
  unit: string;
  barcode: string;
  cost_price: number;
  selling_price: number;
  reorder_threshold: number;
  detail: string | null;
  record_date: string | null;
};

type StockRow = {
  warehouse_id: number;
  warehouse_name: string;
  product_id: number;
  sku: string;
  product_name: string;
  quantity: number;
  stock_date?: string;
};

type WarehouseTabId = 'locations' | 'stock';

type FocusState = { kind: 'warehouse'; id: number } | { kind: 'product'; id: number } | null;

function validateOptionalJSON(raw: string): string | undefined {
  const t = raw.trim();
  if (t === '') return undefined;
  try {
    JSON.parse(t);
  } catch {
    throw new Error('Detail must be valid JSON or empty');
  }
  return t;
}

type ModalBackdropProps = { children: React.ReactNode; title: string; onClose: () => void };

function ModalShell({ title, children, onClose }: ModalBackdropProps) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/55" role="presentation" onMouseDown={(e) => e.target === e.currentTarget && onClose()}>
      <div
        role="dialog"
        aria-labelledby="warehouse-modal-title"
        className="w-full max-w-md max-h-[90vh] overflow-y-auto rounded-[20px] bg-card border border-white/10 shadow-xl p-5 space-y-4"
        onMouseDown={(e) => e.stopPropagation()}
      >
        <div className="flex items-start justify-between gap-3">
          <h2 id="warehouse-modal-title" className="font-medium">
            {title}
          </h2>
          <button type="button" className="text-muted hover:text-text text-sm" onClick={onClose} aria-label="Close">
            Close
          </button>
        </div>
        {children}
      </div>
    </div>
  );
}

function DetailModalShell({ title, children, onClose }: ModalBackdropProps) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/55" role="presentation" onMouseDown={(e) => e.target === e.currentTarget && onClose()}>
      <div
        role="dialog"
        aria-labelledby="detail-modal-title"
        className="w-full max-w-lg max-h-[90vh] overflow-y-auto rounded-[20px] bg-card border border-white/10 shadow-xl p-5 space-y-4"
        onMouseDown={(e) => e.stopPropagation()}
      >
        <div className="flex items-start justify-between gap-3">
          <h2 id="detail-modal-title" className="font-medium">
            {title}
          </h2>
          <button type="button" className="text-muted hover:text-text text-sm" onClick={onClose} aria-label="Close">
            Close
          </button>
        </div>
        {children}
      </div>
    </div>
  );
}

function StockHelpModal({ onClose }: { onClose: () => void }) {
  return (
    <ModalShell title="How stock balances work" onClose={onClose}>
      <div className="space-y-3 text-sm text-text leading-relaxed">
        <p className="text-muted text-xs">
          Use <strong className="text-text">Add inventory</strong> on Locations (or{' '}
          <strong className="text-text">Add inventory</strong> under a warehouse) to record incoming inventory. Optionally set{' '}
          <strong className="text-text">Stock date</strong> — stored as <code className="font-mono text-[10px]">movement_date</code> on the movement.
        </p>
        <p className="text-muted text-xs">
          The grid still only <em>displays</em> balances plus a <strong className="text-text">Stock date</strong> column: last stock-in date for that SKU at that warehouse (or{' '}
          last balance update date if unknown).
        </p>
        <p>
          Each row is one product in one warehouse. The number is the quantity kept in <code className="font-mono text-xs bg-bg px-1.5 py-0.5 rounded">warehouse_stocks</code>
          , updated when you record inventory movements.
        </p>
        <p className="font-medium text-text">How counts change (API)</p>
        <ul className="list-disc pl-5 space-y-2 text-muted">
          <li>
            <strong className="text-text">Stock in</strong> — form in the app posts to{' '}
            <code className="font-mono text-xs text-text">POST /api/v1/warehouse/stock-in</code>:{' '}
            <code className="font-mono text-xs">warehouse_id</code>, <code className="font-mono text-xs">product_id</code>, <code className="font-mono text-xs">quantity</code>{' '}
            . Optional <code className="font-mono text-xs">stock_date</code> (<code className="font-mono text-xs">YYYY-MM-DD</code>) and <code className="font-mono text-xs">reference</code>.
          </li>
          <li>
            <strong className="text-text">Stock out</strong> — <code className="font-mono text-xs text-text">POST /api/v1/warehouse/stock-out</code> same shape; qty must not drop below zero. Optional{' '}
            <code className="font-mono text-xs">stock_date</code>.
          </li>
          <li>
            <strong className="text-text">Transfer</strong> — <code className="font-mono text-xs text-text">POST /api/v1/warehouse/transfers</code> moves quantity from{' '}
            <code className="font-mono text-xs">from_warehouse_id</code> to <code className="font-mono text-xs">to_warehouse_id</code>.
          </li>
        </ul>
        <p className="text-muted text-xs">
          Stock-out and transfers are still API-only (<code className="font-mono text-[10px]">stock-out</code>, <code className="font-mono text-[10px]">transfers</code>).
        </p>
      </div>
      <button type="button" className="w-full rounded-xl bg-violet px-4 py-2 text-sm font-medium text-white" onClick={onClose}>
        Got it
      </button>
    </ModalShell>
  );
}

export function WarehousePage() {
  const token = useAuth((s) => s.accessToken);
  const [tab, setTab] = useState<WarehouseTabId>('locations');
  const [products, setProducts] = useState<ProductRow[]>([]);
  const [warehouses, setWarehouses] = useState<WarehouseRow[]>([]);
  const [stocks, setStocks] = useState<StockRow[]>([]);
  const [msg, setMsg] = useState('');

  const [collapsedWhIds, setCollapsedWhIds] = useState<Set<number>>(new Set());

  const [focused, setFocused] = useState<FocusState>(null);

  const [stockHelpOpen, setStockHelpOpen] = useState(false);

  const [modalWh, setModalWh] = useState(false);
  const [modalProd, setModalProd] = useState(false);

  const [stockInOpen, setStockInOpen] = useState(false);
  const [stockInWhId, setStockInWhId] = useState('');
  const [stockInProductId, setStockInProductId] = useState('');
  const [stockInQty, setStockInQty] = useState('1');
  const [stockInDate, setStockInDate] = useState(() => todayISO());
  const [stockInRef, setStockInRef] = useState('');

  const [whAddName, setWhAddName] = useState('');
  const [whAddCode, setWhAddCode] = useState('');
  const [whAddLoc, setWhAddLoc] = useState('');
  const [whAddDetail, setWhAddDetail] = useState('');
  const [whAddDate, setWhAddDate] = useState(() => todayISO());

  const [pAddSku, setPAddSku] = useState('');
  const [pAddName, setPAddName] = useState('');
  const [pAddDesc, setPAddDesc] = useState('');
  const [pAddCost, setPAddCost] = useState('0');
  const [pAddPrice, setPAddPrice] = useState('0');
  const [pAddDetail, setPAddDetail] = useState('');
  const [pAddDate, setPAddDate] = useState(() => todayISO());

  const [detailKind, setDetailKind] = useState<'warehouse' | 'product' | null>(null);
  const [detailId, setDetailId] = useState<number | null>(null);
  const [detailLoading, setDetailLoading] = useState(false);

  const [whEditName, setWhEditName] = useState('');
  const [whEditCode, setWhEditCode] = useState('');
  const [whEditLoc, setWhEditLoc] = useState('');
  const [whEditDetail, setWhEditDetail] = useState('');
  const [whEditDate, setWhEditDate] = useState('');

  const [pEditSku, setPEditSku] = useState('');
  const [pEditName, setPEditName] = useState('');
  const [pEditDesc, setPEditDesc] = useState('');
  const [pEditCost, setPEditCost] = useState('');
  const [pEditPrice, setPEditPrice] = useState('');
  const [pEditDetail, setPEditDetail] = useState('');
  const [pEditDate, setPEditDate] = useState('');
  const [pEditCat, setPEditCat] = useState('');
  const [pEditUnit, setPEditUnit] = useState('ea');
  const [pEditBc, setPEditBc] = useState('');
  const [pEditReorder, setPEditReorder] = useState('0');

  const stocksByWarehouse = useMemo(() => {
    const m = new Map<number, StockRow[]>();
    for (const s of stocks) {
      const rows = m.get(s.warehouse_id);
      if (rows) rows.push(s);
      else m.set(s.warehouse_id, [s]);
    }
    for (const [, rows] of m) rows.sort((a, b) => a.sku.localeCompare(b.sku));
    return m;
  }, [stocks]);

  const stockedProductIds = useMemo(() => new Set(stocks.map((s) => s.product_id)), [stocks]);

  const orphanProducts = useMemo(() => products.filter((p) => !stockedProductIds.has(p.id)), [products, stockedProductIds]);

  const productsSorted = useMemo(() => [...products].sort((a, b) => a.sku.localeCompare(b.sku)), [products]);

  async function refresh() {
    if (!token) return;
    const [p, w, s] = await Promise.all([
      api<{ products: ProductRow[] | null }>('/api/v1/products', { token }),
      api<{ warehouses: WarehouseRow[] | null }>('/api/v1/warehouses', { token }),
      api<{ stocks: StockRow[] | null }>('/api/v1/warehouse/stock', { token }),
    ]);
    setProducts(p.products ?? []);
    setWarehouses(w.warehouses ?? []);
    setStocks(s.stocks ?? []);
  }

  useEffect(() => {
    refresh().catch((e) => setMsg(String(e)));
  }, [token]);

  useEffect(() => {
    if (!token || detailKind === null || detailId === null) return;
    setDetailLoading(true);
    setMsg('');
    const path =
      detailKind === 'warehouse' ? `/api/v1/warehouses/${detailId}` : `/api/v1/products/${detailId}`;
    api<{ warehouse?: WarehouseRow; product?: ProductRow }>(path, { token })
      .then((res) => {
        if (detailKind === 'warehouse' && res.warehouse) {
          const z = res.warehouse;
          setWhEditName(z.name);
          setWhEditCode(z.code);
          setWhEditLoc(z.location ?? '');
          setWhEditDetail(z.detail ?? '');
          setWhEditDate(z.record_date ?? todayISO());
        }
        if (detailKind === 'product' && res.product) {
          const z = res.product;
          setPEditSku(z.sku);
          setPEditName(z.name);
          setPEditDesc(z.description ?? '');
          setPEditCost(String(z.cost_price ?? 0));
          setPEditPrice(String(z.selling_price ?? 0));
          setPEditDetail(z.detail ?? '');
          setPEditDate(z.record_date ?? todayISO());
          setPEditCat(z.category ?? '');
          setPEditUnit(z.unit ?? 'ea');
          setPEditBc(z.barcode ?? '');
          setPEditReorder(String(z.reorder_threshold ?? 0));
        }
      })
      .catch((e) => setMsg(String(e)))
      .finally(() => setDetailLoading(false));
  }, [detailKind, detailId, token]);

  function closeDetail() {
    setDetailKind(null);
    setDetailId(null);
  }

  function openWarehouse(w: WarehouseRow) {
    setFocused({ kind: 'warehouse', id: w.id });
    setDetailKind('warehouse');
    setDetailId(w.id);
  }

  function openProduct(productId: number) {
    setFocused({ kind: 'product', id: productId });
    setDetailKind('product');
    setDetailId(productId);
  }

  function openStockInModal(preWarehouseId?: number) {
    setMsg('');
    if (preWarehouseId != null && warehouses.some((w) => w.id === preWarehouseId)) {
      setStockInWhId(String(preWarehouseId));
    } else {
      setStockInWhId(warehouses[0] ? String(warehouses[0].id) : '');
    }
    setStockInProductId(productsSorted[0] ? String(productsSorted[0].id) : '');
    setStockInQty('1');
    setStockInDate(todayISO());
    setStockInRef('');
    setStockInOpen(true);
  }

  async function submitStockIn(e: React.FormEvent) {
    e.preventDefault();
    if (!token) return;
    const wid = parseInt(stockInWhId, 10);
    const pid = parseInt(stockInProductId, 10);
    const qty = parseFloat(stockInQty.replace(',', '.'));
    if (!Number.isFinite(wid) || !Number.isFinite(pid) || !Number.isFinite(qty) || qty <= 0) {
      setMsg('Pick a warehouse and product with a quantity greater than 0.');
      return;
    }
    setMsg('');
    try {
      await api('/api/v1/warehouse/stock-in', {
        method: 'POST',
        token,
        body: JSON.stringify({
          warehouse_id: wid,
          product_id: pid,
          quantity: qty,
          stock_date: stockInDate,
          reference: stockInRef.trim(),
        }),
      });
      setStockInOpen(false);
      await refresh();
      setCollapsedWhIds((prev) => {
        const n = new Set(prev);
        n.delete(wid);
        return n;
      });
      if (tab !== 'locations') setTab('locations');
    } catch (e) {
      setMsg(e instanceof Error ? e.message : 'Error');
    }
  }

  async function submitAddWarehouse(e: React.FormEvent) {
    e.preventDefault();
    if (!token) return;
    setMsg('');
    try {
      const detail = validateOptionalJSON(whAddDetail);
      await api('/api/v1/warehouses', {
        method: 'POST',
        token,
        body: JSON.stringify({
          name: whAddName,
          code: whAddCode,
          location: whAddLoc,
          detail: detail ?? '',
          record_date: whAddDate,
        }),
      });
      setWhAddName('');
      setWhAddCode('');
      setWhAddLoc('');
      setWhAddDetail('');
      setWhAddDate(todayISO());
      setModalWh(false);
      await refresh();
    } catch (e) {
      setMsg(e instanceof Error ? e.message : 'Error');
    }
  }

  async function submitAddProduct(e: React.FormEvent) {
    e.preventDefault();
    if (!token) return;
    setMsg('');
    try {
      const detail = validateOptionalJSON(pAddDetail);
      await api('/api/v1/products', {
        method: 'POST',
        token,
        body: JSON.stringify({
          sku: pAddSku,
          name: pAddName,
          description: pAddDesc,
          cost_price: parseFloat(pAddCost) || 0,
          selling_price: parseFloat(pAddPrice) || 0,
          detail: detail ?? '',
          record_date: pAddDate,
        }),
      });
      setPAddSku('');
      setPAddName('');
      setPAddDesc('');
      setPAddCost('0');
      setPAddPrice('0');
      setPAddDetail('');
      setPAddDate(todayISO());
      setModalProd(false);
      await refresh();
    } catch (e) {
      setMsg(e instanceof Error ? e.message : 'Error');
    }
  }

  async function saveWarehouseDetail(e: React.FormEvent) {
    e.preventDefault();
    if (!token || detailId === null) return;
    setMsg('');
    try {
      const detail = validateOptionalJSON(whEditDetail);
      await api(`/api/v1/warehouses/${detailId}`, {
        method: 'PATCH',
        token,
        body: JSON.stringify({
          name: whEditName,
          code: whEditCode,
          location: whEditLoc,
          detail: detail ?? '',
          record_date: whEditDate,
        }),
      });
      closeDetail();
      await refresh();
    } catch (e) {
      setMsg(e instanceof Error ? e.message : 'Error');
    }
  }

  async function removeWarehouse() {
    if (!token || detailId === null) return;
    if (!confirm('Remove this warehouse? Related stock movements may be deleted.')) return;
    setMsg('');
    try {
      await api(`/api/v1/warehouses/${detailId}`, { method: 'DELETE', token });
      if (focused?.kind === 'warehouse' && focused.id === detailId) setFocused(null);
      closeDetail();
      await refresh();
    } catch (e) {
      setMsg(e instanceof Error ? e.message : 'Error');
    }
  }

  async function saveProductDetail(e: React.FormEvent) {
    e.preventDefault();
    if (!token || detailId === null) return;
    setMsg('');
    try {
      const detail = validateOptionalJSON(pEditDetail);
      await api(`/api/v1/products/${detailId}`, {
        method: 'PATCH',
        token,
        body: JSON.stringify({
          sku: pEditSku,
          name: pEditName,
          description: pEditDesc,
          category: pEditCat,
          unit: pEditUnit,
          barcode: pEditBc,
          cost_price: parseFloat(pEditCost) || 0,
          selling_price: parseFloat(pEditPrice) || 0,
          reorder_threshold: parseInt(pEditReorder, 10) || 0,
          detail: detail ?? '',
          record_date: pEditDate,
        }),
      });
      closeDetail();
      await refresh();
    } catch (e) {
      setMsg(e instanceof Error ? e.message : 'Error');
    }
  }

  async function removeProduct() {
    if (!token || detailId === null) return;
    if (!confirm('Remove this product? Related stock rows may be removed.')) return;
    setMsg('');
    try {
      await api(`/api/v1/products/${detailId}`, { method: 'DELETE', token });
      if (focused?.kind === 'product' && focused.id === detailId) setFocused(null);
      closeDetail();
      await refresh();
    } catch (e) {
      setMsg(e instanceof Error ? e.message : 'Error');
    }
  }

  const fmtMoney = (n: number) => (Number.isFinite(n) ? n.toFixed(2) : '—');

  const warehouseFocusClass = (w: WarehouseRow) =>
    focused?.kind === 'warehouse' && focused.id === w.id
      ? 'bg-yellow/15 ring-2 ring-yellow/40 ring-inset border-l-[3px] border-l-yellow'
      : 'border-l-[3px] border-l-transparent hover:bg-white/[0.05]';

  const productFocusClass = (pid: number) =>
    focused?.kind === 'product' && focused.id === pid
      ? 'bg-violet/15 ring-1 ring-violet/50 ring-inset'
      : 'hover:bg-white/[0.05]';

  const tabBtn = (id: WarehouseTabId, label: string) => (
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
    <div className="flex flex-col gap-6">
      <p className="text-muted text-sm">Warehouses, catalog products, and warehouse inventory</p>
      {msg && <p className="text-danger text-sm">{msg}</p>}

      <div className="flex flex-wrap gap-1 p-1 rounded-2xl bg-surface/40 border border-white/5 w-fit max-w-full">
        {tabBtn('locations', 'Locations')}
        {tabBtn('stock', 'Inventory')}
      </div>

      {tab === 'locations' && (
        <>
          <div className="flex flex-wrap gap-3">
            <button type="button" className="rounded-xl bg-yellow px-4 py-2 text-sm font-medium text-[#1e1e24]" onClick={() => setModalWh(true)}>
              New warehouse
            </button>
            <button type="button" className="rounded-xl bg-violet px-4 py-2 text-sm font-medium text-white" onClick={() => setModalProd(true)}>
              New product
            </button>
            <button
              type="button"
              disabled={warehouses.length === 0 || productsSorted.length === 0}
              className="rounded-xl border border-white/15 px-4 py-2 text-sm font-medium text-text hover:bg-white/5 disabled:opacity-40 disabled:pointer-events-none"
              onClick={() => openStockInModal()}
              title={warehouses.length === 0 || productsSorted.length === 0 ? 'Create a warehouse and a product first' : undefined}
            >
              Add inventory
            </button>
          </div>

          <section className="rounded-[20px] border border-white/5 overflow-hidden">
            <div className="bg-surface/60 px-4 py-3 border-b border-white/5 flex items-center justify-between gap-3">
              <h2 className="text-lg font-medium">Warehouse tree</h2>
              <p className="text-xs text-muted max-w-[32ch] md:max-w-none text-right hidden sm:block">
                Expand each site to see stocked SKUs — click to edit
              </p>
            </div>

            <div className="divide-y divide-white/5">
              {warehouses.length === 0 && (
                <div className="p-8 text-muted text-center text-sm">No warehouses — create one above</div>
              )}

              {warehouses.map((w) => {
                const expanded = !collapsedWhIds.has(w.id);
                const rows = stocksByWarehouse.get(w.id) ?? [];

                return (
                  <div key={w.id} className="bg-card/40">
                    <div className={`flex items-stretch gap-0 transition-colors ${warehouseFocusClass(w)}`}>
                      <button
                        type="button"
                        aria-expanded={expanded}
                        aria-label={expanded ? `Collapse ${w.name}` : `Expand ${w.name}`}
                        className="shrink-0 w-11 flex items-center justify-center text-muted hover:text-text hover:bg-white/10 border-0 bg-transparent"
                        onClick={() =>
                          setCollapsedWhIds((prev) => {
                            const n = new Set(prev);
                            if (n.has(w.id)) n.delete(w.id);
                            else n.add(w.id);
                            return n;
                          })
                        }
                      >
                        <span className="text-xs">{expanded ? '▾' : '▸'}</span>
                      </button>
                      <button
                        type="button"
                        className="flex-1 min-w-0 text-left py-3 pr-3 pl-0 flex flex-wrap gap-x-4 gap-y-0.5 items-baseline border-0 bg-transparent text-text"
                        onClick={() => openWarehouse(w)}
                      >
                        <span className="font-mono text-sm opacity-90">{w.code}</span>
                        <span className="font-medium truncate">{w.name}</span>
                        <span className="text-xs text-muted tabular-nums">{w.record_date ?? '—'}</span>
                      </button>
                    </div>

                    {expanded && (
                      <div className="pb-4 pl-[2.85rem] pr-3 md:pr-4 border-t border-white/5 bg-bg/30 space-y-2">
                        <div className="flex justify-end pt-2">
                          <button
                            type="button"
                            disabled={productsSorted.length === 0}
                            className="text-xs rounded-xl border border-white/15 px-3 py-1.5 font-medium text-text hover:bg-white/5 disabled:opacity-40 disabled:pointer-events-none"
                            onClick={(e) => {
                              e.stopPropagation();
                              openStockInModal(w.id);
                            }}
                            title={productsSorted.length === 0 ? 'Add a catalogue product first' : undefined}
                          >
                            Add inventory
                          </button>
                        </div>
                        {rows.length === 0 ? (
                          <p className="py-5 text-muted text-sm pl-2">
                            No inventory yet — use <strong className="text-text">Add inventory</strong> here or from the toolbar above.
                          </p>
                        ) : (
                          <div className="rounded-xl border border-white/5 overflow-x-auto">
                            <table className="w-full text-sm min-w-[460px]">
                              <thead className="bg-surface/80 text-left text-muted text-[10px] uppercase tracking-wide">
                                <tr>
                                  <th className="p-2 pl-3">SKU</th>
                                  <th className="p-2">Product</th>
                                  <th className="p-2 text-right">Qty here</th>
                                  <th className="p-2 whitespace-nowrap">Stock date</th>
                                </tr>
                              </thead>
                              <tbody>
                                {rows.map((s) => (
                                  <tr
                                    key={`${w.id}-${s.product_id}`}
                                    className={`border-t border-white/5 cursor-pointer ${productFocusClass(s.product_id)}`}
                                    onClick={() => openProduct(s.product_id)}
                                  >
                                    <td className="p-2 pl-3 font-mono">{s.sku}</td>
                                    <td className="p-2 text-muted">{s.product_name}</td>
                                    <td className="p-2 text-right tabular-nums">{s.quantity}</td>
                                    <td className="p-2 tabular-nums text-muted">{s.stock_date ?? '—'}</td>
                                  </tr>
                                ))}
                              </tbody>
                            </table>
                          </div>
                        )}
                      </div>
                    )}
                  </div>
                );
              })}
            </div>
          </section>

          {orphanProducts.length > 0 && (
            <section className="rounded-[20px] border border-white/5 overflow-hidden">
              <div className="bg-surface/60 px-4 py-3 border-b border-white/5">
                <h2 className="text-lg font-medium">Products (no balances yet)</h2>
                <p className="text-xs text-muted mt-1">These catalogue items never appear above until you stock them in somewhere.</p>
              </div>
              <div className="overflow-x-auto">
                <table className="w-full text-sm min-w-[480px]">
                  <thead className="bg-surface text-left text-muted text-xs uppercase">
                    <tr>
                      <th className="p-3">SKU</th>
                      <th className="p-3">Name</th>
                      <th className="p-3">Description</th>
                      <th className="p-3 text-right">Cost</th>
                      <th className="p-3 text-right">Price</th>
                      <th className="p-3">Record date</th>
                    </tr>
                  </thead>
                  <tbody>
                    {orphanProducts.map((p) => (
                      <tr
                        key={p.id}
                        className={`border-t border-white/5 cursor-pointer ${productFocusClass(p.id)}`}
                        onClick={() => openProduct(p.id)}
                      >
                        <td className="p-3 font-mono">{p.sku}</td>
                        <td className="p-3">{p.name}</td>
                        <td className="p-3 text-muted max-w-[200px] truncate" title={p.description}>
                          {p.description || '—'}
                        </td>
                        <td className="p-3 text-right tabular-nums">{fmtMoney(p.cost_price)}</td>
                        <td className="p-3 text-right tabular-nums">{fmtMoney(p.selling_price)}</td>
                        <td className="p-3 tabular-nums">{p.record_date ?? '—'}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </section>
          )}
        </>
      )}

      {tab === 'stock' && (
        <>
          <div className="flex flex-wrap items-center gap-3">
            <h2 className="text-lg font-medium">Inventory by warehouse</h2>
            <button
              type="button"
              disabled={warehouses.length === 0 || productsSorted.length === 0}
              className="rounded-xl bg-violet px-4 py-2 text-sm font-medium text-white disabled:opacity-40 disabled:pointer-events-none"
              onClick={() => openStockInModal()}
              title={warehouses.length === 0 || productsSorted.length === 0 ? 'Create a warehouse and a product first' : undefined}
            >
              Add inventory
            </button>
            <button type="button" className="rounded-xl border border-white/15 px-3 py-1.5 text-sm text-muted hover:text-text hover:bg-white/5" onClick={() => setStockHelpOpen(true)}>
              How this works
            </button>
          </div>
          <div className="rounded-[20px] border border-white/5 overflow-x-auto">
            <table className="w-full text-sm min-w-[560px]">
              <thead className="bg-surface text-left text-muted text-xs uppercase">
                <tr>
                  <th className="p-3">Warehouse</th>
                  <th className="p-3">SKU</th>
                  <th className="p-3 text-right">Qty</th>
                  <th className="p-3">Stock date</th>
                </tr>
              </thead>
              <tbody>
                {stocks.length === 0 && (
                  <tr>
                    <td colSpan={4} className="p-4 text-muted text-center">
                      No inventory yet — use <strong className="text-text">Locations → Add inventory</strong> or open the Inventory tab
                    </td>
                  </tr>
                )}
                {stocks.map((s) => (
                  <tr key={`${s.warehouse_id}-${s.product_id}`} className="border-t border-white/5">
                    <td className="p-3">{s.warehouse_name}</td>
                    <td className="p-3 font-mono">{s.sku}</td>
                    <td className="p-3 text-right tabular-nums">{s.quantity}</td>
                    <td className="p-3 tabular-nums text-muted">{s.stock_date ?? '—'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </>
      )}

      {stockHelpOpen && <StockHelpModal onClose={() => setStockHelpOpen(false)} />}

      {stockInOpen && (
        <ModalShell title="Add inventory" onClose={() => setStockInOpen(false)}>
          <form onSubmit={submitStockIn} className="space-y-3">
            <p className="text-xs text-muted">Increases on-hand qty for one product at one warehouse.</p>
            <div>
              <label className="text-xs text-muted block mb-1">Warehouse</label>
              <select
                className="w-full rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm text-text"
                value={stockInWhId}
                onChange={(e) => setStockInWhId(e.target.value)}
                required={warehouses.length > 0}
              >
                {warehouses.length === 0 ? <option value="">No warehouses</option> : null}
                {warehouses.map((w) => (
                  <option key={w.id} value={String(w.id)}>
                    {w.code} — {w.name}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label className="text-xs text-muted block mb-1">Product</label>
              <select
                className="w-full rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm text-text"
                value={stockInProductId}
                onChange={(e) => setStockInProductId(e.target.value)}
                required={productsSorted.length > 0}
              >
                {productsSorted.map((p) => (
                  <option key={p.id} value={String(p.id)}>
                    {p.sku} — {p.name}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label className="text-xs text-muted block mb-1">Quantity</label>
              <input
                className="w-full rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm text-text tabular-nums"
                inputMode="decimal"
                value={stockInQty}
                onChange={(e) => setStockInQty(e.target.value)}
                step="any"
                required
              />
            </div>
            <div>
              <label className="text-xs text-muted block mb-1">Stock date</label>
              <input type="date" className="w-full rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm text-text" value={stockInDate} onChange={(e) => setStockInDate(e.target.value)} required />
            </div>
            <div>
              <label className="text-xs text-muted block mb-1">Reference (optional)</label>
              <input
                className="w-full rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm text-text"
                placeholder="Receipt, PO, note…"
                value={stockInRef}
                onChange={(e) => setStockInRef(e.target.value)}
              />
            </div>
            <div className="flex gap-2 justify-end">
              <button type="button" className="rounded-xl border border-white/15 px-4 py-2 text-sm" onClick={() => setStockInOpen(false)}>
                Cancel
              </button>
              <button type="submit" className="rounded-xl bg-violet px-4 py-2 text-sm font-medium text-white" disabled={warehouses.length === 0 || productsSorted.length === 0}>
                Add stock
              </button>
            </div>
          </form>
        </ModalShell>
      )}

      {modalWh && (
        <ModalShell title="New warehouse" onClose={() => setModalWh(false)}>
          <form onSubmit={submitAddWarehouse} className="space-y-3">
            <input
              className="w-full rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm text-text"
              placeholder="Name"
              value={whAddName}
              onChange={(e) => setWhAddName(e.target.value)}
              required
            />
            <input
              className="w-full rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm text-text"
              placeholder="Code"
              value={whAddCode}
              onChange={(e) => setWhAddCode(e.target.value)}
              required
            />
            <input
              className="w-full rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm text-text"
              placeholder="Location (optional)"
              value={whAddLoc}
              onChange={(e) => setWhAddLoc(e.target.value)}
            />
            <div>
              <label className="text-xs text-muted block mb-1">Record date</label>
              <input type="date" className="w-full rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm text-text" value={whAddDate} onChange={(e) => setWhAddDate(e.target.value)} required />
            </div>
            <div>
              <label className="text-xs text-muted block mb-1">detail (JSON)</label>
              <textarea
                rows={5}
                className="w-full rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm font-mono text-text"
                placeholder='{"notes":""}'
                value={whAddDetail}
                onChange={(e) => setWhAddDetail(e.target.value)}
              />
            </div>
            <div className="flex gap-2 justify-end">
              <button type="button" className="rounded-xl border border-white/15 px-4 py-2 text-sm" onClick={() => setModalWh(false)}>
                Cancel
              </button>
              <button type="submit" className="rounded-xl bg-yellow px-4 py-2 text-sm font-medium text-[#1e1e24]">
                Save
              </button>
            </div>
          </form>
        </ModalShell>
      )}

      {modalProd && (
        <ModalShell title="New product" onClose={() => setModalProd(false)}>
          <form onSubmit={submitAddProduct} className="space-y-3">
            <input
              className="w-full rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm text-text"
              placeholder="SKU"
              value={pAddSku}
              onChange={(e) => setPAddSku(e.target.value)}
              required
            />
            <input
              className="w-full rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm text-text"
              placeholder="Name"
              value={pAddName}
              onChange={(e) => setPAddName(e.target.value)}
              required
            />
            <textarea
              className="w-full rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm text-text min-h-[72px]"
              placeholder="Description"
              value={pAddDesc}
              onChange={(e) => setPAddDesc(e.target.value)}
            />
            <div className="grid grid-cols-2 gap-2">
              <div>
                <label className="text-xs text-muted block mb-1">Cost</label>
                <input className="w-full rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm text-text tabular-nums" inputMode="decimal" value={pAddCost} onChange={(e) => setPAddCost(e.target.value)} />
              </div>
              <div>
                <label className="text-xs text-muted block mb-1">Price</label>
                <input className="w-full rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm text-text tabular-nums" inputMode="decimal" value={pAddPrice} onChange={(e) => setPAddPrice(e.target.value)} />
              </div>
            </div>
            <div>
              <label className="text-xs text-muted block mb-1">Record date</label>
              <input type="date" className="w-full rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm text-text" value={pAddDate} onChange={(e) => setPAddDate(e.target.value)} required />
            </div>
            <div>
              <label className="text-xs text-muted block mb-1">detail (JSON)</label>
              <textarea
                rows={5}
                className="w-full rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm font-mono text-text"
                placeholder='{}'
                value={pAddDetail}
                onChange={(e) => setPAddDetail(e.target.value)}
              />
            </div>
            <div className="flex gap-2 justify-end">
              <button type="button" className="rounded-xl border border-white/15 px-4 py-2 text-sm" onClick={() => setModalProd(false)}>
                Cancel
              </button>
              <button type="submit" className="rounded-xl bg-violet px-4 py-2 text-sm font-medium text-white">
                Save
              </button>
            </div>
          </form>
        </ModalShell>
      )}

      {detailKind === 'warehouse' && detailId !== null && (
        <DetailModalShell title="Warehouse" onClose={closeDetail}>
          {detailLoading ? (
            <p className="text-muted text-sm">Loading…</p>
          ) : (
            <form onSubmit={saveWarehouseDetail} className="space-y-3">
              <input
                className="w-full rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm text-text"
                value={whEditName}
                onChange={(e) => setWhEditName(e.target.value)}
                required
              />
              <input
                className="w-full rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm font-mono text-text"
                value={whEditCode}
                onChange={(e) => setWhEditCode(e.target.value)}
                required
              />
              <input
                className="w-full rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm text-text"
                placeholder="Location"
                value={whEditLoc}
                onChange={(e) => setWhEditLoc(e.target.value)}
              />
              <div>
                <label className="text-xs text-muted block mb-1">Record date</label>
                <input type="date" className="w-full rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm text-text" value={whEditDate} onChange={(e) => setWhEditDate(e.target.value)} required />
              </div>
              <div>
                <label className="text-xs text-muted block mb-1">detail (JSON)</label>
                <textarea rows={6} className="w-full rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm font-mono text-text" value={whEditDetail} onChange={(e) => setWhEditDetail(e.target.value)} />
              </div>
              <div className="flex flex-wrap gap-2 justify-between pt-2">
                <button type="button" className="rounded-xl border border-red-500/40 text-red-400 px-4 py-2 text-sm" onClick={removeWarehouse}>
                  Remove
                </button>
                <div className="flex gap-2">
                  <button type="button" className="rounded-xl border border-white/15 px-4 py-2 text-sm" onClick={closeDetail}>
                    Cancel
                  </button>
                  <button type="submit" className="rounded-xl bg-yellow px-4 py-2 text-sm font-medium text-[#1e1e24]">
                    Save
                  </button>
                </div>
              </div>
            </form>
          )}
        </DetailModalShell>
      )}

      {detailKind === 'product' && detailId !== null && (
        <DetailModalShell title="Product" onClose={closeDetail}>
          {detailLoading ? (
            <p className="text-muted text-sm">Loading…</p>
          ) : (
            <form onSubmit={saveProductDetail} className="space-y-3">
              <input className="w-full rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm font-mono text-text" value={pEditSku} onChange={(e) => setPEditSku(e.target.value)} required />
              <input className="w-full rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm text-text" value={pEditName} onChange={(e) => setPEditName(e.target.value)} required />
              <textarea className="w-full rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm text-text min-h-[72px]" value={pEditDesc} onChange={(e) => setPEditDesc(e.target.value)} />
              <div className="grid grid-cols-2 gap-2">
                <div>
                  <label className="text-xs text-muted block mb-1">Cost</label>
                  <input className="w-full rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm tabular-nums" value={pEditCost} onChange={(e) => setPEditCost(e.target.value)} />
                </div>
                <div>
                  <label className="text-xs text-muted block mb-1">Price</label>
                  <input className="w-full rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm tabular-nums" value={pEditPrice} onChange={(e) => setPEditPrice(e.target.value)} />
                </div>
              </div>
              <div>
                <label className="text-xs text-muted block mb-1">Record date</label>
                <input type="date" className="w-full rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm" value={pEditDate} onChange={(e) => setPEditDate(e.target.value)} required />
              </div>
              <div className="grid grid-cols-2 gap-2">
                <input className="rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm" placeholder="Category" value={pEditCat} onChange={(e) => setPEditCat(e.target.value)} />
                <input className="rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm" placeholder="Unit" value={pEditUnit} onChange={(e) => setPEditUnit(e.target.value)} />
              </div>
              <input className="w-full rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm" placeholder="Barcode" value={pEditBc} onChange={(e) => setPEditBc(e.target.value)} />
              <div>
                <label className="text-xs text-muted block mb-1">Reorder threshold</label>
                <input className="w-full rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm tabular-nums" value={pEditReorder} onChange={(e) => setPEditReorder(e.target.value)} />
              </div>
              <div>
                <label className="text-xs text-muted block mb-1">detail (JSON)</label>
                <textarea rows={6} className="w-full rounded-xl bg-bg border border-white/10 px-3 py-2 text-sm font-mono text-text" value={pEditDetail} onChange={(e) => setPEditDetail(e.target.value)} />
              </div>
              <div className="flex flex-wrap gap-2 justify-between pt-2">
                <button type="button" className="rounded-xl border border-red-500/40 text-red-400 px-4 py-2 text-sm" onClick={removeProduct}>
                  Remove
                </button>
                <div className="flex gap-2">
                  <button type="button" className="rounded-xl border border-white/15 px-4 py-2 text-sm" onClick={closeDetail}>
                    Cancel
                  </button>
                  <button type="submit" className="rounded-xl bg-violet px-4 py-2 text-sm font-medium text-white">
                    Save
                  </button>
                </div>
              </div>
            </form>
          )}
        </DetailModalShell>
      )}
    </div>
  );
}
