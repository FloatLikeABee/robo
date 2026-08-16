<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	onMount(() => {
		void goto('/data-tables', { replaceState: true });
	});
</script>

<p class="p-4 text-sm text-text-secondary">Settings moved to Morph Utils → Settings. Redirecting…</p>
