<script lang="ts">
  let { title = '', children } = $props<{
    title?: string
    children?: import('svelte').Snippet
  }>()
</script>

<div class="data-grid-panel card flex flex-col min-h-0 h-full overflow-hidden">
  {#if title}
    <div class="shrink-0 px-4 py-3 border-b border-white/5 font-semibold text-sm">{title}</div>
  {/if}
  <div class="data-grid-scroll flex-1 min-h-0 overflow-auto">
    {@render children?.()}
  </div>
</div>
