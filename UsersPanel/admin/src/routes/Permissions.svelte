<script lang="ts">
  import { onMount } from 'svelte'
  import { apiJson } from '../lib/api'

  type Perm = {
    id: string
    name: string
    description: string | null
    createdAt: string
    updatedAt: string
  }

  let permissions = $state<Perm[]>([])
  let roleNames = $state<string[]>([])
  let assignments = $state<Record<string, string[]>>({})
  let err = $state('')
  let loading = $state(true)

  let selectedRole = $state('Admin')
  let draftIds = $state<string[]>([])

  let addOpen = $state(false)
  let newName = $state('')
  let newDesc = $state('')

  let editPerm = $state<Perm | null>(null)
  let editDesc = $state('')
  let permTab = $state<'catalog' | 'roles'>('catalog')

  onMount(async () => {
    await loadOverview()
  })

  async function loadOverview() {
    loading = true
    err = ''
    try {
      const data = await apiJson<{
        permissions: Perm[]
        roleNames: string[]
        assignments: Record<string, string[]>
      }>('/api/admin/permissions/overview')
      permissions = data.permissions
      roleNames = data.roleNames
      assignments = data.assignments
      syncDraftForRole()
    } catch (e: unknown) {
      err = e instanceof Error ? e.message : 'Failed to load'
    } finally {
      loading = false
    }
  }

  function syncDraftForRole() {
    draftIds = [...(assignments[selectedRole] ?? [])]
  }

  function togglePermForRole(id: string) {
    if (draftIds.includes(id)) draftIds = draftIds.filter((x) => x !== id)
    else draftIds = [...draftIds, id]
  }

  async function saveRoleAssignments() {
    err = ''
    try {
      await apiJson(`/api/admin/roles/${encodeURIComponent(selectedRole)}/permissions`, {
        method: 'PUT',
        body: JSON.stringify({ permission_ids: draftIds }),
      })
      assignments = { ...assignments, [selectedRole]: [...draftIds] }
    } catch (e: unknown) {
      err = e instanceof Error ? e.message : 'Save failed'
    }
  }

  async function createPerm() {
    err = ''
    try {
      await apiJson('/api/admin/permissions', {
        method: 'POST',
        body: JSON.stringify({
          name: newName.trim(),
          description: newDesc.trim() || null,
        }),
      })
      addOpen = false
      newName = ''
      newDesc = ''
      await loadOverview()
    } catch (e: unknown) {
      err = e instanceof Error ? e.message : 'Create failed'
    }
  }

  async function saveEditDesc() {
    if (!editPerm) return
    err = ''
    try {
      await apiJson(`/api/admin/permissions/${editPerm.id}`, {
        method: 'PATCH',
        body: JSON.stringify({ description: editDesc.trim() || null }),
      })
      editPerm = null
      await loadOverview()
    } catch (e: unknown) {
      err = e instanceof Error ? e.message : 'Update failed'
    }
  }

  async function removePerm(p: Perm) {
    if (!confirm(`Delete permission “${p.name}”? Role links will be removed.`)) return
    err = ''
    try {
      await apiJson<{ message: string }>(`/api/admin/permissions/${p.id}`, { method: 'DELETE' })
      await loadOverview()
    } catch (e: unknown) {
      err = e instanceof Error ? e.message : 'Delete failed'
    }
  }

  function openEdit(p: Perm) {
    editPerm = p
    editDesc = p.description ?? ''
  }
</script>

