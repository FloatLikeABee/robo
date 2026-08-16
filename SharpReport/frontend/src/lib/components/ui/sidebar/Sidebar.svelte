<script lang="ts">
	import { HelpCircle, Sheet, Table } from 'lucide-svelte';
	import { auth } from '$lib/stores/auth.svelte';
	import { page } from '$app/stores';

	const navItems = [
		{ href: '/data-tables', icon: Sheet, label: 'Data tables', requiresAuth: true },
		{ href: '/reports/builder', icon: Table, label: 'Data reports', requiresAuth: true },
		{ href: '/help', icon: HelpCircle, label: 'Help', requiresAuth: false }
	];

	function isActive(href: string, pathname: string): boolean {
		if (href === '/help') return pathname.startsWith('/help');
		return pathname === href || pathname.startsWith(`${href}/`);
	}
</script>

<nav class="flex min-w-0 flex-1 items-center gap-1 overflow-x-auto" aria-label="Sections">
	{#each navItems as item}
		{@const Icon = item.icon}
		{#if !item.requiresAuth || $auth.isAuthenticated}
			<a
				href={item.href}
				class="inline-flex shrink-0 items-center gap-2 rounded-lg px-3 py-1.5 text-sm font-medium transition-colors {isActive(item.href, $page.url.pathname)
					? 'bg-accent-primary/15 text-accent-primary'
					: 'text-text-secondary hover:bg-bg-tertiary hover:text-text-primary'}"
			>
				<Icon class="h-4 w-4 shrink-0" />
				<span>{item.label}</span>
			</a>
		{/if}
	{/each}
</nav>
