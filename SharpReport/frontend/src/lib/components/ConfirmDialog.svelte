<script lang="ts">
  import { onMount } from 'svelte'
  import { finishConfirm, getConfirmPending, subscribeConfirm } from '$lib/confirmDialog'

  let pending = $state(getConfirmPending())

  onMount(() => subscribeConfirm(() => {
    pending = getConfirmPending()
  }))
</script>

{#if pending}
  <div
    class="fixed inset-0 z-[200] flex items-center justify-center p-4 bg-slate-900/50 backdrop-blur-[2px]"
    role="presentation"
    onclick={() => finishConfirm(false)}
  >
    <div
      role="alertdialog"
      aria-modal="true"
      aria-labelledby="confirm-dialog-title"
      aria-describedby="confirm-dialog-desc"
      class="w-full max-w-md rounded-xl border border-border bg-bg-secondary text-text-primary shadow-xl"
      onclick={(e) => e.stopPropagation()}
    >
      <div class="px-4 py-3 border-b border-border">
        <h2 id="confirm-dialog-title" class="font-semibold text-sm">{pending.title}</h2>
      </div>
      <p id="confirm-dialog-desc" class="px-4 py-3 text-sm text-text-secondary whitespace-pre-wrap">{pending.message}</p>
      <div class="px-4 py-3 border-t border-border flex justify-end gap-2">
        {#if !pending.alert}
          <button
            type="button"
            class="px-3 py-1.5 rounded-lg bg-bg-primary border border-border text-sm hover:bg-bg-tertiary"
            onclick={() => finishConfirm(false)}
          >
            {pending.cancelLabel}
          </button>
        {/if}
        <button
          type="button"
          class="px-3 py-1.5 rounded-lg text-sm text-white {pending.danger ? 'bg-red-600 hover:bg-red-500' : 'bg-violet-600 hover:bg-violet-500'}"
          onclick={() => finishConfirm(true)}
        >
          {pending.confirmLabel}
        </button>
      </div>
    </div>
  </div>
{/if}
