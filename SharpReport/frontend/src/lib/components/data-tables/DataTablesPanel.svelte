<script lang="ts">
	import { goto } from '$app/navigation';
	import {
		deleteDataTable,
		importDataTable,
		listDataTables,
		processUploadFile,
		type DataTableSummary
	} from '$lib/dataTables';
	import { whenSessionReady } from '$lib/stores/auth.svelte';
	import { ChevronDown, ChevronUp, Plus, Table2, Trash2, X } from 'lucide-svelte';
	import { onMount } from 'svelte';

	let {
		open = $bindable(false),
		onchange,
		anchor = 'fab'
	}: {
		open?: boolean;
		onchange?: () => void;
		anchor?: 'fab' | 'top-right';
	} = $props();

	const panelPositionClass = $derived(
		anchor === 'top-right'
			? 'fixed top-[4.5rem] right-6 z-[320]'
			: 'fixed bottom-24 right-6 z-[320]'
	);

	let tables = $state<DataTableSummary[]>([]);
	let loading = $state(true);
	let uploading = $state(false);
	let error = $state('');
	let success = $state('');
	let collapsed = $state(false);
	let indexToGraph = $state(false);
	let fileInput: HTMLInputElement | undefined = $state();

	function refresh() {
		loading = true;
		error = '';
		whenSessionReady(async () => {
			try {
				tables = await listDataTables();
				onchange?.();
			} catch (e) {
				error = e instanceof Error ? e.message : 'Failed to load data tables';
				tables = [];
			} finally {
				loading = false;
			}
		});
	}

	onMount(() => {
		void refresh();
	});

	$effect(() => {
		if (open) void refresh();
	});

	async function handleFile(e: Event) {
		const input = e.target as HTMLInputElement;
		const file = input.files?.[0];
		input.value = '';
		if (!file) return;

		const lower = file.name.toLowerCase();
		if (lower.endsWith('.pdf') || lower.endsWith('.md') || lower.endsWith('.markdown')) {
			error =
				'PDF/Markdown are Docs content files. Open Docs in the nav to upload them. Data tables accept CSV, TSV, JSON, or Excel.';
			return;
		}

		uploading = true;
		error = '';
		success = '';
		try {
			const parsed = await processUploadFile(file);
			if (parsed.rows.length === 0) {
				throw new Error('No rows found in file.');
			}
			const baseName = file.name.replace(/\.[^.]+$/, '');
			const created = await importDataTable({
				name: baseName,
				source_filename: file.name,
				source_format: parsed.source_format,
				columns: parsed.columns,
				rows: parsed.rows,
				index_to_graph: indexToGraph
			});
			let msg = parsed.reason
				? `Imported "${created.name}" — ${parsed.reason}`
				: `Imported "${created.name}" (${created.row_count} rows)`;
			if (created.graph_status) {
				msg += ` · ${created.graph_status}`;
			}
			success = msg;
			await refresh();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Import failed';
		} finally {
			uploading = false;
		}
	}

	async function removeTable(id: string, name: string) {
		if (!confirm(`Delete data table "${name}"?`)) return;
		try {
			await deleteDataTable(id);
			await refresh();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Delete failed';
		}
	}

	function openTable(id: string) {
		open = false;
		void goto(`/data-tables/${id}`);
	}
</script>

{#if open}
	<div
		class="{panelPositionClass} flex w-[min(420px,calc(100vw-2rem))] flex-col overflow-hidden rounded-2xl border border-border bg-bg-elevated shadow-2xl"
		role="dialog"
		aria-label="Data tables"
	>
		<header class="flex items-center justify-between gap-2 border-b border-border px-4 py-3">
			<div class="flex min-w-0 items-center gap-2">
				<Table2 class="h-5 w-5 shrink-0 text-accent-primary" />
				<h2 class="truncate text-sm font-semibold text-text-primary">Data tables</h2>
			</div>
			<div class="flex items-center gap-1">
				<button
					type="button"
					class="rounded-md p-1.5 text-text-secondary hover:bg-bg-tertiary hover:text-text-primary"
					aria-label={collapsed ? 'Expand list' : 'Collapse list'}
					onclick={() => (collapsed = !collapsed)}
				>
					{#if collapsed}
						<ChevronUp class="h-4 w-4" />
					{:else}
						<ChevronDown class="h-4 w-4" />
					{/if}
				</button>
				<button
					type="button"
					class="rounded-md p-1.5 text-text-secondary hover:bg-bg-tertiary hover:text-text-primary"
					aria-label="Close"
					onclick={() => (open = false)}
				>
					<X class="h-4 w-4" />
				</button>
			</div>
		</header>

		{#if !collapsed}
			<div class="border-b border-border px-4 py-3">
				<input
					bind:this={fileInput}
					type="file"
					accept=".csv,.tsv,.json,.xlsx,.xls"
					class="hidden"
					onchange={handleFile}
				/>
				<button
					type="button"
					class="flex w-full items-center justify-center gap-2 rounded-xl bg-accent-primary px-3 py-2 text-sm font-medium text-white hover:opacity-95 disabled:opacity-50"
					disabled={uploading}
					onclick={() => fileInput?.click()}
				>
					<Plus class="h-4 w-4" />
					{uploading ? 'Processing…' : 'Upload CSV, Excel, JSON, TXT, MD, PDF'}
				</button>
				<label class="mt-2 flex cursor-pointer items-start gap-2 text-[11px] leading-relaxed text-text-secondary">
					<input
						type="checkbox"
						class="mt-0.5"
						bind:checked={indexToGraph}
						disabled={uploading}
					/>
					<span>Also save into Neo4j graph (Morph Knowledge / faster GraphRAG)</span>
				</label>
				<p class="mt-2 text-[11px] leading-relaxed text-text-tertiary">
					Structured files import directly. Text formats are analyzed by AI when needed.
				</p>
			</div>

			<div class="max-h-[min(50vh,360px)] overflow-y-auto p-2">
				{#if loading}
					<p class="px-2 py-4 text-sm text-text-secondary">Loading…</p>
				{:else if tables.length === 0}
					<p class="px-2 py-4 text-sm text-text-secondary">No data tables yet. Upload a file to create one.</p>
				{:else}
					<ul class="space-y-1">
						{#each tables as table (table.id)}
							<li
								class="group flex items-center gap-2 rounded-lg border border-transparent px-2 py-2 hover:border-border hover:bg-bg-tertiary"
							>
								<button
									type="button"
									class="min-w-0 flex-1 text-left"
									onclick={() => openTable(table.id)}
								>
									<div class="truncate text-sm font-medium text-text-primary">{table.name}</div>
									<div class="truncate text-xs text-text-tertiary">
										{table.row_count} rows · {table.columns.length} columns
										{#if table.source_filename}
											· {table.source_filename}
										{/if}
									</div>
								</button>
								<button
									type="button"
									class="rounded p-1.5 text-text-tertiary opacity-0 transition-opacity hover:bg-danger/10 hover:text-danger group-hover:opacity-100"
									aria-label="Delete {table.name}"
									onclick={() => removeTable(table.id, table.name)}
								>
									<Trash2 class="h-4 w-4" />
								</button>
							</li>
						{/each}
					</ul>
				{/if}
			</div>

			{#if error}
				<p class="border-t border-border px-4 py-2 text-xs text-danger">{error}</p>
			{/if}
			{#if success}
				<p class="border-t border-border px-4 py-2 text-xs text-accent-primary">{success}</p>
			{/if}
		{/if}
	</div>
{/if}
