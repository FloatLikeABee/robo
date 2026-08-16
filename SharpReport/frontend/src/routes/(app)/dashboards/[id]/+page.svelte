<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { whenSessionReady } from '$lib/stores/auth.svelte';
	import { apiUrl, authHeaders } from '$lib/api';
	import { metabaseProxyUrl } from '$lib/metabaseUrl';

	type DashboardDto = {
		id: string;
		name: string;
		description?: string | null;
		metabase_id?: number | null;
		database_id: string;
		is_public: boolean;
	};

	let dashboard: DashboardDto | null = null;
	let loading = true;
	let error = '';

	async function load() {
		const id = $page.params.id;
		if (!id) {
			error = 'Missing dashboard id';
			loading = false;
			return;
		}
		try {
			const response = await fetch(apiUrl(`/api/v1/dashboards/${id}`), {
				headers: authHeaders()
			});
			if (response.status === 404) {
				error = 'Dashboard not found';
				return;
			}
			if (!response.ok) throw new Error('Failed to load dashboard');
			dashboard = await response.json();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load';
		} finally {
			loading = false;
		}
	}

	$: iframeSrc =
		dashboard?.metabase_id != null && dashboard.metabase_id !== undefined
			? metabaseProxyUrl(`dashboard/${dashboard.metabase_id}`)
			: '';

	onMount(() => whenSessionReady(load));
</script>

<div class="flex min-h-0 flex-1 flex-col gap-4">
	<div class="flex shrink-0 flex-wrap items-center gap-4">
		<a href="/dashboards" class="text-sm text-accent-primary hover:underline">← Dashboards</a>
		{#if dashboard}
			<h1 class="text-xl font-bold text-text-primary">{dashboard.name}</h1>
			<a
				href={`/dashboards/${$page.params.id}/embed`}
				class="ml-auto text-sm text-accent-primary hover:underline"
			>
				Embed layout
			</a>
		{/if}
	</div>

	{#if error}
		<div class="shrink-0 rounded bg-error/10 p-3 text-error">{error}</div>
	{/if}

	<div class="min-h-0 flex flex-1 flex-col overflow-hidden rounded-lg border border-border bg-bg-secondary">
		{#if loading}
			<div class="loading-skeleton flex flex-1 flex-col p-6">
				<div class="skeleton-bar mb-4 w-1/3"></div>
				<div class="skeleton-bar flex-1 w-full min-h-[24rem]"></div>
			</div>
		{:else if dashboard && iframeSrc}
			<iframe title={dashboard.name} src={iframeSrc} class="h-[min(80vh,56rem)] w-full flex-1 border-0"></iframe>
			<p class="shrink-0 border-t border-border px-3 py-2 text-xs text-text-tertiary">
				Loaded via Metabase proxy. If you see a login or error page, open Metabase in another tab and sign in,
				or enable static embedding with a signed token (roadmap).
			</p>
		{:else if dashboard}
			<div class="flex flex-1 flex-col items-center justify-center gap-3 p-8 text-center">
				<p class="max-w-md text-text-secondary">
					This dashboard has no <code class="text-text-primary">metabase_id</code> yet. Link it in Metabase
					after you create the dashboard there, or extend the API to store Metabase IDs.
				</p>
				<a
					href="/dashboards"
					class="text-sm text-accent-primary hover:underline"
				>
					Back to list
				</a>
			</div>
		{/if}
	</div>
</div>
