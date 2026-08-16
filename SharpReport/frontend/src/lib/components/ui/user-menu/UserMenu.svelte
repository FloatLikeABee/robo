<script lang="ts">
	import { auth, logout } from '$lib/stores/auth.svelte';
	import { User } from 'lucide-svelte';
</script>

<div class="flex items-center gap-3 border-l border-border pl-4">
	{#if $auth.user}
		<div class="flex min-w-0 max-w-[14rem] items-center gap-2">
			<div
				class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-accent-primary text-sm font-medium text-white"
			>
				{$auth.user.name.charAt(0)}
			</div>
			<span class="truncate text-sm font-medium text-text-primary">{$auth.user.name}</span>
		</div>
		<button
			type="button"
			on:click={() => logout()}
			class="shrink-0 rounded-md border border-border px-3 py-1.5 text-sm font-medium text-text-primary transition-colors hover:bg-bg-tertiary"
		>
			Sign out
		</button>
	{:else}
		<div class="flex items-center gap-2 text-text-secondary">
			<User class="h-5 w-5" />
			<span class="text-sm">Sign in</span>
		</div>
	{/if}
</div>
