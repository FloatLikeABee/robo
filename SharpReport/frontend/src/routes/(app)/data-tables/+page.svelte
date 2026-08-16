<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { listDataTables, type DataTableSummary } from '$lib/dataTables';
	import { whenSessionReady } from '$lib/stores/auth.svelte';
	import DataTablesImportButton from '$lib/components/data-tables/DataTablesImportButton.svelte';

	let tables = $state<DataTableSummary[]>([]);
	let loading = $state(true);
	let error = $state('');

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
			<p class="mt-1 max-w-2xl text-sm text-text-secondary">
				Imported CSV, Excel, JSON, and text files stored as editable tables.
			</p>
		</div>
		<DataTablesImportButton />
	</div>

	{#if error}
		<p class="text-sm text-danger">{error}</p>
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
				<button
					type="button"
					class="rounded-xl border border-border bg-bg-elevated p-4 text-left transition-colors hover:border-accent-primary/40 hover:bg-bg-tertiary"
					onclick={() => goto(`/data-tables/${table.id}`)}
				>
					<div class="font-semibold text-text-primary">{table.name}</div>
					<div class="mt-1 text-xs text-text-tertiary">
						{table.row_count} rows · {table.columns.length} cols
					</div>
					{#if table.source_filename}
						<div class="mt-2 truncate text-xs text-text-secondary">{table.source_filename}</div>
					{/if}
				</button>
			{/each}
		</div>
	{/if}
</div>
