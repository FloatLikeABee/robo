<script lang="ts">
	import { metabaseProxyUrl } from '$lib/metabaseUrl';

	/** When true, shows Metabase UI through the backend proxy (same-origin iframe). */
	export let open = false;
	export let onClose: () => void;

	$: src = metabaseProxyUrl('');
</script>

{#if open}
	<div
		class="fixed inset-0 z-[320] flex items-center justify-center p-3 sm:p-6"
		role="presentation"
	>
		<button
			type="button"
			class="absolute inset-0 bg-black/65 backdrop-blur-[1px]"
			aria-label="Close Metabase"
			on:click={onClose}
		></button>
		<div
			class="relative flex h-[min(92vh,920px)] w-full max-w-[min(96vw,1440px)] flex-col overflow-hidden rounded-xl border border-border bg-bg-elevated shadow-2xl pointer-events-auto"
			role="dialog"
			aria-modal="true"
			aria-labelledby="metabase-modal-title"
		>
			<div
				class="flex shrink-0 items-center justify-between gap-3 border-b border-border bg-bg-secondary/80 px-4 py-3"
			>
				<div>
					<h2 id="metabase-modal-title" class="text-lg font-semibold text-text-primary">Metabase</h2>
					<p class="mt-0.5 text-xs text-text-tertiary">
						Proxied through DataX. Sign in to Metabase here if prompted.
					</p>
				</div>
				<button
					type="button"
					on:click={onClose}
					class="shrink-0 rounded-md border border-border bg-bg-secondary px-3 py-1.5 text-sm text-text-primary hover:bg-bg-tertiary"
				>
					Close
				</button>
			</div>
			<iframe
				title="Metabase"
				class="min-h-0 w-full flex-1 border-0 bg-bg-secondary"
				src={src}
			></iframe>
		</div>
	</div>
{/if}
