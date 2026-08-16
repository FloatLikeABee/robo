<script lang="ts">
    import { onMount } from 'svelte';
    import { whenSessionReady } from '$lib/stores/auth.svelte';
    import { apiUrl, authHeaders } from '$lib/api';
    
    type DashboardCard = {
        id: string;
        name: string;
        description?: string | null;
        metabase_id?: number | null;
        database_id: string;
        is_public: boolean;
    };

    let dashboards: DashboardCard[] = [];
    let loading = true;
    let error = '';
    
    async function fetchDashboards() {
        try {
            const response = await fetch(apiUrl('/api/v1/dashboards'), {
                headers: {
                    ...authHeaders(),
                },
            });
            
            if (!response.ok) {
                throw new Error('Failed to fetch dashboards');
            }
            
            dashboards = await response.json();
        } catch (err) {
            error = err instanceof Error ? err.message : 'Failed to load dashboards';
        } finally {
            loading = false;
        }
    }
    
    onMount(() => whenSessionReady(fetchDashboards));
</script>

<div class="flex min-h-0 flex-1 flex-col gap-6">
    <div class="flex shrink-0 items-center justify-between">
        <h1 class="text-2xl font-bold">Dashboards</h1>
    </div>
    
    {#if error}
        <div class="shrink-0 rounded bg-error/10 p-3 text-error">
            {error}
        </div>
    {/if}

    <div class="min-h-0 flex-1 overflow-y-auto">
    {#if loading}
        <div class="loading-skeleton space-y-3">
            <div class="skeleton-bar w-full"></div>
            <div class="skeleton-bar w-3/4"></div>
            <div class="skeleton-bar w-1/2"></div>
        </div>
    {:else if dashboards.length === 0}
        <div class="py-12 text-center">
            <p class="mb-2 text-text-secondary">No dashboards listed yet.</p>
            <p class="mx-auto max-w-md text-sm text-text-tertiary">
                Create dashboards in Metabase, then register or sync them in DataX when that API is wired up.
            </p>
        </div>
    {:else}
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {#each dashboards as dashboard}
                <div class="bg-bg-elevated p-4 rounded-lg border border-border hover:border-border-hover transition-colors">
                    <div class="flex justify-between items-start mb-2">
                        <h3 class="font-medium text-text-primary">{dashboard.name}</h3>
                        {#if dashboard.is_public}
                            <span class="text-xs px-2 py-1 bg-bg-tertiary rounded-full">Public</span>
                        {/if}
                    </div>
                    {#if dashboard.description}
                        <p class="text-sm text-text-secondary mb-3">{dashboard.description}</p>
                    {/if}
                    <div class="mt-4 flex space-x-2">
                        <a
                            href={`/dashboards/${dashboard.id}`}
                            class="px-3 py-1 text-sm border border-border rounded hover:bg-bg-tertiary transition-colors"
                        >
                            View
                        </a>
                        <a
                            href={`/dashboards/${dashboard.id}/embed`}
                            class="px-3 py-1 text-sm border border-border rounded hover:bg-bg-tertiary transition-colors"
                        >
                            Embed
                        </a>
                    </div>
                </div>
            {/each}
        </div>
    {/if}
    </div>
</div>

<style>
    /* Add any dashboards-specific styles here */
</style>