<script lang="ts">
	import { onMount } from 'svelte';
	import { whenSessionReady } from '$lib/stores/auth.svelte';

	onMount(() => whenSessionReady(() => {}));
</script>

<div class="flex min-h-0 flex-1 flex-col gap-6">
	<h1 class="text-2xl font-bold">Add database connection</h1>
	<p class="max-w-xl text-text-secondary">
		New connections with a password are created through the <strong>setup wizard</strong> (step 3), which calls
		the secure setup API. A generic &quot;create connection&quot; form will be added here later.
	</p>
	<a
		href="/setup"
		class="inline-flex w-fit rounded-md bg-accent-primary px-4 py-2 text-white hover:bg-opacity-90"
	>
		Open setup wizard
	</a>
	<a href="/databases" class="text-sm text-accent-primary hover:underline">Back to databases</a>
</div>
