<script lang="ts">
  import { onMount } from 'svelte'
  import AppLayout from './components/AppLayout.svelte'
  import ModuleShell from './components/ModuleShell.svelte'
  import DataGrid from './components/DataGrid.svelte'
  import { api, ensureSession, loginWithCredentials, setToken, uploadFile } from './lib/api'
  import { buildAiStateExtra } from './lib/aiContext'
  import type { PageId } from './lib/nav'
  import { NAV } from './lib/nav'
  import ProjectDocumentPanel from './components/ProjectDocumentPanel.svelte'

  let authed = $state(false)
  let loading = $state(true)
  let error = $state('')
  let page = $state<PageId>('projects')
  let activeProjectId = $state<number | null>(null)

  let projects = $state<any[]>([])
  let tasks = $state<any[]>([])
  let siteLogs = $state<any[]>([])
  let resourceFiles = $state<any[]>([])
  let contractors = $state<any[]>([])
  let org = $state<any>(null)
  let actionError = $state('')

  let newResourceFile = $state({ name: '', source_type: 'url' as 'url' | 'upload', file_url: '', description: '' })
  let resourceUpload: File | null = $state(null)

  let filesTab = $state('list')

  async function bootstrap() {
    loading = true
    error = ''
    try {
      const session = await ensureSession()
      if (!session.ok) {
        authed = false
        error = session.reason
        return
      }
      await refreshAll()
      authed = true
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load'
      authed = false
    } finally {
      loading = false
    }
  }

  let loginEmail = $state('')
  let loginPassword = $state('')
  let loginBusy = $state(false)

  async function submitLogin() {
    loginBusy = true
    error = ''
    try {
      await loginWithCredentials(loginEmail, loginPassword)
      await refreshAll()
      authed = true
    } catch (e) {
      error = e instanceof Error ? e.message : 'Login failed'
    } finally {
      loginBusy = false
    }
  }

  async function refreshAll() {
    const [proj, orgRes] = await Promise.all([
      api<{ projects: any[] }>('/api/v1/projects'),
      api<any>('/api/v1/organization'),
    ])
    projects = proj.projects ?? []
    org = orgRes
    await refreshPageData()
  }

  async function refreshPageData() {
    if (page === 'files') {
      // Files library lists all retained uploads/pastes for the org (not scoped to header project).
      resourceFiles = (await api<{ resource_files: any[] }>('/api/v1/resource-files')).resource_files ?? []
    }
  }

  $effect(() => {
    if (authed) {
      page
      activeProjectId
      void refreshPageData()
    }
  })

  onMount(() => {
    void bootstrap()
  })

  function signOut() {
    setToken(null)
    authed = false
  }

  function linkedProjectIds(): number[] {
    return activeProjectId ? [activeProjectId] : []
  }

  function slugCode(name: string): string {
    const base = name.trim().toUpperCase().replace(/[^A-Z0-9]+/g, '-').replace(/^-|-$/g, '')
    return base || `PRJ-${Date.now().toString().slice(-6)}`
  }


  function requireProject(): number {
    if (!activeProjectId) throw new Error('Select a project in the header first')
    return activeProjectId
  }




  async function addResourceFile() {
    actionError = ''
    try {
      let file_url = newResourceFile.file_url
      let file_name = ''
      let source_type = newResourceFile.source_type
      if (source_type === 'upload') {
        if (!resourceUpload) throw new Error('Choose a file to upload')
        const up = await uploadFile('/api/v1/resource-files/upload', resourceUpload)
        file_url = up.file_url
        file_name = up.file_name
        source_type = 'upload'
      }
      if (!newResourceFile.name.trim()) throw new Error('Name is required')
      await api('/api/v1/resource-files', {
        method: 'POST',
        body: JSON.stringify({
          name: newResourceFile.name,
          source_type,
          file_url,
          file_name,
          description: newResourceFile.description,
          project_ids: linkedProjectIds(),
        }),
      })
      newResourceFile = { name: '', source_type: 'url', file_url: '', description: '' }
      resourceUpload = null
      filesTab = 'list'
      await refreshPageData()
    } catch (e) {
      actionError = e instanceof Error ? e.message : 'Failed to add file'
    }
  }


  const pageTitle = $derived(NAV.find((n) => n.id === page)?.label ?? 'Project')
  const pageHint = $derived(NAV.find((n) => n.id === page)?.hint ?? '')

  function projectLabel(id: unknown) {
    const p = projects.find((x) => x.id === id)
    return p ? `${p.code} — ${p.name}` : String(id ?? '—')
  }

  function getAiStateExtra() {
    return buildAiStateExtra({
      page,
      activeProjectId,
      organization: org,
      projects,
      tasks,
      siteLogs,
      resourceFiles,
      contractors,
    })
  }
</script>

