<svelte:options runes={false} />

<script lang="ts">
	import { onMount, tick } from 'svelte';
	import { whenSessionReady } from '$lib/stores/auth.svelte';
	import { apiUrl, authHeaders } from '$lib/api';
	import { Table, Trash2 } from 'lucide-svelte';
	import ReportChart from '$lib/components/report/ReportChart.svelte';
	import * as FileData from '$lib/report/fileData';

	type DbOption = { id: string; name: string; engine: string };

	type SchemaColumn = { name: string; data_type: string };
	type SchemaTable = { schema: string; name: string; columns: SchemaColumn[] };
	type SchemaResponse = { database_id: string; engine: string; tables: SchemaTable[] };

	type FieldRow = {
		id: string;
		schema: string;
		table: string;
		column: string;
		alias: string;
		aggregation: '' | 'sum' | 'avg' | 'count' | 'min' | 'max';
	};

	type FilterRow = {
		id: string;
		column: string;
		op: string;
		value: string;
	};

	const aggOptions: { value: FieldRow['aggregation']; label: string }[] = [
		{ value: '', label: 'None (detail)' },
		{ value: 'sum', label: 'Sum' },
		{ value: 'avg', label: 'Average' },
		{ value: 'count', label: 'Count' },
		{ value: 'min', label: 'Min' },
		{ value: 'max', label: 'Max' }
	];

	const filterOps = [
		{ value: 'eq', label: 'Equals' },
		{ value: 'ne', label: 'Not equals' },
		{ value: 'gt', label: '>' },
		{ value: 'gte', label: '>=' },
		{ value: 'lt', label: '<' },
		{ value: 'lte', label: '<=' },
		{ value: 'contains', label: 'Contains' },
		{ value: 'starts_with', label: 'Starts with' },
		{ value: 'isnotnull', label: 'Is not null' }
	];

	let databases: DbOption[] = [];
	let selectedDb = '';
	let schemaLoading = false;
	let schemaError = '';
	let schema: SchemaResponse | null = null;

	let expandedTables = new Set<string>();
	/** `schema\\0name` keys for tables shown in the panel. Use array + reassignment for reliable updates. */
	let schemaSelectedTableKeys: string[] = [];
	let fieldSearch = '';

	let fields: FieldRow[] = [];
	let filters: FilterRow[] = [];
	let sortColumnIndex: number | null = null;
	let sortDir: 'asc' | 'desc' = 'asc';
	let limit = 500;

	let runLoading = false;
	let runError = '';
	let resultColumns: string[] = [];
	let resultRows: Record<string, unknown>[] = [];
	let runMeta = '';

	let chartType: 'none' | 'bar' | 'pie' = 'none';
	let chartCategoryKey = '';
	let chartValueKey = '';

	let dragIndex: number | null = null;

	/** Full-screen results layer so grids and charts are easy to read. */
	let resultsModalOpen = false;

	let tablesFieldsDrawerOpen = false;

	function closeResultsModal() {
		resultsModalOpen = false;
	}

	function openLastResults() {
		if (resultColumns.length === 0 && resultRows.length === 0 && !runMeta) return;
		resultsModalOpen = true;
	}

	type BuilderMode = 'visual' | 'sql' | 'file';
	let builderMode: BuilderMode = 'file';

	let sqlText = 'SELECT 1 AS one';

	let fileStore: Record<string, unknown>[] = [];
	let fileName = '';
	let filePivot = false;
	let fileGroupCol = '';
	let fileSumCol = '';
	let fileSortCol = '';
	let fileSortDir: 'asc' | 'desc' = 'asc';
	let fileFilters: { id: string; column: string; op: string; value: string }[] = [];
	/** Column names to include in the file report (non-pivot). Default: all after load. */
	let fileSelectedColumns: string[] = [];

	$: fileColumnsList = fileStore.length ? FileData.columnKeys(fileStore) : [];

	async function runSqlReport() {
		runError = '';
		runMeta = '';
		resultRows = [];
		resultColumns = [];
		if (!selectedDb) {
			runError = 'Select a database connection.';
			return;
		}
		if (!sqlText.trim()) {
			runError = 'Enter SQL.';
			return;
		}
		resultsModalOpen = true;
		await tick();
		runLoading = true;
		try {
			const response = await fetch(apiUrl('/api/v1/queries/execute'), {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					...authHeaders()
				},
				body: JSON.stringify({ database_id: selectedDb, sql: sqlText })
			});
			if (!response.ok) {
				const t = await response.text();
				runError = t || 'Execute failed';
				return;
			}
			const data = await response.json();
			const rows = (data.results as unknown[]) ?? [];
			const parsed: Record<string, unknown>[] = rows.map((r) => {
				if (r && typeof r === 'object' && !Array.isArray(r)) return r as Record<string, unknown>;
				return { value: r };
			});
			resultRows = parsed;
			resultColumns = parsed.length > 0 ? Object.keys(parsed[0]) : [];
			const n = (data.row_count as number) ?? parsed.length;
			runMeta = `${n} row${n === 1 ? '' : 's'} (SQL)`;
			const g = guessChartKeys(resultColumns, resultRows);
			chartCategoryKey = g.cat;
			chartValueKey = g.val;
		} catch (err) {
			runError = err instanceof Error ? err.message : 'Execute failed';
		} finally {
			runLoading = false;
		}
	}

	async function handleFileInput(e: Event) {
		const inp = e.target as HTMLInputElement;
		const file = inp.files?.[0];
		runError = '';
		if (!file) return;
		fileName = file.name;
		const buf = await file.arrayBuffer();
		const name = file.name.toLowerCase();
		try {
			if (name.endsWith('.csv') || name.endsWith('.tsv')) {
				const text = new TextDecoder().decode(buf);
				fileStore = FileData.parseCsv(text);
			} else if (name.endsWith('.json')) {
				const text = new TextDecoder().decode(buf);
				fileStore = FileData.parseJsonArray(text);
			} else if (name.endsWith('.xlsx') || name.endsWith('.xls')) {
				fileStore = FileData.parseExcelBuffer(buf);
			} else {
				runError = 'Supported: .csv, .tsv, .json, .xlsx';
				fileStore = [];
				return;
			}
			const cols = FileData.columnKeys(fileStore);
			fileSelectedColumns = [...cols];
			if (cols.length) {
				fileSortCol = cols[0];
				fileGroupCol = cols[0];
				fileSumCol = cols.length > 1 ? cols[1] : cols[0];
			}
			fileFilters = [];
		} catch (err) {
			runError = err instanceof Error ? err.message : 'Failed to read file';
			fileStore = [];
		}
	}

	function addFileLineFilter() {
		const cols = fileColumnsList;
		if (cols.length === 0) return;
		fileFilters = [
			...fileFilters,
			{ id: crypto.randomUUID(), column: cols[0], op: 'eq', value: '' }
		];
	}

	function removeFileLineFilter(id: string) {
		fileFilters = fileFilters.filter((f) => f.id !== id);
	}

	function selectAllSchemaTables() {
		if (!schema) return;
		schemaSelectedTableKeys = schema.tables.map((t) => tableKey(t));
	}

	function clearSchemaSelection() {
		schemaSelectedTableKeys = [];
	}

	function toggleFileColumn(col: string) {
		if (fileSelectedColumns.includes(col)) {
			fileSelectedColumns = fileSelectedColumns.filter((c) => c !== col);
		} else {
			fileSelectedColumns = [...fileSelectedColumns, col];
		}
	}

	function selectAllFileColumns() {
		fileSelectedColumns = [...fileColumnsList];
	}

	function selectNoFileColumns() {
		fileSelectedColumns = [];
	}

	function projectFileRows(
		rows: Record<string, unknown>[],
		cols: string[]
	): Record<string, unknown>[] {
		return rows.map((r) => {
			const o: Record<string, unknown> = {};
			for (const c of cols) {
				if (Object.prototype.hasOwnProperty.call(r, c)) o[c] = r[c];
			}
			return o;
		});
	}

	function runFileReport() {
		if (fileStore.length === 0) {
			runError = 'Import a file first.';
			return;
		}
		const colOrder = fileColumnsList.filter((c) => fileSelectedColumns.includes(c));
		if (!filePivot && colOrder.length === 0) {
			runError = 'Select at least one column to include in the report.';
			return;
		}
		runError = '';
		let rows = [...fileStore];
		const fl: FileData.FileFilter[] = fileFilters.map((f) => ({
			column: f.column,
			op: f.op,
			value: f.value
		}));
		rows = FileData.applyFilters(rows, fl);
		if (fileSortCol) rows = FileData.sortRows(rows, fileSortCol, fileSortDir);
		if (filePivot && fileGroupCol && fileSumCol) {
			rows = FileData.groupBySum(rows, fileGroupCol, fileSumCol);
		} else {
			rows = projectFileRows(rows, colOrder);
		}
		resultRows = rows;
		resultColumns = rows.length ? FileData.columnKeys(rows) : [];
		runMeta = `${rows.length} row${rows.length === 1 ? '' : 's'} (file)`;
		const g = guessChartKeys(resultColumns, resultRows);
		chartCategoryKey = g.cat;
		chartValueKey = g.val;
		chartType = 'none';
		resultsModalOpen = true;
	}

	function toggleTable(key: string) {
		const next = new Set(expandedTables);
		if (next.has(key)) next.delete(key);
		else next.add(key);
		expandedTables = next;
	}

	/** Stable, HTML-safe key (avoid `\\0` — null breaks `value=""` in the DOM). */
	function tableKey(t: SchemaTable) {
		return JSON.stringify([t.schema, t.name]);
	}

	function columnsMatchingSearch(t: SchemaTable): SchemaColumn[] {
		const q = fieldSearch.trim().toLowerCase();
		if (!q) return t.columns;
		const tn = `${t.schema}.${t.name}`.toLowerCase();
		if (tn.includes(q)) return t.columns;
		return t.columns.filter((c) => c.name.toLowerCase().includes(q));
	}

	/** Must read `schemaSelectedTableKeys` here (not inside a helper) so Svelte tracks the dependency. */
	$: panelTables =
		schema == null
			? []
			: schema.tables.filter((t) => schemaSelectedTableKeys.includes(tableKey(t)));

	function addField(t: SchemaTable, col: SchemaColumn) {
		if (fields.length > 0) {
			const f0 = fields[0];
			if (f0.schema !== t.schema || f0.table !== t.name) {
				runError =
					'This report uses one table at a time. Clear fields or remove columns to pick another table.';
				return;
			}
		}
		runError = '';
		const dup = fields.some(
			(f) => f.schema === t.schema && f.table === t.name && f.column === col.name
		);
		if (dup) return;
		fields = [
			...fields,
			{
				id: crypto.randomUUID(),
				schema: t.schema,
				table: t.name,
				column: col.name,
				alias: '',
				aggregation: ''
			}
		];
	}

	function removeField(id: string) {
		fields = fields.filter((f) => f.id !== id);
		if (sortColumnIndex !== null && sortColumnIndex >= fields.length) {
			sortColumnIndex = null;
		}
	}

	function moveField(from: number, to: number) {
		if (from === to || from < 0 || to < 0 || from >= fields.length || to >= fields.length) return;
		const next = [...fields];
		const [m] = next.splice(from, 1);
		next.splice(to, 0, m);
		fields = next;
	}

	function addFilter() {
		if (fields.length === 0) {
			runError = 'Add at least one column before adding filters.';
			return;
		}
		runError = '';
		const cols = fields.map((f) => f.column);
		filters = [
			...filters,
			{
				id: crypto.randomUUID(),
				column: cols[0] ?? '',
				op: 'eq',
				value: ''
			}
		];
	}

	function removeFilter(id: string) {
		filters = filters.filter((f) => f.id !== id);
	}

	function guessChartKeys(cols: string[], rows: Record<string, unknown>[]) {
		if (!cols.length || !rows.length) return { cat: '', val: '' };
		let valKey = '';
		for (const c of cols) {
			const v = rows[0][c];
			if (typeof v === 'number') {
				valKey = c;
				break;
			}
			if (typeof v === 'string' && v !== '' && !Number.isNaN(Number(v))) {
				valKey = c;
				break;
			}
		}
		const catKey = cols.find((c) => c !== valKey) ?? cols[0];
		if (!valKey) valKey = cols[1] ?? cols[0];
		return { cat: catKey, val: valKey };
	}

	async function fetchDatabases() {
		try {
			const response = await fetch(apiUrl('/api/v1/databases'), {
				headers: authHeaders()
			});
			if (!response.ok) return;
			databases = await response.json();
			if (!selectedDb && databases.length > 0) selectedDb = databases[0].id;
		} catch {
			/* ignore */
		}
	}

	async function loadSchema() {
		schema = null;
		schemaError = '';
		runError = '';
		fields = [];
		filters = [];
		schemaSelectedTableKeys = [];
		resultRows = [];
		resultColumns = [];
		runMeta = '';
		if (!selectedDb) return;
		schemaLoading = true;
		try {
			const response = await fetch(apiUrl(`/api/v1/databases/${selectedDb}/schema`), {
				headers: authHeaders()
			});
			if (!response.ok) {
				const t = await response.text();
				schemaError = t || 'Failed to load schema';
				return;
			}
			const loaded: SchemaResponse = await response.json();
			schema = loaded;
			if (loaded.tables.length > 0) {
				const next = new Set(expandedTables);
				for (const t of loaded.tables.slice(0, 8)) {
					next.add(tableKey(t));
				}
				expandedTables = next;
			}
		} catch (err) {
			schemaError = err instanceof Error ? err.message : 'Schema request failed';
		} finally {
			schemaLoading = false;
		}
	}

	async function runReport() {
		runError = '';
		runMeta = '';
		resultRows = [];
		resultColumns = [];
		if (!selectedDb) {
			runError = 'Select a database connection.';
			return;
		}
		if (fields.length === 0) {
			runError = 'Add at least one field from the schema tree.';
			return;
		}

		resultsModalOpen = true;
		await tick();

		const f0 = fields[0];
		const body = {
			database_id: selectedDb,
			columns: fields.map((f) => ({
				schema: f.schema,
				table: f.table,
				column: f.column,
				alias: f.alias.trim() || null,
				aggregation: f.aggregation || null
			})),
			filters: filters.map((fl) => ({
				schema: f0.schema,
				table: f0.table,
				column: fl.column,
				op: fl.op,
				value: fl.value
			})),
			order_by:
				sortColumnIndex !== null && sortColumnIndex < fields.length
					? [{ column_index: sortColumnIndex, dir: sortDir }]
					: undefined,
			limit
		};

		runLoading = true;
		try {
			const response = await fetch(apiUrl('/api/v1/reports/builder/execute'), {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					...authHeaders()
				},
				body: JSON.stringify(body)
			});
			if (!response.ok) {
				const t = await response.text();
				runError = t || 'Execute failed';
				return;
			}
			const data = await response.json();
			const cols = (data.columns as string[]) ?? [];
			const rows = (data.rows as Record<string, unknown>[]) ?? [];
			resultColumns = cols;
			resultRows = rows;
			runMeta = `${data.row_count ?? rows.length} row${(data.row_count ?? rows.length) === 1 ? '' : 's'}`;
			const g = guessChartKeys(cols, rows);
			chartCategoryKey = g.cat;
			chartValueKey = g.val;
		} catch (err) {
			runError = err instanceof Error ? err.message : 'Execute failed';
		} finally {
			runLoading = false;
		}
	}

	function handleGlobalKeydown(e: KeyboardEvent) {
		if (e.key !== 'Escape') return;
		if (resultsModalOpen) {
			e.preventDefault();
			closeResultsModal();
			return;
		}
		if (tablesFieldsDrawerOpen) {
			e.preventDefault();
			tablesFieldsDrawerOpen = false;
		}
	}

	function onDragStart(i: number) {
		dragIndex = i;
	}

	function onDragOver(e: DragEvent) {
		e.preventDefault();
	}

	function onDrop(i: number) {
		if (dragIndex === null) return;
		moveField(dragIndex, i);
		dragIndex = null;
	}

	onMount(() =>
		whenSessionReady(async () => {
			/* Databases / Visual / SQL modes removed — file reports only. */
		})
	);

	$: selectedDb, loadSchema();

	$: if (builderMode !== 'visual') tablesFieldsDrawerOpen = false;
