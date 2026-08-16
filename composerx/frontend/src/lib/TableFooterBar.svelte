<script>
  /**
   * @typedef {Object} Props
   * @property {number} [total]
   * @property {number} [offset]
   * @property {number} [limit]
   * @property {() => void} [onApply]
   */

  /** @type {Props} */
  let { total = 0, offset = $bindable(0), limit = $bindable(25), onApply = () => {} } = $props()

  const pageStart = $derived(total === 0 ? 0 : offset + 1)
  const pageEnd = $derived(Math.min(offset + limit, total))
  const canPrev = $derived(offset > 0)
  const canNext = $derived(offset + limit < total)

  function prev() {
    offset = Math.max(0, offset - limit)
    onApply()
  }

  function next() {
    if (offset + limit >= total) return
    offset += limit
    onApply()
  }

  function onLimitChange() {
    limit = typeof limit === 'string' ? parseInt(limit, 10) : limit
    offset = 0
    onApply()
  }
</script>

<div class="table-footer-bar">
  <span class="table-footer-count">
    {#if total === 0}
      0 records
    {:else}
      {pageStart}–{pageEnd} of {total}
    {/if}
  </span>
  <div class="table-footer-controls">
    <label class="table-footer-page-size">
      <span class="sr-only">Rows per page</span>
      <select bind:value={limit} onchange={onLimitChange} aria-label="Rows per page">
        <option value={10}>10 / page</option>
        <option value={25}>25 / page</option>
        <option value={50}>50 / page</option>
        <option value={100}>100 / page</option>
      </select>
    </label>
    <div class="table-footer-pager">
      <button type="button" class="btn-footer" disabled={!canPrev} onclick={prev} aria-label="Previous page">
        Prev
      </button>
      <button type="button" class="btn-footer" disabled={!canNext} onclick={next} aria-label="Next page">
        Next
      </button>
    </div>
  </div>
</div>

<style>
  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
  }

  .table-footer-bar {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: 0.5rem 0.75rem;
    margin-top: 0.5rem;
    padding: 0.45rem 0.55rem;
    border-radius: 0.65rem;
    border: 1px solid var(--color-border-subtle);
    background: var(--color-bg-elevated);
    font-size: 0.78rem;
    color: var(--color-text-subtle);
  }

  .table-footer-count {
    font-variant-numeric: tabular-nums;
  }

  .table-footer-controls {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .table-footer-page-size select {
    border-radius: 0.5rem;
    border: 1px solid var(--color-border-subtle);
    padding: 0.25rem 0.45rem;
    font-size: 0.75rem;
    font-family: inherit;
    background: var(--color-bg);
    color: var(--color-text);
    cursor: pointer;
  }

  .table-footer-pager {
    display: flex;
    gap: 0.25rem;
  }

  .btn-footer {
    border-radius: 0.5rem;
    padding: 0.25rem 0.65rem;
    font-size: 0.75rem;
    font-weight: 500;
    font-family: inherit;
    border: 1px solid var(--color-border-subtle);
    background: var(--color-surface);
    color: var(--color-text);
    cursor: pointer;
    box-shadow: none;
  }

  .btn-footer:hover:not(:disabled) {
    background: var(--color-primary-soft);
    border-color: var(--color-primary-muted);
    color: var(--color-primary);
  }

  .btn-footer:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }
</style>
