<script lang="ts">
	import type { Snippet } from 'svelte';
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { initTheme } from '$lib/stores/theme.svelte';
	import { auth, checkAuth } from '$lib/stores/auth.svelte';
	import { setAuthToken } from '$lib/api';
	import '../app.css';

	let { children }: { children: Snippet } = $props();

	onMount(() => {
		if (!browser) return;
		initTheme();

		try {
			const params = new URLSearchParams(window.location.search);
			const upToken = (params.get('userspanel_token') || '').trim();
			if (upToken) {
				setAuthToken(upToken, { rememberMe: true });
				params.delete('userspanel_token');
				const url = new URL(window.location.href);
				url.search = params.toString();
				window.history.replaceState({}, '', url.toString());
			}
		} catch {
			/* ignore */
		}

		const refreshSession = async () => {
			try {
				const next = await checkAuth();
				auth.set(next);
			} catch (error) {
				console.error('Auth check failed:', error);
				auth.set({
					isAuthenticated: false,
					user: null,
					loading: false
				});
			}
		};

		void refreshSession();

		const onVisible = () => {
			if (document.visibilityState !== 'visible') return;
			void refreshSession();
		};
		document.addEventListener('visibilitychange', onVisible);

		return () => {
			document.removeEventListener('visibilitychange', onVisible);
		};
	});
</script>

<div class="flex h-full min-h-0 w-full flex-col">
	<div class="min-h-0 flex-1">
		{@render children()}
	</div>
</div>
