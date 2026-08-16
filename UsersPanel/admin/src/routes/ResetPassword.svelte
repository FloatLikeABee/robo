<script lang="ts">
  import { get } from 'svelte/store'
  import { onMount } from 'svelte'
  import { apiJson } from '../lib/api'
  import { route, navigate } from '../lib/router'

  let password = $state('')
  let err = $state('')
  let ok = $state('')
  let loading = $state(false)
  let tokenParam = $state('')

  onMount(() => {
    tokenParam = get(route).query.get('token') ?? ''
  })

  async function submit(e: Event) {
    e.preventDefault()
    if (!tokenParam) {
      err = 'Missing reset token in URL.'
      return
    }
    err = ''
    ok = ''
    loading = true
    try {
      const res = await apiJson<{ message: string }>('/api/auth/reset-password', {
        method: 'POST',
        body: JSON.stringify({ token: tokenParam, password }),
      })
      ok = res.message
      navigate('/login')
    } catch (e: unknown) {
      err = e instanceof Error ? e.message : 'Reset failed'
    } finally {
      loading = false
    }
  }
</script>

<div class="route-root route-auth">
  <div class="card stack" style="max-width: 420px; width: 100%">
  <h1>Reset password</h1>
  {#if !tokenParam}
    <p class="error">Open this page from the link in your reset email (with <code>?token=</code> in the URL).</p>
  {:else}
    <form class="stack" onsubmit={submit}>
      <label class="stack">
        <span class="muted">New password (min 8 characters)</span>
        <input type="password" minlength="8" bind:value={password} required />
      </label>
      {#if err}
        <p class="error">{err}</p>
      {/if}
      {#if ok}
        <p class="muted">{ok}</p>
      {/if}
      <button type="submit" disabled={loading}>{loading ? 'Saving…' : 'Update password'}</button>
    </form>
  {/if}
  <p class="mt-2"><a href="#/login">Back to sign in</a></p>
  </div>
</div>
