<script lang="ts">
	import { onMount } from 'svelte';
	import { listDataTables, type DataTableSummary } from '$lib/dataTables';
	import { whenSessionReady } from '$lib/stores/auth.svelte';
	import DataTablesImportButton from '$lib/components/data-tables/DataTablesImportButton.svelte';
	import DataTableAiAnalysisModal from '$lib/components/data-tables/DataTableAiAnalysisModal.svelte';
	import { Sparkles } from 'lucide-svelte';

	let tables = $state<DataTableSummary[]>([]);
	let loading = $state(true);
	let error = $state('');
	let analysisTable = $state<DataTableSummary | null>(null);

	onMount(() =>
		whenSessionReady(async () => {
			try {
				tables = await listDataTables();
			} catch (e) {
				error = e instanceof Error ? e.message : 'Failed to load';
			} finally {
				loading = false;
			}
		})
	);
</script>

<div class="flex min-h-0 flex-1 flex-col gap-6">
	<div class="flex shrink-0 flex-wrap items-start justify-between gap-3">
		<div class="min-w-0">
			<h1 class="text-2xl font-bold text-text-primary">Data tables</h1>
		</div>
		<DataTablesImportButton />
	</div>

	{#if error}
		<p class="text-sm text-error">{error}</p>
	{/if}

	{#if loading}
		<p class="text-sm text-text-secondary">Loading…</p>
	{:else if tables.length === 0}
		<div class="rounded-xl border border-dashed border-border bg-bg-elevated p-8 text-center text-sm text-text-secondary">
			No data tables yet. Use <strong>Import data</strong> (top right) to upload a file.
		</div>
	{:else}
		<div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
			{#each tables as table (table.id)}
				<div
					class="flex flex-col overflow-hidden rounded-xl border border-border bg-bg-elevated transition-colors hover:border-accent-primary/40"
				>
					<a href="/data-tables/{table.id}" class="min-w-0 flex-1 p-4 text-left hover:bg-bg-tertiary">
						<div class="font-semibold text-text-primary">{table.name}</div>
						<div class="mt-1 text-xs text-text-tertiary">
							{table.row_count} rows · {table.columns.length} cols
						</div>
						{#if table.source_filename}
							<div class="mt-2 truncate text-xs text-text-secondary">{table.source_filename}</div>
						{/if}
					</a>
					<div class="border-t border-border px-4 py-2">
						<button
							type="button"
							class="inline-flex items-center gap-1.5 text-sm font-medium text-accent-primary hover:underline"
							onclick={() => (analysisTable = table)}
						>
							<Sparkles class="h-3.5 w-3.5" />
							AI analysis
						</button>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

{#if analysisTable}
	<DataTableAiAnalysisModal
		tableId={analysisTable.id}
		tableName={analysisTable.name}
		onClose={() => (analysisTable = null)}
	/>
{/if}
