<script lang="ts">
	import { onMount } from 'svelte';
	import { auth } from '$lib/stores/auth.svelte';
	import { getAuthToken, morphAiOrigin } from '$lib/api';
	import { goto } from '$app/navigation';

	const morphAi = morphAiOrigin();

	onMount(() => {
		if (getAuthToken() || $auth.isAuthenticated) {
			goto('/');
		}
	});
</script>

<div class="auth-panel space-y-6 p-8">
	<div class="text-center">
		<h1 class="text-2xl font-bold">Data Access</h1>
	</div>

	{#if morphAi}
		<a
			href={morphAi}
			target="_top"
			rel="noopener noreferrer"
			class="block w-full py-2 px-4 bg-accent-primary text-white rounded-md text-center hover:bg-opacity-90 transition-colors"
		>
			Sign in on Morph
		</a>
	{:else}
		<p class="text-center text-sm text-muted">Sign in on Morph, then open Data Access from MorphUtils.</p>
	{/if}
</div>
