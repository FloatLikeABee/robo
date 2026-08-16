<script lang="ts">
  import { apiJson } from '../lib/api'
  import { navigate } from '../lib/router'

  let email = $state('')
  let username = $state('')
  let password = $state('')
  let err = $state('')
  let ok = $state('')
  let loading = $state(false)

  async function submit(e: Event) {
    e.preventDefault()
    err = ''
    ok = ''
    loading = true
    try {
      const res = await apiJson<{ message: string }>('/api/auth/register', {
        method: 'POST',
        body: JSON.stringify({ email, username, password }),
      })
      ok = res.message
    } catch (e: unknown) {
      err = e instanceof Error ? e.message : 'Registration failed'
    } finally {
      loading = false
    }
  }
</script>

<div class="route-root route-auth">
  <div class="card stack" style="max-width: 420px; width: 100%">
  <h1>Register</h1>
  <p class="muted">Password min 8 characters. You’ll need to verify email before login.</p>

  <form class="stack" onsubmit={submit}>
    <label class="stack">
      <span class="muted">Email</span>
      <input type="email" bind:value={email} required />
    </label>
    <label class="stack">
      <span class="muted">Username</span>
      <input bind:value={username} required />
    </label>
    <label class="stack">
      <span class="muted">Password</span>
      <input type="password" minlength="8" bind:value={password} required />
    </label>
    {#if err}
      <p class="error">{err}</p>
    {/if}
    {#if ok}
      <p class="muted">{ok}</p>
    {/if}
    <button type="submit" disabled={loading}>{loading ? 'Submitting…' : 'Register'}</button>
  </form>

  <p class="mt-2"><a href="#/login">Back to sign in</a></p>
  </div>
</div>
