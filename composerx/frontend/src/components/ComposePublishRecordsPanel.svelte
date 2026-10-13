<script>
  import { mergePublishedContents } from '../lib/publishedContents.js'
  import { confirm } from '../lib/confirmDialog.js'

  /** @type {{ apiBase: string, getAuthHeaders: (extra?: Record<string, string>) => Record<string, string>, notify?: (kind?: string, msg?: string) => void }} */
  let { apiBase, getAuthHeaders, notify = () => {} } = $props()

  let drafts = $state([])
  let draftsLoading = $state(false)
  let published = $state([])
  let publishedLoading = $state(false)

  const rows = $derived(mergePublishedContents(drafts, published))
  const loading = $derived(draftsLoading || publishedLoading)

  function endpoint(path) {
    return `${apiBase}${path}`
  }

  async function apiGet(path) {
    const res = await fetch(endpoint(path), { headers: getAuthHeaders() })
    if (!res.ok) {
      let msg = `GET ${path} failed: ${res.status}`
      try {
        const payload = await res.json()
        if (payload?.error) msg = payload.error
      } catch {
        // ignore
      }
      throw new Error(msg)
    }
    return res.json()
  }

  async function apiDelete(path) {
    const res = await fetch(endpoint(path), { method: 'DELETE', headers: getAuthHeaders() })
    if (!res.ok) {
      let msg = `DELETE ${path} failed: ${res.status}`
      try {
        const payload = await res.json()
        if (payload?.error) msg = payload.error
      } catch {
        // ignore
      }
      throw new Error(msg)
    }
  }

  async function loadDrafts() {
    draftsLoading = true
    try {
      const data = await apiGet('/publish-drafts?limit=200&offset=0')
      drafts = data?.items || []
    } catch (err) {
      notify('error', err instanceof Error ? err.message : 'Failed to load drafts')
    } finally {
      draftsLoading = false
    }
  }

  async function loadPublished() {
    publishedLoading = true
    try {
      const data = await apiGet('/publishes/history?limit=200&offset=0')
      published = data?.items || []
    } catch (err) {
      notify('error', err instanceof Error ? err.message : 'Failed to load publish history')
    } finally {
      publishedLoading = false
    }
  }

  async function refreshAll() {
    await Promise.all([loadDrafts(), loadPublished()])
  }

  async function deleteDraft(id) {
    if (!(await confirm({ message: 'Delete this saved HTML draft?', danger: true, confirmLabel: 'Delete' }))) return
    try {
      await apiDelete(`/publish-drafts/${id}`)
      notify('info', 'Draft deleted.')
      await loadDrafts()
    } catch (err) {
      notify('error', err instanceof Error ? err.message : 'Failed to delete draft')
    }
  }

  async function removePublished(row) {
    const msg = row.draftId
      ? 'Remove this published page and its saved draft? The public page will no longer be available.'
      : 'Remove this published page? The public page will no longer be available.'
    if (!(await confirm({ message: msg, danger: true, confirmLabel: 'Remove' }))) return
    try {
      await apiDelete(`/publishes/${row.publishedId}`)
      if (row.draftId) {
        await apiDelete(`/publish-drafts/${row.draftId}`)
      }
      notify('info', 'Published page removed.')
      await refreshAll()
    } catch (err) {
      notify('error', err instanceof Error ? err.message : 'Failed to remove published page')
    }
  }

  $effect(() => {
    void refreshAll()
  })
</script>

