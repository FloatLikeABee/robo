<script lang="ts">
  import { onMount } from 'svelte'
  import { apiJson } from '../lib/api'
  import { showToast } from '../lib/toast'

  type Row = {
    id: string
    email: string
    username: string
    isVerified: boolean
    isAdmin: boolean
    permissions: string[]
    defaultChannelId: string
    createdAt: string
  }

  type MorphRow = {
    id: number
    loginId: string | null
    firstName: string | null
    lastName: string
    email: string | null
    phone: string | null
    administrator: boolean
  }

  const APP_PERMISSIONS = [
    { id: 'morph_util', label: 'Morph Util' },
    { id: 'morph_booki', label: 'Morph Booki' },
    { id: 'morph_engi', label: 'Morph Engi' },
  ] as const

  const DEFAULT_PERMISSIONS = APP_PERMISSIONS.map((p) => p.id)

  function permissionLabel(id: string): string {
    return APP_PERMISSIONS.find((p) => p.id === id)?.label ?? id
  }

  let users = $state<Row[]>([])
  let morphUsers = $state<MorphRow[]>([])
  let err = $state('')

  let accessUser = $state<Row | null>(null)
  let draftIsAdmin = $state(false)
  let draftPermissions = $state<string[]>([])
  let savingAccess = $state(false)

  let profileUser = $state<Row | null>(null)
  let draftEmail = $state('')
  let draftUsername = $state('')
  let savingProfile = $state(false)
  let createOpen = $state(false)
  let savingCreate = $state(false)
  let createEmail = $state('')
  let createUsername = $state('')
  let createPassword = $state('')
  let createIsVerified = $state(true)
  let createIsAdmin = $state(false)
  let createPermissions = $state<string[]>([...DEFAULT_PERMISSIONS])
  let createLoginId = $state('')
  let createFirstName = $state('')
  let createLastName = $state('')
  let createPhone = $state('')
  let createAdministrator = $state(false)

  let morphProfileAccount = $state<Row | null>(null)
  let draftMorphFirstName = $state('')
  let draftMorphLastName = $state('')
  let draftMorphEmail = $state('')
  let draftMorphPhone = $state('')
  let draftMorphAdministrator = $state(false)
  let savingMorph = $state(false)

  function morphForAccount(u: Row): MorphRow | null {
    return (
      morphUsers.find(
        (m) => m.loginId && m.loginId.toLowerCase() === u.username.toLowerCase(),
      ) ?? null
    )
  }

  let morphProfileForModal = $derived(
    morphProfileAccount ? morphForAccount(morphProfileAccount) : null,
  )

  async function reloadAll() {
    await Promise.all([loadUsers(), loadMorphUsers()])
  }

  onMount(() => {
    const reloadUsers = () => void reloadAll()
    window.addEventListener('users-panel:reload-users', reloadUsers)
    void Promise.all([loadUsers(), loadMorphUsers()])
    return () => window.removeEventListener('users-panel:reload-users', reloadUsers)
  })

  async function loadMorphUsers() {
    try {
      const res = await apiJson<{ users: MorphRow[] }>('/api/admin/morph-users')
      morphUsers = res.users
    } catch (e: unknown) {
      err = e instanceof Error ? e.message : 'Failed to load MorphData profiles'
    }
  }

  async function loadUsers() {
    err = ''
    try {
      const res = await apiJson<{ users: Row[] }>('/api/admin/users')
      users = res.users
    } catch (e: unknown) {
      err = e instanceof Error ? e.message : 'Failed to load users'
    }
  }

  function openAccessModal(u: Row) {
    accessUser = u
    draftIsAdmin = u.isAdmin
    draftPermissions = u.isAdmin ? [...DEFAULT_PERMISSIONS] : [...u.permissions]
  }

  function closeAccessModal() {
    accessUser = null
  }

  function openProfileModal(u: Row) {
    profileUser = u
    draftEmail = u.email
    draftUsername = u.username
  }

  function closeProfileModal() {
    profileUser = null
  }

  function togglePermission(id: string) {
    if (draftPermissions.includes(id)) draftPermissions = draftPermissions.filter((x) => x !== id)
    else draftPermissions = [...draftPermissions, id]
  }

  function toggleCreatePermission(id: string) {
    if (createPermissions.includes(id)) createPermissions = createPermissions.filter((x) => x !== id)
    else createPermissions = [...createPermissions, id]
  }

  function openCreateModal() {
    createOpen = true
    createEmail = ''
    createUsername = ''
    createPassword = ''
    createIsVerified = true
    createIsAdmin = false
    createPermissions = [...DEFAULT_PERMISSIONS]
    createLoginId = ''
    createFirstName = ''
    createLastName = ''
    createPhone = ''
    createAdministrator = false
  }

  function closeCreateModal() {
    createOpen = false
  }

  async function saveAccess() {
    if (!accessUser) return
    savingAccess = true
    err = ''
    try {
      await apiJson(`/api/admin/users/${accessUser.id}/access`, {
        method: 'PATCH',
        body: JSON.stringify({
          is_admin: draftIsAdmin,
          permissions: draftIsAdmin ? [] : draftPermissions,
        }),
      })
      closeAccessModal()
      await loadUsers()
    } catch (e: unknown) {
      err = e instanceof Error ? e.message : 'Save failed'
    } finally {
      savingAccess = false
    }
  }

  async function saveProfile() {
    if (!profileUser) return
    savingProfile = true
    err = ''
    try {
      await apiJson(`/api/admin/users/${profileUser.id}`, {
        method: 'PATCH',
        body: JSON.stringify({
          email: draftEmail.trim(),
          username: draftUsername.trim(),
        }),
      })
      closeProfileModal()
      await loadUsers()
    } catch (e: unknown) {
      err = e instanceof Error ? e.message : 'Save failed'
    } finally {
      savingProfile = false
    }
  }

  async function deleteUser(u: Row) {
    if (!confirm(`Delete user ${u.username} (${u.email})? This cannot be undone.`)) return
    err = ''
    try {
      await apiJson<{ message: string }>(`/api/admin/users/${u.id}`, { method: 'DELETE' })
      await loadUsers()
    } catch (e: unknown) {
      err = e instanceof Error ? e.message : 'Delete failed'
    }
  }

  async function createUser() {
    savingCreate = true
    err = ''
    try {
      await apiJson('/api/admin/users', {
        method: 'POST',
        body: JSON.stringify({
          email: createEmail.trim(),
          username: createUsername.trim(),
          password: createPassword,
          is_verified: createIsVerified,
          is_admin: createIsAdmin,
          permissions: createIsAdmin ? [] : createPermissions,
          login_id: createLoginId.trim() || undefined,
          first_name: createFirstName.trim() || undefined,
          last_name: createLastName.trim(),
          phone: createPhone.trim() || undefined,
          administrator: createAdministrator,
        }),
      })
      closeCreateModal()
      await reloadAll()
    } catch (e: unknown) {
      err = e instanceof Error ? e.message : 'Create failed'
    } finally {
      savingCreate = false
    }
  }

  function openMorphProfileModal(u: Row) {
    const morph = morphForAccount(u)
    morphProfileAccount = u
    draftMorphFirstName = morph?.firstName || ''
    draftMorphLastName = morph?.lastName || ''
    draftMorphEmail = morph?.email || u.email
    draftMorphPhone = morph?.phone || ''
    draftMorphAdministrator = morph?.administrator || false
  }

  function closeMorphProfileModal() {
    morphProfileAccount = null
  }

  async function saveMorphProfile() {
    const morph = morphProfileForModal
    if (!morph) return
    savingMorph = true
    err = ''
    try {
      await apiJson(`/api/admin/morph-users/${morph.id}`, {
        method: 'PATCH',
        body: JSON.stringify({
          first_name: draftMorphFirstName.trim() || null,
          last_name: draftMorphLastName.trim(),
          email: draftMorphEmail.trim() || null,
          phone: draftMorphPhone.trim() || null,
          administrator: draftMorphAdministrator,
        }),
      })
      closeMorphProfileModal()
      await loadMorphUsers()
      showToast('Profile saved successfully.')
    } catch (e: unknown) {
      err = e instanceof Error ? e.message : 'Save failed'
    } finally {
      savingMorph = false
    }
  }

  async function deactivateMorphProfile() {
    const morph = morphProfileForModal
    if (!morph) return
    const label =
      [morph.firstName, morph.lastName].filter(Boolean).join(' ') || morph.email || `User ${morph.id}`
    if (!confirm(`Deactivate ${label}? They will no longer appear in MorphData.`)) return
    err = ''
    try {
      await apiJson(`/api/admin/morph-users/${morph.id}`, { method: 'DELETE' })
      closeMorphProfileModal()
      await loadMorphUsers()
    } catch (e: unknown) {
      err = e instanceof Error ? e.message : 'Deactivate failed'
    }
  }

  function onKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      if (accessUser) closeAccessModal()
      if (profileUser) closeProfileModal()
      if (createOpen) closeCreateModal()
      if (morphProfileAccount) closeMorphProfileModal()
    }
  }
