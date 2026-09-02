<script lang="ts">
	import type { Snippet } from 'svelte';
	import { page } from '$app/stores';

	let { children }: { children: Snippet } = $props();

	function authTitle(pathname: string): string {
		if (pathname.startsWith('/setup')) return 'Setup · Data Access';
		if (pathname.startsWith('/login')) return 'Sign in · Data Access';
		return 'Data Access';
	}

	let pageTitle = $derived(authTitle($page.url.pathname));
</script>

<svelte:head>
	<title>{pageTitle}</title>
</svelte:head>

<div
	class="auth-shell relative flex h-full min-h-0 w-full flex-col overflow-hidden"
>

	<div class="flex min-h-0 flex-1 items-center justify-center overflow-hidden p-6">
		{@render children()}
	</div>
</div>
