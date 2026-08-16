import { browser } from '$app/environment';
import { get, writable } from 'svelte/store';
import { apiUrl } from '$lib/api';

export type BackendStatus = 'unknown' | 'online' | 'offline';

export const backendStatus = writable<BackendStatus>('unknown');
export const backendLastError = writable<string>('');

let monitorStarted = false;
let cleanupMonitor: (() => void) | null = null;

/** Lightweight ping — /api/v1/health is public and does not require auth. */
export async function pingBackend(): Promise<boolean> {
	if (!browser) return true;
	try {
		const res = await fetch(apiUrl('/api/v1/health'), {
			cache: 'no-store',
			signal: AbortSignal.timeout(8000)
		});
		if (!res.ok) {
			backendStatus.set('offline');
			backendLastError.set(`DataX API health check failed (${res.status}).`);
			return false;
		}
		backendStatus.set('online');
		backendLastError.set('');
		return true;
	} catch {
		backendStatus.set('offline');
		backendLastError.set(
			'DataX API is not reachable. The backend on port 3050 may have stopped — try Retry or restart sharpreport-api.'
		);
		return false;
	}
}

export function isBackendOffline(): boolean {
	return get(backendStatus) === 'offline';
}

export function startBackendHealthMonitor(): () => void {
	if (!browser) return () => {};
	if (monitorStarted && cleanupMonitor) return cleanupMonitor;
	monitorStarted = true;

	void pingBackend();

	const onVisible = () => {
		if (document.visibilityState === 'visible') void pingBackend();
	};
	document.addEventListener('visibilitychange', onVisible);

	const intervalId = window.setInterval(() => {
		if (document.visibilityState === 'visible') void pingBackend();
	}, 45_000);

	cleanupMonitor = () => {
		document.removeEventListener('visibilitychange', onVisible);
		window.clearInterval(intervalId);
		monitorStarted = false;
		cleanupMonitor = null;
	};
	return cleanupMonitor;
}
