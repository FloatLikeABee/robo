<script lang="ts">
  import { api, uploadFiles } from '../lib/api'

  const MAX_FILES = 3
  const ACCEPT = '.pdf,.txt,.csv,.md,application/pdf,text/plain,text/csv,text/markdown'

  type FileResult = { name: string; mime_type?: string; size_bytes?: number; excerpt?: string; error?: string }

  type ProposedSiteLog = {
    log_date?: string
    weather?: string
    crew_count?: number
    summary?: string
    issues?: string
    include?: boolean
  }
  type ProposedPerson = {
    name?: string
    trade?: string
    contact_name?: string
    phone?: string
    email?: string
    status?: string
    description?: string
    include?: boolean
  }
  type ProposedFlowEntry = {
    entry_date?: string
    direction?: string
    amount?: number
    currency?: string
    category?: string
    status?: string
    title?: string
    notes?: string
    include?: boolean
  }
  type ProposedProject = {
    name?: string
    code?: string
    client?: string
    location?: string
    status?: string
    start_date?: string
    end_date?: string
    budget_total?: number
    description?: string
    include?: boolean
    site_logs: ProposedSiteLog[]
    people: ProposedPerson[]
    flow_log_entries: ProposedFlowEntry[]
  }

  type AnalyzeResponse = {
    session: { id: number; status: string; draft: { projects?: ProposedProject[]; notes?: string } }
    files: FileResult[]
  }
  type ConfirmResponse = {
    created_projects: { id: number; name: string; code: string; site_logs: number; people: number; flow_log_entries: number }[]
    errors: { record?: string; project?: string; error?: string }[]
  }

  let { onCreated }: { onCreated?: () => Promise<void> | void } = $props()

  let files = $state<File[]>([])
  let instruction = $state('')
  let analyzing = $state(false)
  let confirming = $state(false)
  let notice = $state('')
  let errorMsg = $state('')

  let sessionId = $state<number | null>(null)
  let fileResults = $state<FileResult[]>([])
  let draftNotes = $state('')
  let proposals = $state<ProposedProject[]>([])
  let confirmResult = $state<ConfirmResponse | null>(null)

  function pickFiles(e: Event) {
    const input = e.target as HTMLInputElement
    const chosen = Array.from(input.files ?? [])
    input.value = ''
    notice = ''

    const room = MAX_FILES - files.length
    if (chosen.length > room) {
      notice = `Up to ${MAX_FILES} files can be analyzed together. ${
        room > 0 ? `Only the first ${room} of your selection were added.` : 'Remove one before adding another.'
      }`
    }
    files = [...files, ...chosen.slice(0, Math.max(room, 0))]
  }

  function removeFile(i: number) {
    files = files.filter((_, idx) => idx !== i)
    notice = ''
  }

  /** Fill in the flags and arrays the editor binds to. */
  function toEditable(list: ProposedProject[]): ProposedProject[] {
    return list.map((p) => ({
      ...p,
      include: p.include ?? true,
      site_logs: (p.site_logs ?? []).map((l) => ({ ...l, include: l.include ?? true })),
      people: (p.people ?? []).map((x) => ({ ...x, include: x.include ?? true })),
      flow_log_entries: (p.flow_log_entries ?? []).map((x) => ({ ...x, include: x.include ?? true })),
    }))
  }

  async function analyze() {
    if (!files.length || analyzing) return
    analyzing = true
    errorMsg = ''
    notice = ''
    confirmResult = null
    try {
      const out = await uploadFiles<AnalyzeResponse>('/api/v1/project-imports/analyze', files, { instruction })
      sessionId = out.session.id
      fileResults = out.files ?? []
      proposals = toEditable(out.session.draft?.projects ?? [])
      draftNotes = out.session.draft?.notes ?? ''
      if (!proposals.length) {
        notice = 'The AI did not find any project to propose in these files.'
      }
    } catch (e) {
      errorMsg = e instanceof Error ? e.message : 'Analysis failed'
      // A failed analysis may still have per-file results worth showing.
      await loadSessionFiles()
    } finally {
      analyzing = false
    }
  }

  /** Re-read the session so per-file read errors stay visible after a failure. */
  async function loadSessionFiles() {
    if (!sessionId) return
    try {
      const out = await api<{ files: FileResult[] }>(`/api/v1/project-imports/${sessionId}`)
      fileResults = out.files ?? []
    } catch {
      /* the error message already explains the failure */
    }
  }

  async function confirmDraft() {
    if (!sessionId || confirming) return
    confirming = true
    errorMsg = ''
    try {
      const out = await api<ConfirmResponse>(`/api/v1/project-imports/${sessionId}/confirm`, {
        method: 'POST',
        body: JSON.stringify({ projects: proposals }),
      })
      confirmResult = out
      proposals = []
      await onCreated?.()
    } catch (e) {
      errorMsg = e instanceof Error ? e.message : 'Could not create the projects'
    } finally {
      confirming = false
    }
  }

  function startOver() {
    files = []
    instruction = ''
    sessionId = null
    fileResults = []
    proposals = []
    draftNotes = ''
    confirmResult = null
    errorMsg = ''
    notice = ''
  }

  const includedCount = $derived(proposals.filter((p) => p.include).length)
