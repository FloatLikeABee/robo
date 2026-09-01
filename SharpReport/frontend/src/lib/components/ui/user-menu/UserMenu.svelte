<script lang="ts">
	import { auth } from '$lib/stores/auth.svelte';
	import { morphAiOrigin } from '$lib/api';
	import { User } from 'lucide-svelte';

	const morphAi = morphAiOrigin();
</script>

<div class="flex items-center gap-3 border-l border-border pl-4">
	{#if $auth.loading}
		<span class="text-sm text-text-secondary">…</span>
	{:else if $auth.user}
		<div class="flex min-w-0 max-w-[14rem] items-center gap-2">
			<div
				class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-accent-primary text-sm font-medium text-white"
			>
				{$auth.user.name.charAt(0)}
			</div>
			<span class="truncate text-sm font-medium text-text-primary">{$auth.user.name}</span>
		</div>
	{:else}
		<a
			href={morphAi}
			target="_top"
			rel="noopener noreferrer"
			class="flex items-center gap-2 text-sm text-text-secondary hover:text-text-primary"
		>
			<User class="h-5 w-5" />
			<span>Morph AI</span>
		</a>
	{/if}
</div>
