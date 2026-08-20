<script lang="ts">
  import { api, apiUrl, uploadFiles } from '../lib/api'

  const ACCEPT = '.pdf,.txt,.csv,.md,.markdown,application/pdf,text/plain,text/csv,text/markdown'

  let { onCreated }: { onCreated?: () => Promise<void> | void } = $props()

  let files = $state<File[]>([])
  let paste = $state('')
  let titleHint = $state('')
  let warning = $state('')
  let generating = $state(false)
  let publishing = $state(false)
  let deleting = $state(false)
  let previewTab = $state<'md' | 'html'>('md')
  let selected = $state<any | null>(null)
  let projects = $state<any[]>([])
  let loadingList = $state(false)
  let info = $state('')
  let createOpen = $state(false)

  async function loadProjects() {
    loadingList = true
    try {
      const res = await api<{ projects: any[] }>('/api/v1/projects')
      projects = res.projects ?? []
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
      resetCreateForm()
      createOpen = false
      info = 'Project document generated. Source files (and paste, if any) were saved under Files.'
      previewTab = 'md'
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

  async function removeProject() {
    if (!selected?.id || deleting) return
    if (!confirm(`Delete project “${selected.name}”?`)) return
    deleting = true
    warning = ''
    info = ''
    try {
      await api(`/api/v1/projects/${selected.id}`, { method: 'DELETE' })
      selected = null
      info = 'Project deleted.'
      await loadProjects()
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
      <p class="text-sm text-teal px-1">{info}</p>
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
            class="btn-ghost border border-white/10 px-3 py-1.5 rounded-xl text-xs text-rose-300"
            onclick={removeProject}
            disabled={publishing || deleting}
          >
            {deleting ? 'Deleting…' : 'Delete'}
          </button>
        </div>
      </div>
      <div class="shrink-0 flex gap-1 px-4 pt-2">
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
      <div class="flex-1 min-h-0 overflow-auto p-4">
        {#if previewTab === 'md'}
          <pre class="whitespace-pre-wrap text-sm font-mono">{selected.markdown_content || '(no markdown)'}</pre>
        {:else}
          <iframe
            title="Project HTML"
            class="w-full min-h-[28rem] rounded-xl border border-white/10 bg-[#0b1220]"
            srcdoc={selected.html_content || '<p>No HTML</p>'}
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
</style>
