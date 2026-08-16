<script lang="ts">
	import { onMount } from 'svelte';
	import { whenSessionReady } from '$lib/stores/auth.svelte';
	import { apiUrl, authHeaders } from '$lib/api';

	type SavedQuery = {
		id: string;
		name: string;
		description?: string;
		query_text: string;
		database_id: string;
		is_favorite: boolean;
	};

	type DbOption = { id: string; name: string; engine: string };

	let queries: SavedQuery[] = [];
	let databases: DbOption[] = [];
	let loading = true;
	let error = '';

	let selectedDbForRun = '';
	let runQueryText = 'SELECT 1 AS one';
	let runLoading = false;
	let runError = '';
	let runRows: Record<string, unknown>[] = [];
	let runColumns: string[] = [];
	let runMeta = '';

	async function fetchQueries() {
		try {
			const response = await fetch(apiUrl('/api/v1/queries'), {
				headers: authHeaders()
			});
			if (!response.ok) throw new Error('Failed to load queries');
			queries = await response.json();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load queries';
		} finally {
			loading = false;
		}
	}

	async function fetchDatabases() {
		try {
			const response = await fetch(apiUrl('/api/v1/databases'), {
				headers: authHeaders()
			});
			if (!response.ok) return;
			databases = await response.json();
			if (!selectedDbForRun && databases.length > 0) {
				selectedDbForRun = databases[0].id;
			}
		} catch {
			/* ignore */
		}
	}

	async function loadAll() {
		loading = true;
		error = '';
		await Promise.all([fetchQueries(), fetchDatabases()]);
	}

	async function runSql() {
		runError = '';
		runRows = [];
		runColumns = [];
		runMeta = '';
		if (!selectedDbForRun) {
			runError = 'Select a database connection.';
			return;
		}
		if (!runQueryText.trim()) {
			runError = 'Enter a SQL query.';
			return;
		}
		runLoading = true;
		try {
			const response = await fetch(apiUrl('/api/v1/queries/execute'), {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					...authHeaders()
				},
				body: JSON.stringify({
					database_id: selectedDbForRun,
					sql: runQueryText
				})
			});
			if (!response.ok) {
				const t = await response.text();
				runError = t || 'Execute failed';
				return;
			}
			const data = await response.json();
			const results = data.results as Record<string, unknown>[] | undefined;
			runRows = Array.isArray(results) ? results : [];
			if (runRows.length > 0) {
				runColumns = Object.keys(runRows[0]);
			}
			runMeta = `${data.row_count ?? runRows.length} row${(data.row_count ?? runRows.length) === 1 ? '' : 's'}`;
		} catch (err) {
			runError = err instanceof Error ? err.message : 'Execute failed';
		} finally {
			runLoading = false;
		}
	}

	function useSavedQuery(q: SavedQuery) {
		selectedDbForRun = q.database_id;
		runQueryText = q.query_text;
		runRows = [];
		runColumns = [];
		runError = '';
		runMeta = '';
	}

	onMount(() => whenSessionReady(loadAll));
</script>

<div class="flex min-h-0 flex-1 flex-col gap-6">
	<h1 class="shrink-0 text-2xl font-bold">Queries</h1>

	{#if error}
		<div class="shrink-0 rounded bg-error/10 p-3 text-error">{error}</div>
	{/if}

	<div class="flex min-h-0 flex-1 flex-col gap-6 lg:flex-row lg:items-start">
		<div class="flex min-w-0 flex-1 flex-col gap-4 rounded-lg border border-border bg-bg-elevated p-4">
			<h2 class="text-lg font-semibold text-text-primary">Run SQL</h2>

			<div class="flex flex-col gap-2">
				<label for="query-connection" class="text-sm font-medium text-text-secondary">Connection</label>
				<select
					id="query-connection"
					bind:value={selectedDbForRun}
					class="w-full rounded-md border border-border bg-bg-secondary px-3 py-2 text-text-primary"
				>
					{#each databases as d}
						<option value={d.id}>{d.name} ({d.engine})</option>
					{/each}
				</select>
			</div>

			<div class="flex shrink-0 flex-wrap items-center gap-3 border-b border-border pb-3">
				<button
					type="button"
					disabled={runLoading}
					on:click={runSql}
					class="inline-flex shrink-0 items-center rounded-md bg-accent-primary px-4 py-2 text-sm font-medium text-white hover:bg-opacity-90 disabled:opacity-50"
				>
					{runLoading ? 'Running…' : 'Run'}
				</button>
				{#if runMeta}
					<span class="text-sm text-text-tertiary">{runMeta}</span>
				{/if}
			</div>

			<div class="flex flex-col gap-2">
				<label for="query-sql" class="text-sm font-medium text-text-secondary">SQL</label>
				<textarea
					id="query-sql"
					bind:value={runQueryText}
					rows="12"
					spellcheck="false"
					class="box-border min-h-[12rem] w-full resize-y rounded-md border border-border bg-bg-secondary px-3 py-2 font-mono text-sm leading-relaxed text-text-primary"
				></textarea>
			</div>

			{#if runError}
				<div class="rounded bg-error/10 p-2 text-sm text-error">{runError}</div>
			{/if}
			{#if runRows.length > 0}
				<div class="max-h-[24rem] min-h-0 w-full overflow-auto rounded border border-border">
					<table class="w-full text-left text-xs">
						<thead class="sticky top-0 bg-bg-secondary">
							<tr>
								{#each runColumns as col}
									<th class="border-b border-border px-2 py-2 font-medium text-text-primary">{col}</th>
								{/each}
							</tr>
						</thead>
						<tbody>
							{#each runRows as row}
								<tr class="border-b border-border/60 hover:bg-bg-tertiary/40">
									{#each runColumns as col}
										<td class="px-2 py-1.5 text-text-secondary">{String(row[col] ?? '')}</td>
									{/each}
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}
		</div>

		<div class="flex w-full shrink-0 flex-col gap-3 lg:w-80 lg:shrink-0">
			<h2 class="text-lg font-semibold text-text-primary">Saved queries</h2>
			<div class="min-h-0 flex-1 overflow-y-auto">
				{#if loading}
					<div class="loading-skeleton space-y-3">
						<div class="skeleton-bar w-full"></div>
						<div class="skeleton-bar w-3/4"></div>
					</div>
				{:else if queries.length === 0}
					<p class="text-sm text-text-secondary">No saved queries yet.</p>
				{:else}
					<ul class="space-y-2">
						{#each queries as q}
							<li class="rounded-lg border border-border bg-bg-secondary p-3">
								<div class="flex items-start justify-between gap-2">
									<div class="min-w-0">
										<p class="font-medium text-text-primary">{q.name}</p>
										{#if q.description}
											<p class="mt-0.5 text-xs text-text-tertiary">{q.description}</p>
										{/if}
									</div>
									<button
										type="button"
										on:click={() => useSavedQuery(q)}
										class="shrink-0 text-xs text-accent-primary hover:underline"
									>
										Load
									</button>
								</div>
							</li>
						{/each}
					</ul>
				{/if}
			</div>
		</div>
	</div>
</div>
