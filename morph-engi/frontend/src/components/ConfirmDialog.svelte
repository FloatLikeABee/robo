<script lang="ts">
  import { onMount } from 'svelte'
  import { finishConfirm, getConfirmPending, subscribeConfirm } from '../lib/confirmDialog'

  let pending = $state(getConfirmPending())

  onMount(() => subscribeConfirm(() => {
    pending = getConfirmPending()
  }))
</script>

{#if pending}
  <div
    class="fixed inset-0 z-[200] flex items-center justify-center p-4 bg-black/60 backdrop-blur-[2px]"
    role="presentation"
    onclick={() => finishConfirm(false)}
  >
    <div
      role="alertdialog"
      aria-modal="true"
      aria-labelledby="confirm-dialog-title"
      aria-describedby="confirm-dialog-desc"
      class="card w-full max-w-md shadow-2xl"
      onclick={(e) => e.stopPropagation()}
    >
      <div class="px-4 py-3 border-b border-white/10">
        <h2 id="confirm-dialog-title" class="font-semibold text-sm">{pending.title}</h2>
      </div>
      <p id="confirm-dialog-desc" class="px-4 py-3 text-sm text-muted whitespace-pre-wrap">{pending.message}</p>
      <div class="px-4 py-3 border-t border-white/10 flex justify-end gap-2">
        {#if !pending.alert}
          <button type="button" class="btn-ghost border border-white/10 px-3 py-1.5 rounded-xl text-sm" onclick={() => finishConfirm(false)}>
            {pending.cancelLabel}
          </button>
        {/if}
        <button
          type="button"
          class="px-3 py-1.5 rounded-xl text-sm {pending.danger ? 'bg-rose-600 hover:bg-rose-500 text-white' : 'btn-primary'}"
          onclick={() => finishConfirm(true)}
        >
          {pending.confirmLabel}
        </button>
      </div>
    </div>
  </div>
{/if}
