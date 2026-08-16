<script lang="ts">
	import { onMount } from 'svelte';
	import { auth, login } from '$lib/stores/auth.svelte';
	import { goto } from '$app/navigation';

	let email = '';
	let password = '';
	let rememberMe = true;
	let error = '';
	let loading = false;

	async function handleLogin() {
		if (!email || !password) {
			error = 'Please enter email and password';
			return;
		}

		loading = true;
		error = '';

		try {
			await login(email, password, { rememberMe });
			goto('/');
		} catch (err) {
			error = err instanceof Error ? err.message : 'Login failed';
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		if ($auth.isAuthenticated) {
			goto('/');
		}
	});
</script>

<div class="auth-panel space-y-6 p-8">
	<div class="text-center">
		<h1 class="text-2xl font-bold">DataX</h1>
		<p class="text-text-secondary mt-2">Sign in with your platform account</p>
	</div>

	{#if error}
		<div class="p-3 bg-error/10 text-error rounded text-sm">
			{error}
		</div>
	{/if}

	<form on:submit|preventDefault={handleLogin} class="space-y-4">
		<div>
			<label for="email" class="block text-sm font-medium mb-1">Email</label>
			<input
				id="email"
				type="email"
				bind:value={email}
				required
				class="w-full px-3 py-2 bg-bg-secondary border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-accent-primary"
			/>
		</div>

		<div>
			<label for="password" class="block text-sm font-medium mb-1">Password</label>
			<input
				id="password"
				type="password"
				bind:value={password}
				required
				class="w-full px-3 py-2 bg-bg-secondary border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-accent-primary"
			/>
		</div>

		<label class="flex items-center gap-2 text-sm text-text-secondary">
			<input type="checkbox" bind:checked={rememberMe} class="rounded border-border" />
			Remember me on this device
		</label>

		<button
			type="submit"
			disabled={loading}
			class="w-full py-2 px-4 bg-accent-primary text-white rounded-md hover:bg-opacity-90 transition-colors disabled:opacity-50"
		>
			{#if loading}
				Signing In...
			{:else}
				Sign In
			{/if}
		</button>
	</form>

	<div class="text-center text-sm text-text-secondary">
		<p>
			First-time DataX setup?
			<a href="/setup" class="text-accent-primary hover:underline">Configure Metabase</a>
		</p>
	</div>
</div>
