<script lang="ts">
  import { api, apiUrl, uploadFiles } from '../lib/api'
  import { withDarkPreviewSrcDoc } from '../lib/darkPreviewSrcDoc'
  import { renderMarkdownHtml } from '../lib/markdown'
  import { runMermaidIn } from '../lib/runMermaidIn'
  import { confirm } from '../lib/confirmDialog'
  import { tick } from 'svelte'

  const ACCEPT = '.pdf,.txt,.csv,.md,.markdown,application/pdf,text/plain,text/csv,text/markdown'

  function projectPatchBody(p: any, markdown?: string) {
    const body: Record<string, unknown> = {
      code: p.code,
      name: p.name,
      client: p.client || '',
      location: p.location || '',
      status: p.status || 'planning',
      start_date: p.start_date ?? null,
      end_date: p.end_date ?? null,
      budget_total: p.budget_total ?? 0,
      progress_pct: p.progress_pct ?? 0,
      description: p.description || '',
    }
    if (markdown !== undefined) body.markdown_content = markdown
    return body
  }

  let { onCreated }: { onCreated?: () => Promise<void> | void } = $props()

  let files = $state<File[]>([])
  let paste = $state('')
  let titleHint = $state('')
  let warning = $state('')
  let generating = $state(false)
  let publishing = $state(false)
  let deleting = $state(false)
  let savingMd = $state(false)
  let previewTab = $state<'md' | 'html'>('md')
  let mdView = $state<'rendered' | 'raw'>('rendered')
  let draftMd = $state('')
  let selected = $state<any | null>(null)
  let projects = $state<any[]>([])
  let loadingList = $state(false)
  let info = $state('')
  let createOpen = $state(false)

  const renderedMd = $derived(renderMarkdownHtml(selected?.markdown_content || ''))
  let mdPreviewEl = $state<HTMLElement | null>(null)

  $effect(() => {
    renderedMd
    void tick().then(() => runMermaidIn(mdPreviewEl))
  })

  async function loadProjects() {
    loadingList = true
    try {
      const res = await api<{ projects: any[] }>('/api/v1/projects')
      projects = res.projects ?? []
      if (!selected) {
        if (projects[0]) openProject(projects[0])
      } else if (!projects.some((p) => p.id === selected.id)) {
        if (projects[0]) openProject(projects[0])
        else selected = null
      }
    } finally {
      loadingList = false
    }
  }

  $effect(() => {
    void loadProjects()
  })

  function resetCreateForm() {
    files = []
    paste = ''
    titleHint = ''
    warning = ''
  }

  function openCreate() {
    resetCreateForm()
    createOpen = true
  }

  function closeCreate() {
    if (generating) return
    createOpen = false
    warning = ''
  }

  function pickFiles(e: Event) {
    const input = e.target as HTMLInputElement
    const chosen = Array.from(input.files ?? [])
    files = [...files, ...chosen].slice(0, 5)
    warning = ''
    input.value = ''
  }

  function removeFile(i: number) {
    files = files.filter((_, idx) => idx !== i)
  }

  async function generate() {
    if (!files.length && !paste.trim()) {
      warning = 'Provide at least one source: upload a file or paste content.'
      return
    }
    if (generating) return
    generating = true
    warning = ''
    info = ''
    try {
      const fields: Record<string, string> = {}
      if (titleHint.trim()) fields.title = titleHint.trim()
      if (paste.trim()) fields.paste = paste.trim()
      const out = await uploadFiles<{ project: any }>(
        '/api/v1/projects/generate-document',
        files,
        fields
      )
      selected = out.project
      draftMd = out.project?.markdown_content || ''
      resetCreateForm()
      createOpen = false
      info = 'Project document generated. Source files (and paste, if any) were saved under Files.'
      previewTab = 'md'
      mdView = 'rendered'
      await loadProjects()
      await onCreated?.()
    } catch (e) {
      warning = e instanceof Error ? e.message : 'Generation failed'
    } finally {
      generating = false
    }
  }

  async function publish() {
    if (!selected?.id || publishing) return
    publishing = true
    warning = ''
    info = ''
    try {
      const out = await api<any>(`/api/v1/projects/${selected.id}/publish`, {
        method: 'POST',
        body: JSON.stringify({}),
      })
      selected = { ...selected, ...out }
      const path = String(out.published_path || out.published_url || '')
      if (path.startsWith('blob:')) {
        info = 'Published in this browser. Use Open published to view the HTML.'
      } else {
        info = path ? `Published: ${path.startsWith('http') ? path : apiUrl(path)}` : 'Published.'
      }
      await loadProjects()
    } catch (e) {
      warning = e instanceof Error ? e.message : 'Publish failed'
    } finally {
      publishing = false
    }
  }

  async function saveMarkdown() {
    if (!selected?.id || savingMd) return
    savingMd = true
    warning = ''
    info = ''
    try {
      const out = await api<{ project: any }>(`/api/v1/projects/${selected.id}`, {
        method: 'PATCH',
        body: JSON.stringify(projectPatchBody(selected, draftMd)),
      })
      selected = out.project ?? { ...selected, markdown_content: draftMd }
      draftMd = selected.markdown_content || ''
      projects = projects.map((p) => (p.id === selected.id ? selected : p))
      mdView = 'rendered'
      info = 'Markdown saved.'
    } catch (e) {
      warning = e instanceof Error ? e.message : 'Save failed'
    } finally {
      savingMd = false
    }
  }

  async function removeProject() {
    if (!selected?.id || deleting) return
    if (!(await confirm({ message: `Delete project “${selected.name}”?`, danger: true, confirmLabel: 'Delete' }))) return
    deleting = true
    warning = ''
    info = ''
    try {
      await api(`/api/v1/projects/${selected.id}`, { method: 'DELETE' })
      selected = null
      await loadProjects()
      info = 'Project deleted.'
      await onCreated?.()
    } catch (e) {
      warning = e instanceof Error ? e.message : 'Delete failed'
    } finally {
      deleting = false
    }
  }

  function openProject(p: any) {
    selected = p
    previewTab = 'md'
    mdView = 'rendered'
    draftMd = p?.markdown_content || ''
    info = ''
    warning = ''
  }

  function publicHref(p: any) {
    const path = String(p?.published_path || '').trim()
    if (!path) return ''
    if (path.startsWith('blob:') && p?.html_content) {
      return URL.createObjectURL(new Blob([p.html_content], { type: 'text/html' }))
    }
    return path.startsWith('http') ? path : apiUrl(path)
  }

  function onKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape' && createOpen && !generating) closeCreate()
  }
