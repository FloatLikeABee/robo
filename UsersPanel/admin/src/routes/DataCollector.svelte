<script lang="ts">
  import { onDestroy, onMount } from 'svelte'
  import { apiFetch, apiJson } from '../lib/api'

  type TabId = 'data-type' | 'upload' | 'validation' | 'progress'

  type EntitySpec = {
    kind: string
    label: string
    description: string
    required_fields: string[]
    optional_fields: string[]
    template_headers: string[]
    csv_template: string
    csv_example: string
    json_example: string
    instructions: string[]
  }

  type ValidationReport = {
    valid: boolean
    entity: string
    row_count: number
    header_count: number
    uses_template: boolean
    matched_headers: string[]
    extra_headers: string[]
    missing_required: string[]
    warnings: string[]
    sample_issues: string[]
    message: string
  }

  type ImportJob = {
    id: string
    entity: string
    filename: string
    format: string
    status: string
    percent: number
    total_rows: number
    processed_rows: number
    imported: number
    failed: number
    uses_template: boolean
    message: string
    errors: string[]
    row_results: { row_ref: string; success: boolean; message: string; record_id?: number }[]
    finished_at?: string
  }

  const tabs: { id: TabId; label: string }[] = [
    { id: 'data-type', label: 'Data Type' },
    { id: 'upload', label: 'Upload File' },
    { id: 'validation', label: 'Validation' },
    { id: 'progress', label: 'Progress' },
  ]

  let entities = $state<EntitySpec[]>([])
  let activeTab = $state<TabId>('data-type')
  let selectedEntity = $state('member')
  let selectedSpec = $derived(entities.find((e) => e.kind === selectedEntity))
  let file = $state<File | null>(null)
  let err = $state('')
  let validation = $state<ValidationReport | null>(null)
  let validating = $state(false)
  let starting = $state(false)
  let jobId = $state<string | null>(null)
  let job = $state<ImportJob | null>(null)
  let pollTimer: ReturnType<typeof setInterval> | null = null

  onMount(() => {
    void loadEntities()
  })

  onDestroy(() => {
    if (pollTimer) clearInterval(pollTimer)
  })

  async function loadEntities() {
    err = ''
    try {
      const res = await apiJson<{ entities: EntitySpec[] }>('/api/data-collector/entities')
      entities = res.entities
      if (entities.length && !entities.some((e) => e.kind === selectedEntity)) {
        selectedEntity = entities[0].kind
      }
    } catch (e: unknown) {
      err = e instanceof Error ? e.message : 'Failed to load entities'
    }
  }

  function onFileChange(e: Event) {
    const input = e.target as HTMLInputElement
    file = input.files?.[0] ?? null
    validation = null
    job = null
    jobId = null
  }

  function buildFormData(): FormData {
    const fd = new FormData()
    fd.append('entity', selectedEntity)
    if (file) fd.append('file', file)
    return fd
  }

  async function validateSample() {
    if (!file) {
      err = 'Choose a file first'
      activeTab = 'upload'
      return
    }
    validating = true
    err = ''
    validation = null
    try {
      const res = await apiFetch('/api/data-collector/validate', {
        method: 'POST',
        body: buildFormData(),
      })
      const data = await res.json()
      if (!res.ok) throw new Error(data?.error ?? res.statusText)
      validation = data.report as ValidationReport
      activeTab = 'validation'
    } catch (e: unknown) {
      err = e instanceof Error ? e.message : 'Validation failed'
    } finally {
      validating = false
    }
  }

  async function startImport() {
    if (!file) {
      err = 'Choose a file first'
      activeTab = 'upload'
      return
    }
    starting = true
    err = ''
    job = null
    try {
      const res = await apiFetch('/api/data-collector/jobs', {
        method: 'POST',
        body: buildFormData(),
      })
      const data = await res.json()
      if (!res.ok) throw new Error(data?.error ?? res.statusText)
      jobId = data.job_id as string
      activeTab = 'progress'
      startPolling()
    } catch (e: unknown) {
      err = e instanceof Error ? e.message : 'Could not start import'
    } finally {
      starting = false
    }
  }

  function startPolling() {
    if (pollTimer) clearInterval(pollTimer)
    void pollJob()
    pollTimer = setInterval(() => void pollJob(), 1500)
  }

  async function pollJob() {
    if (!jobId) return
    try {
      const j = await apiJson<ImportJob>(`/api/data-collector/jobs/${jobId}`)
      job = j
      if (j.status === 'completed' || j.status === 'failed') {
        if (pollTimer) {
          clearInterval(pollTimer)
          pollTimer = null
        }
      }
    } catch {
      /* keep polling */
    }
  }

  function downloadText(filename: string, content: string) {
    const blob = new Blob([content], { type: 'text/plain;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = filename
    a.click()
    URL.revokeObjectURL(url)
  }

  function downloadTemplate() {
    const spec = selectedSpec
    if (!spec) return
    downloadText(`${spec.kind}-template.csv`, spec.csv_template)
  }

  function downloadExample(format: 'csv' | 'json') {
    const spec = selectedSpec
    if (!spec) return
    if (format === 'csv') {
      downloadText(`${spec.kind}-example.csv`, spec.csv_example)
    } else {
      downloadText(`${spec.kind}-example.json`, spec.json_example)
    }
  }

  function tabBadge(id: TabId): string | null {
    if (id === 'validation' && validation) return validation.valid ? '✓' : '!'
    if (id === 'progress' && job) return `${job.percent}%`
    return null
  }
</script>

<div class="route-root data-collector">
  <header class="dc-header">
    <h1>Data Collector</h1>
    <p class="muted dc-subtitle">
      Import MorphData from CSV, Excel, or JSON. Extra columns are saved as JSON Details.
    </p>
  </header>

  {#if err}
    <p class="error dc-alert" role="alert">{err}</p>
  {/if}

  <div class="dc-shell">
    <div class="dc-tablist" role="tablist" aria-label="Data Collector steps">
      {#each tabs as tab (tab.id)}
        <button
          type="button"
          role="tab"
          class="dc-tab"
          class:active={activeTab === tab.id}
          aria-selected={activeTab === tab.id}
          onclick={() => (activeTab = tab.id)}
        >
          {tab.label}
          {#if tabBadge(tab.id)}
            <span class="dc-tab-badge">{tabBadge(tab.id)}</span>
          {/if}
        </button>
      {/each}
    </div>

    <div
      class="dc-panel themed-scroll"
      role="tabpanel"
      aria-label={tabs.find((t) => t.id === activeTab)?.label}
    >
      {#if activeTab === 'data-type'}
        <div class="dc-panel-inner dc-comfort">
          <section class="dc-block">
            <label class="field dc-field">
              <span class="label">MorphData entity</span>
              <select bind:value={selectedEntity}>
                {#each entities as ent (ent.kind)}
                  <option value={ent.kind}>{ent.label}</option>
                {/each}
              </select>
            </label>
          </section>
          {#if selectedSpec}
            <section class="dc-block">
              <p class="dc-desc muted">{selectedSpec.description}</p>
              <ul class="instructions">
                {#each selectedSpec.instructions as line}
                  <li>{line}</li>
                {/each}
              </ul>
            </section>
            <section class="dc-block">
              <div class="btn-row">
                <button type="button" class="secondary small" onclick={downloadTemplate}>
                  CSV template
                </button>
                <button type="button" class="secondary small" onclick={() => downloadExample('csv')}>
                  CSV example
                </button>
                <button type="button" class="secondary small" onclick={() => downloadExample('json')}>
                  JSON example
                </button>
              </div>
            </section>
          {/if}
        </div>
      {:else if activeTab === 'upload'}
        <div class="dc-panel-inner dc-comfort">
          <section class="dc-block">
            <div class="field dc-field">
              <span class="label">File (CSV, JSON, or XLSX)</span>
              <label class="file-input-wrap">
                <input
                  type="file"
                  class="file-input-native"
                  accept=".csv,.json,.xlsx,.xls,.txt"
                  onchange={onFileChange}
                />
                <span class="btn secondary small file-input-trigger">Choose file</span>
                <span class="file-input-name muted">
                  {#if file}
                    {file.name} · {(file.size / 1024).toFixed(1)} KB
                  {:else}
                    No file chosen
                  {/if}
                </span>
              </label>
            </div>
          </section>

          <section class="dc-block">
            <div class="btn-row">
              <button
                type="button"
                class="secondary"
                disabled={!file || validating}
                onclick={() => validateSample()}
              >
                {validating ? 'Validating…' : 'Validate sample'}
              </button>
              <button
                type="button"
                class="primary"
                disabled={!file || !validation?.valid || starting || !!jobId}
                onclick={() => startImport()}
              >
                {starting ? 'Starting…' : 'Start background import'}
              </button>
            </div>
            {#if !validation?.valid && file}
              <p class="muted dc-hint">Run validation before starting import.</p>
            {/if}
          </section>
        </div>
      {:else if activeTab === 'validation'}
        <div class="dc-panel-inner dc-comfort">
          {#if validation}
            <div class:valid={validation.valid}>
              <p>{validation.message}</p>
              <ul class="meta-list">
                <li>Rows: {validation.row_count}</li>
                <li>Columns: {validation.header_count}</li>
                <li>Uses template: {validation.uses_template ? 'yes' : 'no'}</li>
              </ul>
              {#if validation.sample_issues.length}
                <ul>
                  {#each validation.sample_issues as issue}
                    <li class="error">{issue}</li>
                  {/each}
                </ul>
              {/if}
              {#if validation.warnings.length}
                <ul>
                  {#each validation.warnings as w}
                    <li class="muted">{w}</li>
                  {/each}
                </ul>
              {/if}
            </div>
            {#if validation.valid && file}
              <div class="btn-row">
                <button type="button" class="primary" disabled={starting || !!jobId} onclick={() => startImport()}>
                  {starting ? 'Starting…' : 'Start background import'}
                </button>
              </div>
            {/if}
          {:else}
            <p class="muted">No validation yet. Upload a file and run <strong>Validate sample</strong> on the Upload File tab.</p>
          {/if}
        </div>
      {:else if activeTab === 'progress'}
        <div class="dc-panel-inner dc-comfort">
          {#if job}
            <div class="progress-bar" role="progressbar" aria-valuenow={job.percent} aria-valuemin="0" aria-valuemax="100">
              <div class="progress-fill" style="width: {job.percent}%"></div>
            </div>
            <p>
              <strong>{job.percent}%</strong>
              — {job.status}
              ({job.processed_rows} / {job.total_rows} rows)
            </p>
            <p class="muted dc-message">{job.message}</p>
            {#if job.status === 'completed' || job.status === 'failed'}
              <ul class="meta-list">
                <li>Imported: {job.imported}</li>
                <li>Failed: {job.failed}</li>
              </ul>
              {#if job.row_results?.length}
                <details class="dc-details">
                  <summary>Row details ({job.row_results.length} shown)</summary>
                  <div class="dc-table-scroll themed-scroll">
                    <table class="data-table">
                      <thead>
                        <tr>
                          <th>Row</th>
                          <th>Status</th>
                          <th>Detail</th>
                        </tr>
                      </thead>
                      <tbody>
                        {#each job.row_results as r}
                          <tr class:fail={!r.success}>
                            <td>{r.row_ref}</td>
                            <td>{r.success ? 'OK' : 'Failed'}</td>
                            <td>{r.message}{#if r.record_id} (id {r.record_id}){/if}</td>
                          </tr>
                        {/each}
                      </tbody>
                    </table>
                  </div>
                </details>
              {/if}
            {/if}
          {:else}
            <p class="muted">No import running. Validate and start an import from the Upload File tab.</p>
          {/if}
        </div>
      {/if}
    </div>
  </div>
</div>

<style>
  .data-collector {
    gap: 0;
    overflow: hidden;
  }

  .dc-header {
    flex-shrink: 0;
    margin-bottom: 0.5rem;
  }

  .dc-header h1 {
    margin: 0 0 0.35rem;
    font-size: 1.2rem;
  }

  .dc-subtitle {
    margin: 0;
    font-size: 0.8rem;
    line-height: 1.45;
  }

  .dc-alert {
    flex-shrink: 0;
    margin: 0 0 0.5rem;
  }

  .dc-shell {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 10px;
  }

  .dc-tablist {
    flex-shrink: 0;
    display: flex;
    flex-wrap: wrap;
    gap: 0;
    border-bottom: 1px solid var(--border);
    background: var(--surface-2);
    border-radius: 10px 10px 0 0;
    padding: 0 0.25rem;
  }

  .dc-tab {
    appearance: none;
    border: none;
    background: transparent;
    color: var(--text-muted);
    font: inherit;
    font-size: 0.8rem;
    font-weight: 600;
    padding: 0.7rem 1.1rem;
    cursor: pointer;
    border-bottom: 2px solid transparent;
    margin-bottom: -1px;
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    transition: color 0.15s, border-color 0.15s;
  }

  .dc-tab:hover {
    color: var(--text);
  }

  .dc-tab.active {
    color: var(--text);
    border-bottom-color: var(--primary);
    background: var(--surface);
  }

  .dc-tab-badge {
    font-size: 0.72rem;
    font-weight: 700;
    padding: 0.1rem 0.4rem;
    border-radius: 6px;
    background: var(--accent-soft);
    color: var(--accent);
  }

  .dc-panel {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    overflow-x: hidden;
  }

  .dc-panel-inner {
    padding: 1.5rem 1.75rem 1.75rem;
  }

  .dc-comfort {
    font-size: 0.8125rem;
    line-height: 1.55;
  }

  .dc-comfort .label {
    font-size: 0.72rem;
    margin-bottom: 0.45rem;
  }

  .dc-comfort select,
  .dc-comfort input[type='file'] {
    font-size: 0.8125rem;
  }

  .dc-block {
    margin-bottom: 1.5rem;
  }

  .dc-block:last-child {
    margin-bottom: 0;
  }

  .dc-field {
    margin-bottom: 0;
  }

  .dc-desc {
    margin: 0 0 0.85rem;
    line-height: 1.55;
  }

  .dc-required {
    margin: 0;
    line-height: 1.5;
  }

  .instructions {
    margin: 0;
    padding-left: 1.35rem;
    display: flex;
    flex-direction: column;
    gap: 0.55rem;
  }

  .instructions li {
    padding-left: 0.15rem;
  }

  .btn-row {
    display: flex;
    flex-wrap: wrap;
    gap: 0.65rem;
    margin-top: 0;
  }

  .dc-ai-hint {
    margin-top: 0.25rem;
    padding: 1.1rem 1.2rem;
    background: var(--surface-2);
    border: 1px solid var(--border);
    border-radius: 10px;
  }

  .dc-ai-title {
    margin: 0 0 0.6rem;
    font-size: 0.8rem;
    font-weight: 600;
    color: var(--text);
  }

  .dc-ai-hint p {
    margin: 0 0 0.75rem;
  }

  .dc-ai-hint .instructions {
    margin-top: 0.25rem;
  }

  .meta-list {
    margin: 0.5rem 0;
    padding-left: 1.25rem;
  }

  .dc-hint {
    margin-top: 1rem;
    font-size: 0.78rem;
  }

  .dc-message {
    white-space: pre-wrap;
  }

  .progress-bar {
    height: 10px;
    background: var(--surface-2);
    border-radius: 6px;
    overflow: hidden;
    margin-bottom: 0.5rem;
  }

  .progress-fill {
    height: 100%;
    background: var(--primary);
    transition: width 0.3s ease;
  }

  .valid {
    padding: 0.75rem;
    border: 1px solid var(--success);
    border-radius: 8px;
    margin-bottom: 1rem;
  }

  .dc-details {
    margin-top: 1rem;
  }

  .dc-details summary {
    cursor: pointer;
    font-weight: 600;
    color: var(--text);
  }

  .dc-table-scroll {
    max-height: min(280px, 40vh);
    overflow: auto;
    margin-top: 0.5rem;
    border: 1px solid var(--border);
    border-radius: 8px;
  }

  .data-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.9rem;
  }

  .data-table th,
  .data-table td {
    border: 1px solid var(--border);
    padding: 0.35rem 0.5rem;
    text-align: left;
  }

  tr.fail td {
    color: var(--danger);
  }
</style>
