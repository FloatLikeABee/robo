import type { HandleFetch } from '@sveltejs/kit';

function sharpReportApiOrigin(): string {
	const raw = (process.env.SHARPREPORT_PORT ?? '').trim();
	const port = /^\d+$/.test(raw) ? raw : '3050';
	return `http://127.0.0.1:${port}`;
}

/**
 * SSR and server `fetch` do not use Vite's dev proxy. Rewrite same-origin `/api/*`
 * requests to the API server during development.
 */
export const handleFetch: HandleFetch = async ({ request, fetch }) => {
	const url = new URL(request.url);
	if (!url.pathname.startsWith('/api/')) {
		return fetch(request);
	}

	const backend = sharpReportApiOrigin();
	const backendOrigin = new URL(backend).origin;
	if (url.origin === backendOrigin) {
		return fetch(request);
	}

	const target = new URL(url.pathname + url.search, backend);
	return fetch(new Request(target, request));
};
