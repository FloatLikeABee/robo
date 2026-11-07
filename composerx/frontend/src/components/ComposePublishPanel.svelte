<script>
  import ButtonLeadingIcon from '../lib/ButtonLeadingIcon.svelte'
  import { buildDataSource, dorphParentOrigin } from '../lib/contentDataSource.js'
  import { previewCanEnlarge } from '../lib/previewEnlarge.js'
  import { runAiProgress } from '../lib/aiProgress'
  import AssistantMarkdown from './AssistantMarkdown.svelte'

  /** @type {{ apiBase: string, getAuthHeaders: (extra?: Record<string, string>) => Record<string, string>, notify?: (kind?: string, msg?: string) => void, theme?: 'light' | 'dark' }} */
  let { apiBase, getAuthHeaders, notify = () => {}, theme = 'dark' } = $props()

  let publishName = $state('')
  let publishTheme = $state('')
  let sourceText = $state('')
  let sourceProcessing = $state(false)
  /** @type {Array<{ name: string, kind: string, summary: string, error?: string }>} */
  let sourceMaterials = $state([])
  function starterHtml(mode) {
    const dark = mode === 'dark'
    return `<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>Untitled page</title><style>html,body{margin:0;padding:0}body{min-height:100vh;font-family:Inter,system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;background:${dark ? '#0b1220' : '#ffffff'};color:${dark ? '#e2e8f0' : '#0f172a'}}main{padding:20px}h1{margin:0 0 12px;font-size:18px;line-height:1.25;font-weight:700}p{margin:0;font-size:12px;line-height:1.5}</style></head><body data-composerx-starter="1"><main><h1>Start composing</h1><p>Use the textarea below or ask AI on the right.</p></main></body></html>`
  }

  function buildSafePreviewHtml(rawHtml, mode = 'light') {
    const html = String(rawHtml || '').trim()
    const baseTag = '<base target="_blank">'
    const dark = mode === 'dark'
    const bg = dark ? '#0b1220' : '#ffffff'
    const fg = dark ? '#e2e8f0' : '#0f172a'
    const previewStyle = `<style id="cx-preview-reset">html,body{min-height:100%;}body{background:${bg};color:${fg};}</style>`
    if (!html) {
      return `<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">${baseTag}<style>html,body{margin:0;padding:0}body{min-height:100vh;background:${bg};color:${fg};font-family:Inter,system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif}</style></head><body><main style="padding:20px;font-size:12px;opacity:.8">Start composing</main></body></html>`
    }
    const hasHtml = /<html[\s>]/i.test(html)
    const hasHead = /<head[\s>]/i.test(html)
    if (!hasHtml) {
      return `<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">${baseTag}<style>html,body{margin:0;padding:0}body{min-height:100vh;background:${bg};color:${fg}}</style></head><body>${html}</body></html>`
    }
    if (!hasHead) {
      return html.replace(/<html([^>]*)>/i, `<html$1><head>${baseTag}${previewStyle}</head>`)
    }
    return html.replace(/<head([^>]*)>/i, `<head$1>${baseTag}${previewStyle}`)
  }

  let pageHtml = $state('')
  const previewHtml = $derived(buildSafePreviewHtml(pageHtml, theme))
  const canEnlarge = $derived(previewCanEnlarge(pageHtml))
  let previewOpen = $state(false)
  let dataPickerOpen = $state(false)
  let dataListLoading = $state(false)
  let dataListError = $state('')
  /** @type {Array<{ id: number, title?: string, source_type?: string }>} */
  let dataRecords = $state([])
  let publishPath = $state('')
  let pathLoading = $state(false)
  let publishLoading = $state(false)
  let saveDraftLoading = $state(false)
  let lastPublishedUrl = $state('')
  let pathResolveSeq = 0

  let aiChatLoading = $state(false)
  let aiLoadingStatus = $state('')
  let aiResponseTimeMs = $state(null)
  let aiMsgSeq = 0
  /** @type {Array<{ id: number, role: string, content: string, proposedHtml?: string | null }>} */
  let aiMessages = $state([])
  let aiInput = $state('')
  /** @type {HTMLDivElement | null} */
  let aiLogEl = $state(null)
  /** @type {HTMLDivElement | null} */
  let aiEndEl = $state(null)
  /** @type {HTMLInputElement | null} */
  let sourceFileInput = $state(null)

  function endpoint(path) {
    return `${apiBase}${path}`
  }

  function formatResponseTime(ms) {
    if (ms == null) return ''
    if (ms < 1000) return `${ms} ms`
    return `${(ms / 1000).toFixed(ms < 10000 ? 1 : 0)} s`
  }

  $effect(() => {
    if (String(pageHtml || '').trim() !== '') return
    pageHtml = starterHtml(theme)
  })

  async function apiJson(path, method, body) {
    const res = await fetch(endpoint(path), {
      method,
      headers: getAuthHeaders({ 'Content-Type': 'application/json' }),
      body: JSON.stringify(body),
    })
    if (!res.ok) {
      let msg = `${method} ${path} failed: ${res.status}`
      try {
        const payload = await res.json()
        if (payload?.error) msg = payload.error
      } catch {
        // ignore
      }
      throw new Error(msg)
    }
    return res.status === 204 ? null : res.json()
  }

  async function resolvePublishPath(name, quiet = false) {
    pathLoading = true
    try {
      const url = `${endpoint('/publishes/resolve-path')}?name=${encodeURIComponent(name)}`
      const res = await fetch(url, { headers: getAuthHeaders() })
      if (!res.ok) throw new Error(`Failed to resolve path (${res.status})`)
      const data = await res.json()
      publishPath = data?.public_path || ''
      if (!quiet) notify('info', `Resolved path: ${publishPath}`)
      return publishPath
    } catch (err) {
      if (!quiet) notify('error', err instanceof Error ? err.message : 'Failed to resolve publish path')
      return ''
    } finally {
      pathLoading = false
    }
  }

  $effect(() => {
    const name = publishName.trim()
    if (!name) {
      publishPath = ''
      return
    }
    const seq = ++pathResolveSeq
    const timer = setTimeout(async () => {
      const path = await resolvePublishPath(name, true)
      if (seq !== pathResolveSeq) return
      publishPath = path
    }, 220)
    return () => clearTimeout(timer)
  })

  $effect(() => {
    const current = String(pageHtml || '')
    if (!current.includes('data-composerx-starter="1"')) return
    pageHtml = starterHtml(theme)
  })

  async function processSourceFiles(event) {
    const fileList = event.currentTarget.files
    if (!fileList?.length) return
    sourceProcessing = true
    try {
      const form = new FormData()
      for (const file of fileList) form.append('files', file)
      const res = await fetch(endpoint('/ai/publish-sources/process'), {
        method: 'POST',
        headers: getAuthHeaders(),
        body: form,
      })
      if (!res.ok) {
        const data = await res.json().catch(() => ({}))
        throw new Error(data.error || `Process failed (${res.status})`)
      }
      const payload = await res.json()
      const items = payload?.items || []
      const combined = String(payload?.combined_text || '').trim()
      const processed = items.filter((it) => String(it?.summary || '').trim() || it?.error)
      sourceMaterials = [...sourceMaterials, ...processed]
      if (combined) {
        const existing = sourceText.trim()
        sourceText = existing ? `${existing}\n\n---\n\n${combined}` : combined
      }
      const ok = items.filter((it) => String(it?.summary || '').trim()).length
      const failed = items.filter((it) => it?.error).length
      if (ok) notify('info', `Processed ${ok} file(s)${failed ? ` (${failed} failed)` : ''}. Summaries added to source text.`)
      else notify('error', 'No files could be processed. Try PDF, text, CSV, or images.')
    } catch (err) {
      notify('error', err instanceof Error ? err.message : 'Failed to process sources')
    } finally {
      sourceProcessing = false
      event.target.value = ''
    }
  }

  function clearSourceMaterials() {
    sourceMaterials = []
  }

  async function runPublishAI(userMessage) {
    aiChatLoading = true
    aiResponseTimeMs = null
    aiLoadingStatus = 'Reading your question…'
    const startedAt = performance.now()
    const stopProgress = runAiProgress(
      {
        app: 'composerx-publish',
        userText: userMessage,
        webSearch: true,
        hasAttachments: sourceMaterials.some((m) => String(m?.summary || '').trim()),
      },
      (status) => {
        aiLoadingStatus = status
      },
    )
    try {
      const payload = {
        messages: [...aiMessages.map((m) => ({ role: m.role, content: m.content })), { role: 'user', content: userMessage }],
        current_html: pageHtml,
        source_text: sourceText,
        source_materials: sourceMaterials
          .filter((m) => String(m?.summary || '').trim())
          .map((m) => ({
            name: m.name,
            kind: m.kind,
            summary: m.summary,
          })),
        theme: publishTheme,
        use_web_search: true,
      }
      aiMessages = [...aiMessages, { id: ++aiMsgSeq, role: 'user', content: userMessage }]
      const res = await apiJson('/ai/publish-chat', 'POST', payload)
      aiResponseTimeMs = Math.round(performance.now() - startedAt)
      const assistant = String(res?.assistant_message || 'Draft updated.')
      const proposed = res?.proposed_page_html ?? null
      aiMessages = [...aiMessages, { id: ++aiMsgSeq, role: 'assistant', content: assistant, proposedHtml: proposed }]
      if (proposed && !pageHtml.trim()) {
        pageHtml = proposed
      }
    } catch (err) {
      aiResponseTimeMs = Math.round(performance.now() - startedAt)
      notify('error', err instanceof Error ? err.message : 'Failed to generate content')
    } finally {
      stopProgress()
      aiChatLoading = false
      aiLoadingStatus = ''
      setTimeout(() => aiEndEl?.scrollIntoView({ behavior: 'smooth', block: 'end' }), 0)
    }
  }

  async function sendAI() {
    const msg = aiInput.trim()
    if (!msg || aiChatLoading) return
    aiInput = ''
    await runPublishAI(msg)
  }

  $effect(() => {
    if (!previewOpen) return undefined
    const onKey = (event) => {
      if (event.key === 'Escape') previewOpen = false
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  })

  async function openDataPicker() {
    dataPickerOpen = true
    dataListError = ''
    dataRecords = []
    const origin = dorphParentOrigin(document.referrer)
    if (!origin) {
      dataListError = 'Open Content from DorphNotes to pick a data record.'
      return
    }
    dataListLoading = true
    try {
      const res = await fetch(`${origin}/api/tran/generic-data`, { headers: getAuthHeaders() })
      if (!res.ok) throw new Error(`Could not load data records (${res.status})`)
      const rows = await res.json()
      dataRecords = Array.isArray(rows) ? rows : []
    } catch (err) {
      dataListError = err instanceof Error ? err.message : 'Could not load data records.'
    } finally {
      dataListLoading = false
    }
  }

  async function generateFromRecord(row) {
    const origin = dorphParentOrigin(document.referrer)
    if (!origin || !row?.id) return
    dataListLoading = true
    dataListError = ''
    try {
      const res = await fetch(`${origin}/api/tran/generic-data/${row.id}/full`, { headers: getAuthHeaders() })
      if (!res.ok) throw new Error(`Could not load that record (${res.status})`)
      const full = await res.json()
      const built = buildDataSource(full)
      sourceText = built.text
      if (!publishName.trim()) publishName = built.title
      if (!publishTheme.trim()) publishTheme = 'twilight glass'
      dataPickerOpen = false
      const shape = built.layout === 'table'
        ? 'Include the data table and the summary counts. Do not turn the rows into an essay.'
        : 'Write an article with headings and prose, not a dashboard of rows.'
      await runPublishAI(`Generate a polished HTML page from the provided source. ${shape} Theme: ${publishTheme}.`)
    } catch (err) {
      dataListError = err instanceof Error ? err.message : 'Could not load that record.'
    } finally {
      dataListLoading = false
    }
  }

  async function generateFromSourceText() {
    const src = sourceText.trim()
    const hasMaterials = sourceMaterials.some((m) => String(m?.summary || '').trim())
    if (!src && !hasMaterials) {
      notify('error', 'Add source text or upload files to summarize first.')
      return
    }
    await runPublishAI(`Generate a polished HTML page from the provided source text and file summaries. Theme: ${publishTheme || 'default'}.`)
  }

  function applyProposedHtml(html) {
    const next = String(html || '').trim()
    if (!next) return
    pageHtml = next
    notify('info', 'Applied AI HTML to preview.')
  }

  async function publishPage() {
    const name = publishName.trim()
    if (!name) {
      notify('error', 'Publish name is required.')
      return
    }
    if (!pageHtml.trim()) {
      notify('error', 'Please generate or write page HTML first.')
      return
    }
    publishLoading = true
    try {
      await resolvePublishPath(name, true)
      const created = await apiJson('/publishes', 'POST', {
        name,
        theme: publishTheme.trim(),
        html_content: pageHtml,
        created_by: 1,
      })
      publishPath = created?.public_path || publishPath
      if (publishPath) {
        lastPublishedUrl = `${window.location.origin}${publishPath}`
      } else if (created?.slug) {
        lastPublishedUrl = `${window.location.origin}/public/p/${created.slug}`
      } else {
        lastPublishedUrl = ''
      }
      notify('info', `Published page at ${lastPublishedUrl || publishPath}`)
    } catch (err) {
      notify('error', err instanceof Error ? err.message : 'Publish failed')
    } finally {
      publishLoading = false
    }
  }

  async function saveDraft() {
    const name = publishName.trim()
    if (!name) {
      notify('error', 'Publish name is required to save HTML.')
      return
    }
    if (!pageHtml.trim()) {
      notify('error', 'Please generate or write page HTML first.')
      return
    }
    saveDraftLoading = true
    try {
      await apiJson('/publish-drafts', 'POST', {
        name,
        theme: publishTheme.trim(),
        html_content: pageHtml,
        created_by: 1,
      })
      notify('info', 'HTML draft saved.')
    } catch (err) {
      notify('error', err instanceof Error ? err.message : 'Failed to save HTML draft')
    } finally {
      saveDraftLoading = false
    }
  }
</script>

<section class="panel panel-wide compose-publish">
  <header class="panel-header compose-publish-header">
    <div class="compose-publish-head-fields">
      <div class="field-group">
        <label for="publish-name">Publish name</label>
        <input id="publish-name" placeholder="Summer Launch Landing" bind:value={publishName} />
      </div>
      <div class="field-group">
        <label for="publish-theme">Theme</label>
        <input id="publish-theme" placeholder="e.g. modern minimal, fintech blue, playful" bind:value={publishTheme} />
      </div>
    </div>
    <div class="compose-publish-actions">
      <button type="button" class="btn-secondary" onclick={saveDraft} disabled={saveDraftLoading}>
        <ButtonLeadingIcon name="save" />
        {saveDraftLoading ? 'Saving…' : 'Save'}
      </button>
      <button type="button" class="primary-action" onclick={publishPage} disabled={publishLoading}>
        <ButtonLeadingIcon name="send" />
        {publishLoading ? 'Publishing…' : 'Publish'}
      </button>
    </div>
  </header>

  {#if lastPublishedUrl}
    <div class="publish-success-strip">
      Published:
      <a href={lastPublishedUrl} target="_blank" rel="noreferrer">{lastPublishedUrl}</a>
    </div>
  {/if}

  <div class="compose-publish-main">
    <div class="compose-publish-left">
      <div class="publish-preview-shell">
        <div class="publish-path-row">
          <span class="publish-path-label">Public path:</span>
          <code>{publishPath || '/public/p/<slug>'}</code>
          {#if pathLoading}
            <span class="publish-path-status">checking…</span>
          {/if}
        </div>
        <div class="publish-preview-head">
          <span class="pill-label">HTML Preview</span>
          {#if canEnlarge}
            <button type="button" class="btn-ghost btn-sm" onclick={() => (previewOpen = true)}>Enlarge</button>
          {/if}
        </div>
        <div class="publish-preview-frame-wrap">
          <iframe
            title="Published page preview"
            srcdoc={previewHtml}
            class="publish-preview-frame"
            sandbox="allow-scripts allow-forms allow-popups allow-popups-to-escape-sandbox"
          ></iframe>
        </div>
      </div>
      <div class="publish-source-box">
        <div class="publish-source-head">
          <label for="publish-source-text">Source text &amp; files</label>
          <input
            bind:this={sourceFileInput}
            class="publish-source-file"
            type="file"
            multiple
            accept=".pdf,.txt,.md,.csv,.png,.jpg,.jpeg,.gif,.webp,image/*"
            onchange={processSourceFiles}
            disabled={sourceProcessing}
          />
        </div>
        {#if sourceMaterials.length > 0}
          <ul class="publish-source-summaries">
            {#each sourceMaterials as item, i (item.name ?? i)}
              <li class:publish-source-summaries-error={!!item.error}>
                <div class="publish-source-summaries-title">{item.name} <span>({item.kind})</span></div>
                {#if item.error}
                  <p class="publish-source-summaries-err">{item.error}</p>
                {:else}
                  <p>{item.summary}</p>
                {/if}
              </li>
            {/each}
          </ul>
          <button type="button" class="btn-ghost btn-sm" onclick={clearSourceMaterials}>Clear processed files</button>
        {/if}
        <textarea
          id="publish-source-text"
          placeholder="Paste rough content. PDFs and text files are summarized; images get an AI description. All of this feeds HTML generation."
          bind:value={sourceText}
        ></textarea>
        <div class="publish-source-actions">
          <button
            type="button"
            class="btn-secondary"
            onclick={() => sourceFileInput?.click()}
            disabled={sourceProcessing}
          >
            <ButtonLeadingIcon name="upload" />
            {sourceProcessing ? 'Summarizing…' : 'Upload'}
          </button>
          <button
            type="button"
            class="btn-secondary"
            onclick={generateFromSourceText}
            disabled={aiChatLoading || (!sourceText.trim() && !sourceMaterials.some((m) => String(m?.summary || '').trim()))}
          >
            From text
          </button>
          <button type="button" class="primary-action" onclick={openDataPicker} disabled={aiChatLoading}>
            <ButtonLeadingIcon name="ai" />
            From data
          </button>
        </div>
      </div>
    </div>

    <aside class="composer-ai-rail" aria-label="AI assistant for publish">
      <div class="composer-ai-rail-head">
        <h2 class="composer-ai-rail-title">AI assistant</h2>
      </div>
      <div class="ai-chat-log composer-ai-chat-log" bind:this={aiLogEl} role="log" aria-live="polite">
        {#if aiMessages.length === 0}
          <p class="sidebar-hint">Ask AI to build or refine your public page.</p>
        {:else}
          {#each aiMessages as msg (msg.id)}
            <div class="ai-chat-bubble" class:ai-chat-user={msg.role === 'user'} class:ai-chat-asst={msg.role !== 'user'}>
              <div class="ai-chat-role">{msg.role === 'user' ? 'You' : 'Assistant'}</div>
              <div class="ai-chat-text"><AssistantMarkdown text={msg.content} /></div>
              {#if msg.role === 'assistant' && msg.proposedHtml}
                <div class="ai-chat-apply-row">
                  <button type="button" class="btn-secondary" onclick={() => applyProposedHtml(msg.proposedHtml)}>
                    <ButtonLeadingIcon name="code" />
                    Apply
                  </button>
                </div>
              {/if}
            </div>
          {/each}
        {/if}
        {#if aiChatLoading}
          <p class="sidebar-hint ai-loading-status">{aiLoadingStatus || 'Working…'}</p>
        {/if}
        <div class="ai-chat-end-anchor" bind:this={aiEndEl} aria-hidden="true"></div>
      </div>
      <div class="ai-chat-compose composer-ai-compose">
        <textarea
          class="ai-chat-input composer-ai-chat-input"
          rows="4"
          placeholder="e.g. Build a modern hero section and pricing blocks."
          bind:value={aiInput}
          onkeydown={(e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
              e.preventDefault()
              sendAI()
            }
          }}
        ></textarea>
        <div class="composer-ai-actions">
          <button type="button" class="btn-ghost" onclick={() => (aiMessages = [])}>Clear chat</button>
          <button
            type="button"
            class="btn-send-fancy"
            onclick={sendAI}
            disabled={aiChatLoading || !aiInput.trim()}
            aria-label="Send AI message"
            title="Send"
          >
            <ButtonLeadingIcon name="send" />
          </button>
        </div>
        {#if aiResponseTimeMs != null}
          <div class="ai-response-time">Response {formatResponseTime(aiResponseTimeMs)}</div>
        {/if}
      </div>
    </aside>
  </div>

  {#if previewOpen}
    <div class="preview-modal" role="presentation" onclick={() => (previewOpen = false)}>
      <div
        class="preview-modal-panel"
        role="dialog"
        aria-modal="true"
        aria-label="HTML preview"
        onclick={(event) => event.stopPropagation()}
      >
        <button type="button" class="btn-ghost" onclick={() => (previewOpen = false)}>Close</button>
        <iframe
          title="Enlarged HTML preview"
          srcdoc={previewHtml}
          class="preview-modal-frame"
          sandbox="allow-scripts allow-forms allow-popups allow-popups-to-escape-sandbox"
        ></iframe>
      </div>
    </div>
  {/if}

  {#if dataPickerOpen}
    <div class="preview-modal" role="presentation" onclick={() => (dataPickerOpen = false)}>
      <div
        class="data-picker"
        role="dialog"
        aria-modal="true"
        aria-label="Choose a data record"
        onclick={(event) => event.stopPropagation()}
      >
        <div class="data-picker-head">
          <strong>Choose a data record</strong>
          <button type="button" class="btn-ghost" onclick={() => (dataPickerOpen = false)}>Close</button>
        </div>
        {#if dataListLoading}
          <p class="sidebar-hint">Loading records…</p>
        {:else if dataListError}
          <p class="publish-source-summaries-err">{dataListError}</p>
        {:else if dataRecords.length === 0}
          <p class="sidebar-hint">No data records yet.</p>
        {:else}
          <ul class="data-picker-list">
            {#each dataRecords as row (row.id)}
              <li>
                <button type="button" class="btn-secondary" onclick={() => generateFromRecord(row)} disabled={dataListLoading}>
                  {row.title || 'Untitled data'} <span>({row.source_type || 'file'})</span>
                </button>
              </li>
            {/each}
          </ul>
        {/if}
      </div>
    </div>
  {/if}
</section>

<style>
  .btn-secondary {
    display: inline-flex;
    align-items: center;
    gap: 0.42rem;
    background: var(--color-surface);
    color: var(--color-text);
    border: 1px solid transparent;
  }

  .btn-ghost {
    background: transparent;
    color: var(--color-text-subtle);
    border: 1px solid transparent;
    box-shadow: none;
  }

  .btn-ghost:hover {
    background: var(--color-primary-soft);
    color: var(--color-text);
  }

  .primary-action {
    background: var(--color-primary);
    color: #f4f0fa;
    border: 1px solid transparent;
  }

  .pill-label {
    display: inline-flex;
    align-items: center;
    border-radius: 999px;
    border: 1px solid var(--color-border-subtle);
    background: var(--color-primary-soft);
    color: var(--color-primary);
    font-size: 0.67rem;
    letter-spacing: 0.05em;
    text-transform: uppercase;
    font-weight: 650;
    padding: 0.2rem 0.55rem;
  }

  .sidebar-hint {
    font-size: 0.75rem;
    color: var(--color-text-subtle);
    margin: 0;
  }

  .compose-publish {
    grid-column: 1 / -1;
    padding: 0.9rem 1rem;
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .compose-publish-header {
    display: flex;
    gap: 1rem;
    align-items: flex-end;
    justify-content: space-between;
    flex-wrap: wrap;
  }

  .compose-publish-head-fields {
    display: flex;
    gap: 0.75rem;
    flex-wrap: wrap;
    flex: 1;
  }

  .compose-publish-head-fields .field-group {
    min-width: 14rem;
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
  }

  .field-group label {
    font-size: 0.76rem;
    color: var(--color-text-subtle);
  }

  .field-group input {
    border-radius: 0.65rem;
    border: 1px solid var(--color-border-subtle);
    background: var(--color-bg);
    color: var(--color-text);
    padding: 0.5rem 0.65rem;
  }

  .publish-path-row {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 0.45rem;
    font-size: 0.78rem;
    padding: 0.5rem 0.65rem;
    border-bottom: 1px solid var(--color-border-subtle);
    background: linear-gradient(90deg, var(--color-primary-soft), transparent);
  }

  .publish-path-label {
    color: var(--color-text-subtle);
  }

  .publish-path-status {
    font-size: 0.72rem;
    color: var(--color-text-subtle);
  }

  .compose-publish-actions {
    display: flex;
    gap: 0.5rem;
    align-items: center;
  }

  .publish-success-strip {
    border-radius: 0.7rem;
    border: 1px solid var(--flash-info-border);
    background: var(--flash-info-bg);
    color: var(--flash-info-text);
    padding: 0.45rem 0.6rem;
    font-size: 0.76rem;
    display: flex;
    gap: 0.4rem;
    flex-wrap: wrap;
  }

  .publish-success-strip a {
    color: inherit;
    text-decoration: underline;
    word-break: break-all;
  }

  .compose-publish-main {
    flex: 1;
    min-height: 0;
    display: grid;
    grid-template-columns: minmax(0, 3.2fr) minmax(320px, 1.1fr);
    gap: 0.9rem;
  }

  .compose-publish-left {
    min-height: 0;
    display: grid;
    grid-template-columns: minmax(0, 1.4fr) minmax(280px, 1fr);
    gap: 16px;
  }

  .publish-preview-shell {
    min-height: 0;
    border: 1px solid var(--color-border-subtle);
    border-radius: 0.95rem;
    background: var(--color-bg-elevated);
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .publish-preview-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
    border-bottom: 1px solid var(--color-border-subtle);
    padding: 0.45rem 0.65rem;
    background: linear-gradient(90deg, var(--color-primary-soft), transparent);
  }

  .preview-modal {
    position: fixed;
    inset: 0;
    z-index: 40;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 1.25rem;
    background: rgba(26, 16, 40, 0.72);
    backdrop-filter: blur(8px);
  }

  .preview-modal-panel {
    width: min(96vw, 1100px);
    height: min(92dvh, 900px);
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    padding: 0.75rem;
    border-radius: 1rem;
    border: 1px solid rgba(255, 255, 255, 0.28);
    background: rgba(53, 46, 74, 0.94);
  }

  .preview-modal-frame {
    flex: 1;
    min-height: 0;
    width: 100%;
    border: 0;
    border-radius: 0.6rem;
    background: #2c2640;
  }

  .data-picker {
    width: min(32rem, 96vw);
    max-height: min(80dvh, 640px);
    overflow: auto;
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
    padding: 0.9rem;
    border-radius: 1rem;
    border: 1px solid rgba(255, 255, 255, 0.28);
    background: rgba(53, 46, 74, 0.96);
    color: #f4f0fa;
  }

  .data-picker-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
  }

  .data-picker-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
  }

  .data-picker-list button {
    width: 100%;
    justify-content: flex-start;
  }

  .publish-preview-frame-wrap {
    flex: 1;
    min-height: 18rem;
  }

  .publish-preview-frame {
    width: 100%;
    height: 100%;
    border: 0;
    background: white;
  }

  :global([data-theme='dark']) .publish-preview-frame {
    background: #0b1220;
  }

  .publish-source-box {
    border: 1px solid var(--color-border-subtle);
    border-radius: 0.9rem;
    background: var(--color-bg-elevated);
    padding: 16px;
    display: flex;
    flex-direction: column;
    gap: 16px;
    height: 100%;
    min-height: 0;
    overflow: auto;
  }

  .publish-source-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
    flex-wrap: wrap;
  }

  .publish-source-head > label {
    font-size: 0.76rem;
    color: var(--color-text-subtle);
  }

  .btn-sm {
    font-size: 0.72rem;
    padding: 0.28rem 0.5rem;
  }

  .publish-source-file {
    display: none;
  }

  .publish-source-summaries {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    max-height: 7rem;
    overflow: auto;
  }

  .publish-source-summaries li {
    border-radius: 0.55rem;
    border: 1px solid var(--color-border-subtle);
    background: var(--color-bg);
    padding: 0.35rem 0.5rem;
    font-size: 0.72rem;
    line-height: 1.35;
  }

  .publish-source-summaries-title {
    font-weight: 600;
    margin-bottom: 0.15rem;
  }

  .publish-source-summaries-title span {
    font-weight: 400;
    color: var(--color-text-subtle);
  }

  .publish-source-summaries li p {
    margin: 0;
    white-space: pre-wrap;
    display: -webkit-box;
    -webkit-line-clamp: 4;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .publish-source-summaries-error {
    border-color: var(--flash-error-border, #fecaca);
    background: var(--flash-error-bg, #fef2f2);
  }

  .publish-source-summaries-err {
    margin: 0;
    color: var(--flash-error-text, #b91c1c);
  }

  .publish-source-box textarea {
    flex: 1;
    min-height: 12rem;
    border-radius: 0.65rem;
    border: 1px solid var(--color-border-subtle);
    background: var(--color-bg);
    color: var(--color-text);
    padding: 0.55rem 0.65rem;
    resize: none;
    font-family: inherit;
  }

  .publish-source-actions {
    display: flex;
    flex-direction: column;
    gap: 12px;
    flex-shrink: 0;
  }

  .publish-source-actions button {
    width: 100%;
    justify-content: center;
  }

  .publish-source-actions .primary-action {
    padding: 0.8rem 1.1rem;
    font-weight: 650;
  }

  .composer-ai-rail {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    min-height: 0;
    border-radius: 0.85rem;
    border: 1px dashed var(--color-border-subtle);
    background: var(--color-bg-elevated);
    padding: 0.55rem 0.65rem;
  }

  .composer-ai-rail-head {
    flex-shrink: 0;
  }

  .composer-ai-rail-title {
    margin: 0;
    font-size: 0.82rem;
    font-weight: 650;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    color: var(--color-text-subtle);
  }

  .ai-chat-log {
    flex: 1 1 0;
    min-height: 0;
    overflow-y: auto;
    border: 1px solid var(--color-border-subtle);
    border-radius: 0.75rem;
    padding: 0.65rem;
    display: flex;
    flex-direction: column;
    gap: 0.65rem;
    background: var(--color-bg);
  }

  .composer-ai-chat-log {
    flex: 1;
  }

  .ai-chat-bubble {
    border-radius: 0.65rem;
    padding: 0.5rem 0.65rem;
    font-size: 0.85rem;
    line-height: 1.45;
  }

  .ai-chat-user {
    align-self: flex-end;
    background: rgba(59, 130, 246, 0.14);
    max-width: 92%;
  }

  .ai-chat-asst {
    align-self: stretch;
    background: rgba(15, 23, 42, 0.06);
  }

  :global([data-theme='dark']) .ai-chat-asst {
    background: rgba(255, 255, 255, 0.07);
  }

  .ai-chat-role {
    font-size: 0.65rem;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    opacity: 0.75;
    margin-bottom: 0.25rem;
  }

  .ai-chat-text {
    white-space: pre-wrap;
    word-break: break-word;
  }

  .ai-chat-apply-row {
    margin-top: 0.5rem;
  }

  .ai-chat-compose {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    flex-shrink: 0;
  }

  .ai-chat-input {
    width: 100%;
    resize: none;
    height: 5.6rem;
    min-height: 5.6rem;
    max-height: 5.6rem;
    border-radius: 0.75rem;
    border: 1px solid var(--color-border-subtle);
    padding: 0.55rem 0.65rem;
    font-size: 0.85rem;
    font-family: inherit;
    background: var(--color-bg);
    color: var(--color-text);
  }

  .composer-ai-chat-input {
    height: 5.6rem;
    min-height: 5.6rem;
    max-height: 5.6rem;
  }

  .composer-ai-actions {
    display: flex;
    flex-wrap: wrap;
    justify-content: flex-end;
    gap: 0.5rem;
  }

  .ai-response-time {
    text-align: right;
    color: var(--color-accent, #8b5cf6);
    font-size: 0.7rem;
    font-variant-numeric: tabular-nums;
    opacity: 0.9;
  }

  .btn-send-fancy {
    width: 3rem;
    height: 3rem;
    min-width: 3rem;
    border-radius: 999px;
    padding: 0;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border: 1px solid transparent;
    color: #f4f0fa;
    background: var(--color-primary);
    box-shadow:
      0 8px 22px rgba(124, 58, 237, 0.35),
      inset 0 1px 0 rgba(255, 255, 255, 0.2);
    transition:
      transform 160ms ease,
      box-shadow 180ms ease,
      filter 160ms ease;
  }

  .btn-send-fancy :global(.btn-leading-icon) {
    width: 1.15rem;
    height: 1.15rem;
    margin: 0;
    transform: translateX(0.02rem) rotate(-8deg);
  }

  .btn-send-fancy:hover:not(:disabled) {
    transform: translateY(-1px) scale(1.03);
    filter: saturate(1.08);
    box-shadow:
      0 12px 26px rgba(124, 58, 237, 0.42),
      0 0 0 4px rgba(139, 92, 246, 0.2);
  }

  .btn-send-fancy:active:not(:disabled) {
    transform: translateY(0) scale(0.99);
  }

  .btn-send-fancy:disabled {
    opacity: 0.5;
    cursor: not-allowed;
    box-shadow: none;
    filter: grayscale(0.2);
  }

  .ai-chat-end-anchor {
    height: 1px;
  }

  @media (max-width: 1080px) {
    .compose-publish-main,
    .compose-publish-left {
      grid-template-columns: 1fr;
    }
  }
</style>
