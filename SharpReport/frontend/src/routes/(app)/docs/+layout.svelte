<script lang="ts">
	import type { Snippet } from 'svelte';
	import { page } from '$app/stores';

	let { children }: { children: Snippet } = $props();

	function isActive(href: string, exact: boolean): boolean {
		const p = $page.url.pathname;
		if (exact) return p === href;
		return p === href || p.startsWith(`${href}/`);
	}
</script>

<div class="flex min-h-0 flex-1 flex-col gap-4">
	<div class="shrink-0">
		<h1 class="text-2xl font-bold text-text-primary">Docs</h1>
		<p class="mt-1 max-w-2xl text-sm text-text-secondary">
			Content documents (PDF, TXT) and Docs AI — separate from Data tables (CSV/JSON).
		</p>
		<nav class="mt-3 flex flex-wrap gap-2" aria-label="Docs sections">
			<a
				href="/docs"
				class="rounded-lg px-3 py-1.5 text-sm font-medium transition-colors {isActive('/docs', true)
					? 'bg-accent-primary/15 text-accent-primary'
					: 'text-text-secondary hover:bg-bg-tertiary hover:text-text-primary'}">Library</a
			>
			<a
				href="/docs/ai"
				class="rounded-lg px-3 py-1.5 text-sm font-medium transition-colors {isActive('/docs/ai', false)
					? 'bg-accent-primary/15 text-accent-primary'
					: 'text-text-secondary hover:bg-bg-tertiary hover:text-text-primary'}">Docs AI</a
			>
			<a
				href="/docs/publish"
				class="rounded-lg px-3 py-1.5 text-sm font-medium transition-colors {isActive('/docs/publish', false)
					? 'bg-accent-primary/15 text-accent-primary'
					: 'text-text-secondary hover:bg-bg-tertiary hover:text-text-primary'}">Publish</a
			>
		</nav>
	</div>
	<div class="flex min-h-0 flex-1 flex-col">
		{@render children()}
	</div>
</div>
