<script lang="ts">
	import DataTablesPanel from './DataTablesPanel.svelte';
	import { Table2 } from 'lucide-svelte';

	let {
		label = 'Import data',
		panelAnchor = 'top-right'
	}: {
		label?: string;
		panelAnchor?: 'top-right' | 'fab';
	} = $props();

	let panelOpen = $state(false);
</script>

<DataTablesPanel bind:open={panelOpen} anchor={panelAnchor} />

<button
	type="button"
	class="inline-flex shrink-0 items-center gap-1.5 rounded-lg bg-accent-primary px-3 py-1.5 text-sm font-medium text-white shadow-sm transition-opacity hover:opacity-90"
	aria-label={label}
	aria-expanded={panelOpen}
	onclick={() => (panelOpen = !panelOpen)}
>
	<Table2 class="h-4 w-4" />
	<span>{label}</span>
</button>
