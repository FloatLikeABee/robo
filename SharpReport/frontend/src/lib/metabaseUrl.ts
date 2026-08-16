import { getApiOrigin } from '$lib/api';

const envMetabase = import.meta.env.PUBLIC_METABASE_URL as string | undefined;

/**
 * Metabase origin for a normal browser tab (HTML/JS from Metabase itself).
 * Set `PUBLIC_METABASE_URL` if Metabase is not on the default below.
 */
export function getMetabaseDirectOrigin(): string {
	if (envMetabase !== undefined && envMetabase !== '') {
		return envMetabase.replace(/\/$/, '');
	}
	/* Default matches backend `config/default.toml` [metabase] when not overridden. */
	return 'http://127.0.0.1:8001';
}

/** Open Metabase in a new tab at root or under `path` (e.g. `question`). */
export function metabaseDirectUrl(path = ''): string {
	const base = getMetabaseDirectOrigin();
	const p = path.replace(/^\//, '');
	return p ? `${base}/${p}` : base;
}

/**
 * Iframe `src` for Metabase through the `/metabase/*` proxy.
 * Always prefer same-origin paths so localhost vs 127.0.0.1 does not break iframe security in dev.
 * Vite and the API both proxy `/metabase` to Metabase (see `vite.config.ts`, backend `metabase.rs`).
 */
export function metabaseProxyUrl(path: string): string {
	const p = path.replace(/^\//, '');
	if (typeof window !== 'undefined') {
		return `/metabase/${p}`;
	}
	const origin = getApiOrigin();
	if (origin) {
		return `${origin.replace(/\/$/, '')}/metabase/${p}`;
	}
	return `/metabase/${p}`;
}
