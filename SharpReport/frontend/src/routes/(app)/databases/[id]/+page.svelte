<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { whenSessionReady } from '$lib/stores/auth.svelte';
	import { requestAssistantOpen } from '$lib/stores/assistantContext.svelte';
	import { apiUrl, authHeaders } from '$lib/api';

	let db: {
		id: string;
		name: string;
		engine: string;
		host: string;
		port: number;
		database_name: string;
		username: string;
		ssl_enabled: boolean;
	} | null = null;
	let loading = true;
	let error = '';

	async function load() {
		const id = $page.params.id;
		if (!id) {
			error = 'Missing id';
			loading = false;
			return;
		}
		try {
			const response = await fetch(apiUrl(`/api/v1/databases/${id}`), {
				headers: authHeaders()
			});
			if (response.status === 404) {
				error = 'Connection not found';
				return;
			}
			if (!response.ok) throw new Error('Failed to load connection');
			db = await response.json();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load';
		} finally {
			loading = false;
		}
	}

	onMount(() => whenSessionReady(load));
</script>

<div class="flex min-h-0 flex-1 flex-col gap-6">
	<div class="flex shrink-0 items-center gap-4">
		<a href="/databases" class="text-sm text-accent-primary hover:underline">← Databases</a>
	</div>

	{#if error}
		<div class="shrink-0 rounded bg-error/10 p-3 text-error">{error}</div>
	{/if}

	<div class="min-h-0 flex-1 overflow-y-auto">
		{#if loading}
			<div class="loading-skeleton space-y-3">
				<div class="skeleton-bar w-1/2"></div>
				<div class="skeleton-bar w-full"></div>
			</div>
		{:else if db}
			<h1 class="mb-4 text-2xl font-bold">{db.name}</h1>
			<dl class="grid max-w-lg gap-3 text-sm">
				<dt class="text-text-secondary">Engine</dt>
				<dd class="text-text-primary">{db.engine}</dd>
				<dt class="text-text-secondary">Host</dt>
				<dd class="text-text-primary">{db.host}:{db.port}</dd>
				<dt class="text-text-secondary">Database</dt>
				<dd class="text-text-primary">{db.database_name}</dd>
				<dt class="text-text-secondary">Username</dt>
				<dd class="text-text-primary">{db.username}</dd>
				<dt class="text-text-secondary">SSL</dt>
				<dd class="text-text-primary">{db.ssl_enabled ? 'Enabled' : 'Disabled'}</dd>
			</dl>

			<div class="mt-8 max-w-xl rounded-lg border border-border bg-bg-elevated p-4">
				<h2 class="text-lg font-semibold text-text-primary">Data access</h2>
				<p class="mt-2 text-sm text-text-secondary">
					Run read-only SQL against this connection from the Queries page (PostgreSQL or MySQL/MariaDB). Full schema browsing and
					visualizations stay in Metabase once this database is added there.
				</p>
				<div class="mt-4 flex flex-wrap gap-3">
					<a
						href="/queries"
						class="rounded-md bg-accent-primary px-3 py-2 text-sm font-medium text-white hover:bg-opacity-90"
					>
						Open queries
					</a>
					<button
						type="button"
						class="rounded-md border border-border px-3 py-2 text-sm text-text-primary hover:bg-bg-tertiary"
						onclick={() => requestAssistantOpen()}
					>
						AI Assistant
					</button>
				</div>
			</div>
		{/if}
	</div>
</div>
