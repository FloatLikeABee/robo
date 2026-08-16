<script lang="ts">
    import { onMount } from 'svelte';
    import { whenSessionReady } from '$lib/stores/auth.svelte';
    import { apiUrl, authHeaders } from '$lib/api';
    
    type DbRow = {
        id: string;
        name: string;
        engine: string;
        host: string;
        port: number;
        database_name: string;
        username: string;
        ssl_enabled: boolean;
    };

    let databases: DbRow[] = [];
    let loading = true;
    let error = '';
    let testMessage: { id: string; text: string; ok: boolean } | null = null;
    let testingId: string | null = null;

    async function testConnection(id: string) {
        testMessage = null;
        testingId = id;
        try {
            const response = await fetch(apiUrl(`/api/v1/databases/${id}/test`), {
                method: 'POST',
                headers: authHeaders()
            });
            const data = await response.json().catch(() => ({}));
            const msg =
                typeof data.message === 'string'
                    ? data.message
                    : response.ok
                      ? 'Test finished'
                      : 'Request failed';
            testMessage = {
                id,
                text: msg,
                ok: response.ok && data.status === 'ok'
            };
        } catch (err) {
            testMessage = {
                id,
                text: err instanceof Error ? err.message : 'Test failed',
                ok: false
            };
        } finally {
            testingId = null;
        }
    }
    
    async function fetchDatabases() {
        try {
            const response = await fetch(apiUrl('/api/v1/databases'), {
                headers: {
                    ...authHeaders(),
                },
            });
            
            if (!response.ok) {
                throw new Error('Failed to fetch databases');
            }
            
            databases = await response.json();
        } catch (err) {
            error = err instanceof Error ? err.message : 'Failed to load databases';
        } finally {
            loading = false;
        }
    }
    
    onMount(() => whenSessionReady(fetchDatabases));
</script>

<div class="flex min-h-0 flex-1 flex-col gap-6">
    <div class="flex shrink-0 items-center justify-between">
        <h1 class="text-2xl font-bold">Database Connections</h1>
        <a
            href="/databases/new"
            class="px-4 py-2 bg-accent-primary text-white rounded-md hover:bg-opacity-90 transition-colors"
        >
            Add Connection
        </a>
    </div>
    
    {#if error}
        <div class="shrink-0 rounded bg-error/10 p-3 text-error">
            {error}
        </div>
    {/if}
    {#if testMessage}
        <div
            class="shrink-0 rounded p-3 text-sm {testMessage.ok
                ? 'bg-emerald-500/10 text-emerald-200'
                : 'bg-error/10 text-error'}"
            role="status"
        >
            <strong class="font-medium">Connection test ({testMessage.id.slice(0, 8)}…):</strong>
            {testMessage.text}
        </div>
    {/if}

    <div class="min-h-0 flex-1 overflow-y-auto">
    {#if loading}
        <div class="loading-skeleton space-y-3">
            <div class="skeleton-bar w-full"></div>
            <div class="skeleton-bar w-3/4"></div>
            <div class="skeleton-bar w-1/2"></div>
        </div>
    {:else if databases.length === 0}
        <div class="text-center py-12">
            <p class="text-text-secondary mb-4">No database connections found</p>
            <a
                href="/databases/new"
                class="px-4 py-2 bg-accent-primary text-white rounded-md hover:bg-opacity-90 transition-colors"
            >
                Add Your First Connection
            </a>
        </div>
    {:else}
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {#each databases as db}
                <div class="bg-bg-elevated p-4 rounded-lg border border-border hover:border-border-hover transition-colors">
                    <div class="flex justify-between items-start mb-2">
                        <h3 class="font-medium text-text-primary">{db.name}</h3>
                        <span class="text-xs px-2 py-1 bg-bg-tertiary rounded-full">{db.engine}</span>
                    </div>
                    <div class="text-sm text-text-secondary space-y-1">
                        <p><span class="font-medium">Host:</span> {db.host}:{db.port}</p>
                        <p><span class="font-medium">Database:</span> {db.database_name}</p>
                        <p><span class="font-medium">User:</span> {db.username}</p>
                        <p><span class="font-medium">SSL:</span> {db.ssl_enabled ? 'Enabled' : 'Disabled'}</p>
                    </div>
                    <div class="mt-4 flex space-x-2">
                        <a
                            href={`/databases/${db.id}`}
                            class="px-3 py-1 text-sm border border-border rounded hover:bg-bg-tertiary transition-colors"
                        >
                            Details
                        </a>
                        <button
                            type="button"
                            disabled={testingId === db.id}
                            on:click={() => testConnection(db.id)}
                            class="px-3 py-1 text-sm border border-border rounded hover:bg-bg-tertiary transition-colors disabled:opacity-50"
                        >
                            {testingId === db.id ? 'Testing…' : 'Test'}
                        </button>
                    </div>
                </div>
            {/each}
        </div>
    {/if}
    </div>
</div>

<style>
    /* Add any databases-specific styles here */
</style>