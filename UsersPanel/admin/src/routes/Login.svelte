<script lang="ts">
  import { apiJson } from '../lib/api'
  import { setToken } from '../lib/auth'
  import { navigate } from '../lib/router'

  let email = $state('')
  let password = $state('')
  let err = $state('')
  let loading = $state(false)

  async function submit(e: Event) {
    e.preventDefault()
    err = ''
    loading = true
    try {
      const res = await apiJson<{ token: string }>('/api/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
      })
      setToken(res.token)
      document.body.style.removeProperty('overflow')
      document.body.style.removeProperty('padding-right')
      navigate('/users')
    } catch (e: unknown) {
      err = e instanceof Error ? e.message : 'Login failed'
    } finally {
      loading = false
    }
  }
</script>

<div class="route-root route-auth">
  <div class="card stack" style="max-width: 420px; width: 100%">
  <h1>Sign in</h1>
  <form class="stack" onsubmit={submit}>
    <label class="stack">
      <span class="muted">Email</span>
      <input type="email" autocomplete="username" bind:value={email} required />
    </label>
    <label class="stack">
      <span class="muted">Password</span>
      <input type="password" autocomplete="current-password" bind:value={password} required />
    </label>
    {#if err}
      <p class="error">{err}</p>
    {/if}
    <button type="submit" disabled={loading}>{loading ? 'Signing in…' : 'Sign in'}</button>
  </form>

  <div class="row mt-2">
    <a href="#/register">Create account</a>
    <span class="muted">·</span>
    <a href="#/reset-password">Forgot password</a>
  </div>
  </div>
</div>
