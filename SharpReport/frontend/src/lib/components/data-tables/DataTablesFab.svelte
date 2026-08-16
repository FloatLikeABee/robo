<script lang="ts">
	import DataTablesPanel from './DataTablesPanel.svelte';
	import { Table2 } from 'lucide-svelte';
	import { auth } from '$lib/stores/auth.svelte';

	let panelOpen = $state(false);
</script>

{#if $auth.isAuthenticated}
	<DataTablesPanel bind:open={panelOpen} />

	<button
		type="button"
		class="fixed bottom-6 right-6 z-[310] flex h-14 w-14 items-center justify-center rounded-full border border-border bg-accent-primary text-white shadow-lg transition-transform hover:scale-105 hover:opacity-95"
		aria-label="Data tables"
		aria-expanded={panelOpen}
		onclick={() => (panelOpen = !panelOpen)}
	>
		<Table2 class="h-6 w-6" />
	</button>
{/if}
