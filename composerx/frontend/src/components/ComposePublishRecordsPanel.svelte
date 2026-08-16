<script>
  /** @type {{ apiBase: string, getAuthHeaders: (extra?: Record<string, string>) => Record<string, string>, notify?: (kind?: string, msg?: string) => void }} */
  let { apiBase, getAuthHeaders, notify = () => {} } = $props()

  let drafts = $state([])
  let draftsLoading = $state(false)
  let draftsTotal = $state(0)
  let published = $state([])
  let publishedLoading = $state(false)
  let publishedTotal = $state(0)

  /** @type {{ id: number, name: string, html_content: string, theme: string, updated_at: string } | null} */
  let draftDetail = $state(null)
  let draftDetailOpen = $state(false)
  let draftDetailLoading = $state(false)

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
      draftsTotal = data?.total || drafts.length
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
      publishedTotal = data?.total || published.length
    } catch (err) {
      notify('error', err instanceof Error ? err.message : 'Failed to load publish history')
    } finally {
      publishedLoading = false
    }
  }

  async function openDraftDetail(id) {
    draftDetailLoading = true
    draftDetailOpen = true
    try {
      draftDetail = await apiGet(`/publish-drafts/${id}`)
    } catch (err) {
      notify('error', err instanceof Error ? err.message : 'Failed to load draft')
      draftDetailOpen = false
    } finally {
      draftDetailLoading = false
    }
  }

  async function deleteDraft(id) {
    if (!confirm('Delete this saved HTML draft?')) return
    try {
      await apiDelete(`/publish-drafts/${id}`)
      notify('info', 'Draft deleted.')
      await loadDrafts()
    } catch (err) {
      notify('error', err instanceof Error ? err.message : 'Failed to delete draft')
    }
  }

  $effect(() => {
    void loadDrafts()
    void loadPublished()
  })
</script>

