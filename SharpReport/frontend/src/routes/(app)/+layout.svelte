<script lang="ts">
	import type { Snippet } from 'svelte';
	import { page } from '$app/stores';
	import Sidebar from '$lib/components/ui/sidebar/Sidebar.svelte';
	import ThemeToggle from '$lib/components/ui/theme-toggle/ThemeToggle.svelte';
	import UserMenu from '$lib/components/ui/user-menu/UserMenu.svelte';
	import PlatformAssistantDrawer from '$lib/components/report/PlatformAssistantDrawer.svelte';
	import { assistantCtx } from '$lib/stores/assistantContext.svelte';
	import { Sparkles } from 'lucide-svelte';

	let { children }: { children: Snippet } = $props();

	let assistantOpen = $state(false);

	function moduleTitle(pathname: string): string {
		const p = pathname || '/';
		if (p.startsWith('/reports/builder')) return 'Data reports';
		if (p.startsWith('/data-tables/') && p !== '/data-tables') return 'Data table';
		if (p.startsWith('/data-tables')) return 'Data tables';
		if (p.startsWith('/settings')) return 'Settings';
		if (p.startsWith('/help')) return 'Help';
		if (p.startsWith('/embed')) return 'Embed';
		return 'Data Access';
	}

	let pageTitle = $derived(`${moduleTitle($page.url.pathname)} · Data Access`);

	$effect(() => {
		if (assistantCtx.openRequest > 0) {
			assistantOpen = true;
		}
	});
</script>

<svelte:head>
	<title>{pageTitle}</title>
</svelte:head>

<div class="flex h-full min-h-0 w-full flex-col overflow-hidden">
	<header
		class="box-border shrink-0 border-b border-border bg-bg-secondary"
		style="padding-top: env(safe-area-inset-top);"
	>
		<div class="flex items-center justify-between gap-2 px-3 py-2.5 sm:gap-3 sm:px-4">
			<div class="flex min-w-0 items-center gap-2.5">
				<img src="/datax-logo.svg" alt="" width="28" height="28" class="h-7 w-7 shrink-0 rounded-lg object-cover" />
				<div class="min-w-0">
					<div class="truncate text-sm font-bold text-text-primary sm:text-base">Data Access</div>
				</div>
			</div>
			<div class="flex shrink-0 items-center gap-1.5 sm:gap-3">
				<button
					type="button"
					aria-expanded={assistantOpen}
					aria-controls="sharpreport-ai-assistant-drawer"
					onclick={() => (assistantOpen = !assistantOpen)}
					class="inline-flex items-center gap-2 rounded-full border border-accent-primary/50 bg-accent-primary/10 px-3 py-1.5 text-sm font-semibold text-accent-primary shadow-sm transition-colors hover:bg-accent-primary/20 sm:px-4"
				>
					<Sparkles class="h-4 w-4 shrink-0" />
					<span class="hidden sm:inline">AI Assistant</span>
					<span class="sm:hidden">AI</span>
				</button>
				<ThemeToggle />
				<UserMenu />
			</div>
		</div>

		<div class="border-t border-border px-3 py-2 sm:px-4">
			<Sidebar />
		</div>
	</header>

	<main class="flex min-h-0 flex-1 flex-col overflow-auto p-3 sm:overflow-hidden sm:p-6">
		{@render children()}
	</main>
</div>

<PlatformAssistantDrawer open={assistantOpen} onClose={() => (assistantOpen = false)} />
