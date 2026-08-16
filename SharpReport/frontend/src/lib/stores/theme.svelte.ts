import { writable } from 'svelte/store';

export const theme = writable<'light' | 'dark'>('dark');

// Theme management functions
export function toggleTheme() {
    theme.update(current => {
        const newTheme = current === 'light' ? 'dark' : 'light';
        localStorage.setItem('theme', newTheme);
        document.documentElement.setAttribute('data-theme', newTheme);
        return newTheme;
    });
}

// Initialize theme from localStorage
export function initTheme() {
    const savedTheme = localStorage.getItem('theme') as 'light' | 'dark' | null;
    const systemPrefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
    
    const initialTheme = savedTheme || (systemPrefersDark ? 'dark' : 'light');
    theme.set(initialTheme);
    document.documentElement.setAttribute('data-theme', initialTheme);
}