<div class="route-root">
  {#if err}
    <p class="error">{err}</p>
  {/if}

  {#if loading}
    <p class="muted">Loading…</p>
  {:else}
    <div class="perms-panel">
      <div class="subtab-bar" role="tablist" aria-label="Permissions sections">
      <button
        type="button"
        class="subtab-trigger"
        class:active={permTab === 'catalog'}
        role="tab"
        aria-selected={permTab === 'catalog'}
        tabindex={permTab === 'catalog' ? 0 : -1}
        onclick={() => (permTab = 'catalog')}
      >
        Permission catalog
      </button>
      <button
        type="button"
        class="subtab-trigger"
        class:active={permTab === 'roles'}
        role="tab"
        aria-selected={permTab === 'roles'}
        tabindex={permTab === 'roles' ? 0 : -1}
        onclick={() => (permTab = 'roles')}
      >
        Role ↔ permissions
      </button>
    </div>

    {#if permTab === 'catalog'}
      <div class="card page-table-card">
        <div class="toolbar-row">
          <button type="button" onclick={() => (addOpen = true)}>Add permission</button>
        </div>

        <div class="page-table-scroll">
          <div class="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Description</th>
                  <th class="text-right">Actions</th>
                </tr>
              </thead>
              <tbody>
                {#each permissions as p}
                  <tr>
                    <td><code style="font-size: 0.85rem;">{p.name}</code></td>
                    <td class="muted">{p.description ?? '—'}</td>
                    <td class="text-right actions-inline">
                      <button type="button" class="secondary small" onclick={() => openEdit(p)}
                        >Edit</button>
                      <button type="button" class="secondary small" onclick={() => removePerm(p)}
                        >Delete</button>
                    </td>
                  </tr>
                {:else}
                  <tr>
                    <td colspan="3" class="muted">No permissions yet.</td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    {:else}
      <div class="card page-table-card">
        <div
          class="row page-table-meta"
          style="align-items: flex-end; gap: 0.75rem; flex-wrap: wrap;"
        >
          <div class="form-field" style="max-width: 280px; margin: 0; flex: 1 1 200px;">
            <label for="role-pick">Role</label>
            <select id="role-pick" bind:value={selectedRole} onchange={syncDraftForRole}>
              {#each roleNames as rn}
                <option value={rn}>{rn}</option>
              {/each}
            </select>
          </div>
          <button type="button" onclick={saveRoleAssignments}>Save for {selectedRole}</button>
        </div>

        <div class="page-table-scroll">
          <div class="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Description</th>
                  <th class="text-center">Granted</th>
                </tr>
              </thead>
              <tbody>
                {#each permissions as p}
                  <tr>
                    <td><code style="font-size: 0.85rem;">{p.name}</code></td>
                    <td class="muted">{p.description ?? '—'}</td>
                    <td class="checkbox-cell">
                      <input
                        type="checkbox"
                        aria-label={`Grant ${p.name} to role ${selectedRole}`}
                        checked={draftIds.includes(p.id)}
                        onchange={() => togglePermForRole(p.id)}
                      />
                    </td>
                  </tr>
                {:else}
                  <tr>
                    <td colspan="3" class="muted">Add permissions in the catalog tab first.</td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    {/if}
    </div>
  {/if}
</div>

{#if addOpen}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <div class="modal-backdrop" onclick={() => (addOpen = false)} role="presentation">
    <div
      class="modal card"
      onclick={(e) => e.stopPropagation()}
      role="dialog"
      aria-modal="true"
      tabindex="-1"
    >
      <h2>Add permission</h2>
      <p class="modal-sub">Use lowercase <code>snake_case</code>, e.g. <code>export_data</code>.</p>
      <div class="form-field">
        <label for="pn">Slug</label>
        <input id="pn" bind:value={newName} placeholder="my_permission" autocomplete="off" />
      </div>
      <div class="form-field">
        <label for="pd">Description</label>
        <textarea id="pd" rows="2" bind:value={newDesc} placeholder="What this allows"></textarea>
      </div>
      <div class="modal-actions">
        <button type="button" class="secondary" onclick={() => (addOpen = false)}>Cancel</button>
        <button type="button" onclick={createPerm} disabled={!newName.trim()}>Create</button>
      </div>
    </div>
  </div>
{/if}

{#if editPerm}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <div class="modal-backdrop" onclick={() => (editPerm = null)} role="presentation">
    <div
      class="modal card"
      onclick={(e) => e.stopPropagation()}
      role="dialog"
      aria-modal="true"
      tabindex="-1"
    >
      <h2>Edit description</h2>
      <p class="modal-sub"><code>{editPerm.name}</code></p>
      <div class="form-field">
        <label for="ed">Description</label>
        <textarea id="ed" rows="3" bind:value={editDesc}></textarea>
      </div>
      <div class="modal-actions">
        <button type="button" class="secondary" onclick={() => (editPerm = null)}>Cancel</button>
        <button type="button" onclick={saveEditDesc}>Save</button>
      </div>
    </div>
  </div>
{/if}
