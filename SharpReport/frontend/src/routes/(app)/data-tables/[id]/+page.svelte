<script lang="ts">
	import { page } from '$app/stores';
	import DataTablesImportButton from '$lib/components/data-tables/DataTablesImportButton.svelte';
	import {
		deleteDataTableRow,
		getDataTable,
		getDataTableRows,
		updateDataTableRow,
		type DataTableSummary
	} from '$lib/dataTables';
	import {
		assistantCtx,
		attachDataTable,
		requestAssistantOpen
	} from '$lib/stores/assistantContext.svelte';
	import { whenSessionReady } from '$lib/stores/auth.svelte';
	import {
		ArrowDown,
		ArrowUp,
		ArrowUpDown,
		ChevronLeft,
		ChevronRight,
		Globe,
		Search,
		Save,
		Sparkles,
		Trash2
	} from 'lucide-svelte';

	const PAGE_SIZE = 50;

	let table = $state<DataTableSummary | null>(null);
	let columns = $state<string[]>([]);
	let rows = $state<Record<string, unknown>[]>([]);
	let total = $state(0);
	let offset = $state(0);
	let loading = $state(true);
	let saving = $state(false);
	let deletingIndex = $state<number | null>(null);
	let error = $state('');
	let success = $state('');
	let editBuffer = $state<Record<string, Record<string, string>>>({});
	let searchQuery = $state('');
	let sortBy = $state('');
	let sortDir = $state<'asc' | 'desc'>('asc');
	let searchTimer: ReturnType<typeof setTimeout> | undefined;

	const tableId = $derived($page.params.id ?? '');
	const isAttached = $derived(
		table ? assistantCtx.attachedDataTables.some((t) => t.id === table.id) : false
	);

	function loadPage() {
		if (!tableId) return;
		loading = true;
		error = '';
		whenSessionReady(async () => {
			try {
				table = await getDataTable(tableId);
				const data = await getDataTableRows(tableId, {
					limit: PAGE_SIZE,
					offset,
					search: searchQuery,
					sortBy,
					sortDir
				});
				columns = data.columns;
				rows = data.rows;
				total = data.total;
				editBuffer = {};
				for (let i = 0; i < rows.length; i++) {
					const row = rows[i];
					const idx =
						typeof row._row_index === 'number' ? row._row_index : offset + i;
					const buf: Record<string, string> = {};
					for (const col of columns) {
						buf[col] = row[col] == null ? '' : String(row[col]);
					}
					editBuffer[String(idx)] = buf;
				}
			} catch (e) {
				error = e instanceof Error ? e.message : 'Failed to load table';
			} finally {
				loading = false;
			}
		});
	}

	$effect(() => {
		if (tableId) void loadPage();
	});

	function onSearchInput(value: string) {
		searchQuery = value;
		offset = 0;
		clearTimeout(searchTimer);
		searchTimer = setTimeout(() => loadPage(), 300);
	}

	function toggleSort(column: string) {
		if (sortBy === column) {
			sortDir = sortDir === 'asc' ? 'desc' : 'asc';
		} else {
			sortBy = column;
			sortDir = 'asc';
		}
		offset = 0;
		loadPage();
	}

	function rowKey(row: Record<string, unknown>, index: number): number {
		const v = row._row_index;
		return typeof v === 'number' ? v : offset + index;
	}

	async function saveRow(rowIndex: number) {
		if (!tableId) return;
		const buf = editBuffer[String(rowIndex)];
		if (!buf) return;
		saving = true;
		error = '';
		success = '';
		try {
			const data: Record<string, unknown> = {};
			for (const col of columns) {
				const raw = buf[col] ?? '';
				const n = Number(raw);
				if (raw !== '' && !Number.isNaN(n) && String(n) === raw.trim()) {
					data[col] = n;
				} else if (raw.toLowerCase() === 'true') {
					data[col] = true;
				} else if (raw.toLowerCase() === 'false') {
					data[col] = false;
				} else {
					data[col] = raw;
				}
			}
			await updateDataTableRow(tableId, rowIndex, data);
			success = `Saved row ${rowIndex + 1}`;
			await loadPage();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Save failed';
		} finally {
			saving = false;
		}
	}

	function prevPage() {
		offset = Math.max(0, offset - PAGE_SIZE);
		void loadPage();
	}

	function nextPage() {
		if (offset + PAGE_SIZE < total) {
			offset += PAGE_SIZE;
			void loadPage();
		}
	}

	async function deleteRow(rowIndex: number) {
		if (!tableId) return;
		if (!confirm(`Delete row ${rowIndex + 1}?`)) return;
		deletingIndex = rowIndex;
		error = '';
		success = '';
		try {
			await deleteDataTableRow(tableId, rowIndex);
			success = `Deleted row ${rowIndex + 1}`;
			if (rows.length === 1 && offset > 0) {
				offset = Math.max(0, offset - PAGE_SIZE);
			}
			loadPage();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Delete failed';
		} finally {
			deletingIndex = null;
		}
	}

	function addToAssistant() {
		if (!table) return;
		attachDataTable({ id: table.id, name: table.name });
		requestAssistantOpen();
		success = `Added "${table.name}" to AI conversation`;
	}
