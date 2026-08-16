<script lang="ts">
	import { backendLastError, backendStatus, pingBackend } from '$lib/backendHealth';
	import { checkAuth, auth } from '$lib/stores/auth.svelte';

	let retrying = $state(false);

	async function onRetry() {
		retrying = true;
		try {
			const ok = await pingBackend();
			if (ok) {
				const next = await checkAuth();
				auth.set(next);
			}
		} finally {
			retrying = false;
		}
	}
</script>

{#if $backendStatus === 'offline'}
	<div
		class="backend-status-banner flex shrink-0 flex-wrap items-center justify-between gap-2 border-b px-4 py-2 text-sm"
		role="status"
	>
		<span class="min-w-0">
			<strong class="font-semibold">DataX API offline.</strong>
			{$backendLastError || ' The backend is not responding.'}
		</span>
		<button
			type="button"
			class="shrink-0 rounded-md border px-3 py-1 text-xs font-semibold disabled:opacity-60"
			disabled={retrying}
			onclick={onRetry}
		>
			{retrying ? 'Checking…' : 'Retry connection'}
		</button>
	</div>
{/if}

<style>
	.backend-status-banner {
		border-color: rgba(245, 158, 11, 0.45);
		background: rgba(245, 158, 11, 0.12);
		color: var(--color-text-primary, #f8fafc);
	}
	.backend-status-banner button {
		border-color: rgba(245, 158, 11, 0.55);
		color: inherit;
	}
	.backend-status-banner button:hover:not(:disabled) {
		background: rgba(245, 158, 11, 0.18);
	}
	:global([data-theme='light']) .backend-status-banner {
		background: #fff7ed;
		color: #78350f;
	}
</style>