</script>

<svelte:window onkeydown={onKeydown} />

<div class="route-root">
  {#if err}
    <p class="error">{err}</p>
  {/if}

  <div class="section-actions" style="margin-bottom: 12px;">
    <button type="button" class="small" onclick={openCreateModal}>Add user</button>
  </div>

  <h2 class="section-title">Users</h2>
  <p class="muted section-sub" style="margin-bottom: 12px;">
    Add accounts as Admin or non-admin. Admins can open Users Panel and have every app permission. Non-admins only get the apps you enable below.
  </p>

  <div class="page-table-scroll">
    <div class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Account</th>
            <th>Verified</th>
            <th>Admin</th>
            <th>Apps</th>
            <th class="text-right">Actions</th>
          </tr>
        </thead>
        <tbody>
          {#each users as u}
            <tr>
              <td>
                <strong>{u.username}</strong>
                <div class="muted">{u.email}</div>
              </td>
              <td>{u.isVerified ? 'Yes' : 'No'}</td>
              <td>{u.isAdmin ? 'Yes' : 'No'}</td>
              <td>
                <div class="row" style="flex-wrap: wrap;">
                  {#if u.isAdmin}
                    <span class="pill">All apps</span>
                  {:else}
                    {#each u.permissions as p}
                      <span class="pill">{permissionLabel(p)}</span>
                    {:else}
                      <span class="muted">—</span>
                    {/each}
                  {/if}
                </div>
              </td>
              <td class="text-right">
                <div class="user-actions-grid">
                  <button type="button" class="secondary small" onclick={() => openMorphProfileModal(u)}>Profile</button>
                  <button type="button" class="secondary small" onclick={() => openProfileModal(u)}>Edit</button>
                  <button type="button" class="secondary small" onclick={() => openAccessModal(u)}>Access</button>
                  <button type="button" class="secondary small" onclick={() => deleteUser(u)}>Delete</button>
                </div>
              </td>
            </tr>
          {:else}
            <tr>
              <td colspan="5" class="muted">No users yet.</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  </div>
</div>

{#if accessUser}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <div class="modal-backdrop" onclick={closeAccessModal} role="presentation">
    <div
      class="modal card wide"
      onclick={(e) => e.stopPropagation()}
      role="dialog"
      aria-modal="true"
      aria-labelledby="access-modal-title"
      tabindex="-1"
    >
      <h2 id="access-modal-title">User access</h2>
      <p class="modal-sub">
        <strong>{accessUser.username}</strong>
        <span class="muted"> · {accessUser.email}</span>
      </p>

      <label class="checkbox-row">
        <input
          type="checkbox"
          checked={draftIsAdmin}
          onchange={() => {
            draftIsAdmin = !draftIsAdmin
            if (!draftIsAdmin && draftPermissions.length === 0) {
              draftPermissions = [...DEFAULT_PERMISSIONS]
            }
          }}
        />
        <span>Admin (Users Panel + all apps)</span>
      </label>

      {#if !draftIsAdmin}
        <div class="form-field" style="margin-top: 1rem;">
          <span id="access-perms-label">App permissions</span>
          <div class="role-check-grid" aria-labelledby="access-perms-label">
            {#each APP_PERMISSIONS as p}
              <label>
                <input type="checkbox" checked={draftPermissions.includes(p.id)} onchange={() => togglePermission(p.id)} />
                <span>{p.label}</span>
              </label>
            {/each}
          </div>
        </div>
      {/if}

      <div class="modal-actions">
        <button type="button" class="secondary" onclick={closeAccessModal} disabled={savingAccess}>Cancel</button>
        <button type="button" onclick={saveAccess} disabled={savingAccess}>{savingAccess ? 'Saving…' : 'Save'}</button>
      </div>
    </div>
  </div>
{/if}

{#if profileUser}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <div class="modal-backdrop" onclick={closeProfileModal} role="presentation">
    <div
      class="modal card wide"
      onclick={(e) => e.stopPropagation()}
      role="dialog"
      aria-modal="true"
      aria-labelledby="profile-modal-title"
      tabindex="-1"
    >
      <h2 id="profile-modal-title">Edit account</h2>
      <p class="modal-sub">Update sign-in email and username for <strong>{profileUser.username}</strong>.</p>

      <div class="form-field">
        <label for="pe">Email</label>
        <input id="pe" type="email" bind:value={draftEmail} autocomplete="off" />
      </div>
      <div class="form-field">
        <label for="pu">Username</label>
        <input id="pu" bind:value={draftUsername} autocomplete="off" />
      </div>

      <div class="modal-actions">
        <button type="button" class="secondary" onclick={closeProfileModal} disabled={savingProfile}>Cancel</button>
        <button type="button" onclick={saveProfile} disabled={savingProfile}>{savingProfile ? 'Saving…' : 'Save'}</button>
      </div>
    </div>
  </div>
{/if}

{#if createOpen}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <div class="modal-backdrop" onclick={closeCreateModal} role="presentation">
    <div
      class="modal card wide"
      onclick={(e) => e.stopPropagation()}
      role="dialog"
      aria-modal="true"
      aria-labelledby="create-modal-title"
      tabindex="-1"
    >
      <h2 id="create-modal-title">Add user</h2>
      <p class="modal-sub">Create the sign-in account, MorphData profile, and access settings in one step.</p>

      <div class="form-field">
        <label for="cu-email">Email</label>
        <input id="cu-email" type="email" bind:value={createEmail} autocomplete="off" />
      </div>
      <div class="form-field">
        <label for="cu-username">Username</label>
        <input id="cu-username" bind:value={createUsername} autocomplete="off" />
      </div>
      <div class="form-field">
        <label for="cu-password">Password</label>
        <input id="cu-password" type="password" bind:value={createPassword} autocomplete="new-password" />
      </div>
      <div class="form-field">
        <label for="cu-login-id">MorphData Login ID</label>
        <input id="cu-login-id" bind:value={createLoginId} autocomplete="off" />
      </div>
      <div class="row" style="gap: 12px;">
        <div class="form-field" style="flex: 1;">
          <label for="cu-first-name">First name</label>
          <input id="cu-first-name" bind:value={createFirstName} autocomplete="off" />
        </div>
        <div class="form-field" style="flex: 1;">
          <label for="cu-last-name">Last name</label>
          <input id="cu-last-name" bind:value={createLastName} autocomplete="off" />
        </div>
      </div>
      <div class="row" style="gap: 12px;">
        <div class="form-field" style="flex: 1;">
          <label for="cu-phone">Phone</label>
          <input id="cu-phone" bind:value={createPhone} autocomplete="off" />
        </div>
      </div>
      <div class="row" style="gap: 12px; flex-wrap: wrap;">
        <label class="checkbox-row">
          <input type="checkbox" bind:checked={createIsVerified} />
          <span>Mark as verified</span>
        </label>
        <label class="checkbox-row">
          <input type="checkbox" bind:checked={createAdministrator} />
          <span>MorphData administrator</span>
        </label>
      </div>

      <div class="form-field" style="margin-top: 0.5rem;">
        <label class="checkbox-row">
          <input
            type="checkbox"
            checked={createIsAdmin}
            onchange={() => {
              createIsAdmin = !createIsAdmin
              if (!createIsAdmin && createPermissions.length === 0) {
                createPermissions = [...DEFAULT_PERMISSIONS]
              }
            }}
          />
          <span>Admin (Users Panel + all apps)</span>
        </label>
      </div>

      {#if !createIsAdmin}
        <div class="form-field">
          <span id="create-perms-label">App permissions</span>
          <div class="role-check-grid" aria-labelledby="create-perms-label">
            {#each APP_PERMISSIONS as p}
              <label>
                <input type="checkbox" checked={createPermissions.includes(p.id)} onchange={() => toggleCreatePermission(p.id)} />
                <span>{p.label}</span>
              </label>
            {/each}
          </div>
        </div>
      {/if}

      <div class="modal-actions">
        <button type="button" class="secondary" onclick={closeCreateModal} disabled={savingCreate}>Cancel</button>
        <button type="button" onclick={createUser} disabled={savingCreate}>
          {savingCreate ? 'Creating…' : 'Create user'}
        </button>
      </div>
    </div>
  </div>
{/if}

{#if morphProfileAccount}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <div class="modal-backdrop" onclick={closeMorphProfileModal} role="presentation">
    <div
      class="modal card wide"
      onclick={(e) => e.stopPropagation()}
      role="dialog"
      aria-modal="true"
      aria-labelledby="morph-profile-modal-title"
      tabindex="-1"
    >
      <h2 id="morph-profile-modal-title">MorphData profile</h2>
      <p class="modal-sub">
        <strong>{morphProfileAccount.username}</strong>
        <span class="muted"> · {morphProfileAccount.email}</span>
      </p>

      {#if morphProfileForModal}
        <div class="row" style="gap: 12px;">
          <div class="form-field" style="flex: 1;">
            <label for="me-first-name">First name</label>
            <input id="me-first-name" bind:value={draftMorphFirstName} autocomplete="off" />
          </div>
          <div class="form-field" style="flex: 1;">
            <label for="me-last-name">Last name</label>
            <input id="me-last-name" bind:value={draftMorphLastName} autocomplete="off" />
          </div>
        </div>
        <div class="form-field">
          <label for="me-email">Email</label>
          <input id="me-email" type="email" bind:value={draftMorphEmail} autocomplete="off" />
        </div>
        <div class="form-field">
          <label for="me-phone">Phone</label>
          <input id="me-phone" bind:value={draftMorphPhone} autocomplete="off" />
        </div>
        <label class="checkbox-row">
          <input type="checkbox" bind:checked={draftMorphAdministrator} />
          <span>MorphData administrator</span>
        </label>

        <div class="modal-actions">
          <button type="button" class="secondary" onclick={deactivateMorphProfile}>Deactivate</button>
          <span style="flex: 1;"></span>
          <button type="button" class="secondary" onclick={closeMorphProfileModal} disabled={savingMorph}>Cancel</button>
          <button type="button" onclick={saveMorphProfile} disabled={savingMorph}>{savingMorph ? 'Saving…' : 'Save'}</button>
        </div>
      {:else}
        <p class="muted">No MorphData profile is linked to this account.</p>
        <div class="modal-actions">
          <button type="button" class="secondary" onclick={closeMorphProfileModal}>Close</button>
        </div>
      {/if}
    </div>
  </div>
{/if}

<style>
  .user-actions-grid {
    display: grid;
    grid-template-columns: repeat(2, 5.5rem);
    gap: 0.35rem;
    justify-content: end;
    margin-left: auto;
    width: fit-content;
  }

  .user-actions-grid button {
    width: 100%;
    min-width: 0;
    padding-inline: 0.35rem;
    text-align: center;
  }

  .checkbox-row {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    cursor: pointer;
    font-size: 0.9rem;
    font-weight: 500;
    color: var(--text);
  }

  .checkbox-row input {
    width: auto;
    flex-shrink: 0;
    margin: 0;
  }
</style>
