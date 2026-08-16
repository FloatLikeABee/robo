<script lang="ts">
    import { onMount } from 'svelte';
    import { page } from '$app/stores';
    import { metabaseProxyUrl } from '$lib/metabaseUrl';

    let embedUrl = '';
    let loading = true;
    let error = '';

    async function loadEmbed() {
        const token = $page.params.token;

        try {
            // Token is treated as Metabase path segment (e.g. signed token or id placeholder).
            embedUrl = metabaseProxyUrl(
                `embed/dashboard/${token}#bordered=true&titled=true`
            );
        } catch (err) {
            error = err instanceof Error ? err.message : 'Failed to load embed';
        } finally {
            loading = false;
        }
    }
    
    onMount(() => {
        loadEmbed();
    });
</script>

<div class="flex min-h-0 flex-1 flex-col overflow-hidden">
    {#if error}
        <div class="flex min-h-0 flex-1 items-center justify-center">
            <div class="text-center">
                <h2 class="text-xl font-semibold mb-2">Embed Error</h2>
                <p class="text-text-secondary">{error}</p>
            </div>
        </div>
    {:else if loading}
        <div class="flex min-h-0 flex-1 items-center justify-center">
            <div class="loading-skeleton">
                <div class="skeleton-chart w-96 h-64"></div>
            </div>
        </div>
    {:else}
        <iframe
            title="Embedded dashboard"
            src={embedUrl}
            class="min-h-0 w-full flex-1 border-none"
        ></iframe>
    {/if}
</div>

<style>
    /* Add any embed-specific styles here */
</style>