{#if loading}
  <div class="h-full flex items-center justify-center text-muted">Loading Project…</div>
{:else if !authed}
  <div class="h-full flex flex-col items-center justify-center gap-4 p-8 text-center auth-flow-shell">
    <img src="/morph-engi-icon.svg" alt="" width="72" height="72" class="rounded-2xl" />
    <h1 class="text-2xl font-semibold">Project</h1>
    <p class="text-muted max-w-md text-sm">
      Simple project documents from files or paste. Sign in with your Morph account.
    </p>
    {#if error}<p class="text-rose-400 text-sm max-w-md">{error}</p>{/if}
    <form class="w-full max-w-sm space-y-3 text-left card p-5" onsubmit={(e) => { e.preventDefault(); void submitLogin() }}>
      <h2 class="font-semibold text-sm">Sign in</h2>
      <input class="input" type="email" placeholder="Email" bind:value={loginEmail} required autocomplete="email" />
      <input class="input" type="password" placeholder="Password" bind:value={loginPassword} required autocomplete="current-password" />
      <button type="submit" class="btn-primary w-full" disabled={loginBusy}>{loginBusy ? 'Signing in…' : 'Sign in'}</button>
    </form>
    <div class="flex gap-3 flex-wrap justify-center">
      <button type="button" class="btn-ghost border border-white/10 px-4 py-2 rounded-xl" onclick={() => bootstrap()}>Retry SSO</button>
      <a class="btn-ghost border border-white/10 px-4 py-2 rounded-xl" href="http://localhost:3031" target="_blank" rel="noopener">Open Morph AI</a>
    </div>
  </div>
{:else}
  <div class="h-full min-h-0 max-h-dvh overflow-hidden">
  <AppLayout bind:page bind:activeProjectId {projects} getStateExtra={getAiStateExtra} onSignOut={signOut}>
    <div class="h-full min-h-0">
      {#if page === 'projects'}
        <div class="h-full min-h-0">
          <ProjectDocumentPanel onCreated={refreshAll} />
        </div>

      {:else if page === 'files'}
        <ModuleShell
          title={pageTitle}
          hint={pageHint}
          tabs={[
            { id: 'create', label: 'Add' },
            { id: 'list', label: 'Files' },
          ]}
          bind:activeTab={filesTab}
          {actionError}
        >
          {#snippet children()}
            {#if filesTab === 'create'}
              <div class="card p-5 form-compact h-full overflow-auto">
                <h2 class="font-semibold mb-3">Add file</h2>
                <p class="text-xs text-muted mb-3">
                  Keep uploads in your library. Paste-generated sources from Projects also appear here.
                </p>
                <div class="grid grid-cols-2 md:grid-cols-3 gap-3 items-start">
                  <div class="form-field-md"><input class="input" placeholder="Name *" bind:value={newResourceFile.name} /></div>
                  <div class="form-field-sm">
                    <select class="input" bind:value={newResourceFile.source_type}>
                      <option value="url">URL / link</option>
                      <option value="upload">Upload file</option>
                    </select>
                  </div>
                  {#if newResourceFile.source_type === 'url'}
                    <div class="md:col-span-1 min-w-0"><input class="input" placeholder="https://…" bind:value={newResourceFile.file_url} /></div>
                  {:else}
                    <div class="form-field-md"><input class="input" type="file" onchange={(e) => { resourceUpload = (e.currentTarget as HTMLInputElement).files?.[0] ?? null }} /></div>
                  {/if}
                </div>
                <span class="block mt-4 text-xs text-muted">Notes</span>
                <textarea class="input mt-1 min-h-[8rem]" placeholder="What this file is for…" bind:value={newResourceFile.description}></textarea>
                <button type="button" class="btn-primary mt-4" onclick={addResourceFile}>Save file</button>
              </div>
            {:else if resourceFiles.length === 0}
              <div class="card p-6 text-sm text-muted">
                No files yet. Upload here, or generate a project from file/paste — those sources are saved into this list.
              </div>
            {:else}
              <DataGrid title="Files">
                <table class="data-table">
                  <thead><tr><th>Name</th><th>Type</th><th>Link</th><th>Notes</th></tr></thead>
                  <tbody>
                    {#each resourceFiles as d}
                      <tr>
                        <td>{d.name}</td>
                        <td>{d.source_type}</td>
                        <td class="max-w-[12rem] truncate"><a class="text-teal underline" href={d.file_url} target="_blank" rel="noopener">{d.file_name || d.file_url}</a></td>
                        <td class="max-w-[18rem] whitespace-pre-wrap text-muted">{d.description || '—'}</td>
                      </tr>
                    {/each}
                  </tbody>
                </table>
              </DataGrid>
            {/if}
          {/snippet}
        </ModuleShell>

      {:else if page === 'settings'}
        <ModuleShell title="Settings moved" hint="Use Morph Utils Settings" tabs={[]} {actionError}>
          {#snippet children()}
            <div class="card p-5 space-y-2 max-w-lg">
              <h2 class="font-semibold">Shared Settings</h2>
              <p class="text-sm text-muted">
                Module settings now live in Morph Utils → Settings (same level as the apps).
              </p>
              <button type="button" class="btn btn-secondary mt-2" onclick={() => (page = 'projects')}>
                Back to Projects
              </button>
            </div>
          {/snippet}
        </ModuleShell>
      {/if}
    </div>
  </AppLayout>
  </div>

{/if}

<style>
  :global(.auth-flow-shell) {
    background:
      radial-gradient(circle at 22% 18%, rgba(91, 63, 214, 0.18), transparent 48%),
      radial-gradient(circle at 78% 82%, rgba(45, 212, 191, 0.12), transparent 42%),
      linear-gradient(165deg, #26262f 0%, #30303c 48%, #2a2a34 100%);
  }
</style>
