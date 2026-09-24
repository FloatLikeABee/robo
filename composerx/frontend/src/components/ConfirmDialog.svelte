<script>
  import { onMount } from 'svelte'
  import ButtonLeadingIcon from '../lib/ButtonLeadingIcon.svelte'
  import { finishConfirm, getConfirmPending, subscribeConfirm } from '../lib/confirmDialog.js'

  let pending = $state(getConfirmPending())

  onMount(() => subscribeConfirm(() => {
    pending = getConfirmPending()
  }))
</script>

{#if pending}
  <div
    class="modal-backdrop confirm-dialog-backdrop"
    role="presentation"
    tabindex="-1"
    onclick={() => finishConfirm(false)}
    onkeydown={(e) => e.key === 'Escape' && finishConfirm(false)}
  >
    <div
      class="modal-panel confirm-dialog-panel"
      role="alertdialog"
      aria-modal="true"
      aria-labelledby="confirm-dialog-title"
      aria-describedby="confirm-dialog-desc"
      tabindex="-1"
      onclick={(e) => e.stopPropagation()}
      onkeydown={(e) => {
        if (e.key === 'Escape') {
          e.preventDefault()
          finishConfirm(false)
        }
        e.stopPropagation()
      }}
    >
      <header class="modal-header">
        <h2 id="confirm-dialog-title" class="modal-title">{pending.title}</h2>
      </header>
      <div class="modal-body">
        <p id="confirm-dialog-desc" class="confirm-dialog-message">{pending.message}</p>
        <div class="modal-footer-row confirm-dialog-actions">
          {#if !pending.alert}
            <button type="button" class="btn-ghost" onclick={() => finishConfirm(false)}>
              {pending.cancelLabel}
            </button>
          {/if}
          <button
            type="button"
            class="btn-secondary"
            class:btn-confirm-danger={pending.danger}
            onclick={() => finishConfirm(true)}
          >
            {#if pending.danger}
              <ButtonLeadingIcon name="danger" />
            {:else}
              <ButtonLeadingIcon name="confirmOk" />
            {/if}
            {pending.confirmLabel}
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}