<section class="panel panel-wide publish-records">
  <header class="panel-header">
    <h2>Published contents</h2>
    <div class="publish-records-actions">
      <button type="button" class="btn-secondary" onclick={refreshAll} disabled={loading}>Refresh</button>
    </div>
  </header>

  <div class="panel-body publish-records-body">
    <div class="table-scroll">
      <table class="grid">
        <thead>
          <tr>
            <th>Name</th>
            <th>Status</th>
            <th>Path</th>
            <th>Updated</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          {#if loading && rows.length === 0}
            <tr class="grid-empty"><td colspan="5">Loading…</td></tr>
          {:else if rows.length === 0}
            <tr class="grid-empty"><td colspan="5">No saved or published HTML yet. Use <strong>Save HTML</strong> or <strong>Publish</strong> in Compose &amp; Publish.</td></tr>
          {:else}
            {#each rows as row (row.key)}
              <tr>
                <td>{row.name}</td>
                <td>{row.status}</td>
                <td>{#if row.path}<code>{row.path}</code>{:else}—{/if}</td>
                <td>{row.updated_at ? new Date(row.updated_at).toLocaleString() : '—'}</td>
                <td class="grid-actions">
                  <span class="grid-actions-inner">
                    {#if row.canDelete && row.publishedId}
                      <button type="button" class="btn-ghost btn-danger-lite" onclick={() => removePublished(row)}>Remove</button>
                    {:else if row.canDelete}
                      <button type="button" class="btn-ghost btn-danger-lite" onclick={() => deleteDraft(row.draftId)}>Delete</button>
                    {/if}
                    {#if row.canOpen}
                      <a class="btn-ghost publish-open-link" href={row.path} target="_blank" rel="noreferrer">Open</a>
                    {/if}
                  </span>
                </td>
              </tr>
            {/each}
          {/if}
        </tbody>
      </table>
    </div>
  </div>
</section>

<style>
  .btn-secondary {
    background: var(--color-bg-elevated);
    color: var(--color-text);
    border: 1px solid var(--color-border-subtle);
    border-radius: 0.65rem;
    padding: 0.38rem 0.58rem;
    font-size: 0.79rem;
    cursor: pointer;
  }

  .btn-ghost {
    background: transparent;
    color: var(--color-text-subtle);
    border: 1px solid transparent;
    border-radius: 0.6rem;
    padding: 0.3rem 0.45rem;
    font-size: 0.76rem;
    cursor: pointer;
    text-decoration: none;
    display: inline-flex;
    align-items: center;
    line-height: 1;
    text-box: trim-both cap alphabetic;
  }

  .btn-ghost:hover {
    background: var(--color-primary-soft);
    color: var(--color-text);
  }

  .publish-records {
    grid-column: 1 / -1;
    border: 1px solid var(--color-border-subtle);
    border-radius: 0.95rem;
    background: var(--color-bg-elevated);
    min-height: 0;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .panel-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 0.9rem;
    flex-wrap: wrap;
    border-bottom: 1px solid var(--color-border-subtle);
    padding: 0.7rem 0.8rem;
  }

  .panel-header h2 {
    margin: 0;
    font-size: 1rem;
  }

  .panel-body {
    min-height: 0;
    flex: 1;
    padding: 0.7rem;
  }

  .publish-records-actions {
    display: flex;
    gap: 0.5rem;
    flex-wrap: wrap;
    align-items: center;
  }

  .publish-records-body {
    overflow: hidden;
    display: flex;
    flex-direction: column;
  }

  .table-scroll {
    overflow: auto;
    min-height: 0;
    flex: 1;
  }

  table.grid {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.8rem;
  }

  table.grid th,
  table.grid td {
    padding: 0.48rem 0.55rem;
    text-align: left;
    vertical-align: middle;
    border-bottom: 1px solid var(--color-border-subtle);
  }

  table.grid th {
    font-size: 0.72rem;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--color-text-subtle);
  }

  .grid-empty td {
    color: var(--color-text-subtle);
    font-size: 0.78rem;
  }

  .grid-actions {
    white-space: nowrap;
  }

  .grid-actions-inner {
    display: inline-flex;
    align-items: center;
    gap: 0.2rem;
  }

  .publish-open-link {
    display: inline-flex;
    align-items: center;
    justify-content: center;
  }

  .btn-danger-lite {
    color: var(--color-danger);
  }
</style>
