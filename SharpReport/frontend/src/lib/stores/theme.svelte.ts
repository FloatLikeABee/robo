import { writable } from 'svelte/store';

export const theme = writable<'light' | 'dark'>('dark');

export function toggleTheme() {
	/* Dark only — no light mode. */
}

export function initTheme() {
	theme.set('dark');
	try {
		localStorage.setItem('theme', 'dark');
	} catch {
		/* ignore */
	}
	document.documentElement.setAttribute('data-theme', 'dark');
}
