<script lang="ts">
  import { onMount } from 'svelte'
  import { apiJson } from '../lib/api'

  type Role = {
    id: string
    name: string
    description: string | null
    createdAt: string
    updatedAt: string
  }

  let roles = $state<Role[]>([])
  let err = $state('')
  let loading = $state(true)

  let addOpen = $state(false)
  let newName = $state('')
  let newDesc = $state('')

  let editRole = $state<Role | null>(null)
  let editName = $state('')
  let editDesc = $state('')

  onMount(() => {
    void load()
    const reloadRoles = () => void load()
    window.addEventListener('users-panel:reload-roles', reloadRoles)
    return () => window.removeEventListener('users-panel:reload-roles', reloadRoles)
  })

  async function load() {
    loading = true
    err = ''
    try {
      const res = await apiJson<{ roles: Role[] }>('/api/admin/roles')
      roles = res.roles
    } catch (e: unknown) {
      err = e instanceof Error ? e.message : 'Failed to load roles'
    } finally {
      loading = false
    }
  }

  async function createRole() {
    err = ''
    try {
      await apiJson('/api/admin/roles', {
        method: 'POST',
        body: JSON.stringify({
          name: newName.trim(),
          description: newDesc.trim() || null,
        }),
      })
      addOpen = false
      newName = ''
      newDesc = ''
      await load()
    } catch (e: unknown) {
      err = e instanceof Error ? e.message : 'Create failed'
    }
  }

  async function saveEdit() {
    if (!editRole) return
    err = ''
    try {
      await apiJson(`/api/admin/roles/${editRole.id}`, {
        method: 'PATCH',
        body: JSON.stringify({
          name: editName.trim(),
          description: editDesc.trim() || null,
        }),
      })
      editRole = null
      await load()
    } catch (e: unknown) {
      err = e instanceof Error ? e.message : 'Update failed'
    }
  }

  async function removeRole(r: Role) {
    if (
      !confirm(
        `Delete role “${r.name}”? Permission mappings are removed; users lose this role from their profile.`,
      )
    )
      return
    err = ''
    try {
      await apiJson<{ message: string }>(`/api/admin/roles/${r.id}`, { method: 'DELETE' })
      await load()
    } catch (e: unknown) {
      err = e instanceof Error ? e.message : 'Delete failed'
    }
  }

  function openEdit(r: Role) {
    editRole = r
    editName = r.name
    editDesc = r.description ?? ''
  }
</script>

<div class="route-root">
  {#if err}
    <p class="error">{err}</p>
  {/if}

  {#if loading}
    <p class="muted">Loading…</p>
  {:else}
    <div class="card page-table-card">
      <div class="toolbar-row">
        <button type="button" onclick={() => (addOpen = true)}>Add role</button>
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
            {#each roles as r}
              <tr>
                <td><strong>{r.name}</strong></td>
                <td class="muted">{r.description ?? '—'}</td>
                <td class="text-right actions-inline">
                  <button type="button" class="secondary small" onclick={() => openEdit(r)}>Edit</button>
                  <button type="button" class="secondary small" onclick={() => removeRole(r)}>Delete</button>
                </td>
              </tr>
            {:else}
              <tr>
                <td colspan="3" class="muted">No roles.</td>
              </tr>
            {/each}
          </tbody>
        </table>
        </div>
      </div>
    </div>
  {/if}
</div>

{#if addOpen}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <div class="modal-backdrop" onclick={() => (addOpen = false)} role="presentation">
    <div
      class="modal card wide"
      onclick={(e) => e.stopPropagation()}
      role="dialog"
      aria-modal="true"
      tabindex="-1"
    >
      <h2>Add role</h2>
      <p class="modal-sub">Must match how you reference it in permission assignments (case-sensitive).</p>
      <div class="form-field">
        <label for="rn">Display name</label>
        <input id="rn" bind:value={newName} placeholder="e.g. Support Agent" autocomplete="off" />
      </div>
      <div class="form-field">
        <label for="rd">Description</label>
        <textarea id="rd" rows="2" bind:value={newDesc} placeholder="Optional"></textarea>
      </div>
      <div class="modal-actions">
        <button type="button" class="secondary" onclick={() => (addOpen = false)}>Cancel</button>
        <button type="button" onclick={createRole} disabled={!newName.trim()}>Create</button>
      </div>
    </div>
  </div>
{/if}

{#if editRole}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <div class="modal-backdrop" onclick={() => (editRole = null)} role="presentation">
    <div
      class="modal card wide"
      onclick={(e) => e.stopPropagation()}
      role="dialog"
      aria-modal="true"
      tabindex="-1"
    >
      <h2>Edit role</h2>
      <p class="modal-sub">Renaming updates all users and permission rows.</p>
      <div class="form-field">
        <label for="en">Display name</label>
        <input id="en" bind:value={editName} autocomplete="off" />
      </div>
      <div class="form-field">
        <label for="ed">Description</label>
        <textarea id="ed" rows="3" bind:value={editDesc}></textarea>
      </div>
      <div class="modal-actions">
        <button type="button" class="secondary" onclick={() => (editRole = null)}>Cancel</button>
        <button type="button" onclick={saveEdit} disabled={!editName.trim()}>Save</button>
      </div>
    </div>
  </div>
{/if}
