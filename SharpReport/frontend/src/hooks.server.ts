import type { HandleFetch } from '@sveltejs/kit';

const BACKEND = 'http://127.0.0.1:3050';

/**
 * SSR and server `fetch` do not use Vite's dev proxy. Rewrite same-origin `/api/*`
 * requests to the API server during development.
 */
export const handleFetch: HandleFetch = async ({ request, fetch }) => {
	const url = new URL(request.url);
	if (!url.pathname.startsWith('/api/')) {
		return fetch(request);
	}

	const backendOrigin = new URL(BACKEND).origin;
	if (url.origin === backendOrigin) {
		return fetch(request);
	}

	const target = new URL(url.pathname + url.search, BACKEND);
	return fetch(new Request(target, request));
};
