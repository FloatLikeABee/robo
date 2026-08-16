<script lang="ts">
  import { api } from '../lib/api'

  type FlowEntry = {
    id: number
    entry_date: string
    direction: 'income' | 'expense'
    amount: number
    currency: string
    category: string
    status: string
    title: string
    notes: string
  }

  type FlowSummary = {
    from: string
    to: string
    income: number
    expense: number
    net: number
    entry_count: number
    by_category: Record<string, number>
  }

  const CATEGORY_SUGGESTIONS = ['Food', 'Travel', 'Salary', 'Rent', 'Software', 'Marketing', 'Materials', 'Other']
  const STATUS_SUGGESTIONS = ['logged', 'pending', 'cleared', 'reimbursed', 'planned', 'ignored']

  function todayISO(): string {
    return new Date().toISOString().slice(0, 10)
  }

  function monthStartISO(): string {
    const d = new Date()
    return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-01`
  }

  let entries = $state<FlowEntry[]>([])
  let summary = $state<FlowSummary | null>(null)
  let msg = $state('')
  let filterFrom = $state(monthStartISO())
  let filterTo = $state(todayISO())

  let entryDate = $state(todayISO())
  let direction = $state<'income' | 'expense'>('expense')
  let amount = $state('')
  let title = $state('')
  let category = $state('')
  let status = $state('logged')
  let notes = $state('')

  const topCategories = $derived(
    summary?.by_category
      ? Object.entries(summary.by_category)
          .sort((a, b) => Math.abs(b[1]) - Math.abs(a[1]))
          .slice(0, 6)
      : [],
  )

  async function load() {
    const q = new URLSearchParams({ from: filterFrom, to: filterTo, limit: '300' })
    const [listRes, sumRes] = await Promise.all([
      api<{ entries: FlowEntry[] | null }>(`/api/v1/flow-log/entries?${q}`),
      api<FlowSummary>(
        `/api/v1/flow-log/summary?from=${encodeURIComponent(filterFrom)}&to=${encodeURIComponent(filterTo)}`,
      ),
    ])
    entries = listRes.entries ?? []
    summary = sumRes
  }

  async function refresh() {
    msg = ''
    try {
      await load()
    } catch (e) {
      msg = e instanceof Error ? e.message : 'Failed to load Flow Log.'
    }
  }

  $effect(() => {
    filterFrom
    filterTo
    void refresh()
  })

  async function submit(e: Event) {
    e.preventDefault()
    msg = ''
    const amt = parseFloat(amount)
    if (!Number.isFinite(amt) || amt <= 0) {
      msg = 'Enter a positive amount.'
      return
    }
    try {
      await api('/api/v1/flow-log/entries', {
        method: 'POST',
        body: JSON.stringify({
          entry_date: entryDate,
          direction,
          amount: amt,
          title: title.trim() || (direction === 'income' ? 'Income' : 'Expense'),
          category: category.trim(),
          status: status.trim() || 'logged',
          notes: notes.trim(),
        }),
      })
      amount = ''
      title = ''
      notes = ''
      await load()
    } catch (err) {
      msg = err instanceof Error ? err.message : 'Failed to save entry.'
    }
  }

  async function removeEntry(id: number) {
    if (!window.confirm('Delete this Flow Log entry?')) return
    try {
      await api(`/api/v1/flow-log/entries/${id}`, { method: 'DELETE' })
      await load()
    } catch (err) {
      msg = err instanceof Error ? err.message : 'Delete failed.'
    }
  }
</script>

<div class="h-full min-h-0 flex flex-col gap-3">
  <div class="shrink-0 flex flex-wrap items-center justify-between gap-x-4 gap-y-2 border-b border-white/8 pb-2">
    <div class="min-w-0">
      <h1 class="text-base font-semibold text-text leading-tight">Flow Log</h1>
      <p class="text-xs text-muted leading-snug">
        Free-form money notes — category is free text. No double-entry rules.
      </p>
    </div>
    <div class="flex flex-wrap items-center gap-2 text-xs shrink-0">
      <div class="rounded-lg border border-white/8 bg-card/80 px-2.5 py-1.5">
        <span class="text-muted">Income </span>
        <span class="tabular-nums font-medium text-emerald-300">{(summary?.income ?? 0).toFixed(2)}</span>
      </div>
      <div class="rounded-lg border border-white/8 bg-card/80 px-2.5 py-1.5">
        <span class="text-muted">Expense </span>
        <span class="tabular-nums font-medium text-rose-300">{(summary?.expense ?? 0).toFixed(2)}</span>
      </div>
      <div class="rounded-lg border border-white/8 bg-card/80 px-2.5 py-1.5">
        <span class="text-muted">Net </span>
        <span class="tabular-nums font-medium text-teal">{(summary?.net ?? 0).toFixed(2)}</span>
      </div>
    </div>
  </div>

  {#if msg}
    <p class="shrink-0 text-rose-400 text-xs">{msg}</p>
  {/if}

  <div class="flex-1 min-h-0 grid gap-4 xl:grid-cols-[minmax(0,1fr)_240px] overflow-hidden">
    <div class="min-h-0 flex flex-col gap-3 overflow-hidden">
      <form class="shrink-0 rounded-2xl bg-card border border-white/5 p-3 space-y-2" onsubmit={submit}>
        <div class="flex flex-wrap gap-2 items-end">
          <div>
            <label class="text-[11px] text-muted block mb-0.5" for="fl-date">Date</label>
            <input id="fl-date" type="date" class="input !py-1.5 !text-sm" bind:value={entryDate} required />
          </div>
          <div>
            <label class="text-[11px] text-muted block mb-0.5" for="fl-dir">Type</label>
            <select id="fl-dir" class="input !py-1.5 !text-sm" bind:value={direction}>
              <option value="expense">Expense</option>
              <option value="income">Income</option>
            </select>
          </div>
          <div>
            <label class="text-[11px] text-muted block mb-0.5" for="fl-amt">Amount</label>
            <input id="fl-amt" class="input !py-1.5 !text-sm w-24" bind:value={amount} type="number" step="0.01" min="0.01" required />
          </div>
          <div class="min-w-[8rem] flex-1">
            <label class="text-[11px] text-muted block mb-0.5" for="fl-title">Title</label>
            <input id="fl-title" class="input !py-1.5 !text-sm w-full" bind:value={title} placeholder="What was this?" />
          </div>
          <div>
            <label class="text-[11px] text-muted block mb-0.5" for="fl-cat">Category</label>
            <input id="fl-cat" list="flow-categories" class="input !py-1.5 !text-sm w-28" bind:value={category} placeholder="Any text" />
            <datalist id="flow-categories">
              {#each CATEGORY_SUGGESTIONS as c}
                <option value={c}></option>
              {/each}
            </datalist>
          </div>
          <div>
            <label class="text-[11px] text-muted block mb-0.5" for="fl-status">Status</label>
            <input id="fl-status" list="flow-statuses" class="input !py-1.5 !text-sm w-24" bind:value={status} />
            <datalist id="flow-statuses">
              {#each STATUS_SUGGESTIONS as s}
                <option value={s}></option>
              {/each}
            </datalist>
          </div>
          <div class="min-w-[8rem] flex-1">
            <label class="text-[11px] text-muted block mb-0.5" for="fl-notes">Notes</label>
            <input id="fl-notes" class="input !py-1.5 !text-sm w-full" bind:value={notes} />
          </div>
          <button type="submit" class="btn-primary !py-1.5">Add</button>
        </div>
      </form>

      <div class="shrink-0 flex flex-wrap items-end gap-2">
        <input type="date" aria-label="Filter from" class="input !py-1.5 !text-sm" bind:value={filterFrom} />
        <span class="text-xs text-muted pb-1">→</span>
        <input type="date" aria-label="Filter to" class="input !py-1.5 !text-sm" bind:value={filterTo} />
        <span class="text-xs text-muted pb-1">{summary?.entry_count ?? 0} entries</span>
      </div>

      <div class="flex-1 min-h-0 rounded-2xl border border-white/5 overflow-auto">
        <table class="data-table w-full text-sm">
          <thead class="bg-surface text-left text-muted text-xs uppercase sticky top-0 z-10">
            <tr>
              <th class="p-2">Date</th>
              <th class="p-2">Type</th>
              <th class="p-2">Title</th>
              <th class="p-2 hidden md:table-cell">Category</th>
              <th class="p-2 hidden sm:table-cell">Status</th>
              <th class="p-2 text-right">Amount</th>
              <th class="p-2 w-12"></th>
            </tr>
          </thead>
          <tbody>
            {#each entries as e}
              <tr class="border-t border-white/5">
                <td class="p-2 whitespace-nowrap text-xs">{e.entry_date}</td>
                <td class="p-2 capitalize text-xs">
                  <span class={e.direction === 'income' ? 'text-emerald-300' : 'text-rose-300'}>{e.direction}</span>
                </td>
                <td class="p-2">
                  <div class="text-sm">{e.title}</div>
                  {#if e.notes}<div class="text-xs text-muted">{e.notes}</div>{/if}
                </td>
                <td class="p-2 text-muted text-xs hidden md:table-cell">{e.category || '—'}</td>
                <td class="p-2 hidden sm:table-cell">
                  <span class="rounded-full border border-white/10 px-1.5 py-0.5 text-[10px]">{e.status}</span>
                </td>
                <td class="p-2 text-right tabular-nums text-sm whitespace-nowrap">
                  {e.direction === 'expense' ? '−' : '+'}{e.amount.toFixed(2)}
                </td>
                <td class="p-2">
                  <button type="button" class="text-[11px] text-muted hover:text-rose-300" onclick={() => void removeEntry(e.id)}>×</button>
                </td>
              </tr>
            {:else}
              <tr>
                <td colspan="7" class="p-4 text-center text-muted text-sm">No entries yet.</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </div>

    <aside class="min-h-0 overflow-auto">
      {#if topCategories.length > 0}
        <div class="rounded-2xl border border-white/5 bg-card p-3">
          <h3 class="text-xs font-semibold mb-2">Top categories</h3>
          <ul class="space-y-1 text-xs">
            {#each topCategories as [cat, val]}
              <li class="flex justify-between gap-2">
                <span class="text-muted truncate">{cat}</span>
                <span class="tabular-nums shrink-0">{val.toFixed(2)}</span>
              </li>
            {/each}
          </ul>
        </div>
      {/if}
    </aside>
  </div>
</div>
