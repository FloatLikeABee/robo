<script lang="ts">
	import type { Snippet } from 'svelte';
	import { page } from '$app/stores';
	import ThemeToggle from '$lib/components/ui/theme-toggle/ThemeToggle.svelte';

	let { children }: { children: Snippet } = $props();

	function authTitle(pathname: string): string {
		if (pathname.startsWith('/setup')) return 'Setup · DataX';
		if (pathname.startsWith('/login')) return 'Sign in · DataX';
		return 'DataX';
	}

	let pageTitle = $derived(authTitle($page.url.pathname));
</script>

<svelte:head>
	<title>{pageTitle}</title>
</svelte:head>

<div
	class="auth-shell relative flex h-full min-h-0 w-full flex-col overflow-hidden"
>
	<div class="absolute right-4 top-4 z-10">
		<ThemeToggle />
	</div>

	<div class="flex min-h-0 flex-1 items-center justify-center overflow-hidden p-6">
		{@render children()}
	</div>
</div>