</script>

<svelte:window onkeydown={onKeydown} />

<div class="h-full min-h-0 flex flex-col md:flex-row gap-4">
  <div class="md:w-72 shrink-0 flex flex-col gap-3 min-h-0">
    <button type="button" class="btn-primary w-full shrink-0" onclick={openCreate}>Create new</button>

    {#if info}
      <p class="text-sm text-sky-300 px-1">{info}</p>
    {/if}
    {#if warning && !createOpen}
      <p class="text-sm text-amber-300 px-1">{warning}</p>
    {/if}

    <div class="card p-3 flex-1 min-h-0 overflow-auto">
      <h3 class="text-xs font-semibold uppercase text-muted mb-2">Projects</h3>
      {#if loadingList}
        <p class="text-xs text-muted">Loading…</p>
      {:else if projects.length === 0}
        <p class="text-xs text-muted">No projects yet. Use Create new to generate one from a file or paste.</p>
      {:else}
        <ul class="space-y-1">
          {#each projects as p}
            <li>
              <button
                type="button"
                class="w-full text-left rounded-lg px-2 py-1.5 text-sm hover:bg-white/5 {selected?.id === p.id
                  ? 'bg-violet/30'
                  : ''}"
                onclick={() => openProject(p)}
              >
                <div class="font-medium truncate">{p.name}</div>
                <div class="text-[11px] text-muted truncate">
                  {p.code}{#if p.published_path} · published{/if}
                </div>
              </button>
            </li>
          {/each}
        </ul>
      {/if}
    </div>
  </div>

  <div class="flex-1 min-w-0 min-h-0 card flex flex-col overflow-hidden">
    {#if !selected}
      <div class="p-6 text-muted text-sm">Select a project or create a new document.</div>
    {:else}
      <div class="shrink-0 flex items-start justify-between gap-3 p-4 border-b border-white/5">
        <div class="min-w-0">
          <h2 class="font-semibold truncate">{selected.name}</h2>
          <p class="text-xs text-muted truncate">
            {selected.code} · Sources: {selected.source_summary || selected.description || '—'}
          </p>
        </div>
        <div class="flex gap-2 shrink-0">
          {#if publicHref(selected)}
            <a class="btn-ghost border border-white/10 px-3 py-1.5 rounded-xl text-xs" href={publicHref(selected)} target="_blank" rel="noopener">
              Open published
            </a>
          {/if}
          <button type="button" class="btn-primary text-xs px-3 py-1.5" onclick={publish} disabled={publishing || deleting}>
            {publishing ? 'Publishing…' : 'Publish'}
          </button>
          <button
            type="button"
            class="btn-ghost border border-white/10 px-3 py-1.5 rounded-xl text-xs text-muted"
            onclick={removeProject}
            disabled={publishing || deleting}
          >
            {deleting ? 'Deleting…' : 'Delete'}
          </button>
        </div>
      </div>
      <div class="shrink-0 flex items-center justify-between gap-2 px-4 pt-2">
        <div class="flex gap-1">
          <button
            type="button"
            class="px-3 py-1 rounded-lg text-xs {previewTab === 'md' ? 'bg-violet/30' : 'text-muted'}"
            onclick={() => (previewTab = 'md')}>Markdown</button
          >
          <button
            type="button"
            class="px-3 py-1 rounded-lg text-xs {previewTab === 'html' ? 'bg-violet/30' : 'text-muted'}"
            onclick={() => (previewTab = 'html')}>HTML</button
          >
        </div>
        {#if previewTab === 'md'}
          <div class="flex gap-1 items-center">
            <button
              type="button"
              class="px-3 py-1 rounded-lg text-xs {mdView === 'rendered' ? 'bg-violet/30' : 'text-muted'}"
              onclick={() => (mdView = 'rendered')}>Rendered</button
            >
            <button
              type="button"
              class="px-3 py-1 rounded-lg text-xs {mdView === 'raw' ? 'bg-violet/30' : 'text-muted'}"
              onclick={() => (mdView = 'raw')}>Raw</button
            >
            {#if mdView === 'raw'}
              <button
                type="button"
                class="btn-primary text-xs px-3 py-1.5"
                onclick={saveMarkdown}
                disabled={savingMd || publishing || deleting}
              >
                {savingMd ? 'Saving…' : 'Save'}
              </button>
            {/if}
          </div>
        {/if}
      </div>
      <div class="flex-1 min-h-0 overflow-hidden">
        {#if previewTab === 'md'}
          {#if mdView === 'raw'}
            <textarea
              class="preview-scroll h-full min-h-0 w-full resize-none border-0 bg-transparent p-4 text-sm font-mono text-white outline-none"
              bind:value={draftMd}
              spellcheck="false"
            ></textarea>
          {:else if selected.markdown_content}
            <div class="md-prose preview-scroll h-full min-h-0 overflow-auto p-4 text-sm" bind:this={mdPreviewEl}>{@html renderedMd}</div>
          {:else}
            <div class="p-4 text-sm text-muted">(no markdown)</div>
          {/if}
        {:else}
          <iframe
            title="Project HTML"
            class="preview-scroll w-full h-full min-h-0 border-0 bg-[#0b1220]"
            style="color-scheme: dark"
            srcdoc={withDarkPreviewSrcDoc(selected.html_content || '<p>No HTML</p>')}
            sandbox=""
          ></iframe>
        {/if}
      </div>
    {/if}
  </div>
</div>

{#if createOpen}
  <div class="create-modal-backdrop" role="presentation" onclick={closeCreate}></div>
  <div class="create-modal" role="dialog" aria-labelledby="create-modal-title" aria-modal="true">
    <div class="create-modal-head">
      <h2 id="create-modal-title" class="font-semibold text-sm">New from sources</h2>
      <button type="button" class="create-modal-close" aria-label="Close" onclick={closeCreate} disabled={generating}>
        ✕
      </button>
    </div>
    <div class="create-modal-body space-y-3">
      <p class="text-xs text-muted">
        Upload requirements/specs and/or paste content. AI organizes them into markdown + HTML you can publish.
      </p>
      {#if warning}
        <p class="text-sm text-amber-300">{warning}</p>
      {/if}
      <input class="input" placeholder="Title (optional)" bind:value={titleHint} disabled={generating} />
      <label class="btn-ghost border border-white/10 px-3 py-2 rounded-xl text-sm inline-block cursor-pointer">
        Choose files (.pdf .txt .md .csv)
        <input class="hidden" type="file" multiple accept={ACCEPT} onchange={pickFiles} disabled={generating} />
      </label>
      {#if files.length}
        <ul class="text-xs text-muted space-y-1">
          {#each files as f, i}
            <li class="flex justify-between gap-2">
              <span class="truncate">{f.name}</span>
              <button type="button" class="text-rose-300" onclick={() => removeFile(i)} disabled={generating}>Remove</button>
            </li>
          {/each}
        </ul>
      {/if}
      <textarea
        class="input min-h-[8rem] text-sm"
        placeholder="Or paste requirements / specification text…"
        bind:value={paste}
        disabled={generating}
      ></textarea>
      <div class="flex justify-end gap-2 pt-1">
        <button type="button" class="btn-ghost border border-white/10 px-4 py-2 rounded-xl" onclick={closeCreate} disabled={generating}>
          Cancel
        </button>
        <button type="button" class="btn-primary" onclick={generate} disabled={generating}>
          {generating ? 'Generating…' : 'Generate project document'}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .create-modal-backdrop {
    position: fixed;
    inset: 0;
    z-index: 60;
    background: rgba(2, 8, 23, 0.65);
  }

  .create-modal {
    position: fixed;
    left: 50%;
    top: 50%;
    z-index: 70;
    width: min(92vw, 32rem);
    max-height: min(90vh, 40rem);
    overflow: auto;
    transform: translate(-50%, -50%);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 16px;
    background: #1a1a22;
    box-shadow: 0 24px 64px rgba(0, 0, 0, 0.45);
  }

  .create-modal-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 14px 16px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.06);
    position: sticky;
    top: 0;
    background: #1a1a22;
  }

  .create-modal-close {
    width: 32px;
    height: 32px;
    border: none;
    border-radius: 8px;
    background: rgba(255, 255, 255, 0.06);
    color: inherit;
    cursor: pointer;
  }

  .create-modal-close:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .create-modal-body {
    padding: 16px;
  }

  .md-prose {
    color: #e8e8f0;
    line-height: 1.55;
  }

  .md-prose :global(h1),
  .md-prose :global(h2),
  .md-prose :global(h3) {
    color: #fff;
    line-height: 1.25;
    margin: 1.1em 0 0.45em;
  }

  .md-prose :global(h1) { font-size: 1.45rem; }
  .md-prose :global(h2) { font-size: 1.2rem; }
  .md-prose :global(h3) { font-size: 1.05rem; }

  .md-prose :global(p),
  .md-prose :global(ul),
  .md-prose :global(ol) {
    margin: 0.7em 0;
  }

  .md-prose :global(ul),
  .md-prose :global(ol) {
    padding-left: 1.4em;
  }

  .md-prose :global(a) {
    color: #38bdf8;
  }

  .md-prose :global(code) {
    font-family: ui-monospace, monospace;
    background: rgba(255, 255, 255, 0.06);
    padding: 0.1em 0.35em;
    border-radius: 4px;
  }

  .md-prose :global(pre) {
    overflow: auto;
    padding: 1em;
    border-radius: 10px;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.08);
  }

  .md-prose :global(blockquote) {
    margin: 0.7em 0;
    padding-left: 0.9em;
    border-left: 3px solid rgba(255, 255, 255, 0.2);
    color: #b8b8c7;
  }

  .md-prose :global(table) {
    width: 100%;
    border-collapse: collapse;
    margin: 0.75rem 0;
  }

  .md-prose :global(th),
  .md-prose :global(td) {
    border: 1px solid rgba(255, 255, 255, 0.08);
    padding: 0.4rem 0.55rem;
    text-align: left;
  }

  .md-prose :global(strong) {
    color: #fff;
  }

  .md-prose :global(hr) {
    border: 0;
    border-top: 1px solid rgba(255, 255, 255, 0.1);
    margin: 1.2em 0;
  }
</style>
