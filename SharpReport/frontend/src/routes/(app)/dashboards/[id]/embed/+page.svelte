<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { whenSessionReady } from '$lib/stores/auth.svelte';
	import { apiUrl, authHeaders } from '$lib/api';
	import { metabaseProxyUrl } from '$lib/metabaseUrl';

	type DashboardDto = {
		id: string;
		name: string;
		metabase_id?: number | null;
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
			? metabaseProxyUrl(`dashboard/${dashboard.metabase_id}#bordered=true&titled=true`)
			: '';

	onMount(() => whenSessionReady(load));
</script>

<div class="flex min-h-0 flex-1 flex-col overflow-hidden">
	<div
		class="flex shrink-0 items-center justify-between border-b border-border bg-bg-elevated px-4 py-2"
	>
		<a href="/dashboards/{$page.params.id}" class="text-sm text-accent-primary hover:underline"
			>← Normal view</a
		>
		<span class="truncate text-sm font-medium text-text-primary">{dashboard?.name ?? '…'}</span>
		<span class="w-24"></span>
	</div>

	{#if error}
		<div class="m-4 shrink-0 rounded bg-error/10 p-3 text-error">{error}</div>
	{/if}

	{#if loading}
		<div class="loading-skeleton flex flex-1 items-center justify-center p-8">
			<div class="skeleton-chart h-64 w-full max-w-3xl"></div>
		</div>
	{:else if iframeSrc}
		<iframe title="Embedded dashboard" src={iframeSrc} class="min-h-0 w-full flex-1 border-0"></iframe>
	{:else}
		<div class="flex flex-1 items-center justify-center p-8 text-text-secondary">
			No Metabase dashboard id configured for embedding.
		</div>
	{/if}
</div>
