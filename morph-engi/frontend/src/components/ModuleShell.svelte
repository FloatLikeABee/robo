<script lang="ts">
  import type { Snippet } from 'svelte'

  export type ModuleTab = { id: string; label: string }

  let {
    title,
    hint = '',
    tabs,
    activeTab = $bindable(''),
    actionError = '',
    header,
    children,
  } = $props<{
    title: string
    hint?: string
    tabs: ModuleTab[]
    activeTab?: string
    actionError?: string
    header?: Snippet
    children?: Snippet
  }>()
</script>

<div class="module-shell flex flex-col h-full min-h-0 gap-3">
  <div class="shrink-0">
    <h1 class="text-2xl font-semibold">{title}</h1>
    {#if hint}<p class="text-sm text-muted mt-1">{hint}</p>{/if}
    {#if actionError}<p class="text-rose-400 text-sm mt-2">{actionError}</p>{/if}
    {@render header?.()}
  </div>

  {#if tabs.length > 1}
    <div class="module-tabs shrink-0 flex flex-wrap gap-1 p-1 rounded-2xl bg-white/[0.04] border border-white/5 w-fit max-w-full">
      {#each tabs as tab}
        <button
          type="button"
          class="module-tab px-3 py-1.5 rounded-xl text-sm transition-colors {activeTab === tab.id
            ? 'bg-violet/30 text-text font-medium'
            : 'text-muted hover:text-text hover:bg-white/5'}"
          onclick={() => (activeTab = tab.id)}
        >
          {tab.label}
        </button>
      {/each}
    </div>
  {/if}

  <div class="module-body flex-1 min-h-0 overflow-hidden flex flex-col">
    <div class="flex-1 min-h-0 h-full">
      {@render children?.()}
    </div>
  </div>
</div>
