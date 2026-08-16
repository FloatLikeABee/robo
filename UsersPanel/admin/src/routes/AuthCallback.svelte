<script lang="ts">
  import { get } from 'svelte/store'
  import { onMount } from 'svelte'
  import { setToken } from '../lib/auth'
  import { route, navigate } from '../lib/router'

  let msg = $state('Completing sign-in…')

  onMount(() => {
    const { query } = get(route)
    const t = query.get('token')
    const err = query.get('error')
    if (err) {
      msg = `Sign-in issue: ${err}`
      return
    }
    if (t) {
      setToken(t)
      navigate('/users')
      return
    }
    msg = 'Missing token in callback URL.'
  })
</script>

<div class="route-root route-auth">
  <div class="card" style="max-width: 480px; width: 100%">
    <p>{msg}</p>
    <p class="mt-2"><a href="#/login">Back to login</a></p>
  </div>
</div>
