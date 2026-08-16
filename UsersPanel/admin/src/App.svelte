<script lang="ts">
  import { onMount } from 'svelte'
  import { token, logout } from './lib/auth'
  import { route, navigate } from './lib/router'
  import { theme, toggleTheme, applyTheme, readTheme } from './lib/theme'
  import { rolesFromToken } from './lib/jwt'
  import PlatformAssistantDrawer from './components/PlatformAssistantDrawer.svelte'
  import ToastHost from './components/ToastHost.svelte'

  import Login from './routes/Login.svelte'
  import Register from './routes/Register.svelte'
  import Users from './routes/Users.svelte'
  import DataCollector from './routes/DataCollector.svelte'
  import AuthCallback from './routes/AuthCallback.svelte'
  import ResetPassword from './routes/ResetPassword.svelte'

  onMount(() => {
    applyTheme(readTheme())
  })

  const isAuthed = $derived(!!$token)
  const isAdmin = $derived(rolesFromToken($token).includes('Admin'))

  const currentYear = new Date().getFullYear()

  let assistantOpen = $state(false)
</script>

<div class="app-frame">
  <header class="app-header">
    <div class="app-header-inner">
      <div class="app-header-brand">
        <a href={isAuthed ? '#/users' : '#/login'} class="brand">
          <img class="brand-mark" src="/userspanel-logo.svg" width="36" height="36" alt="" aria-hidden="true" />
          Users Panel
        </a>
      </div>
      <nav class="nav app-header-nav" aria-label="Main">
        {#if isAuthed}
          {#if isAdmin}
            <a href="#/users" class:active={$route.path === '/users'}>Users</a>
            <a
              href="#/data-collector"
              class="nav-data-collector"
              class:active={$route.path === '/data-collector'}
            >Data Collector</a>
          {/if}
        {:else}
          <a href="#/login" class:active={$route.path === '/login'}>Login</a>
          <a href="#/register" class:active={$route.path === '/register'}>Register</a>
        {/if}
      </nav>
      <div class="app-header-actions">
        {#if isAuthed}
          <button
            type="button"
            class="secondary small"
            aria-expanded={assistantOpen}
            aria-controls="users-panel-ai-assistant"
            onclick={() => (assistantOpen = !assistantOpen)}
          >
            AI Assistant
          </button>
          <button
            type="button"
            class="ghost btn-icon"
            aria-label="Log out"
            title="Log out"
            onclick={() => {
              logout()
              navigate('/login')
            }}
          >
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M15.75 9V5.25A2.25 2.25 0 0 0 13.5 3h-6a2.25 2.25 0 0 0-2.25 2.25v13.5A2.25 2.25 0 0 0 7.5 21h6a2.25 2.25 0 0 0 2.25-2.25V15M18 9l3 3m0 0-3 3m3-3H9"
              />
            </svg>
          </button>
        {/if}
        <button
          type="button"
          class="secondary btn-icon"
          onclick={() => toggleTheme()}
          title={$theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
          aria-label={$theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
        >
          {#if $theme === 'dark'}
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M12 3v2.25m6.364.386-1.591 1.591M21 12h-2.25m-.386 6.364-1.591-1.591M12 18.75V21m-4.773-4.227-1.591 1.591M5.25 12H3m4.227-4.773L5.636 5.636M15.75 12a3.75 3.75 0 1 1-7.5 0 3.75 3.75 0 0 1 7.5 0Z"
              />
            </svg>
          {:else}
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M21.752 15.002A9.72 9.72 0 0 1 18 15.75c-5.385 0-9.75-4.365-9.75-9.75 0-1.33.266-2.597.748-3.752A9.753 9.753 0 0 0 3 11.25C3 16.635 7.365 21 12.75 21a9.753 9.753 0 0 0 9.002-5.998Z"
              />
            </svg>
          {/if}
        </button>
      </div>
    </div>
  </header>

  <main class="app-main">
    <div class="app-main-inner">
      {#if $route.path === '/login'}
        <Login />
      {:else if $route.path === '/register'}
        <Register />
      {:else if $route.path === '/auth/callback'}
        <AuthCallback />
      {:else if $route.path === '/reset-password'}
        <ResetPassword />
      {:else if $route.path === '/users'}
        {#if !isAuthed}
          <div class="route-root">
            <p class="muted">Sign in required. <a href="#/login">Login</a></p>
          </div>
        {:else if !isAdmin}
          <div class="route-root">
            <p class="error">Admin access required to manage users.</p>
          </div>
        {:else}
          <Users />
        {/if}
      {:else if $route.path === '/data-collector'}
        {#if !isAuthed}
          <div class="route-root">
            <p class="muted">Sign in required. <a href="#/login">Login</a></p>
          </div>
        {:else if !isAdmin}
          <div class="route-root">
            <p class="error">Admin access required.</p>
          </div>
        {:else}
          <DataCollector />
        {/if}
      {:else}
        <div class="route-root">
          <p class="muted">Not found. <a href="#/login">Home</a></p>
        </div>
      {/if}
    </div>
  </main>

  <footer class="app-footer">
    <div class="app-footer-inner">
      Users Panel · {currentYear}
    </div>
  </footer>

  <PlatformAssistantDrawer bind:open={assistantOpen} />
  <ToastHost />
</div>
