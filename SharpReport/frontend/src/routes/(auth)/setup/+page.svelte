<script lang="ts">
    import { onMount } from 'svelte';
    import { auth } from '$lib/stores/auth.svelte';
    import { apiUrl } from '$lib/api';
    
    let step = 1;
    let loading = false;
    let error = '';
    
    // Form data
    let adminData = {
        email: '',
        password: '',
        name: ''
    };
    
    let dbData = {
        name: '',
        engine: 'postgres',
        host: '',
        port: '5432',
        databaseName: '',
        username: '',
        password: '',
        sslEnabled: false
    };
    
    async function checkSetupStatus() {
        try {
            const response = await fetch(apiUrl('/api/v1/setup/status'));
            if (response.ok) {
                const data = await response.json();
                if (data.is_completed) {
                    // Redirect to main app
                    window.location.href = '/';
                }
            }
        } catch (err) {
            console.error('Setup check failed:', err);
        }
    }
    
    async function initializeMetabase() {
        loading = true;
        error = '';
        
        try {
            const response = await fetch(apiUrl('/api/v1/setup/initialize'), {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({}),
            });
            
            if (!response.ok) {
                const text = await response.text();
                let msg = `Request failed (${response.status})`;
                try {
                    const body = JSON.parse(text);
                    if (typeof body?.error === 'string') msg = body.error;
                } catch {
                    if (text) msg = text.slice(0, 200);
                }
                throw new Error(msg);
            }
            
            step = 2;
        } catch (err) {
            error = err instanceof Error ? err.message : 'Initialization failed';
        } finally {
            loading = false;
        }
    }
    
    async function createAdmin() {
        if (!adminData.email || !adminData.password || !adminData.name) {
            error = 'Please fill all fields';
            return;
        }
        
        loading = true;
        error = '';
        
        try {
            const response = await fetch(apiUrl('/api/v1/setup/admin'), {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(adminData),
            });
            
            if (!response.ok) {
                throw new Error('Failed to create admin user');
            }
            
            step = 3;
        } catch (err) {
            error = err instanceof Error ? err.message : 'Admin creation failed';
        } finally {
            loading = false;
        }
    }
    
    async function addDatabase() {
        if (!dbData.name || !dbData.host || !dbData.databaseName || !dbData.username) {
            error = 'Please fill all required fields';
            return;
        }
        
        loading = true;
        error = '';
        
        try {
            const response = await fetch(apiUrl('/api/v1/setup/database'), {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    name: dbData.name,
                    engine: dbData.engine,
                    host: dbData.host,
                    port: Number.parseInt(String(dbData.port), 10) || 5432,
                    databaseName: dbData.databaseName,
                    username: dbData.username,
                    password: dbData.password,
                    sslEnabled: dbData.sslEnabled,
                }),
            });
            
            if (!response.ok) {
                throw new Error('Failed to add database');
            }
            
            // Complete setup
            step = 4;
        } catch (err) {
            error = err instanceof Error ? err.message : 'Database addition failed';
        } finally {
            loading = false;
        }
    }
    
    function completeSetup() {
        window.location.href = '/login';
    }
    
    onMount(() => {
        checkSetupStatus();
    });
</script>

<div
    class="auth-setup-panel w-full max-w-2xl max-h-[min(100%,calc(100dvh-7rem))] overflow-y-auto p-8"
