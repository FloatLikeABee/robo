<script lang="ts">
  import { toasts } from '../lib/toast'
</script>

<div class="toast-host" aria-live="polite" aria-atomic="false">
  {#each $toasts as t (t.id)}
    <div class="toast toast-{t.type}" role="status">{t.message}</div>
  {/each}
</div>

<style>
  .toast-host {
    position: fixed;
    top: 1rem;
    right: 1rem;
    z-index: 2000;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    max-width: min(22rem, calc(100vw - 2rem));
    pointer-events: none;
  }

  .toast {
    pointer-events: auto;
    padding: 0.75rem 1rem;
    border-radius: 10px;
    font-size: 0.9rem;
    font-weight: 600;
    box-shadow: var(--shadow);
    animation: toast-in 0.2s ease;
  }

  .toast-success {
    background: var(--surface);
    color: var(--success);
    border: 1px solid color-mix(in srgb, var(--success) 45%, var(--border));
  }

  .toast-error {
    background: var(--surface);
    color: var(--danger);
    border: 1px solid color-mix(in srgb, var(--danger) 45%, var(--border));
  }

  @keyframes toast-in {
    from {
      opacity: 0;
      transform: translateY(-6px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
</style>