<section class="panel panel-wide publish-records">
  <header class="panel-header">
    <div>
      <h2>Published contents</h2>
      <span class="panel-meta">Saved HTML drafts and published page history.</span>
    </div>
    <div class="publish-records-actions">
      <button type="button" class="btn-secondary" onclick={loadDrafts} disabled={draftsLoading}>Refresh drafts</button>
      <button type="button" class="btn-secondary" onclick={loadPublished} disabled={publishedLoading}>Refresh history</button>
    </div>
  </header>

  <div class="panel-body publish-records-body">
    <div class="publish-records-grid">
      <section class="publish-records-card">
        <div class="publish-records-card-head">
          <h3>Saved HTML files</h3>
          <span>{draftsTotal}</span>
        </div>
        <div class="table-scroll">
          <table class="grid">
            <thead>
              <tr>
                <th>Name</th>
                <th>Theme</th>
                <th>Updated</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {#if draftsLoading}
                <tr class="grid-empty"><td colspan="4">Loading drafts…</td></tr>
              {:else if drafts.length === 0}
                <tr class="grid-empty"><td colspan="4">No saved HTML yet. Use <strong>Save HTML</strong> in Compose &amp; Publish.</td></tr>
              {:else}
                {#each drafts as row}
                  <tr>
                    <td>{row.name}</td>
                    <td>{row.theme || 'default'}</td>
                    <td>{row.updated_at ? new Date(row.updated_at).toLocaleString() : '—'}</td>
                    <td class="grid-actions">
                      <button type="button" class="btn-ghost" onclick={() => openDraftDetail(row.id)}>View</button>
                      <button type="button" class="btn-ghost btn-danger-lite" onclick={() => deleteDraft(row.id)}>Delete</button>
                    </td>
                  </tr>
                {/each}
              {/if}
            </tbody>
          </table>
        </div>
      </section>

      <section class="publish-records-card">
        <div class="publish-records-card-head">
          <h3>Published history</h3>
          <span>{publishedTotal}</span>
        </div>
        <div class="table-scroll">
          <table class="grid">
            <thead>
              <tr>
                <th>Name</th>
                <th>Path</th>
                <th>Updated</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {#if publishedLoading}
                <tr class="grid-empty"><td colspan="4">Loading history…</td></tr>
              {:else if published.length === 0}
                <tr class="grid-empty"><td colspan="4">No published pages yet.</td></tr>
              {:else}
                {#each published as row}
                  <tr>
                    <td>{row.name}</td>
                    <td><code>/public/p/{row.slug}</code></td>
                    <td>{row.updated_at ? new Date(row.updated_at).toLocaleString() : '—'}</td>
                    <td class="grid-actions">
                      <a class="btn-ghost publish-open-link" href={`/public/p/${row.slug}`} target="_blank" rel="noreferrer">Open</a>
                    </td>
                  </tr>
                {/each}
              {/if}
            </tbody>
          </table>
        </div>
      </section>
    </div>
  </div>
</section>

{#if draftDetailOpen}
  <div
    class="modal-backdrop"
    role="presentation"
    tabindex="-1"
    onclick={() => (draftDetailOpen = false)}
    onkeydown={(e) => e.key === 'Escape' && (draftDetailOpen = false)}
  >
    <div
      class="modal-panel publish-draft-modal"
      role="dialog"
      aria-modal="true"
      aria-labelledby="publish-draft-modal-title"
      tabindex="-1"
      onclick={(e) => e.stopPropagation()}
      onkeydown={(e) => e.stopPropagation()}
    >
      <header class="modal-header">
        <h2 id="publish-draft-modal-title" class="modal-title">{draftDetail?.name || 'Draft HTML'}</h2>
        <button type="button" class="icon-btn modal-close" aria-label="Close" onclick={() => (draftDetailOpen = false)}>×</button>
      </header>
      <div class="modal-body">
        {#if draftDetailLoading}
          <p class="sidebar-hint">Loading…</p>
        {:else if draftDetail}
          <div class="publish-draft-meta">Theme: <strong>{draftDetail.theme || 'default'}</strong></div>
          <textarea class="publish-draft-html-view" readonly value={draftDetail.html_content}></textarea>
        {/if}
      </div>
    </div>
  </div>
{/if}

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
    align-items: flex-end;
    gap: 0.9rem;
    flex-wrap: wrap;
    border-bottom: 1px solid var(--color-border-subtle);
    padding: 0.7rem 0.8rem;
  }

  .panel-header h2 {
    margin: 0;
    font-size: 1rem;
  }

  .panel-meta {
    font-size: 0.76rem;
    color: var(--color-text-subtle);
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
  }

  .publish-records-body {
    overflow: hidden;
  }

  .publish-records-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 0.85rem;
    min-height: 0;
  }

  .publish-records-card {
    min-height: 0;
    display: flex;
    flex-direction: column;
    border: 1px solid var(--color-border-subtle);
    border-radius: 0.85rem;
    background: var(--color-bg-elevated);
    overflow: hidden;
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
    vertical-align: top;
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

  .publish-records-card-head {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.55rem 0.7rem;
    border-bottom: 1px solid var(--color-border-subtle);
    background: linear-gradient(90deg, var(--color-primary-soft), transparent);
  }

  .publish-records-card-head h3 {
    margin: 0;
    font-size: 0.84rem;
  }

  .publish-open-link {
    display: inline-flex;
    align-items: center;
    justify-content: center;
  }

  .btn-danger-lite {
    color: var(--color-danger);
  }

  .publish-draft-modal {
    max-width: min(64rem, 96vw);
    width: min(64rem, 96vw);
  }

  .modal-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(2, 8, 23, 0.55);
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 1rem;
    z-index: 70;
  }

  .modal-panel {
    border: 1px solid var(--color-border-subtle);
    border-radius: 0.95rem;
    background: var(--color-bg-elevated);
    width: 100%;
    max-height: min(92vh, 56rem);
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .modal-header {
    border-bottom: 1px solid var(--color-border-subtle);
    padding: 0.65rem 0.8rem;
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .modal-title {
    margin: 0;
    font-size: 0.95rem;
  }

  .modal-body {
    padding: 0.7rem;
    min-height: 0;
    overflow: auto;
  }

  .icon-btn {
    border: 1px solid var(--color-border-subtle);
    background: var(--color-bg);
    color: var(--color-text);
    width: 1.9rem;
    height: 1.9rem;
    border-radius: 0.55rem;
    cursor: pointer;
    line-height: 1;
    font-size: 1rem;
  }

  .publish-draft-meta {
    font-size: 0.8rem;
    color: var(--color-text-subtle);
    margin-bottom: 0.45rem;
  }

  .publish-draft-html-view {
    width: 100%;
    min-height: min(58vh, 32rem);
    resize: vertical;
    border-radius: 0.75rem;
    border: 1px solid var(--color-border-subtle);
    padding: 0.55rem 0.65rem;
    background: var(--color-bg);
    color: var(--color-text);
    font-size: 0.78rem;
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  }

  @media (max-width: 1080px) {
    .publish-records-grid {
      grid-template-columns: 1fr;
    }
  }
</style>
