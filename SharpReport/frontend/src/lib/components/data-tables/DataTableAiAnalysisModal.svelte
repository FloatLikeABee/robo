<script lang="ts">
	import { analysisDownloadFilename, renderAnalysisMarkdown } from '$lib/analysisMarkdown';
	import { requestTableAiAnalysis } from '$lib/dataTables';
	import { Download, Loader2, X } from 'lucide-svelte';
	import { onMount } from 'svelte';

	let {
		tableId,
		tableName,
		onClose
	}: {
		tableId: string;
		tableName: string;
		onClose: () => void;
	} = $props();

	let dialogEl = $state<HTMLDialogElement | undefined>();
	let loading = $state(true);
	let error = $state('');
	let markdown = $state('');
	let rendered = $state('');
	let abort: AbortController | null = null;

	const downloadName = $derived(analysisDownloadFilename(tableName));

	$effect(() => {
		const src = markdown;
		if (!src) {
			rendered = '';
			return;
		}
		let cancelled = false;
		void renderAnalysisMarkdown(src).then((html) => {
			if (!cancelled) rendered = html;
		});
		return () => {
			cancelled = true;
		};
	});

	function onBackdropClick(event: MouseEvent) {
		const el = dialogEl;
		if (!el || event.target !== el) return;
		const rect = el.getBoundingClientRect();
		const inContent =
			rect.top <= event.clientY &&
			event.clientY <= rect.top + rect.height &&
			rect.left <= event.clientX &&
			event.clientX <= rect.left + rect.width;
		if (!inContent) el.close();
	}

	function handleClose() {
		abort?.abort();
		onClose();
	}

	function downloadMarkdown() {
		if (!markdown) return;
		const blob = new Blob([markdown], { type: 'text/markdown;charset=utf-8' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = downloadName;
		a.click();
		URL.revokeObjectURL(url);
	}

	onMount(() => {
		const el = dialogEl;
		if (!el) return;
		el.showModal();
		if (!('closedBy' in HTMLDialogElement.prototype)) {
			el.addEventListener('click', onBackdropClick);
		}
		const controller = new AbortController();
		abort = controller;
		void (async () => {
			loading = true;
			error = '';
			markdown = '';
			try {
				const result = await requestTableAiAnalysis(tableId, controller.signal);
				if (controller.signal.aborted) return;
				markdown = result.markdown;
			} catch (e) {
				if (controller.signal.aborted) return;
				error = e instanceof Error ? e.message : 'Could not generate analysis.';
			} finally {
				if (!controller.signal.aborted) loading = false;
			}
		})();
		return () => {
			abort?.abort();
			el.removeEventListener('click', onBackdropClick);
		};
	});
</script>

<dialog
	bind:this={dialogEl}
	class="analysis-dialog m-auto flex max-h-[min(90vh,640px)] w-[calc(100%-1.5rem)] max-w-2xl flex-col overflow-hidden rounded-xl border border-border bg-bg-elevated p-0 text-text-primary shadow-2xl"
	closedby="any"
	aria-labelledby="ai-analysis-title"
	aria-busy={loading}
	onclose={handleClose}
>
	<div class="flex shrink-0 items-center justify-between gap-3 border-b border-border px-4 py-3">
		<div class="min-w-0">
			<h2 id="ai-analysis-title" class="truncate text-lg font-semibold text-text-primary">
				AI analysis · {tableName}
			</h2>
		</div>
		<div class="flex shrink-0 items-center gap-2">
			{#if markdown && !error}
				<button
					type="button"
					class="inline-flex items-center gap-1.5 rounded-md border border-border px-2.5 py-1.5 text-sm text-text-secondary hover:bg-bg-tertiary hover:text-text-primary"
					onclick={downloadMarkdown}
				>
					<Download class="h-4 w-4" />
					Download .md
				</button>
			{/if}
			<form method="dialog">
				<button
					type="submit"
					class="rounded-md border border-border p-1.5 text-text-secondary hover:bg-bg-tertiary"
					aria-label="Close"
				>
					<X class="h-4 w-4" />
				</button>
			</form>
		</div>
	</div>
	<div class="min-h-0 flex-1 overflow-y-auto p-4">
		{#if loading}
			<p class="flex items-center gap-2 text-sm text-text-secondary">
				<Loader2 class="h-4 w-4 animate-spin" />
				Generating analysis…
			</p>
		{:else if error}
			<p class="text-sm text-error">{error}</p>
		{:else}
			<div class="analysis-markdown text-sm text-text-primary">
				{@html rendered}
			</div>
		{/if}
	</div>
</dialog>

<style>
	.analysis-dialog {
		color-scheme: dark;
	}

	.analysis-dialog::backdrop {
		background: rgba(0, 0, 0, 0.65);
		backdrop-filter: blur(2px);
	}

	.analysis-markdown :global(h1),
	.analysis-markdown :global(h2),
	.analysis-markdown :global(h3),
	.analysis-markdown :global(h4) {
		color: var(--color-text-primary);
		font-weight: 600;
		margin: 0.85em 0 0.4em;
		line-height: 1.3;
	}

	.analysis-markdown :global(h1) {
		font-size: 1.25rem;
	}

	.analysis-markdown :global(h2) {
		font-size: 1.1rem;
	}

	.analysis-markdown :global(h3) {
		font-size: 1rem;
	}

	.analysis-markdown :global(p),
	.analysis-markdown :global(ul),
	.analysis-markdown :global(ol) {
		margin: 0.5em 0;
	}

	.analysis-markdown :global(ul),
	.analysis-markdown :global(ol) {
		padding-left: 1.25rem;
	}

	.analysis-markdown :global(li) {
		margin: 0.2em 0;
	}

	.analysis-markdown :global(table) {
		border-collapse: collapse;
		width: 100%;
		margin: 0.75em 0;
		font-size: 0.875rem;
	}

	.analysis-markdown :global(th),
	.analysis-markdown :global(td) {
		border: 1px solid var(--color-border);
		padding: 0.35rem 0.5rem;
		text-align: left;
	}

	.analysis-markdown :global(th) {
		background: var(--color-bg-tertiary);
		font-weight: 600;
	}

	.analysis-markdown :global(a) {
		color: var(--color-accent-primary);
		text-decoration: underline;
	}

	.analysis-markdown :global(code) {
		font-family: var(--font-mono);
		font-size: 0.875em;
		background: var(--color-bg-primary);
		padding: 0.1em 0.35em;
		border-radius: 0.25rem;
	}

	.analysis-markdown :global(pre) {
		overflow-x: auto;
		padding: 0.75rem;
		background: var(--color-bg-primary);
		border-radius: 0.5rem;
		margin: 0.75em 0;
	}

	.analysis-markdown :global(pre code) {
		padding: 0;
		background: transparent;
	}
</style>
