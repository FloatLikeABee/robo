<script lang="ts">
    import { theme, toggleTheme } from '$lib/stores/theme.svelte';
    import { Moon, Sun } from 'lucide-svelte';
</script>

<button
    onclick={toggleTheme}
    class="p-2 rounded-full hover:bg-bg-tertiary transition-colors"
    aria-label="Toggle theme"
>
    {#if $theme === 'dark'}
        <Sun class="w-5 h-5 text-text-primary" />
    {:else}
        <Moon class="w-5 h-5 text-text-primary" />
    {/if}
</button>

<style>
    button:focus-visible {
        outline: 2px solid var(--accent-primary);
        outline-offset: 2px;
    }
</style>