</script>

<div class="flex min-h-0 flex-1 flex-col gap-4">
	<div class="flex shrink-0 flex-wrap items-start justify-between gap-3">
		<div class="min-w-0">
			<a href="/data-tables" class="text-sm text-accent-primary hover:underline">← Data tables</a>
			<h1 class="mt-1 text-2xl font-bold text-text-primary">{table?.name ?? 'Data table'}</h1>
			{#if table}
				<p class="text-sm text-text-secondary">
					{table.row_count} rows · {table.columns.length} columns
					{#if table.source_filename}
						· from {table.source_filename}
					{/if}
				</p>
			{/if}
		</div>
		<div class="flex flex-wrap items-center gap-2">
			<a
				href="/data-tables/{tableId}/publish"
				class="inline-flex items-center gap-1.5 rounded-lg border border-border px-3 py-1.5 text-sm font-medium text-text-secondary transition-colors hover:bg-bg-tertiary hover:text-text-primary"
				title="AI builds a public web page with charts and analytics"
			>
				<Globe class="h-4 w-4" />
				Build & publish page
			</a>
			<button
				type="button"
				class="inline-flex items-center gap-1.5 rounded-lg border px-3 py-1.5 text-sm font-medium transition-colors {isAttached
					? 'border-accent-primary/50 bg-accent-primary/10 text-accent-primary'
					: 'border-border text-text-secondary hover:bg-bg-tertiary hover:text-text-primary'}"
				disabled={!table}
				title="Add this table to the AI assistant conversation"
				onclick={addToAssistant}
			>
				<Sparkles class="h-4 w-4" />
				{isAttached ? 'In AI chat' : 'Add to AI'}
			</button>
			<DataTablesImportButton label="Import data" />
			<div class="flex items-center gap-2 text-sm text-text-secondary">
				<button
					type="button"
					class="rounded-lg border border-border px-2 py-1 hover:bg-bg-tertiary disabled:opacity-40"
					disabled={offset <= 0 || loading}
					onclick={prevPage}
				>
					<ChevronLeft class="inline h-4 w-4" />
				</button>
				<span>
					{total === 0 ? 0 : offset + 1}–{Math.min(offset + PAGE_SIZE, total)} of {total}
				</span>
				<button
					type="button"
					class="rounded-lg border border-border px-2 py-1 hover:bg-bg-tertiary disabled:opacity-40"
					disabled={offset + PAGE_SIZE >= total || loading}
					onclick={nextPage}
				>
					<ChevronRight class="inline h-4 w-4" />
				</button>
			</div>
		</div>
	</div>

	<div class="relative max-w-md">
		<Search class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-text-tertiary" />
		<input
			type="search"
			class="w-full rounded-lg border border-border bg-bg-primary py-2 pl-9 pr-3 text-sm text-text-primary placeholder:text-text-tertiary"
			placeholder="Search all columns…"
			value={searchQuery}
			oninput={(e) => onSearchInput((e.currentTarget as HTMLInputElement).value)}
		/>
	</div>

	{#if error}
		<p class="text-sm text-danger">{error}</p>
	{/if}
	{#if success}
		<p class="text-sm text-accent-primary">{success}</p>
	{/if}

	{#if loading}
		<p class="text-sm text-text-secondary">Loading rows…</p>
	{:else if columns.length === 0}
		<p class="text-sm text-text-secondary">This table has no columns.</p>
	{:else}
		<div class="min-h-0 flex-1 overflow-auto rounded-xl border border-border">
			<table class="min-w-full border-separate border-spacing-0 text-left text-sm">
				<thead class="text-text-secondary">
					<tr>
						<th
							class="sticky left-0 top-0 z-30 min-w-[3rem] border-b border-r border-border bg-bg-elevated px-3 py-2 font-medium shadow-[4px_0_8px_-4px_rgba(0,0,0,0.35)]"
						>
							#
						</th>
						{#each columns as col}
							<th
								class="sticky top-0 z-20 min-w-[8rem] border-b border-border bg-bg-elevated px-3 py-2 font-medium"
							>
								<button
									type="button"
									class="inline-flex items-center gap-1 hover:text-text-primary"
									onclick={() => toggleSort(col)}
								>
									{col}
									{#if sortBy === col}
										{#if sortDir === 'asc'}
											<ArrowUp class="h-3.5 w-3.5 text-accent-primary" />
										{:else}
											<ArrowDown class="h-3.5 w-3.5 text-accent-primary" />
										{/if}
									{:else}
										<ArrowUpDown class="h-3.5 w-3.5 opacity-40" />
									{/if}
								</button>
							</th>
						{/each}
						<th
							class="sticky right-0 top-0 z-30 min-w-[5.5rem] border-b border-l border-border bg-bg-elevated px-3 py-2 font-medium shadow-[-4px_0_8px_-4px_rgba(0,0,0,0.35)]"
						>
							Actions
						</th>
					</tr>
				</thead>
				<tbody>
					{#each rows as row, i (offset + i)}
						{@const idx = rowKey(row, i)}
						<tr class="group border-b border-border/60 hover:bg-bg-tertiary/40">
							<td
								class="sticky left-0 z-10 border-r border-border bg-bg-primary px-3 py-2 text-text-tertiary shadow-[4px_0_8px_-4px_rgba(0,0,0,0.2)] group-hover:bg-bg-tertiary/40"
							>
								{idx + 1}
							</td>
							{#each columns as col}
								<td class="bg-bg-primary px-3 py-1 group-hover:bg-bg-tertiary/40">
									<input
										class="w-full min-w-[8rem] rounded-md border border-border bg-bg-primary px-2 py-1 text-text-primary"
										bind:value={editBuffer[String(idx)][col]}
									/>
								</td>
							{/each}
							<td
								class="sticky right-0 z-10 border-l border-border bg-bg-primary px-3 py-2 shadow-[-4px_0_8px_-4px_rgba(0,0,0,0.2)] group-hover:bg-bg-tertiary/40"
							>
								<div class="flex items-center gap-1.5">
									<button
										type="button"
										class="inline-flex h-8 w-8 items-center justify-center rounded-md bg-accent-primary text-white transition-opacity hover:opacity-90 disabled:opacity-45"
										disabled={saving || deletingIndex !== null}
										aria-label="Save row {idx + 1}"
										title="Save row"
										onclick={() => saveRow(idx)}
									>
										<Save class="h-4 w-4" />
									</button>
									<button
										type="button"
										class="inline-flex h-8 w-8 items-center justify-center rounded-md bg-error text-white transition-opacity hover:opacity-90 disabled:opacity-45"
										disabled={saving || deletingIndex !== null}
										aria-label="Delete row {idx + 1}"
										title="Delete row"
										onclick={() => deleteRow(idx)}
									>
										<Trash2 class="h-4 w-4" />
									</button>
								</div>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>
