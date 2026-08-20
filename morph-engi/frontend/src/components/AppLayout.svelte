<script lang="ts">
  import type { PageId } from '../lib/nav'
  import { NAV } from '../lib/nav'
  import PlatformAssistantDrawer from './PlatformAssistantDrawer.svelte'
  import type { AssistantState } from '@robo/platform-chat/usePlatformChat'

  let {
    page = $bindable('projects' as PageId),
    getStateExtra,
    onSignOut,
    children,
  } = $props<{
    page?: PageId
    getStateExtra?: () => AssistantState
    onSignOut?: () => void
    children?: import('svelte').Snippet
  }>()

  let assistantOpen = $state(false)

  const pageLabel = $derived(NAV.find((n) => n.id === page)?.label ?? 'Project')
  const pageHint = $derived(NAV.find((n) => n.id === page)?.hint ?? '')
</script>

<div class="app-frame flex flex-col h-dvh w-full overflow-hidden bg-bg text-text">
  <header class="app-header shrink-0 border-b border-white/5 bg-surface/30">
    <div class="flex items-center justify-between gap-3 px-4 py-3 md:px-6 flex-wrap">
      <div class="flex items-center gap-2.5 min-w-0">
        <img src="/morph-engi-icon.svg" alt="" width="32" height="32" class="h-8 w-8 shrink-0 rounded-xl shadow-[0_0_20px_rgba(45,212,191,0.25)]" />
        <div class="min-w-0">
          <div class="font-semibold leading-tight">Project</div>
        </div>
      </div>
      <div class="flex items-center gap-2 flex-wrap">
        <button type="button" class="px-3 py-1.5 rounded-xl text-xs font-medium bg-violet/30 hover:bg-violet/40" onclick={() => (assistantOpen = true)}>
          AI Assistant
        </button>
        <button type="button" class="px-3 py-1.5 rounded-xl text-xs font-medium text-muted hover:bg-white/5" onclick={onSignOut}>
          Sign out
        </button>
      </div>
    </div>

    <nav class="app-header-tabs flex gap-1 px-4 pb-3 md:px-6 overflow-x-auto" aria-label="Sections">
      {#each NAV as item}
        <button
          type="button"
          class="app-header-tab shrink-0 px-3 py-1.5 rounded-xl text-sm transition-colors {page === item.id
            ? 'bg-violet/30 text-text font-medium'
            : 'text-muted hover:text-text hover:bg-white/5'}"
          onclick={() => (page = item.id)}
        >
          {item.label}
        </button>
      {/each}
    </nav>
  </header>

  <div class="app-pane flex flex-1 flex-col min-w-0 min-h-0 overflow-hidden">
    <main class="app-content flex-1 min-h-0 overflow-hidden px-4 py-4 md:px-6 md:py-5">
      <div class="h-full min-h-0 max-w-6xl mx-auto">
        {@render children?.()}
      </div>
    </main>

    <footer class="app-footer shrink-0 border-t border-white/5 bg-surface/40 hidden md:block">
      <div class="flex items-center justify-between gap-3 px-6 py-2.5 text-xs text-muted">
        <span class="truncate"><span class="text-text font-medium">{pageLabel}</span>{#if pageHint}<span class="mx-2 opacity-40">·</span>{pageHint}{/if}</span>
        <span class="shrink-0 opacity-60">Project</span>
      </div>
    </footer>
  </div>

  <PlatformAssistantDrawer bind:open={assistantOpen} {getStateExtra} />
</div>

<style>
  .app-header-tabs {
    scrollbar-width: none;
  }

  .app-header-tabs::-webkit-scrollbar {
    display: none;
  }
</style>
