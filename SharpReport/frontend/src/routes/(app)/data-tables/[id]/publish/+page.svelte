<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import DataPagePublishPanel from '$lib/components/data-tables/DataPagePublishPanel.svelte';
	import { getDataTable } from '$lib/dataTables';
	import { whenSessionReady } from '$lib/stores/auth.svelte';

	let tableName = $state('Data table');
	let loading = $state(true);
	let error = $state('');

	const tableId = $derived($page.params.id ?? '');

	$effect(() => {
		if (!tableId) return;
		loading = true;
		error = '';
		whenSessionReady(async () => {
			try {
				const table = await getDataTable(tableId);
				tableName = table.name;
			} catch (e) {
				error = e instanceof Error ? e.message : 'Failed to load table';
			} finally {
				loading = false;
			}
		});
	});

	function goBack() {
		goto(`/data-tables/${tableId}`);
	}
</script>

{#if loading}
	<p class="text-sm text-text-secondary">Loading…</p>
{:else if error}
	<p class="text-sm text-danger">{error}</p>
	<button type="button" class="text-sm text-accent-primary hover:underline" onclick={goBack}>
		← Back
	</button>
{:else}
	<div class="flex min-h-0 flex-1 flex-col overflow-hidden">
		<DataPagePublishPanel tableId={tableId} tableName={tableName} onBack={goBack} />
	</div>
{/if}