>
        <h1 class="text-2xl font-bold mb-6">DataX Setup Wizard</h1>
        
        {#if error}
            <div class="mb-4 p-3 bg-error/10 text-error rounded">
                {error}
            </div>
        {/if}
        
        {#if step === 1}
            <div class="space-y-4">
                <h2 class="text-xl font-semibold">Step 1: Connect to Metabase</h2>
                <p class="text-text-secondary">
                    Start Metabase separately (default <code class="text-text-primary">http://127.0.0.1:8001</code> per backend config), ensure the API is running on port 3050, then continue. This step checks that Metabase is reachable.
                </p>
                <button
                    on:click={initializeMetabase}
                    disabled={loading}
                    class="px-4 py-2 bg-accent-primary text-white rounded-md hover:bg-opacity-90 transition-colors disabled:opacity-50"
                >
                    {#if loading}
                        Initializing...
                    {:else}
                        Initialize Metabase
                    {/if}
                </button>
            </div>
        {:else if step === 2}
            <div class="space-y-4">
                <h2 class="text-xl font-semibold">Step 2: Create Admin User</h2>
                <p class="text-text-secondary">
                    Create your administrator account to manage DataX.
                </p>
                <div class="space-y-3">
                    <div>
                        <label class="block text-sm font-medium mb-1">Name</label>
                        <input
                            type="text"
                            bind:value={adminData.name}
                            class="w-full px-3 py-2 bg-bg-secondary border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-accent-primary"
                        >
                    </div>
                    <div>
                        <label class="block text-sm font-medium mb-1">Email</label>
                        <input
                            type="email"
                            bind:value={adminData.email}
                            class="w-full px-3 py-2 bg-bg-secondary border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-accent-primary"
                        >
                    </div>
                    <div>
                        <label class="block text-sm font-medium mb-1">Password</label>
                        <input
                            type="password"
                            bind:value={adminData.password}
                            class="w-full px-3 py-2 bg-bg-secondary border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-accent-primary"
                        >
                    </div>
                    <button
                        on:click={createAdmin}
                        disabled={loading}
                        class="px-4 py-2 bg-accent-primary text-white rounded-md hover:bg-opacity-90 transition-colors disabled:opacity-50"
                    >
                        {#if loading}
                            Creating...
                        {:else}
                            Create Admin
                        {/if}
                    </button>
                </div>
            </div>
        {:else if step === 3}
            <div class="space-y-4">
                <h2 class="text-xl font-semibold">Step 3: Add Database Connection</h2>
                <p class="text-text-secondary">
                    Connect your first database to start analyzing data.
                </p>
                <div class="space-y-3">
                    <div>
                        <label class="block text-sm font-medium mb-1">Connection Name</label>
                        <input
                            type="text"
                            bind:value={dbData.name}
                            class="w-full px-3 py-2 bg-bg-secondary border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-accent-primary"
                        >
                    </div>
                    <div>
                        <label class="block text-sm font-medium mb-1">Database Type</label>
                        <select
                            bind:value={dbData.engine}
                            class="w-full px-3 py-2 bg-bg-secondary border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-accent-primary"
                        >
                            <option value="postgres">PostgreSQL</option>
                            <option value="mysql">MySQL</option>
                            <option value="sqlite">SQLite</option>
                            <option value="mongo">MongoDB</option>
                            <option value="bigquery">BigQuery</option>
                        </select>
                    </div>
                    <div>
                        <label class="block text-sm font-medium mb-1">Host</label>
                        <input
                            type="text"
                            bind:value={dbData.host}
                            class="w-full px-3 py-2 bg-bg-secondary border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-accent-primary"
                        >
                    </div>
                    <div>
                        <label class="block text-sm font-medium mb-1">Port</label>
                        <input
                            type="number"
                            bind:value={dbData.port}
                            class="w-full px-3 py-2 bg-bg-secondary border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-accent-primary"
                        >
                    </div>
                    <div>
                        <label class="block text-sm font-medium mb-1">Database Name</label>
                        <input
                            type="text"
                            bind:value={dbData.databaseName}
                            class="w-full px-3 py-2 bg-bg-secondary border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-accent-primary"
                        >
                    </div>
                    <div>
                        <label class="block text-sm font-medium mb-1">Username</label>
                        <input
                            type="text"
                            bind:value={dbData.username}
                            class="w-full px-3 py-2 bg-bg-secondary border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-accent-primary"
                        >
                    </div>
                    <div>
                        <label class="block text-sm font-medium mb-1">Password</label>
                        <input
                            type="password"
                            bind:value={dbData.password}
                            class="w-full px-3 py-2 bg-bg-secondary border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-accent-primary"
                        >
                    </div>
                    <div class="flex items-center space-x-2">
                        <input
                            type="checkbox"
                            bind:checked={dbData.sslEnabled}
                            class="rounded border-border text-accent-primary focus:ring-accent-primary"
                        >
                        <label class="text-sm">Use SSL</label>
                    </div>
                    <button
                        on:click={addDatabase}
                        disabled={loading}
                        class="px-4 py-2 bg-accent-primary text-white rounded-md hover:bg-opacity-90 transition-colors disabled:opacity-50"
                    >
                        {#if loading}
                            Connecting...
                        {:else}
                            Add Database
                        {/if}
                    </button>
                </div>
            </div>
        {:else if step === 4}
            <div class="text-center space-y-4">
                <div class="w-16 h-16 mx-auto bg-success/10 rounded-full flex items-center justify-center">
                    <svg class="w-8 h-8 text-success" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path>
                    </svg>
                </div>
                <h2 class="text-xl font-semibold">Setup Complete!</h2>
                <p class="text-text-secondary">
                    DataX is ready to use. You can now sign in with your admin account.
                </p>
                <button
                    on:click={completeSetup}
                    class="px-4 py-2 bg-accent-primary text-white rounded-md hover:bg-opacity-90 transition-colors"
                >
                    Go to Sign In
                </button>
            </div>
        {/if}
</div>

<style>
    /* Add any setup-specific styles here */
</style>