</script>

<div class="flex min-h-0 flex-1 flex-col gap-4">
	<div class="shrink-0">
		<h1 class="text-lg font-semibold text-text-primary">Data reports</h1>
		<p class="text-sm text-text-secondary">Upload a spreadsheet or JSON file and shape it into a report.</p>
	</div>

		<div class="flex min-h-0 flex-1 flex-col gap-4 rounded-lg border border-border bg-bg-elevated p-4">
			<h2 class="text-sm font-semibold text-text-primary">Import &amp; shape</h2>
			<p class="text-xs text-text-secondary">
				Parsing runs in your browser. Apply filters, sort, then optionally group by one column and sum another. Results
				open in a results overlay when you run the report.
			</p>
			{#if runError && !resultsModalOpen}
				<div class="rounded bg-error/10 p-3 text-sm text-error">{runError}</div>
			{/if}
			<div class="flex flex-col gap-2">
				<label for="rb-file" class="text-xs font-medium text-text-secondary">File (.csv, .tsv, .json, .xlsx)</label>
				<input
					id="rb-file"
					type="file"
					accept=".csv,.tsv,.json,.xlsx,.xls"
					on:change={handleFileInput}
					class="text-sm text-text-secondary file:mr-3 file:rounded-md file:border-0 file:bg-accent-primary file:px-3 file:py-1.5 file:text-sm file:text-white"
				/>
				{#if fileName}
					<p class="text-xs text-text-tertiary">Loaded: {fileName} ({fileStore.length} rows)</p>
				{/if}
			</div>

			{#if fileStore.length > 0}
				<div class="rounded border border-border p-3">
					<div class="mb-2 flex flex-wrap items-center justify-between gap-2">
						<h3 class="text-xs font-semibold uppercase text-text-tertiary">Columns in report</h3>
						<div class="flex gap-2">
							<button
								type="button"
								class="text-xs text-accent-primary hover:underline"
								on:click={selectAllFileColumns}
							>
								All
							</button>
							<button
								type="button"
								class="text-xs text-accent-primary hover:underline"
								on:click={selectNoFileColumns}
							>
								None
							</button>
						</div>
					</div>
					{#if filePivot}
						<p class="mb-2 text-xs text-text-tertiary">
							With group by + sum, the result uses the group and sum columns. Column selection below applies when
							group by + sum is off.
						</p>
					{:else}
						<p class="mb-2 text-xs text-text-tertiary">Uncheck to omit columns from the table output.</p>
					{/if}
					<div class="flex max-h-36 flex-wrap gap-x-4 gap-y-2 overflow-y-auto">
						{#each fileColumnsList as c}
							<label class="flex max-w-full cursor-pointer items-center gap-2 text-sm text-text-primary">
								<input
									type="checkbox"
									checked={fileSelectedColumns.includes(c)}
									on:change={() => toggleFileColumn(c)}
								/>
								<span class="truncate" title={c}>{c}</span>
							</label>
						{/each}
					</div>
				</div>

				<div class="flex flex-wrap items-center justify-between gap-2">
					<h3 class="text-xs font-semibold uppercase text-text-tertiary">Filters</h3>
					<button
						type="button"
						class="rounded border border-border bg-bg-secondary px-2 py-1 text-xs"
						on:click={addFileLineFilter}
					>
						Add filter
					</button>
				</div>
				{#each fileFilters as ff (ff.id)}
					<div class="flex flex-wrap items-end gap-2">
						<div class="min-w-[8rem]">
							<label for="ffcol-{ff.id}" class="text-xs text-text-tertiary">Column</label>
							<select
								id="ffcol-{ff.id}"
								bind:value={ff.column}
								class="mt-0.5 w-full rounded-md border border-border bg-bg-secondary px-2 py-1.5 text-sm"
							>
								{#each fileColumnsList as c}
									<option value={c}>{c}</option>
								{/each}
							</select>
						</div>
						<div class="min-w-[8rem]">
							<label for="ffop-{ff.id}" class="text-xs text-text-tertiary">Operator</label>
							<select
								id="ffop-{ff.id}"
								bind:value={ff.op}
								class="mt-0.5 w-full rounded-md border border-border bg-bg-secondary px-2 py-1.5 text-sm"
							>
								{#each filterOps as op}
									<option value={op.value}>{op.label}</option>
								{/each}
							</select>
						</div>
						{#if ff.op !== 'isnotnull'}
							<div class="min-w-[10rem] flex-1">
								<label for="ffval-{ff.id}" class="text-xs text-text-tertiary">Value</label>
								<input
									id="ffval-{ff.id}"
									bind:value={ff.value}
									class="mt-0.5 w-full rounded-md border border-border bg-bg-secondary px-2 py-1.5 text-sm"
								/>
							</div>
						{:else}
							<div class="flex min-w-[10rem] flex-1 flex-col justify-end pb-2 pt-5">
								<span class="text-xs text-text-tertiary">—</span>
							</div>
						{/if}
						<button
							type="button"
							class="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-md border border-border text-error hover:bg-error/10"
							aria-label="Remove filter"
							title="Remove"
							on:click={() => removeFileLineFilter(ff.id)}
						>
							<Trash2 class="h-4 w-4" />
						</button>
					</div>
				{/each}

				<div class="flex flex-wrap gap-4">
					<div class="flex flex-col gap-1">
						<label for="rb-file-sort" class="text-xs text-text-secondary">Sort by</label>
						<select
							id="rb-file-sort"
							bind:value={fileSortCol}
							class="rounded-md border border-border bg-bg-secondary px-3 py-2 text-sm"
						>
							{#each fileColumnsList as c}
								<option value={c}>{c}</option>
							{/each}
						</select>
					</div>
					<div class="flex flex-col gap-1">
						<label for="rb-file-sdir" class="text-xs text-text-secondary">Direction</label>
						<select
							id="rb-file-sdir"
							bind:value={fileSortDir}
							class="rounded-md border border-border bg-bg-secondary px-3 py-2 text-sm"
						>
							<option value="asc">Ascending</option>
							<option value="desc">Descending</option>
						</select>
					</div>
				</div>

				<div class="flex flex-wrap items-center gap-3 rounded border border-border/60 bg-bg-secondary/50 p-3">
					<label class="flex items-center gap-2 text-sm text-text-primary">
						<input type="checkbox" bind:checked={filePivot} />
						Group by + sum
					</label>
					{#if filePivot}
						<select bind:value={fileGroupCol} class="rounded-md border border-border bg-bg-secondary px-2 py-1.5 text-sm">
							{#each fileColumnsList as c}
								<option value={c}>{c}</option>
							{/each}
						</select>
						<span class="text-text-tertiary">sum</span>
						<select bind:value={fileSumCol} class="rounded-md border border-border bg-bg-secondary px-2 py-1.5 text-sm">
							{#each fileColumnsList as c}
								<option value={c}>{c}</option>
							{/each}
						</select>
					{/if}
				</div>

				<div class="flex justify-end">
					<button
						type="button"
						on:click={runFileReport}
						class="inline-flex items-center rounded-md bg-accent-primary px-5 py-2 text-sm font-medium text-white hover:bg-opacity-90"
					>
						Run file report
					</button>
				</div>
			{/if}
		</div>
</div>

<svelte:window on:keydown={handleGlobalKeydown} />

{#if resultsModalOpen}
	<!-- Full-screen results: fixed above app chrome -->
	<div
		class="fixed inset-0 z-[300] flex items-center justify-center p-3 sm:p-6"
		role="presentation"
	>
		<button
			type="button"
			class="absolute inset-0 bg-black/65 backdrop-blur-[2px]"
			aria-label="Close results"
			on:click={closeResultsModal}
		></button>

		<div
			class="relative flex max-h-[min(94vh,920px)] w-full max-w-[min(96vw,1440px)] flex-col overflow-hidden rounded-xl border border-border bg-bg-elevated shadow-2xl pointer-events-auto"
			role="dialog"
			aria-modal="true"
			aria-labelledby="rb-results-title"
			tabindex="-1"
		>
			<div
				class="flex shrink-0 flex-wrap items-center justify-between gap-3 border-b border-border bg-bg-secondary/80 px-4 py-3"
			>
				<div class="min-w-0">
					<h2 id="rb-results-title" class="text-lg font-semibold text-text-primary">Report results</h2>
					{#if runMeta && !runLoading}
						<p class="mt-0.5 text-sm text-text-tertiary">{runMeta}</p>
					{/if}
				</div>
				<button
					type="button"
					on:click={closeResultsModal}
					class="shrink-0 rounded-md border border-border bg-bg-secondary px-3 py-1.5 text-sm text-text-primary hover:bg-bg-tertiary"
				>
					Close
				</button>
			</div>

			<div class="flex min-h-0 flex-1 flex-col gap-4 overflow-hidden p-4 sm:p-5">
				{#if runLoading}
					<div class="flex flex-1 flex-col items-center justify-center gap-3 py-24">
						<div
							class="h-10 w-10 animate-spin rounded-full border-2 border-border border-t-accent-primary"
						></div>
						<p class="text-sm text-text-secondary">Running query…</p>
					</div>
				{:else if runError}
					<div class="rounded-lg bg-error/10 p-4 text-sm text-error">{runError}</div>
				{:else}
					<div class="flex flex-wrap items-center gap-2 border-b border-border/60 pb-3">
						<span class="text-xs font-medium uppercase text-text-tertiary">View</span>
						<button
							type="button"
							class="rounded-md border px-3 py-1.5 text-sm {chartType === 'none'
								? 'border-accent-primary bg-accent-primary/10 text-text-primary'
								: 'border-border bg-bg-secondary text-text-secondary'}"
							on:click={() => (chartType = 'none')}
						>
							Table
						</button>
						<button
							type="button"
							class="rounded-md border px-3 py-1.5 text-sm {chartType === 'bar'
								? 'border-accent-primary bg-accent-primary/10 text-text-primary'
								: 'border-border bg-bg-secondary text-text-secondary'}"
							on:click={() => (chartType = 'bar')}
						>
							Bar chart
						</button>
						<button
							type="button"
							class="rounded-md border px-3 py-1.5 text-sm {chartType === 'pie'
								? 'border-accent-primary bg-accent-primary/10 text-text-primary'
								: 'border-border bg-bg-secondary text-text-secondary'}"
							on:click={() => (chartType = 'pie')}
						>
							Pie chart
						</button>
					</div>

					{#if chartType !== 'none' && resultColumns.length > 0}
						<div class="flex flex-wrap gap-4">
							<div class="flex min-w-[12rem] flex-col gap-1">
								<label for="rb-chart-cat" class="text-xs text-text-secondary">Category (labels)</label>
								<select
									id="rb-chart-cat"
									bind:value={chartCategoryKey}
									class="rounded-md border border-border bg-bg-secondary px-2 py-2 text-sm"
								>
									{#each resultColumns as c}
										<option value={c}>{c}</option>
									{/each}
								</select>
							</div>
							<div class="flex min-w-[12rem] flex-col gap-1">
								<label for="rb-chart-val" class="text-xs text-text-secondary">Value (numeric)</label>
								<select
									id="rb-chart-val"
									bind:value={chartValueKey}
									class="rounded-md border border-border bg-bg-secondary px-2 py-2 text-sm"
								>
									{#each resultColumns as c}
										<option value={c}>{c}</option>
									{/each}
								</select>
							</div>
						</div>
					{/if}

					<div class="flex min-h-0 flex-1 flex-col overflow-hidden">
						{#if resultRows.length > 0}
							{#if chartType === 'none'}
								<div class="h-full max-h-[min(72vh,760px)] min-h-[240px] overflow-auto rounded-lg border border-border">
									<table class="w-full text-left text-sm">
										<thead class="sticky top-0 z-[1] bg-bg-secondary shadow-sm">
											<tr>
												{#each resultColumns as col}
													<th class="whitespace-nowrap border-b border-border px-3 py-2.5 font-medium text-text-primary">
														{col}
													</th>
												{/each}
											</tr>
										</thead>
										<tbody>
											{#each resultRows as row}
												<tr class="border-b border-border/60 hover:bg-bg-tertiary/40">
													{#each resultColumns as col}
														<td class="whitespace-nowrap px-3 py-2 text-text-secondary">
															{String(row[col] ?? '')}
														</td>
													{/each}
												</tr>
											{/each}
										</tbody>
									</table>
								</div>
							{:else}
								<div class="h-[min(72vh,760px)] min-h-[280px] w-full">
									<ReportChart
										{chartType}
										rows={resultRows}
										categoryKey={chartCategoryKey}
										valueKey={chartValueKey}
									/>
								</div>
							{/if}
						{:else}
							<p class="py-12 text-center text-sm text-text-secondary">No rows returned.</p>
						{/if}
					</div>
				{/if}
			</div>
		</div>
	</div>
{/if}