</script>

<div class="h-full overflow-auto flex flex-col gap-3">
  <div class="card p-5 form-compact">
    <h2 class="font-semibold">Import projects from description files</h2>
    <p class="text-xs text-muted mt-1">
      Upload up to {MAX_FILES} files (PDF, TXT, CSV, MD). The AI reads them and proposes projects with their site logs,
      people, and money movements. Nothing is created until you confirm.
    </p>

    <div class="mt-3 flex flex-wrap items-center gap-2">
      <label class="btn btn-secondary cursor-pointer">
        Choose files
        <input type="file" multiple accept={ACCEPT} class="hidden" onchange={pickFiles} />
      </label>
      <span class="text-xs text-muted">{files.length} of {MAX_FILES} selected</span>
    </div>

    {#if files.length}
      <ul class="mt-2 flex flex-col gap-1">
        {#each files as f, i}
          <li class="flex items-center gap-2 text-sm">
            <span class="truncate">{f.name}</span>
            <span class="text-xs text-muted">{Math.max(1, Math.round(f.size / 1024))} KB</span>
            <button type="button" class="text-xs text-rose-400 underline" onclick={() => removeFile(i)}>remove</button>
          </li>
        {/each}
      </ul>
    {/if}

    <span class="block mt-4 text-xs text-muted">What should the AI focus on? (optional)</span>
    <textarea
      class="input mt-1 min-h-[5rem]"
      placeholder="e.g. Treat each site address as its own project; costs are in EUR."
      bind:value={instruction}
    ></textarea>

    <div class="mt-4 flex flex-wrap gap-2">
      <button type="button" class="btn-primary" disabled={analyzing || !files.length} onclick={analyze}>
        {analyzing ? 'Analyzing…' : 'Analyze files'}
      </button>
      {#if sessionId}
        <button type="button" class="btn btn-secondary" onclick={startOver}>Start over</button>
      {/if}
    </div>

    {#if notice}<p class="text-amber-400 text-sm mt-3">{notice}</p>{/if}
    {#if errorMsg}<p class="text-rose-400 text-sm mt-3">{errorMsg}</p>{/if}
  </div>

  {#if fileResults.length}
    <div class="card p-5">
      <h3 class="font-semibold text-sm">Files read</h3>
      <ul class="mt-2 flex flex-col gap-2">
        {#each fileResults as f}
          <li class="text-sm">
            <div class="flex items-center gap-2">
              <span class="truncate">{f.name}</span>
              {#if f.error}
                <span class="text-xs text-rose-400">could not be read</span>
              {:else}
                <span class="text-xs text-teal">read</span>
              {/if}
            </div>
            {#if f.error}
              <p class="text-xs text-rose-400 mt-0.5">{f.error}</p>
            {:else if f.excerpt}
              <details class="mt-1">
                <summary class="text-xs text-muted cursor-pointer">What the AI read</summary>
                <pre class="mt-1 max-h-40 overflow-auto whitespace-pre-wrap text-xs text-muted">{f.excerpt}</pre>
              </details>
            {/if}
          </li>
        {/each}
      </ul>
    </div>
  {/if}

  {#if confirmResult}
    <div class="card p-5">
      <h3 class="font-semibold text-sm">Created</h3>
      <ul class="mt-2 flex flex-col gap-1 text-sm">
        {#each confirmResult.created_projects as c}
          <li>
            <span class="font-medium">{c.name}</span>
            <span class="text-xs text-muted">
              ({c.code}) — {c.site_logs} log(s), {c.people} person/people, {c.flow_log_entries} flow entry/entries
            </span>
          </li>
        {/each}
      </ul>
      {#if confirmResult.errors?.length}
        <h4 class="text-sm mt-3 text-rose-400">Skipped</h4>
        <ul class="mt-1 flex flex-col gap-1 text-xs text-rose-400">
          {#each confirmResult.errors as e}
            <li>{e.record}: {e.error}</li>
          {/each}
        </ul>
      {/if}
      <button type="button" class="btn btn-secondary mt-4" onclick={startOver}>Import more files</button>
    </div>
  {/if}

  {#if proposals.length}
    <div class="card p-5">
      <div class="flex flex-wrap items-center justify-between gap-2">
        <h3 class="font-semibold text-sm">
          Proposed projects ({includedCount} of {proposals.length} selected)
        </h3>
        <button type="button" class="btn-primary" disabled={confirming || !includedCount} onclick={confirmDraft}>
          {confirming ? 'Creating…' : `Create ${includedCount} project(s)`}
        </button>
      </div>
      {#if draftNotes}
        <p class="text-xs text-muted mt-2 whitespace-pre-wrap">AI notes: {draftNotes}</p>
      {/if}
    </div>

    {#each proposals as project, pi}
      <div class="card p-5 form-compact {project.include ? '' : 'opacity-60'}">
        <label class="flex items-center gap-2 text-sm font-medium">
          <input type="checkbox" bind:checked={project.include} />
          Include this project
        </label>

        <div class="grid grid-cols-2 md:grid-cols-4 gap-3 items-start mt-3">
          <div class="form-field-sm"><input class="input" placeholder="Code (optional)" bind:value={project.code} /></div>
          <div class="form-field-md"><input class="input" placeholder="Name *" bind:value={project.name} /></div>
          <div class="form-field-md"><input class="input" placeholder="Client" bind:value={project.client} /></div>
          <div class="form-field-sm">
            <select class="input" bind:value={project.status}>
              <option value="planning">Planning</option>
              <option value="active">Active</option>
              <option value="on_hold">On hold</option>
              <option value="completed">Completed</option>
            </select>
          </div>
          <div class="form-field-md"><input class="input" placeholder="Location" bind:value={project.location} /></div>
          <div class="form-field-sm"><input class="input" type="date" bind:value={project.start_date} /></div>
          <div class="form-field-sm"><input class="input" type="date" bind:value={project.end_date} /></div>
          <div class="form-field-sm">
            <input class="input" type="number" step="0.01" placeholder="Budget" bind:value={project.budget_total} />
          </div>
        </div>

        <span class="block mt-3 text-xs text-muted">Description</span>
        <textarea class="input mt-1 min-h-[5rem]" bind:value={project.description}></textarea>

        {#if project.site_logs.length}
          <h4 class="text-sm font-medium mt-4">Site logs ({project.site_logs.length})</h4>
          {#each project.site_logs as log, li}
            <div class="flex flex-wrap items-start gap-2 mt-2">
              <input type="checkbox" class="mt-2" bind:checked={log.include} aria-label="Include site log {li + 1}" />
              <input class="input w-36" type="date" bind:value={log.log_date} />
              <input class="input flex-1 min-w-[14rem]" placeholder="Summary *" bind:value={log.summary} />
              <input class="input w-28" placeholder="Weather" bind:value={log.weather} />
              <input class="input w-20" type="number" placeholder="Crew" bind:value={log.crew_count} />
            </div>
          {/each}
        {/if}

        {#if project.people.length}
          <h4 class="text-sm font-medium mt-4">People ({project.people.length})</h4>
          {#each project.people as person, xi}
            <div class="flex flex-wrap items-start gap-2 mt-2">
              <input type="checkbox" class="mt-2" bind:checked={person.include} aria-label="Include person {xi + 1}" />
              <input class="input flex-1 min-w-[12rem]" placeholder="Name *" bind:value={person.name} />
              <input class="input w-32" placeholder="Trade" bind:value={person.trade} />
              <input class="input w-36" placeholder="Contact" bind:value={person.contact_name} />
              <input class="input w-36" placeholder="Phone" bind:value={person.phone} />
              <input class="input w-44" placeholder="Email" bind:value={person.email} />
            </div>
          {/each}
        {/if}

        {#if project.flow_log_entries.length}
          <h4 class="text-sm font-medium mt-4">Flow log ({project.flow_log_entries.length})</h4>
          {#each project.flow_log_entries as entry, ei}
            <div class="flex flex-wrap items-start gap-2 mt-2">
              <input type="checkbox" class="mt-2" bind:checked={entry.include} aria-label="Include entry {ei + 1}" />
              <input class="input w-36" type="date" bind:value={entry.entry_date} />
              <select class="input w-28" bind:value={entry.direction}>
                <option value="expense">Expense</option>
                <option value="income">Income</option>
              </select>
              <input class="input w-28" type="number" step="0.01" placeholder="Amount *" bind:value={entry.amount} />
              <input class="input w-20" placeholder="Cur" bind:value={entry.currency} />
              <input class="input flex-1 min-w-[10rem]" placeholder="Title" bind:value={entry.title} />
              <input class="input w-32" placeholder="Category" bind:value={entry.category} />
            </div>
          {/each}
        {/if}
      </div>
    {/each}

    <div class="card p-5">
      <button type="button" class="btn-primary" disabled={confirming || !includedCount} onclick={confirmDraft}>
        {confirming ? 'Creating…' : `Create ${includedCount} project(s)`}
      </button>
      <p class="text-xs text-muted mt-2">
        Records that fail validation are reported and skipped; their valid siblings are still created.
      </p>
    </div>
  {/if}
</div>
